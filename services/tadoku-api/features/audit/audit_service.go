package audit

import (
	"context"
	"log/slog"
	"maps"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

const recordOutcomeTimeout = 5 * time.Second

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Record(ctx context.Context, event Event) error {
	event.recordedAt = timex.Now()
	return s.repository.Create(ctx, event)
}

// RecordOutcome records a change that already happened, so call it after the
// change and outside a transaction that may roll back. A cause that refused the
// request before it changed anything records nothing; any other cause records
// the attempt as failed. Recording failures are reported to the default logger
// and never returned.
func (s *Service) RecordOutcome(ctx context.Context, event Event, cause error) {
	if cause != nil {
		switch errx.KindOf(cause) {
		case errx.InvalidInput, errx.Unauthorized, errx.Forbidden, errx.NotFound, errx.Conflict:
			return
		}

		metadata := make(map[string]any, len(event.Metadata)+1)
		maps.Copy(metadata, event.Metadata)
		metadata["error"] = cause.Error()

		event.Action += "_failed"
		event.Metadata = metadata
	}

	recordCtx, cancel := recordContext(ctx)
	defer cancel()

	if err := s.Record(recordCtx, event); err != nil {
		slog.ErrorContext(
			recordCtx,
			"privileged change not audited",
			"action", event.Action,
			"actor_id", event.ActorID,
			"metadata", event.Metadata,
			"error", err,
		)
	}
}

// recordContext detaches the write from the caller's cancellation, so a change
// that already happened is still recorded once the caller is gone.
func recordContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), recordOutcomeTimeout)
}
