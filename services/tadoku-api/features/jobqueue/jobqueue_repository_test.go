package jobqueue

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestClaimReturnsTenantForFencedTransitions(t *testing.T) {
	db := jobTestDB(t)
	key, err := tenant.Parse("e2e/worker-0123abcd")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool.Exec(t.Context(), `insert into tenants (key, kind) values ($1, 'test')`, key.String()); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Pool.Exec(t.Context(), `
		insert into jobs (tenant, task_type, payload)
		select $1, $2, '{"year":2026}'::jsonb from generate_series(1, 4)`,
		key.String(), string(jobs.LeaderboardInvalidateOfficialV1),
	); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db.Pool)
	claimed, err := repo.Claim(tenant.WithKey(t.Context(), key), jobs.LeaderboardInvalidateOfficialV1, 4, time.Minute, 3)
	if err != nil || len(claimed) != 4 {
		t.Fatalf("claim test-tenant jobs: count=%d error=%v", len(claimed), err)
	}
	for _, task := range claimed {
		if task.Tenant != key {
			t.Errorf("job %d tenant = %s; want %s", task.ID, task.Tenant, key)
		}
	}

	ctx := tenant.WithKey(t.Context(), key)
	if ok, err := repo.Complete(ctx, claimed[0]); err != nil || !ok {
		t.Fatalf("complete test-tenant job: held=%t error=%v", ok, err)
	}
	expires, held, err := repo.Renew(ctx, claimed[1], 2*time.Minute)
	if err != nil || !held || !expires.After(claimed[1].LeaseExpiresAt) {
		t.Fatalf("renew test-tenant job: expires=%s held=%t error=%v", expires, held, err)
	}

	if ok, err := repo.Retry(ctx, claimed[2], timex.Now().Add(time.Minute), "temporary", 3); err != nil || !ok {
		t.Fatalf("retry test-tenant job: held=%t error=%v", ok, err)
	}
	if ok, err := repo.Fail(ctx, claimed[3], "invalid_payload"); err != nil || !ok {
		t.Fatalf("fail test-tenant job: held=%t error=%v", ok, err)
	}

	for i, want := range []string{"completed", "running", "pending", "failed"} {
		var state, stored string
		if err := db.Pool.QueryRow(t.Context(), `select state, tenant from jobs where id=$1`, claimed[i].ID).Scan(&state, &stored); err != nil {
			t.Fatal(err)
		}
		if state != want || stored != key.String() {
			t.Errorf("job %d state=%s tenant=%s; want %s tenant=%s", claimed[i].ID, state, stored, want, key)
		}
	}
}

func TestInsertSharesBusinessTransaction(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()}
	var err error
	if _, err := db.Pool.Exec(tenantCtx, `create table business_mutation (id integer primary key)`); err != nil {
		t.Fatal(err)
	}
	rollback := errors.New("roll back business write")

	at(t, jobTestTime, func() {
		err = postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
			executor, err := postgres.Executor(ctx, db.Pool)
			if err != nil {
				return err
			}

			if _, err := executor.Exec(ctx, `insert into business_mutation (id) values (1)`); err != nil {
				return err
			}
			if _, err := insertJob(repo, ctx, task); err != nil {
				return err
			}
			return rollback
		})
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transaction error = %v, want rollback", err)
	}

	var business, queued int
	if err := db.Pool.QueryRow(tenantCtx, `select count(*) from business_mutation`).Scan(&business); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(tenantCtx, `select count(*) from jobs`).Scan(&queued); err != nil {
		t.Fatal(err)
	}

	if business != 0 || queued != 0 {
		t.Errorf("rollback left business=%d queued=%d", business, queued)
	}
}

func TestClaimLimitsKnownTypesAndFencesExpiredLeases(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	contest := jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()}
	official := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}

	at(t, jobTestTime, func() {
		for _, task := range []jobs.Job{contest, contest, official} {
			if _, err := insertJob(repo, tenantCtx, task); err != nil {
				t.Fatal(err)
			}
		}
	})

	if _, err := db.Pool.Exec(tenantCtx, `
		insert into jobs (tenant, task_type, payload, created_at, next_attempt_at)
		values ('tadoku/prod', 'future.task.v1', '{}', $1, $1)`, jobTestTime); err != nil {
		t.Fatal(err)
	}

	var first, second, third ClaimedJob
	at(t, jobTestTime, func() {
		claims, err := repo.Claim(tenantCtx, jobs.LeaderboardInvalidateContestV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("first claim = %v, %v", claims, err)
		}
		first = claims[0]
		if first.Attempts != 1 || first.Type != jobs.LeaderboardInvalidateContestV1 || first.Reclaimed || first.LeaseExpiresAt.IsZero() {
			t.Errorf("first claim = %+v", first)
		}

		claims, err = repo.Claim(tenantCtx, jobs.LeaderboardInvalidateContestV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("second claim = %v, %v", claims, err)
		}
		second = claims[0]
		if second.ID == first.ID {
			t.Error("claim reused active row")
		}

		claims, err = repo.Claim(tenantCtx, jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("other type claim = %v, %v", claims, err)
		}
		third = claims[0]
		if expires, ok, err := repo.Renew(tenantCtx, third, 2*time.Minute); err != nil || !ok || !expires.After(third.LeaseExpiresAt) {
			t.Errorf("renew = %v, %v, %v", expires, ok, err)
		}
	})

	if _, err := db.Pool.Exec(tenantCtx, `update jobs
		set lease_expires_at=clock_timestamp()-interval '1 second' where id=$1`, first.ID); err != nil {
		t.Fatal(err)
	}

	at(t, jobTestTime.Add(time.Minute), func() {
		if ok, err := repo.Complete(tenantCtx, first); err != nil || ok {
			t.Errorf("expired complete = %v, %v", ok, err)
		}
		if expires, ok, err := repo.Renew(tenantCtx, first, time.Minute); err != nil || ok || !expires.IsZero() {
			t.Errorf("expired renew = %v, %v, %v", expires, ok, err)
		}
		if ok, err := repo.Retry(tenantCtx, first, jobTestTime.Add(2*time.Minute), "temporary", 2); err != nil || ok {
			t.Errorf("expired retry = %v, %v", ok, err)
		}

		claims, err := repo.Claim(tenantCtx, jobs.LeaderboardInvalidateContestV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("reclaim = %v, %v", claims, err)
		}
		if claims[0].ID != first.ID || claims[0].Token == first.Token || claims[0].Attempts != 2 || !claims[0].Reclaimed {
			t.Errorf("reclaim = %+v, previous %+v", claims[0], first)
		}

		if ok, err := repo.Complete(tenantCtx, first); err != nil || ok {
			t.Errorf("stale token complete = %v, %v", ok, err)
		}
	})

	if _, err := db.Pool.Exec(tenantCtx, `update jobs
		set lease_expires_at=clock_timestamp()-interval '1 second' where state='running'`); err != nil {
		t.Fatal(err)
	}

	at(t, jobTestTime.Add(2*time.Minute), func() {
		if claims, err := repo.Claim(tenantCtx, jobs.LeaderboardInvalidateContestV1, 2, time.Minute, 2); err != nil || len(claims) != 1 || claims[0].ID != second.ID || claims[0].Attempts != 2 {
			t.Errorf("exhausted crash claims = %v, %v; want only second task", claims, err)
		}
		if ok, err := repo.Complete(tenantCtx, third); err != nil || ok {
			t.Errorf("other expired claim completed = %v, %v", ok, err)
		}
	})

	var failed, unknown string
	if err := db.Pool.QueryRow(tenantCtx, `select state from jobs where id=$1`, first.ID).Scan(&failed); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(tenantCtx, `select state from jobs where task_type='future.task.v1'`).Scan(&unknown); err != nil {
		t.Fatal(err)
	}
	if failed != "failed" || unknown != "pending" {
		t.Errorf("states failed=%q unknown=%q", failed, unknown)
	}

	stats, err := repo.UnsupportedStats(tenantCtx, []jobs.Type{jobs.LeaderboardInvalidateContestV1, jobs.LeaderboardInvalidateOfficialV1})
	if err != nil || stats.Pending != 1 || stats.OldestDueAt == nil || !stats.OldestDueAt.Equal(jobTestTime) {
		t.Errorf("unsupported stats = %+v, %v", stats, err)
	}
}

func TestReducedAttemptLimitFailsDuePendingTask(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}

	at(t, jobTestTime, func() {
		if _, err := insertJob(repo, tenantCtx, task); err != nil {
			t.Fatal(err)
		}

		claims, err := repo.Claim(tenantCtx, jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("claim = %v, %v", claims, err)
		}
		if ok, err := repo.Retry(tenantCtx, claims[0], jobTestTime, "temporary", 2); err != nil || !ok {
			t.Fatalf("retry = %v, %v", ok, err)
		}

		claims, err = repo.Claim(tenantCtx, jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 1)
		if err != nil || len(claims) != 0 {
			t.Errorf("reduced limit claim = %v, %v", claims, err)
		}
	})

	var state, code string
	if err := db.Pool.QueryRow(tenantCtx, `select state,last_error from jobs`).Scan(&state, &code); err != nil {
		t.Fatal(err)
	}
	if state != "failed" || code != "attempts_exhausted" {
		t.Errorf("reduced limit left state=%q code=%q", state, code)
	}
}

func TestTransitionsRejectDatabaseExpiredLease(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var err error
	var claims []ClaimedJob

	at(t, jobTestTime, func() {
		for range 4 {
			if _, err := insertJob(repo, tenantCtx, task); err != nil {
				t.Fatal(err)
			}
		}
		claims, err = repo.Claim(tenantCtx, jobs.LeaderboardInvalidateOfficialV1, 4, time.Minute, 2)
		if err != nil || len(claims) != 4 {
			t.Fatalf("claim = %v, %v", claims, err)
		}
	})

	if _, err := db.Pool.Exec(tenantCtx, `update jobs
		set lease_expires_at=clock_timestamp()-interval '1 second' where state='running'`); err != nil {
		t.Fatal(err)
	}

	at(t, jobTestTime, func() {
		if ok, err := repo.Complete(tenantCtx, claims[0]); err != nil || ok {
			t.Errorf("complete expired = %v, %v", ok, err)
		}
		if _, ok, err := repo.Renew(tenantCtx, claims[1], time.Minute); err != nil || ok {
			t.Errorf("renew expired = %v, %v", ok, err)
		}
		if ok, err := repo.Retry(tenantCtx, claims[2], jobTestTime.Add(time.Minute), "temporary", 2); err != nil || ok {
			t.Errorf("retry expired = %v, %v", ok, err)
		}
		if ok, err := repo.Fail(tenantCtx, claims[3], "invalid_payload"); err != nil || ok {
			t.Errorf("fail expired = %v, %v", ok, err)
		}
	})
}

func TestCompleteSkipsLockedClaim(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var err error
	var claim ClaimedJob

	at(t, jobTestTime, func() {
		if _, err := insertJob(repo, tenantCtx, task); err != nil {
			t.Fatal(err)
		}
		claims, err := repo.Claim(tenantCtx, jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("claim = %v, %v", claims, err)
		}
		claim = claims[0]
	})

	tx, err := db.Pool.Begin(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(tenantCtx)
	if _, err := tx.Exec(tenantCtx, `select id from jobs where id=$1 for update`, claim.ID); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(tenantCtx, time.Second)
	defer cancel()
	if ok, err := repo.Complete(ctx, claim); err != nil || ok {
		t.Errorf("complete locked claim = %v, %v", ok, err)
	}
}

func TestClaimSkipsLockedRows(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()}
	var err error
	var lockedID int64

	at(t, jobTestTime, func() {
		lockedID, err = insertJob(repo, tenantCtx, task)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := insertJob(repo, tenantCtx, task); err != nil {
			t.Fatal(err)
		}
	})

	tx, err := db.Pool.Begin(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(tenantCtx)
	if _, err := tx.Exec(tenantCtx, `select id from jobs where id=$1 for update`, lockedID); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(tenantCtx, time.Second)
	defer cancel()
	at(t, jobTestTime, func() {
		claims, err := repo.Claim(ctx, jobs.LeaderboardInvalidateContestV1, 2, time.Minute, 2)
		if err != nil || len(claims) != 1 || claims[0].ID == lockedID {
			t.Errorf("skip locked claim = %v, %v", claims, err)
		}
	})
}

func TestInsertWithoutExplicitTransaction(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	job := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}

	id, err := insertJob(repo, tenantCtx, job)
	if err != nil {
		t.Fatal(err)
	}

	var typ string
	var payload []byte
	if err := db.Pool.QueryRow(tenantCtx, `select task_type, payload from jobs where id=$1`, id).Scan(&typ, &payload); err != nil {
		t.Fatal(err)
	}
	var persisted jobs.InvalidateOfficialLeaderboardV1
	if err := json.Unmarshal(payload, &persisted); err != nil {
		t.Fatal(err)
	}
	if typ != string(job.Type()) || persisted != job {
		t.Errorf("persisted job = %s %+v, want %s %+v", typ, persisted, job.Type(), job)
	}
}

func TestInsertRejectsInvalidTransactionScope(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	job := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}

	var ended context.Context
	if err := postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
		ended = ctx
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := insertJob(repo, ended, job); err == nil {
		t.Fatal("accepted ended transaction")
	}

	other := jobTestDB(t)
	err := postgres.RunInTransaction(tenantCtx, other.Pool, func(ctx context.Context) error {
		_, err := insertJob(repo, ctx, job)
		return err
	})
	if !errors.Is(err, postgres.ErrWrongDatabase) {
		t.Errorf("wrong database = %v", err)
	}
}

func TestInsertFailureRollsBackBusinessWrite(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)

	if _, err := db.Pool.Exec(tenantCtx, `create table business_mutation (id integer primary key)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool.Exec(tenantCtx, `
		alter table jobs add constraint reject_official_for_test
		check (task_type <> 'leaderboard.invalidate_official.v1')`); err != nil {
		t.Fatal(err)
	}

	err := postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
		executor, err := postgres.Executor(ctx, db.Pool)
		if err != nil {
			return err
		}

		if _, err := executor.Exec(ctx, `insert into business_mutation (id) values (1)`); err != nil {
			return err
		}
		if _, err := insertJob(repo, ctx, jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()}); err != nil {
			return err
		}
		_, err = insertJob(repo, ctx, jobs.InvalidateOfficialLeaderboardV1{Year: 2026})
		return err
	})
	if err == nil {
		t.Fatal("expected insertion failure")
	}

	var business, queued int
	if err := db.Pool.QueryRow(tenantCtx, `select count(*) from business_mutation`).Scan(&business); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(tenantCtx, `select count(*) from jobs`).Scan(&queued); err != nil {
		t.Fatal(err)
	}
	if business != 0 || queued != 0 {
		t.Errorf("failed enqueue left business=%d queued=%d", business, queued)
	}
}
