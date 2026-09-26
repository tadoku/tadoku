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
		prefix + "leaderboard:global",
		prefix + "leaderboard:global:last_updated",
		prefix + "leaderboard:global:generation",
		prefix + "leaderboard:yearly:2025",
		prefix + "leaderboard:yearly:2025:last_updated",
		prefix + "leaderboard:yearly:2025:generation",
		"other:" + prefix + "leaderboard:global:last_updated",
	}
	t.Cleanup(func() {
		ctx, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		if err := client.Do(ctx, client.B().Del().Key(keys...).Build()).Error(); err != nil {
			t.Error(err)
		}
	})

	for _, marker := range []string{keys[1], keys[6]} {
		if err := client.Do(t.Context(), client.B().Set().Key(marker).Value("native:0").Build()).Error(); err != nil {
			t.Fatal(err)
		}
	}

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
		return runner.Ready(), nil
	})
	for _, item := range []struct {
		key  string
		want int64
	}{{keys[1], 0}, {keys[6], 1}} {
		count, err := client.Do(t.Context(), client.B().Exists().Key(item.key).Build()).AsInt64()
		if err != nil || count != item.want {
			t.Fatalf("startup marker %s: count=%d error=%v; want %d", item.key, count, err, item.want)
		}
	}

	for _, marker := range []string{keys[1], keys[4]} {
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

	for _, marker := range []string{keys[1], keys[4]} {
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
