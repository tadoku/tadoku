package worker

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestBaseWorkerHousekeepingUsesAllTenants(t *testing.T) {
	f, err := newWorkerFixture(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	if _, err := f.db.Exec(t.Context(), `
		insert into tenants (key, kind) values ('e2e/worker-0123abcd', 'test');

		insert into jobs (tenant, task_type, payload, state, created_at, completed_at)
		select
			key, 'leaderboard.invalidate_official.v1', '{"year":2026}'::jsonb, 'completed',
			now()-interval '1 year', now()-interval '1 year'
		from tenants
		where key in ('tadoku/prod', 'e2e/worker-0123abcd');

		insert into jobs (tenant, task_type, payload, state, failed_at)
		select key, 'leaderboard.invalidate_official.v1', '{"year":2026}'::jsonb, 'failed', now()
		from tenants
		where key in ('tadoku/prod', 'e2e/worker-0123abcd');

		insert into jobs (tenant, task_type, payload)
		select key, 'future.worker.v1', '{}'::jsonb
		from tenants
		where key in ('tadoku/prod', 'e2e/worker-0123abcd');`); err != nil {
		t.Fatal(err)
	}

	pool, tracer := tracedWorkerPool(t, f.dsn)
	registry := prometheus.NewRegistry()
	handlers, err := newRegistry(handle(
		func(context.Context, jobs.InvalidateOfficialLeaderboardV1) error { return nil },
		policy{Concurrency: 1, Timeout: time.Second, MaxAttempts: 3},
	))
	if err != nil {
		t.Fatal(err)
	}

	application := &Application{runner: &runner{
		scope:           baseWorkerScope(t),
		queue:           jobqueue.NewService(jobqueue.NewRepository(pool)),
		handlers:        handlers,
		logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		metrics:         NewMetrics(registry),
		concurrency:     1,
		shutdownTimeout: time.Second,
	}}

	startWorker(t, application)
	if err := waitFor(t.Context(), func() (bool, error) { return application.Ready(), nil }); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"Claim", "CleanupCompleted", "Stats", "UnsupportedStats"} {
		observed := tracer.scopes(name)
		if len(observed) == 0 {
			t.Errorf("%s did not reach PostgreSQL", name)
		}
		for _, scope := range observed {
			if scope.key != (tenant.Key{}) || !scope.all {
				t.Errorf("%s scope = %+v; want only all-tenants", name, scope)
			}
		}
	}

	var completed int
	if err := f.db.QueryRow(t.Context(), `select count(*) from jobs where state='completed'`).Scan(&completed); err != nil {
		t.Fatal(err)
	}
	if completed != 0 {
		t.Errorf("expired completed jobs remaining = %d; want 0 across both tenants", completed)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"tadoku_worker_failed_jobs", "tadoku_worker_unsupported_pending_jobs"} {
		found := false
		for _, family := range families {
			if family.GetName() == name {
				found = true
				if len(family.Metric) != 1 || family.Metric[0].GetGauge().GetValue() != 2 {
					t.Errorf("%s = %v; want 2 across both tenants", name, family.Metric)
				}
			}
		}
		if !found {
			t.Errorf("missing metric %s", name)
		}
	}
}

func TestWorkerCleansUpExpiredCompletedJobsAtStartup(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	var id int64
	if err := f.db.QueryRow(tenantCtx, `
		insert into jobs (tenant, task_type, payload, state, created_at, completed_at)
		values ('tadoku/prod', $1, '{"year":2025}', 'completed', now() - interval '1 year', now() - interval '1 year')
		returning id`,
		string(jobs.LeaderboardInvalidateOfficialV1),
	).Scan(&id); err != nil {
		t.Fatal(err)
	}

	startWorker(t, f.runner(t, f.client, time.Second, time.Second))
	if err := waitFor(tenantCtx, func() (bool, error) {
		var exists bool
		err := f.db.QueryRow(tenantCtx, `select exists(select 1 from jobs where id = $1)`, id).Scan(&exists)
		return !exists, err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerCleanupRetainsThreeMonthsAndFailures(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	runtime := &runner{
		scope:  baseWorkerScope(t),
		queue:  jobqueue.NewService(jobqueue.NewRepository(f.db)),
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	now := time.Date(2026, time.May, 31, 12, 0, 0, 0, time.UTC)
	recent := time.Date(2026, time.April, 30, 12, 0, 0, 0, time.UTC)
	old := now.AddDate(-1, 0, 0)
	var recentID, oldID, failedID, replayID, pendingID, runningID int64

	for _, item := range []struct {
		id        *int64
		completed time.Time
	}{{&recentID, recent}, {&oldID, old}} {
		if err := f.db.QueryRow(tenantCtx, `
			insert into jobs (tenant, task_type, payload, state, created_at, completed_at)
			values ('tadoku/prod', $1, '{"year":2025}', 'completed', $2, $2)
			returning id`,
			string(jobs.LeaderboardInvalidateOfficialV1), item.completed,
		).Scan(item.id); err != nil {
			t.Fatal(err)
		}
	}

	if err := f.db.QueryRow(tenantCtx, `
		insert into jobs (tenant, task_type, payload, state, created_at, failed_at)
		values ('tadoku/prod', $1, '{"year":2025}', 'failed', $2, $2)
		returning id`,
		string(jobs.LeaderboardInvalidateOfficialV1), old,
	).Scan(&failedID); err != nil {
		t.Fatal(err)
	}

	if err := f.db.QueryRow(tenantCtx, `
		insert into jobs (
			tenant, task_type, payload, state, created_at, completed_at,
			replay_of_id, replay_actor, replay_reason
		)
		values ('tadoku/prod', $1, '{"year":2025}', 'completed', $2, $2, $3, 'retention-test', 'repaired')
		returning id`,
		string(jobs.LeaderboardInvalidateOfficialV1), old, failedID,
	).Scan(&replayID); err != nil {
		t.Fatal(err)
	}

	if _, err := f.db.Exec(tenantCtx, `
		insert into jobs (tenant, task_type, payload, state, created_at, completed_at)
		select 'tadoku/prod', $1, '{"year":2025}'::jsonb, 'completed', $2, $2
		from generate_series(1, 101)`,
		string(jobs.LeaderboardInvalidateOfficialV1), old,
	); err != nil {
		t.Fatal(err)
	}

	if err := f.db.QueryRow(tenantCtx, `
		insert into jobs (tenant, task_type, payload, created_at)
		values ('tadoku/prod', $1, '{"year":2025}', $2)
		returning id`,
		string(jobs.LeaderboardInvalidateOfficialV1), old,
	).Scan(&pendingID); err != nil {
		t.Fatal(err)
	}

	if err := f.db.QueryRow(tenantCtx, `
		insert into jobs (tenant, task_type, payload, state, created_at, claim_token, lease_expires_at)
		values ('tadoku/prod', $1, '{"year":2025}', 'running', $2, $3, now() + interval '1 hour')
		returning id`,
		string(jobs.LeaderboardInvalidateOfficialV1), old, uuid.New(),
	).Scan(&runningID); err != nil {
		t.Fatal(err)
	}

	timex.TheWorld(now, func() { runtime.cleanupCompleted(tenantCtx) })

	var expired int
	if err := f.db.QueryRow(tenantCtx, `select count(*) from jobs where state = 'completed' and completed_at = $1`, old).Scan(&expired); err != nil {
		t.Fatal(err)
	}
	if expired != 0 {
		t.Errorf("completed jobs remaining after batched cleanup = %d; want 0", expired)
	}

	for _, item := range []struct {
		name   string
		id     int64
		retain bool
	}{
		{"one-month-old success", recentID, true},
		{"expired success", oldID, false},
		{"failed original", failedID, true},
		{"expired successful replay", replayID, false},
		{"pending job", pendingID, true},
		{"running job", runningID, true},
	} {
		var exists bool
		if err := f.db.QueryRow(tenantCtx, `select exists(select 1 from jobs where id = $1)`, item.id).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists != item.retain {
			t.Errorf("%s retained=%t; want %t", item.name, exists, item.retain)
		}
	}
}
