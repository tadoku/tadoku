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
	application, err := worker.NewApplication(
		workerDependencies(t, pool),
		worker.Config{
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
	eligible, arguments := eligibleJobs(forTenant)
	var queued int
	err := s.db.Pool.QueryRow(t.Context(), "select count(*) "+eligible, arguments...).Scan(&queued)
	if err != nil {
		t.Fatal(err)
	}
	if queued == 0 {
		t.Fatal("API write did not enqueue jobs")
	}

	stop := startWorker(t, s, forTenant)
	defer stop()
	if failed := waitForWorker(t, s, forTenant); failed != 0 {
		t.Fatalf("worker failed %d jobs; queued=%d", failed, queued)
	}
}

// startWorker runs the base worker, or with forTenant only that tenant's worker, until stop returns. A caller that fixes
// business time keeps its timex.TheWorld scope open until stop returns.
func startWorker(t *testing.T, s *suite, forTenant string) (stop func()) {
	t.Helper()
	scope, err := jobqueue.AllTenantsExcept(jobqueue.WorkerComponent)
	if err != nil {
		t.Fatal(err)
	}
	if forTenant != "" {
		key, err := tenant.Parse(forTenant)
		if err != nil {
			t.Fatal(err)
		}
		scope, err = jobqueue.OnlyTenant(key)
		if err != nil {
			t.Fatal(err)
		}
	}

	application := newWorkerApplication(t, s.db.AppPool, scope)

	workerContext, cancelWorker := context.WithCancel(t.Context())
	workerDone := make(chan error, 1)
	go func() { workerDone <- application.Run(workerContext) }()
	return func() {
		cancelWorker()
		select {
		case err := <-workerDone:
			if err != nil {
				t.Errorf("standalone worker shutdown: %v", err)
			}
		case <-time.After(10 * time.Second):
			t.Error("standalone worker did not stop")
		}
	}
}

// waitForWorker blocks until a job that the worker started for forTenant may claim has failed or none is pending or
// running, and returns how many have failed.
func waitForWorker(t *testing.T, s *suite, forTenant string) (failed int) {
	t.Helper()
	eligible, arguments := eligibleJobs(forTenant)

	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()

	for {
		var outstanding int
		err := s.db.Pool.QueryRow(t.Context(), `select
			count(*) filter (where state in ('pending', 'running')),
			count(*) filter (where state = 'failed') `+eligible,
			arguments...,
		).Scan(&outstanding, &failed)
		if err != nil {
			t.Fatal(err)
		}
		if failed != 0 || outstanding == 0 {
			return failed
		}
		select {
		case <-deadline.C:
			t.Fatalf("worker left %d jobs outstanding", outstanding)
		case <-tick.C:
		}
	}
}

func eligibleJobs(forTenant string) (string, []any) {
	if forTenant != "" {
		return "from jobs where tenant = $1", []any{forTenant}
	}
	return `from jobs
		where not exists (
			select 1 from tenant_overrides as o
			where o.tenant = jobs.tenant and o.component = 'tadoku-worker'
		)`, nil
}

// workerDependencies gives email no sender, so the worker holds every message.
func workerDependencies(t *testing.T, pool *pgxpool.Pool) worker.Dependencies {
	t.Helper()
	model, err := permissionmodel.New()
	if err != nil {
		t.Fatal(err)
	}

	return worker.Dependencies{
		DB:       pool,
		JobQueue: jobqueue.NewService(jobqueue.NewRepository(pool)),
		Leaderboard: leaderboard.NewService(
			leaderboard.NewRepository(pool),
			leaderboard.NewCache(leaderboardValkey.client, time.Second),
		),
		Email:            email.NewService(email.NewRepository(pool), nil),
		IdentityExpiry:   identity.NewExpiry(identity.NewRepository(pool)),
		IdentityDeletion: identity.NewDeletion(identity.NewRepository(pool)),
		Profile:          featureprofile.NewService(featureprofile.NewRepository(pool), identity.NewReader(pool)),
		Logs:             logs.NewService(logs.NewLogsRepository(pool), false),
		Contests:         contests.NewService(contests.NewContestsRepository(pool)),
		Relationships:    permissions.NewRelationships(featureauthz.NewRepository(pool), model),
		Audit:            featureaudit.NewService(featureaudit.NewRepository(pool)),
		ContestEvents:    contestevents.NewService(contestevents.NewRepository(pool)),
	}
}
