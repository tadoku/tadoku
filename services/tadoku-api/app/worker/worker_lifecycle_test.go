package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestWorkerJobLifecycle(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, err := newWorkerFixture(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	tenantCtx = tenant.WithKey(t.Context(), f.key)
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	db := f.database
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner, err := NewApplication(
		workerDependencies(t, db.Pool),
		Config{
			Scope:           baseWorkerScope(t),
			Concurrency:     4,
			Logger:          logger,
			Metrics:         NewMetrics(prometheus.NewRegistry()),
			ShutdownTimeout: 2 * time.Second,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); runner.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Error("worker did not stop")
		}
	})

	if err := waitFor(tenantCtx, func() (bool, error) {
		return runner.Ready(), nil
	}); err != nil {
		t.Fatal(err)
	}

	validID, err := insertJob(tenantCtx, db.Pool, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":2025}`, false)
	if err != nil {
		t.Fatal(err)
	}
	invalidID, err := insertJob(tenantCtx, db.Pool, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":0}`, false)
	if err != nil {
		t.Fatal(err)
	}
	unknownID, err := insertJob(tenantCtx, db.Pool, "future.job.v1", `{}`, false)
	if err != nil {
		t.Fatal(err)
	}
	reclaimedID, err := insertJob(tenantCtx, db.Pool, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":2025}`, true)
	if err != nil {
		t.Fatal(err)
	}

	if err := waitFor(tenantCtx, func() (bool, error) {
		var completed, failed, reclaimed, unknown string
		err := db.Pool.QueryRow(tenantCtx, `select state from jobs where id = $1`, validID).Scan(&completed)
		if err != nil {
			return false, err
		}
		err = db.Pool.QueryRow(tenantCtx, `select state from jobs where id = $1`, invalidID).Scan(&failed)
		if err != nil {
			return false, err
		}
		err = db.Pool.QueryRow(tenantCtx, `select state from jobs where id = $1`, reclaimedID).Scan(&reclaimed)
		if err != nil {
			return false, err
		}
		err = db.Pool.QueryRow(tenantCtx, `select state from jobs where id = $1`, unknownID).Scan(&unknown)
		return completed == "completed" && failed == "failed" && reclaimed == "completed" && unknown == "pending", err
	}); err != nil {
		t.Fatal(err)
	}

	var code string
	if err := db.Pool.QueryRow(tenantCtx, `select last_error from jobs where id = $1`, invalidID).Scan(&code); err != nil {
		t.Fatal(err)
	}
	if code != "invalid_payload" {
		t.Errorf("invalid job failure code = %q", code)
	}
}

func insertJob(ctx context.Context, db *pgxpool.Pool, jobType, payload string, expired bool) (int64, error) {
	key, ok := tenant.FromContext(ctx)
	if !ok {
		return 0, fmt.Errorf("worker fixture job requires a tenant")
	}

	var id int64
	if expired {
		err := db.QueryRow(ctx, `
			insert into jobs (tenant, task_type, payload, state, attempts, claim_token, lease_expires_at)
			values ($4, $1, $2::jsonb, 'running', 1, $3, now() - interval '1 second')
			returning id`,
			jobType, payload, uuid.New(), key.String(),
		).Scan(&id)
		return id, err
	}

	err := db.QueryRow(ctx, `
		insert into jobs (tenant, task_type, payload)
		values ($3, $1, $2::jsonb)
		returning id`,
		jobType, payload, key.String(),
	).Scan(&id)
	return id, err
}

func waitFor(ctx context.Context, check func() (bool, error)) error {
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		ok, err := check()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		select {
		case <-deadline.C:
			return fmt.Errorf("worker result did not converge: %w", context.DeadlineExceeded)
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}
	}
}
