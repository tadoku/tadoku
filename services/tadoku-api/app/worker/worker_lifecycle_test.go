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
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestWorkerJobLifecycle(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, err := newWorkerFixture(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	db, client, prefix := f.database, f.client, f.prefix
	keys := []string{
		prefix + "leaderboard:global",
		prefix + "leaderboard:global:last_updated",
		prefix + "leaderboard:global:generation",
		prefix + "leaderboard:yearly:2025",
		prefix + "leaderboard:yearly:2025:last_updated",
		prefix + "leaderboard:yearly:2025:generation",
		"other:" + prefix + "leaderboard:global:last_updated",
	}

	t.Cleanup(func() {
		ctx, stop := context.WithTimeout(tenant.WithKey(context.Background(), tenant.Production()), time.Second)
		defer stop()
		if err := client.Do(ctx, client.B().Del().Key(keys[6]).Build()).Error(); err != nil {
			t.Error(err)
		}
	})

	for _, marker := range []string{keys[1], keys[6]} {
		if err := client.Do(tenantCtx, client.B().Set().Key(marker).Value("native:0").Build()).Error(); err != nil {
			t.Fatal(err)
		}
	}

	service := leaderboard.NewService(
		leaderboard.NewRepository(db.Pool),
		leaderboard.NewCache(client, time.Second, prefix),
	)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner, err := NewApplication(jobqueue.NewService(jobqueue.NewRepository(db.Pool)), service, Config{
		Concurrency:     4,
		Logger:          logger,
		Metrics:         NewMetrics(prometheus.NewRegistry()),
		ShutdownTimeout: 2 * time.Second,
	})
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

	for _, item := range []struct {
		key  string
		want int64
	}{{keys[1], 1}, {keys[6], 1}} {
		count, err := client.Do(tenantCtx, client.B().Exists().Key(item.key).Build()).AsInt64()
		if err != nil || count != item.want {
			t.Fatalf("startup marker %s: count=%d error=%v; want %d", item.key, count, err, item.want)
		}
	}

	for _, marker := range []string{keys[1], keys[4]} {
		if err := client.Do(tenantCtx, client.B().Set().Key(marker).Value("native:0").Build()).Error(); err != nil {
			t.Fatal(err)
		}
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

	for _, item := range []struct {
		key  string
		want int64
	}{{keys[1], 0}, {keys[4], 0}, {keys[6], 1}} {
		count, err := client.Do(tenantCtx, client.B().Exists().Key(item.key).Build()).AsInt64()
		if err != nil || count != item.want {
			t.Errorf("marker %s after completed job: count=%d error=%v; want %d", item.key, count, err, item.want)
		}
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
	var id int64
	if expired {
		err := db.QueryRow(ctx, `
			insert into jobs (tenant, task_type, payload, state, attempts, claim_token, lease_expires_at)
			values ('tadoku/prod', $1, $2::jsonb, 'running', 1, $3, now() - interval '1 second')
			returning id`,
			jobType, payload, uuid.New(),
		).Scan(&id)
		return id, err
	}

	err := db.QueryRow(ctx, `
		insert into jobs (tenant, task_type, payload)
		values ('tadoku/prod', $1, $2::jsonb)
		returning id`,
		jobType, payload,
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
