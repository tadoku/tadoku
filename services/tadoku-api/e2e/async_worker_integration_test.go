package e2e_test

import (
	"context"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestAPIWriteStandaloneWorkerLeaderboardRead(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	directory := filepath.Join(journeysDir, "LeaderboardOutbox")
	resetJourney(t, api, directory)
	seedLeaderboardCache(t, "leaderboard:global", "hit")

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	apiLeaderboard := leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), leaderboardValkey.client, time.Second, "")
	handler, _, _, err := newTestRouterWithLeaderboardService(t.Context(), api.db.Pool, api.db.Pool, api.keto, api.kratos, logger, true, apiLeaderboard)
	if err != nil {
		t.Fatal(err)
	}

	steps := []step{
		{request: "create_log", as: user, want: http.StatusOK},
		{request: "after_worker_cache_miss", as: guest, want: http.StatusOK},
	}
	tokens, err := loadCast(steps)
	if err != nil {
		t.Fatal(err)
	}
	previous := jwt.TimeFunc
	jwt.TimeFunc = func() time.Time { return fixtureInstant }
	defer func() { jwt.TimeFunc = previous }()

	timex.TheWorld(fixtureInstant, func() {
		runRequestStep(t, api, handler, filepath.Join(directory, "01_create_log"), steps[0], tokens)
	})

	var queued int
	if err := api.db.Pool.QueryRow(t.Context(), `select count(*) from jobs`).Scan(&queued); err != nil {
		t.Fatal(err)
	}
	if queued == 0 {
		t.Fatal("API write did not enqueue generic work")
	}

	workerLeaderboard := leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), leaderboardValkey.client, time.Second, "")
	application, err := worker.NewApplication(jobqueue.NewService(jobqueue.NewRepository(api.db.Pool)), workerLeaderboard, worker.Config{
		Concurrency:     4,
		ShutdownTimeout: 2 * time.Second,
		Logger:          logger,
		Metrics:         worker.NewMetrics(prometheus.NewRegistry()),
	})
	if err != nil {
		t.Fatal(err)
	}
	workerContext, cancelWorker := context.WithCancel(context.Background())
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- application.Run(workerContext)
	}()
	t.Cleanup(func() {
		cancelWorker()
		select {
		case err := <-workerDone:
			if err != nil {
				t.Errorf("standalone worker shutdown: %v", err)
			}
		case <-time.After(10 * time.Second):
			t.Error("standalone worker did not stop")
		}
	})

	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		var completed, failed int
		if err := api.db.Pool.QueryRow(t.Context(), `select count(*) filter (where state = 'completed'), count(*) filter (where state = 'failed') from jobs`).Scan(&completed, &failed); err != nil {
			t.Fatal(err)
		}
		if failed != 0 {
			t.Fatalf("worker failed %d tasks; completed=%d queued=%d", failed, completed, queued)
		}
		if completed == queued {
			break
		}
		select {
		case <-deadline.C:
			t.Fatalf("worker did not complete %d tasks; completed=%d failed=%d outstanding=%d", queued, completed, failed, queued-completed-failed)
		case <-tick.C:
		}
	}

	timex.TheWorld(fixtureInstant, func() {
		runRequestStep(t, api, handler, filepath.Join(directory, "04_after_worker_cache_miss"), steps[1], tokens)
	})
	if score := cachedLeaderboardScore(t, "leaderboard:global"); score != 10 {
		t.Errorf("rebuilt leaderboard cache score = %v, want 10", score)
	}
}
