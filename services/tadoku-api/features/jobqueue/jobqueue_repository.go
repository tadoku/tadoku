package jobqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Insert(ctx context.Context, typ jobs.Type, payload json.RawMessage) (int64, error) {
	if typ == "" || !json.Valid(payload) {
		return 0, errors.New("invalid outbox task")
	}
	executor, err := postgres.TransactionExecutor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	id, err := queries.New(executor).Enqueue(ctx, queries.EnqueueParams{
		TaskType: string(typ),
		Payload:  payload,
		Now:      timestamp(timex.Now()),
	})
	if err != nil {
		return 0, fmt.Errorf("enqueue outbox task: %w", err)
	}
	return id, nil
}

func (r *Repository) Claim(ctx context.Context, typ jobs.Type, limit int, lease time.Duration, maxAttempts int) ([]ClaimedJob, error) {
	if typ == "" || limit < 1 || limit > 100 || lease < time.Microsecond || maxAttempts < 1 || maxAttempts > math.MaxInt32 {
		return nil, errors.New("invalid outbox claim parameters")
	}
	now := timex.Now()
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).Claim(ctx, queries.ClaimParams{
		TaskType:    string(typ),
		Now:         timestamp(now),
		MaxAttempts: int32(maxAttempts),
		BatchSize:   int32(limit),
		LeaseMicros: lease.Microseconds(),
	})
	if err != nil {
		return nil, fmt.Errorf("claim outbox tasks: %w", err)
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
	if lease < time.Microsecond {
		return time.Time{}, false, errors.New("outbox lease must be positive")
	}
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
		Now:        timestamp(timex.Now()),
		ID:         task.ID,
		ClaimToken: postgres.UUID(task.Token),
	})
	return count == 1, err
}

func (r *Repository) Retry(ctx context.Context, task ClaimedJob, next time.Time, code string, maxAttempts int) (bool, error) {
	now := timex.Now()
	if err := validateErrorCode(code); err != nil {
		return false, err
	}
	if next.Before(now) || maxAttempts < 1 || maxAttempts > math.MaxInt32 {
		return false, errors.New("invalid outbox retry parameters")
	}
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return false, err
	}
	count, err := queries.New(executor).Retry(ctx, queries.RetryParams{
		MaxAttempts:   int32(maxAttempts),
		Now:           timestamp(now),
		NextAttemptAt: timestamp(next),
		ErrorCode:     code,
		ID:            task.ID,
		ClaimToken:    postgres.UUID(task.Token),
	})
	return count == 1, err
}

func (r *Repository) Fail(ctx context.Context, task ClaimedJob, code string) (bool, error) {
	if err := validateErrorCode(code); err != nil {
		return false, err
	}
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return false, err
	}
	count, err := queries.New(executor).Fail(ctx, queries.FailParams{
		Now:        timestamp(timex.Now()),
		ErrorCode:  code,
		ID:         task.ID,
		ClaimToken: postgres.UUID(task.Token),
	})
	return count == 1, err
}

func (r *Repository) Replay(ctx context.Context, failedID int64, actor, reason string, supported []jobs.Type) (int64, error) {
	if failedID < 1 || len(strings.TrimSpace(actor)) < 1 || len(strings.TrimSpace(actor)) > 200 ||
		len(strings.TrimSpace(reason)) < 1 || len(strings.TrimSpace(reason)) > 500 {
		return 0, errors.New("invalid outbox replay parameters")
	}
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	names := make([]string, len(supported))
	for i, typ := range supported {
		names[i] = string(typ)
	}
	id, err := queries.New(executor).Replay(ctx, queries.ReplayParams{
		Now:            timestamp(timex.Now()),
		Actor:          actor,
		Reason:         reason,
		FailedID:       failedID,
		SupportedTypes: names,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFailed
	}
	if err != nil {
		return 0, fmt.Errorf("replay outbox task: %w", err)
	}
	return id, nil
}

func (r *Repository) CleanupCompleted(ctx context.Context, before time.Time, limit int) (int64, error) {
	if before.IsZero() || limit < 1 || limit > 1000 {
		return 0, errors.New("invalid outbox cleanup parameters")
	}
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	return queries.New(executor).CleanupCompleted(ctx, queries.CleanupCompletedParams{
		Before:    timestamp(before),
		BatchSize: int32(limit),
	})
}

func (r *Repository) Outstanding(ctx context.Context, typ jobs.Type) (int64, error) {
	if typ == "" {
		return 0, errors.New("unknown outbox task type")
	}
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	return queries.New(executor).Outstanding(ctx, string(typ))
}

func (r *Repository) Stats(ctx context.Context, typ jobs.Type) (Stats, error) {
	if typ == "" {
		return Stats{}, errors.New("unknown outbox task type")
	}
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
		if typ == "" {
			return UnsupportedStats{}, errors.New("unknown registered outbox task type")
		}
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

func timestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func validateErrorCode(code string) error {
	if len(code) == 0 || len(code) > 100 {
		return errors.New("outbox error code must be 1 to 100 characters")
	}
	for _, r := range code {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			return errors.New("outbox error code must be redacted lowercase letters, digits, or underscores")
		}
	}
	return nil
}
