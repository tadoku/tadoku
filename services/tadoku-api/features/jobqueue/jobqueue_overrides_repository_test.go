package jobqueue

import (
	"context"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

const (
	frontendTenant = "e2e/frontend-0123abcd"
	workerTenant   = "e2e/worker-0123abcd"
	otherTenant    = "e2e/other-0123abcd"
)

func overrideTestDB(t *testing.T) (*testpostgres.Database, *Repository) {
	t.Helper()
	db := jobTestDB(t)
	_, err := db.Pool.Exec(t.Context(), `
		insert into tenants (key, kind) values
			('e2e/frontend-0123abcd', 'test'),
			('e2e/worker-0123abcd', 'test'),
			('e2e/other-0123abcd', 'test');

		insert into tenant_overrides (tenant, component) values
			('e2e/worker-0123abcd', 'tadoku-worker'),
			('e2e/other-0123abcd', 'other-component')`)
	if err != nil {
		t.Fatal(err)
	}
	return db, NewRepository(db.AppPool)
}

func baseQueueContext(t *testing.T) context.Context {
	t.Helper()
	scope, err := AllTenantsExcept("tadoku-worker")
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := scope.Context(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}

func branchQueueContext(t *testing.T) context.Context {
	t.Helper()
	key, err := tenant.Parse(workerTenant)
	if err != nil {
		t.Fatal(err)
	}
	scope, err := OnlyTenant(key)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := scope.Context(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}

func seedDueJobs(t *testing.T, db *testpostgres.Database, keys ...string) {
	t.Helper()
	_, err := db.Pool.Exec(t.Context(), `
		insert into jobs (tenant, task_type, payload, created_at, next_attempt_at)
		select key, 'leaderboard.invalidate_official.v1', '{"year":2026}'::jsonb, $2, $2
		from unnest($1::text[]) as key`,
		keys,
		jobTestTime,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBaseClaimIncludesFrontendTenant(t *testing.T) {
	db, repo := overrideTestDB(t)
	seedDueJobs(t, db, tenant.Production().String(), frontendTenant)

	at(t, jobTestTime, func() {
		claimed, err := repo.Claim(baseQueueContext(t), jobs.LeaderboardInvalidateOfficialV1, 10, time.Minute, 5)
		if err != nil || len(claimed) != 2 {
			t.Fatalf("base frontend claim count=%d error=%v; want 2", len(claimed), err)
		}
		seen := map[string]bool{}
		for _, task := range claimed {
			seen[task.Tenant.String()] = true
		}
		if !seen[tenant.Production().String()] || !seen[frontendTenant] {
			t.Errorf("claimed tenants=%v", seen)
		}
	})
}

func TestBaseClaimSkipsWorkerOverride(t *testing.T) {
	db, repo := overrideTestDB(t)
	seedDueJobs(t, db, tenant.Production().String(), workerTenant)

	at(t, jobTestTime, func() {
		claimed, err := repo.Claim(baseQueueContext(t), jobs.LeaderboardInvalidateOfficialV1, 10, time.Minute, 5)
		if err != nil || len(claimed) != 1 || claimed[0].Tenant != tenant.Production() {
			t.Fatalf("base claimed=%v error=%v; want only canonical job", claimed, err)
		}
	})

	var state string
	var attempts int
	err := db.Pool.QueryRow(t.Context(), "select state, attempts from jobs where tenant = $1", workerTenant).
		Scan(&state, &attempts)
	if err != nil || state != "pending" || attempts != 0 {
		t.Errorf("overridden row state=%s attempts=%d error=%v", state, attempts, err)
	}
}

func TestBaseClaimIgnoresOtherComponentOverride(t *testing.T) {
	db, repo := overrideTestDB(t)
	seedDueJobs(t, db, otherTenant)

	at(t, jobTestTime, func() {
		claimed, err := repo.Claim(baseQueueContext(t), jobs.LeaderboardInvalidateOfficialV1, 10, time.Minute, 5)
		if err != nil || len(claimed) != 1 || claimed[0].Tenant.String() != otherTenant {
			t.Fatalf("other-component claimed=%v error=%v; want own job", claimed, err)
		}
	})
}

func TestWorkerOverridePausesExhaustedJobs(t *testing.T) {
	db, repo := overrideTestDB(t)
	seedDueJobs(t, db, workerTenant)
	_, err := db.Pool.Exec(t.Context(), `
		insert into jobs (tenant, task_type, payload, state, attempts, next_attempt_at, claim_token, lease_expires_at)
		values
			($1, 'leaderboard.invalidate_official.v1', '{"year":2026}', 'pending', 5, $2, null, null),
			($1, 'leaderboard.invalidate_official.v1', '{"year":2026}', 'running', 5, $2,
				gen_random_uuid(), clock_timestamp() - interval '1 minute')`,
		workerTenant,
		jobTestTime,
	)
	if err != nil {
		t.Fatal(err)
	}
	const snapshotSQL = "select jsonb_agg(to_jsonb(jobs) order by id)::text from jobs"
	var before, after string
	if err := db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&before); err != nil {
		t.Fatal(err)
	}

	at(t, jobTestTime, func() {
		claimed, err := repo.Claim(baseQueueContext(t), jobs.LeaderboardInvalidateOfficialV1, 10, time.Minute, 5)
		if err != nil || len(claimed) != 0 {
			t.Errorf("paused claim count=%d error=%v; want no claims", len(claimed), err)
		}
	})

	if err := db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Errorf("base changed paused rows, including exhausted rows: before=%s after=%s", before, after)
	}
}

func TestBranchClaimLeavesOtherTenantsUntouched(t *testing.T) {
	db, repo := overrideTestDB(t)
	seedDueJobs(t, db, tenant.Production().String(), frontendTenant, workerTenant, otherTenant)
	_, err := db.Pool.Exec(t.Context(), `
		update jobs set attempts = 5
		where tenant = 'tadoku/prod'`)
	if err != nil {
		t.Fatal(err)
	}
	const snapshotSQL = "select jsonb_agg(to_jsonb(jobs) order by id)::text from jobs where tenant <> $1"
	var before, after string
	if err := db.Pool.QueryRow(t.Context(), snapshotSQL, workerTenant).Scan(&before); err != nil {
		t.Fatal(err)
	}

	at(t, jobTestTime, func() {
		claimed, err := repo.Claim(branchQueueContext(t), jobs.LeaderboardInvalidateOfficialV1, 10, time.Minute, 5)
		if err != nil || len(claimed) != 1 || claimed[0].Tenant.String() != workerTenant {
			t.Fatalf("branch claimed=%v error=%v; want own listed tenant", claimed, err)
		}
	})

	if err := db.Pool.QueryRow(t.Context(), snapshotSQL, workerTenant).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Errorf("branch changed other tenant rows: before=%s after=%s", before, after)
	}
}

func TestQueueCleanupIncludesOverriddenTenants(t *testing.T) {
	db, repo := overrideTestDB(t)
	_, err := db.Pool.Exec(t.Context(), `
		insert into jobs (tenant, task_type, payload, state, created_at, completed_at)
		select key, 'leaderboard.invalidate_official.v1', '{"year":2026}'::jsonb, 'completed',
			$1::timestamptz - interval '1 year', $1::timestamptz - interval '1 year'
		from (select key from tenants union all select 'e2e/worker-0123abcd') as retained`,
		jobTestTime,
	)
	if err != nil {
		t.Fatal(err)
	}

	deleted, err := repo.CleanupCompleted(branchQueueContext(t), jobTestTime, 1)
	if err != nil || deleted != 1 {
		t.Fatalf("branch cleanup deleted=%d error=%v; want 1", deleted, err)
	}
	var canonical, frontend, branch, other int
	err = db.Pool.QueryRow(t.Context(), `
		select
			count(*) filter (where tenant = 'tadoku/prod'),
			count(*) filter (where tenant = $1),
			count(*) filter (where tenant = $2),
			count(*) filter (where tenant = $3)
		from jobs`,
		frontendTenant,
		workerTenant,
		otherTenant,
	).Scan(&canonical, &frontend, &branch, &other)
	if err != nil || canonical != 1 || frontend != 1 || branch != 1 || other != 1 {
		t.Fatalf("branch cleanup rows canonical=%d frontend=%d branch=%d other=%d error=%v",
			canonical, frontend, branch, other, err)
	}

	deleted, err = repo.CleanupCompleted(baseQueueContext(t), jobTestTime, 10)
	if err != nil || deleted != 4 {
		t.Errorf("base cleanup deleted=%d error=%v; want 4 including listed tenant", deleted, err)
	}
}

func TestQueueStatsRespectWorkerOverrides(t *testing.T) {
	db, repo := overrideTestDB(t)
	_, err := db.Pool.Exec(t.Context(), `
		insert into jobs (tenant, task_type, payload, state, next_attempt_at, failed_at, claim_token, lease_expires_at)
		select key, typ, '{}'::jsonb, state,
			case when key = $1 then $2::timestamptz - interval '1 year' else $2::timestamptz end,
			case when state = 'failed' then $2::timestamptz else null end,
			case when state = 'running' then gen_random_uuid() else null end,
			case when state = 'running' then $2::timestamptz else null end
		from tenants cross join (values
			('leaderboard.invalidate_official.v1', 'pending'),
			('leaderboard.invalidate_official.v1', 'failed'),
			('future.worker.v2', 'pending'),
			('future.worker.v2', 'running'),
			('future.worker.v2', 'failed')
		) as kinds(typ, state)`,
		workerTenant,
		jobTestTime,
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name   string
		ctx    context.Context
		count  int64
		oldest time.Time
	}{
		{"base", baseQueueContext(t), 3, jobTestTime},
		{"branch", branchQueueContext(t), 1, jobTestTime.AddDate(-1, 0, 0)},
	} {
		t.Run(test.name, func(t *testing.T) {
			stats, err := repo.Stats(test.ctx, jobs.LeaderboardInvalidateOfficialV1)
			if err != nil || stats.Pending != test.count || stats.Failed != test.count ||
				stats.OldestDueAt == nil || !stats.OldestDueAt.Equal(test.oldest) {
				t.Errorf("supported stats=%+v error=%v; want count=%d oldest=%s", stats, err, test.count, test.oldest)
			}

			unknown, err := repo.UnsupportedStats(test.ctx, []jobs.Type{jobs.LeaderboardInvalidateOfficialV1})
			if err != nil || unknown.Pending != test.count || unknown.Running != test.count ||
				unknown.Failed != test.count || unknown.OldestDueAt == nil || !unknown.OldestDueAt.Equal(test.oldest) {
				t.Errorf("unsupported stats=%+v error=%v; want count=%d oldest=%s", unknown, err, test.count, test.oldest)
			}
		})
	}
}
