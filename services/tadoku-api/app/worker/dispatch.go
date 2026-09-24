package worker

import (
	"context"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/asyncwork"
	"github.com/tadoku/tadoku/services/tadoku-api/storage/postgres/asyncoutbox"
)

type Application struct{ leaderboard *leaderboard.Service }

func NewApplication(leaderboard *leaderboard.Service) *Application {
	return &Application{leaderboard: leaderboard}
}

type PermanentError struct{ Err error }

func (e *PermanentError) Error() string { return e.Err.Error() }
func (e *PermanentError) Unwrap() error { return e.Err }

type UnknownTypeError struct{ Type asyncwork.Type }

func (e *UnknownTypeError) Error() string { return fmt.Sprintf("unsupported task type %q", e.Type) }

func (a *Application) Dispatch(ctx context.Context, task asyncoutbox.ClaimedTask) error {
	switch task.Type {
	case asyncwork.InvalidateContest:
		payload, err := asyncwork.DecodeContestInvalidation(task.Payload)
		if err != nil {
			return &PermanentError{Err: fmt.Errorf("decode contest invalidation: %w", err)}
		}
		return a.leaderboard.InvalidateContest(ctx, payload.ContestID)
	case asyncwork.InvalidateOfficial:
		payload, err := asyncwork.DecodeOfficialInvalidation(task.Payload)
		if err != nil {
			return &PermanentError{Err: fmt.Errorf("decode official invalidation: %w", err)}
		}
		return a.leaderboard.InvalidateOfficial(ctx, payload.Year)
	default:
		return &UnknownTypeError{Type: task.Type}
	}
}

func failureCode(err error) string {
	var permanent *PermanentError
	if errors.As(err, &permanent) {
		return "invalid_payload"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline_exceeded"
	}
	return "handler_error"
}
