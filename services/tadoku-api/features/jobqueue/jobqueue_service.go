package jobqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }

// Enqueue persists the complete batch in the application's active transaction.
// Callers must return any error from their transaction callback.
func (s *Service) Enqueue(ctx context.Context, batch ...jobs.Job) error {
	if _, err := postgres.TransactionExecutor(ctx, s.repository.db); err != nil {
		return err
	}
	payloads := make([]json.RawMessage, len(batch))
	for i, job := range batch {
		if job == nil || (reflect.ValueOf(job).Kind() == reflect.Pointer && reflect.ValueOf(job).IsNil()) {
			return errors.New("job must not be nil")
		}
		if err := job.Validate(); err != nil {
			return fmt.Errorf("validate job %d: %w", i, err)
		}
		payload, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("encode job %d: %w", i, err)
		}
		payloads[i] = payload
	}
	for i, job := range batch {
		if _, err := s.repository.Insert(ctx, job.Type(), payloads[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Claim(ctx context.Context, typ jobs.Type, limit int, lease time.Duration, maxAttempts int) ([]ClaimedJob, error) {
	return s.repository.Claim(ctx, typ, limit, lease, maxAttempts)
}

func (s *Service) Renew(ctx context.Context, task ClaimedJob, lease time.Duration) (time.Time, bool, error) {
	return s.repository.Renew(ctx, task, lease)
}

func (s *Service) Complete(ctx context.Context, task ClaimedJob) (bool, error) {
	return s.repository.Complete(ctx, task)
}

func (s *Service) Retry(ctx context.Context, task ClaimedJob, next time.Time, code string, maxAttempts int) (bool, error) {
	return s.repository.Retry(ctx, task, next, code, maxAttempts)
}

func (s *Service) Fail(ctx context.Context, task ClaimedJob, code string) (bool, error) {
	return s.repository.Fail(ctx, task, code)
}

func (s *Service) Replay(ctx context.Context, failedID int64, actor, reason string, supported []jobs.Type) (int64, error) {
	return s.repository.Replay(ctx, failedID, actor, reason, supported)
}

func (s *Service) CleanupCompleted(ctx context.Context, limit int) (int64, error) {
	return s.repository.CleanupCompleted(ctx, timex.Now(), limit)
}

func (s *Service) Outstanding(ctx context.Context, typ jobs.Type) (int64, error) {
	return s.repository.Outstanding(ctx, typ)
}

func (s *Service) Stats(ctx context.Context, typ jobs.Type) (Stats, error) {
	return s.repository.Stats(ctx, typ)
}

func (s *Service) UnsupportedStats(ctx context.Context, known []jobs.Type) (UnsupportedStats, error) {
	return s.repository.UnsupportedStats(ctx, known)
}
