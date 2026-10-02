package e2e_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func runWorkerStep(t *testing.T, s *suite) {
	t.Helper()
	var queued int
	if err := s.db.Pool.QueryRow(t.Context(), `select count(*) from jobs`).Scan(&queued); err != nil {
		t.Fatal(err)
	}
	if queued == 0 {
		t.Fatal("API write did not enqueue jobs")
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	leaderboardService := leaderboard.NewService(leaderboard.NewRepository(s.db.Pool), leaderboardValkey.client, time.Second, "")
	application, err := worker.NewApplication(jobqueue.NewService(jobqueue.NewRepository(s.db.Pool)), leaderboardService, worker.Config{
		Concurrency:     4,
		ShutdownTimeout: 2 * time.Second,
		Logger:          logger,
		Metrics:         worker.NewMetrics(prometheus.NewRegistry()),
	})
	if err != nil {
		t.Fatal(err)
	}
	workerContext, cancelWorker := context.WithCancel(tenant.WithKey(t.Context(), tenant.Production()))
	workerDone := make(chan error, 1)
	go func() { workerDone <- application.Run(workerContext) }()
	defer func() {
		cancelWorker()
		select {
		case err := <-workerDone:
			if err != nil {
				t.Errorf("standalone worker shutdown: %v", err)
			}
		case <-time.After(10 * time.Second):
			t.Error("standalone worker did not stop")
		}
	}()

	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		var completed, failed int
		if err := s.db.Pool.QueryRow(t.Context(), `select count(*) filter (where state = 'completed'), count(*) filter (where state = 'failed') from jobs`).Scan(&completed, &failed); err != nil {
			t.Fatal(err)
		}
		if failed != 0 {
			t.Fatalf("worker failed %d jobs; completed=%d queued=%d", failed, completed, queued)
		}
		if completed == queued {
			return
		}
		select {
		case <-deadline.C:
			t.Fatalf("worker did not complete %d jobs; completed=%d failed=%d outstanding=%d", queued, completed, failed, queued-completed-failed)
		case <-tick.C:
		}
	}
}
