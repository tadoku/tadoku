package e2e_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func runWorkerStep(t *testing.T, s *suite, forTenant string) {
	t.Helper()
	scope, err := jobqueue.AllTenantsExcept("tadoku-worker")
	if err != nil {
		t.Fatal(err)
	}
	eligibleSQL := `from jobs
		where not exists (
			select 1 from tenant_overrides as o
			where o.tenant = jobs.tenant and o.component = 'tadoku-worker'
		)`
	var arguments []any
	if forTenant != "" {
		key, err := tenant.Parse(forTenant)
		if err != nil {
			t.Fatal(err)
		}
		scope, err = jobqueue.OnlyTenant(key)
		if err != nil {
			t.Fatal(err)
		}
		eligibleSQL = "from jobs where tenant = $1"
		arguments = []any{forTenant}
	}

	var queued int
	err = s.db.Pool.QueryRow(t.Context(), "select count(*) "+eligibleSQL, arguments...).Scan(&queued)
	if err != nil {
		t.Fatal(err)
	}
	if queued == 0 {
		t.Fatal("API write did not enqueue jobs")
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application, err := worker.NewApplication(
		workerDependencies(s.db.AppPool),
		worker.Config{
			Scope:           scope,
			Concurrency:     4,
			ShutdownTimeout: 2 * time.Second,
			Logger:          logger,
			Metrics:         worker.NewMetrics(prometheus.NewRegistry()),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	workerContext, cancelWorker := context.WithCancel(t.Context())
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
		err := s.db.Pool.QueryRow(t.Context(), `select
			count(*) filter (where state = 'completed'),
			count(*) filter (where state = 'failed') `+eligibleSQL,
			arguments...,
		).Scan(&completed, &failed)
		if err != nil {
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
			t.Fatalf(
				"worker did not complete %d jobs; completed=%d failed=%d outstanding=%d",
				queued,
				completed,
				failed,
				queued-completed-failed,
			)
		case <-tick.C:
		}
	}
}

// workerDependencies gives email no sender, so the worker holds every message.
func workerDependencies(pool *pgxpool.Pool) worker.Dependencies {
	return worker.Dependencies{
		JobQueue: jobqueue.NewService(jobqueue.NewRepository(pool)),
		Leaderboard: leaderboard.NewService(
			leaderboard.NewRepository(pool),
			leaderboard.NewCache(leaderboardValkey.client, time.Second),
		),
		Email:          email.NewService(email.NewRepository(pool), nil),
		IdentityExpiry: identity.NewExpiry(identity.NewRepository(pool)),
	}
}
