package e2e_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app/permissionmodel"
	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	featureaudit "github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	featureauthz "github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contestevents"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	featureprofile "github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func newWorkerApplication(t *testing.T, pool *pgxpool.Pool, scope jobqueue.Scope) *worker.Application {
	t.Helper()
	model, err := permissionmodel.New()
	if err != nil {
		t.Fatal(err)
	}

	application, err := worker.NewApplication(
		workerDependencies(pool),
		worker.Config{
			AccountDeletion: &worker.AccountDeletion{
				DB:            pool,
				Profile:       featureprofile.NewService(featureprofile.NewRepository(pool), identity.NewReader(pool)),
				Logs:          logs.NewService(logs.NewLogsRepository(pool), false),
				Contests:      contests.NewService(contests.NewContestsRepository(pool)),
				Relationships: permissions.NewRelationships(featureauthz.NewRepository(pool), model),
				Identity:      identity.NewDeletion(identity.NewRepository(pool)),
				Audit:         featureaudit.NewService(featureaudit.NewRepository(pool)),
				ContestEvents: contestevents.NewService(contestevents.NewRepository(pool)),
			},
			Scope:           scope,
			Concurrency:     4,
			ShutdownTimeout: 2 * time.Second,
			Logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
			Metrics:         worker.NewMetrics(prometheus.NewRegistry()),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return application
}

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

	application := newWorkerApplication(t, s.db.AppPool, scope)

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
