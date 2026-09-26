package worker

import (
	"context"
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
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	valkeygo "github.com/valkey-io/valkey-go"
)

func TestWorkerOutboxJourney(t *testing.T) {
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	rawURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}
	option, err := valkeygo.ParseURL(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	option.SelectDB = 13
	option.ForceSingleClient = true
	option.DisableRetry = true
	client, err := valkeygo.NewClient(option)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)

	prefix := "test:" + uuid.NewString() + ":"
	keys := []string{
		prefix + "leaderboard:ready",
		prefix + "leaderboard:global",
		prefix + "leaderboard:global:last_updated",
		prefix + "leaderboard:global:generation",
		prefix + "leaderboard:yearly:2025",
		prefix + "leaderboard:yearly:2025:last_updated",
		prefix + "leaderboard:yearly:2025:generation",
	}
	t.Cleanup(func() {
		ctx, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		if err := client.Do(ctx, client.B().Del().Key(keys...).Build()).Error(); err != nil {
			t.Error(err)
		}
	})

	service := leaderboard.NewService(leaderboard.NewRepository(db.Pool), client, time.Second, prefix)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner, err := NewApplication(jobqueue.NewService(jobqueue.NewRepository(db.Pool)), service, Config{Concurrency: 4, Logger: logger, Metrics: NewMetrics(prometheus.NewRegistry()), ShutdownTimeout: 2 * time.Second})
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

	waitFor(t, func() (bool, error) {
		return serviceCacheReady(t.Context(), client, keys[0])
	})
	for _, marker := range []string{keys[2], keys[5]} {
		if err := client.Do(t.Context(), client.B().Set().Key(marker).Value("native:0").Build()).Error(); err != nil {
			t.Fatal(err)
		}
	}

	validID := insertTask(t, db.Pool, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":2025}`, false)
	invalidID := insertTask(t, db.Pool, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":0}`, false)
	unknownID := insertTask(t, db.Pool, "future.task.v1", `{}`, false)
	reclaimedID := insertTask(t, db.Pool, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":2025}`, true)

	waitFor(t, func() (bool, error) {
		var completed, failed, reclaimed, unknown string
		err := db.Pool.QueryRow(t.Context(), `select state from jobs where id = $1`, validID).Scan(&completed)
		if err != nil {
			return false, err
		}
		err = db.Pool.QueryRow(t.Context(), `select state from jobs where id = $1`, invalidID).Scan(&failed)
		if err != nil {
			return false, err
		}
		err = db.Pool.QueryRow(t.Context(), `select state from jobs where id = $1`, reclaimedID).Scan(&reclaimed)
		if err != nil {
			return false, err
		}
		err = db.Pool.QueryRow(t.Context(), `select state from jobs where id = $1`, unknownID).Scan(&unknown)
		return completed == "completed" && failed == "failed" && reclaimed == "completed" && unknown == "pending", err
	})

	for _, marker := range []string{keys[2], keys[5]} {
		count, err := client.Do(t.Context(), client.B().Exists().Key(marker).Build()).AsInt64()
		if err != nil || count != 0 {
			t.Errorf("marker %s remains after completed task: count=%d error=%v", marker, count, err)
		}
	}
	var code string
	if err := db.Pool.QueryRow(t.Context(), `select last_error from jobs where id = $1`, invalidID).Scan(&code); err != nil {
		t.Fatal(err)
	}
	if code != "invalid_payload" {
		t.Errorf("invalid task failure code = %q", code)
	}
	for _, state := range []string{"pending", "running", "failed"} {
		_, err := db.Pool.Exec(t.Context(), `update jobs set state = $1,
   claim_token = case when $1 = 'running' then $2::uuid else null end,
   lease_expires_at = case when $1 = 'running' then now() + interval '1 minute' else null end,
   failed_at = case when $1 = 'failed' then now() else null end where id = $3`, state, uuid.New(), unknownID)
		if err != nil {
			t.Fatal(err)
		}
		if err := runner.refreshReadiness(t.Context()); err != nil {
			t.Fatal(err)
		}
		ready, err := serviceCacheReady(t.Context(), client, keys[0])
		if err != nil {
			t.Fatal(err)
		}
		if ready {
			t.Fatalf("cache readiness published while unsupported future.task.v1 remains %s", state)
		}
	}
	if _, err := db.Pool.Exec(t.Context(), `delete from jobs where id = $1`, unknownID); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() (bool, error) {
		return serviceCacheReady(t.Context(), client, keys[0])
	})
}

func insertTask(t *testing.T, db *pgxpool.Pool, taskType, payload string, expired bool) int64 {
	t.Helper()
	var id int64
	if expired {
		err := db.QueryRow(t.Context(), `insert into jobs (task_type, payload, state, attempts, claim_token, lease_expires_at)
			values ($1, $2::jsonb, 'running', 1, $3, now() - interval '1 second') returning id`, taskType, payload, uuid.New()).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	if err := db.QueryRow(t.Context(), `insert into jobs (task_type, payload) values ($1, $2::jsonb) returning id`, taskType, payload).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func serviceCacheReady(ctx context.Context, client valkeygo.Client, key string) (bool, error) {
	count, err := client.Do(ctx, client.B().Exists().Key(key).Build()).AsInt64()
	return count == 1, err
}

func waitFor(t *testing.T, check func() (bool, error)) {
	t.Helper()
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		ok, err := check()
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			return
		}
		select {
		case <-deadline.C:
			t.Fatal("worker result did not converge")
		case <-tick.C:
		}
	}
}
