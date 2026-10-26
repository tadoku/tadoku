package jobqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }

func (s *Service) Enqueue(ctx context.Context, batch ...jobs.Job) error {
	_, err := s.enqueue(ctx, batch)
	return err
}

func (s *Service) EnqueueOne(ctx context.Context, job jobs.Job) (int64, error) {
	ids, err := s.enqueue(ctx, []jobs.Job{job})
	if err != nil {
		return 0, err
	}
	return ids[0], nil
}

func (s *Service) enqueue(ctx context.Context, batch []jobs.Job) ([]int64, error) {
	payloads := make([]json.RawMessage, len(batch))
	for i, job := range batch {
		if job == nil || (reflect.ValueOf(job).Kind() == reflect.Pointer && reflect.ValueOf(job).IsNil()) {
			return nil, errx.NewInvalidInputError("job must not be nil")
		}
		if err := job.Validate(); err != nil {
			return nil, errx.NewInvalidInputError(fmt.Sprintf("job %d: %s", i, err))
		}
		if job.Type() == "" {
			return nil, errx.NewInvalidInputError("job type must not be empty")
		}

		payload, err := json.Marshal(job)
		if err != nil {
			return nil, fmt.Errorf("encode job %d: %w", i, err)
		}
		payloads[i] = payload
	}

	ids := make([]int64, len(batch))
	for i, job := range batch {
		id, err := s.repository.Insert(ctx, job.Type(), payloads[i])
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}
	return ids, nil
}

func (s *Service) Claim(ctx context.Context, typ jobs.Type, limit int, lease time.Duration, maxAttempts int) ([]ClaimedJob, error) {
	if typ == "" {
		return nil, errx.NewInvalidInputError("job type must not be empty")
	}
	if limit < 1 || limit > 100 {
		return nil, errx.NewInvalidInputError("claim limit must be between 1 and 100")
	}
	if lease < time.Microsecond {
		return nil, errx.NewInvalidInputError("job lease must be at least one microsecond")
	}
	if maxAttempts < 1 || maxAttempts > math.MaxInt32 {
		return nil, errx.NewInvalidInputError("max attempts must be between 1 and 2147483647")
	}

	return s.repository.Claim(ctx, typ, limit, lease, maxAttempts)
}

// FailExhausted fails up to 100 due jobs of typ that already used maxAttempts, including
// running jobs whose final lease expired. Claim never picks these jobs.
func (s *Service) FailExhausted(ctx context.Context, typ jobs.Type, maxAttempts int) ([]ExhaustedJob, error) {
	if typ == "" {
		return nil, errx.NewInvalidInputError("job type must not be empty")
	}
	if maxAttempts < 1 || maxAttempts > math.MaxInt32 {
		return nil, errx.NewInvalidInputError("max attempts must be between 1 and 2147483647")
	}

	return s.repository.FailExhausted(ctx, typ, maxAttempts)
}

func (s *Service) Renew(ctx context.Context, task ClaimedJob, lease time.Duration) (time.Time, bool, error) {
	if lease < time.Microsecond {
		return time.Time{}, false, errx.NewInvalidInputError("job lease must be at least one microsecond")
	}

	return s.repository.Renew(ctx, task, lease)
}

func (s *Service) Complete(ctx context.Context, task ClaimedJob) (bool, error) {
	return s.repository.Complete(ctx, task)
}

func (s *Service) Retry(ctx context.Context, task ClaimedJob, next time.Time, code string, maxAttempts int) (bool, error) {
	if err := validateErrorCode(code); err != nil {
		return false, err
	}
	if next.Before(timex.Now()) {
		return false, errx.NewInvalidInputError("next attempt must not be before the current time")
	}
	if maxAttempts < 1 || maxAttempts > math.MaxInt32 {
		return false, errx.NewInvalidInputError("max attempts must be between 1 and 2147483647")
	}

	return s.repository.Retry(ctx, task, next, code, maxAttempts)
}

func (s *Service) Fail(ctx context.Context, task ClaimedJob, code string) (bool, error) {
	if err := validateErrorCode(code); err != nil {
		return false, err
	}

	return s.repository.Fail(ctx, task, code)
}

func (s *Service) Replay(ctx context.Context, failedID int64, actor, reason string, supported []jobs.Type) (int64, error) {
	if failedID < 1 {
		return 0, errx.NewInvalidInputError("failed job ID must be positive")
	}
	actor = strings.TrimSpace(actor)
	reason = strings.TrimSpace(reason)
	if len(actor) < 1 || len(actor) > 200 {
		return 0, errx.NewInvalidInputError("replay actor must be 1 to 200 characters after trimming whitespace")
	}
	if len(reason) < 1 || len(reason) > 500 {
		return 0, errx.NewInvalidInputError("replay reason must be 1 to 500 characters after trimming whitespace")
	}

	return s.repository.Replay(ctx, failedID, actor, reason, supported)
}

func (s *Service) CleanupCompleted(ctx context.Context, limit int) (int64, error) {
	now := timex.Now()
	if now.IsZero() {
		return 0, errx.NewInvalidInputError("cleanup time must not be zero")
	}
	if limit < 1 || limit > 1000 {
		return 0, errx.NewInvalidInputError("cleanup limit must be between 1 and 1000")
	}

	return s.repository.CleanupCompleted(ctx, now, limit)
}

func (s *Service) Tenants(ctx context.Context) ([]tenant.Key, error) {
	return s.repository.Tenants(ctx)
}

func (s *Service) Stats(ctx context.Context, typ jobs.Type) (Stats, error) {
	if typ == "" {
		return Stats{}, errx.NewInvalidInputError("job type must not be empty")
	}

	return s.repository.Stats(ctx, typ)
}

func (s *Service) UnsupportedStats(ctx context.Context, known []jobs.Type) (UnsupportedStats, error) {
	for _, typ := range known {
		if typ == "" {
			return UnsupportedStats{}, errx.NewInvalidInputError("registered job type must not be empty")
		}
	}

	return s.repository.UnsupportedStats(ctx, known)
}
