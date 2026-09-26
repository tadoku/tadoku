package jobqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Insert(ctx context.Context, typ jobs.Type, payload json.RawMessage) (int64, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	id, err := queries.New(executor).Enqueue(ctx, queries.EnqueueParams{
		TaskType: string(typ),
		Payload:  payload,
		Now:      postgres.Timestamptz(timex.Now()),
	})
	if err != nil {
		return 0, fmt.Errorf("enqueue job: %w", err)
	}
	return id, nil
}

func (r *Repository) Claim(ctx context.Context, typ jobs.Type, limit int, lease time.Duration, maxAttempts int) ([]ClaimedJob, error) {
	now := timex.Now()
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).Claim(ctx, queries.ClaimParams{
		TaskType:    string(typ),
		Now:         postgres.Timestamptz(now),
		MaxAttempts: int32(maxAttempts),
		BatchSize:   int32(limit),
		LeaseMicros: lease.Microseconds(),
	})
	if err != nil {
		return nil, fmt.Errorf("claim jobs: %w", err)
	}
	tasks := make([]ClaimedJob, len(rows))
	for i, row := range rows {
		tasks[i] = ClaimedJob{
			ID:             row.ID,
			Type:           jobs.Type(row.TaskType),
			Payload:        row.Payload,
			Attempts:       int(row.Attempts),
			Token:          uuid.UUID(row.ClaimToken.Bytes),
			LeaseExpiresAt: row.LeaseExpiresAt.Time,
			Reclaimed:      row.Reclaimed,
		}
	}
	return tasks, nil
}

func (r *Repository) Renew(ctx context.Context, task ClaimedJob, lease time.Duration) (time.Time, bool, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return time.Time{}, false, err
	}
	expires, err := queries.New(executor).Renew(ctx, queries.RenewParams{
		LeaseMicros: lease.Microseconds(),
		ID:          task.ID,
		ClaimToken:  postgres.UUID(task.Token),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	return expires.Time, true, nil
}

func (r *Repository) Complete(ctx context.Context, task ClaimedJob) (bool, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return false, err
	}
	count, err := queries.New(executor).Complete(ctx, queries.CompleteParams{
		Now:        postgres.Timestamptz(timex.Now()),
		ID:         task.ID,
		ClaimToken: postgres.UUID(task.Token),
	})
	return count == 1, err
}

func (r *Repository) Retry(ctx context.Context, task ClaimedJob, next time.Time, code string, maxAttempts int) (bool, error) {
	now := timex.Now()
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return false, err
	}
	count, err := queries.New(executor).Retry(ctx, queries.RetryParams{
		MaxAttempts:   int32(maxAttempts),
		Now:           postgres.Timestamptz(now),
		NextAttemptAt: postgres.Timestamptz(next),
		ErrorCode:     code,
		ID:            task.ID,
		ClaimToken:    postgres.UUID(task.Token),
	})
	return count == 1, err
}

func (r *Repository) Fail(ctx context.Context, task ClaimedJob, code string) (bool, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return false, err
	}
	count, err := queries.New(executor).Fail(ctx, queries.FailParams{
		Now:        postgres.Timestamptz(timex.Now()),
		ErrorCode:  code,
		ID:         task.ID,
		ClaimToken: postgres.UUID(task.Token),
	})
	return count == 1, err
}

func (r *Repository) Replay(ctx context.Context, failedID int64, actor, reason string, supported []jobs.Type) (int64, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	names := make([]string, len(supported))
	for i, typ := range supported {
		names[i] = string(typ)
	}
	id, err := queries.New(executor).Replay(ctx, queries.ReplayParams{
		Now:            postgres.Timestamptz(timex.Now()),
		Actor:          actor,
		Reason:         reason,
		FailedID:       failedID,
		SupportedTypes: names,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFailed
	}
	if err != nil {
		return 0, fmt.Errorf("replay job: %w", err)
	}
	return id, nil
}

func (r *Repository) CleanupCompleted(ctx context.Context, now time.Time, limit int) (int64, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	return queries.New(executor).CleanupCompleted(ctx, queries.CleanupCompletedParams{
		Now:       postgres.Timestamptz(now),
		BatchSize: int32(limit),
	})
}

func (r *Repository) Outstanding(ctx context.Context, typ jobs.Type) (int64, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	return queries.New(executor).Outstanding(ctx, string(typ))
}

func (r *Repository) Stats(ctx context.Context, typ jobs.Type) (Stats, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return Stats{}, err
	}
	row, err := queries.New(executor).Stats(ctx, string(typ))
	if err != nil {
		return Stats{}, err
	}
	stats := Stats{Pending: row.Pending, Failed: row.Failed}
	if row.OldestDueAt.Valid {
		instant := row.OldestDueAt.Time
		stats.OldestDueAt = &instant
	}
	return stats, nil
}

func (r *Repository) UnsupportedStats(ctx context.Context, known []jobs.Type) (UnsupportedStats, error) {
	names := make([]string, len(known))
	for i, typ := range known {
		names[i] = string(typ)
	}
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return UnsupportedStats{}, err
	}
	row, err := queries.New(executor).UnsupportedStats(ctx, names)
	if err != nil {
		return UnsupportedStats{}, err
	}
	stats := UnsupportedStats{Pending: row.Pending, Running: row.Running, Failed: row.Failed}
	if row.OldestDueAt.Valid {
		instant := row.OldestDueAt.Time
		stats.OldestDueAt = &instant
	}
	return stats, nil
}
