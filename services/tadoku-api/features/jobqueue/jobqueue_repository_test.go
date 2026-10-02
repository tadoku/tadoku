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
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

var jobTestTime = time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

func TestClaimReturnsTenantForFencedTransitions(t *testing.T) {
	db := jobTestDB(t)
	key, err := tenant.Parse("e2e/worker-0123abcd")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool.Exec(t.Context(), `insert into tenants (key, kind) values ($1, 'test')`, key.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool.Exec(t.Context(), `insert into jobs (tenant, task_type, payload)
		select $1, $2, '{"year":2026}'::jsonb from generate_series(1, 4)`, key.String(), string(jobs.LeaderboardInvalidateOfficialV1)); err != nil {
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

func jobTestDB(t *testing.T) *testpostgres.Database {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	t.Helper()
	db, err := testpostgres.New(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}

func at(t *testing.T, instant time.Time, work func()) {
	t.Helper()
	timex.TheWorld(instant, work)
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
	if _, err := db.Pool.Exec(tenantCtx, `insert into jobs (tenant, task_type,payload,created_at,next_attempt_at)
		values ('tadoku/prod', 'future.task.v1','{}',$1,$1)`, jobTestTime); err != nil {
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

func TestRetryReplayAndRetention(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var err error
	var id int64
	at(t, jobTestTime, func() { id, err = insertJob(repo, tenantCtx, task) })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Replay(tenantCtx, id, "operator", "repair", []jobs.Type{jobs.LeaderboardInvalidateOfficialV1}); !errors.Is(err, ErrNotFailed) {
		t.Errorf("replay pending = %v", err)
	}
	var claim ClaimedJob
	at(t, jobTestTime, func() {
		claims, err := repo.Claim(tenantCtx, jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("claim = %v, %v", claims, err)
		}
		claim = claims[0]
		if ok, err := repo.Retry(tenantCtx, claim, jobTestTime.Add(time.Minute), "temporary", 2); err != nil || !ok {
			t.Fatalf("retry = %v, %v", ok, err)
		}
		if claims, err := repo.Claim(tenantCtx, jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2); err != nil || len(claims) != 0 {
			t.Errorf("early claim = %v, %v", claims, err)
		}
	})
	at(t, jobTestTime.Add(time.Minute), func() {
		claims, err := repo.Claim(tenantCtx, jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 || claims[0].Attempts != 2 {
			t.Fatalf("retry claim = %v, %v", claims, err)
		}
		claim = claims[0]
		if ok, err := repo.Retry(tenantCtx, claim, jobTestTime.Add(2*time.Minute), "temporary", 2); err != nil || !ok {
			t.Fatalf("exhausted retry = %v, %v", ok, err)
		}
	})
	var replayID int64
	at(t, jobTestTime.Add(2*time.Minute), func() {
		replayID, err = repo.Replay(tenantCtx, id, "operator", "repair", []jobs.Type{jobs.LeaderboardInvalidateOfficialV1})
	})
	if err != nil || replayID == 0 {
		t.Fatalf("replay = %d, %v", replayID, err)
	}
	var replayOf int64
	if err := db.Pool.QueryRow(tenantCtx, `select replay_of_id from jobs where id=$1`, replayID).Scan(&replayOf); err != nil {
		t.Fatal(err)
	}
	if replayOf != id {
		t.Errorf("replay source = %d, want %d", replayOf, id)
	}
	at(t, jobTestTime.Add(2*time.Minute), func() {
		claims, err := repo.Claim(tenantCtx, jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("replay claim = %v, %v", claims, err)
		}
		if ok, err := repo.Complete(tenantCtx, claims[0]); err != nil || !ok {
			t.Fatalf("complete replay = %v, %v", ok, err)
		}
	})
	if _, err := db.Pool.Exec(tenantCtx, `insert into jobs (tenant, task_type,payload,state,completed_at,created_at) values ('tadoku/prod', 'leaderboard.invalidate_official.v1','{}','completed',$1,$1)`, jobTestTime); err != nil {
		t.Fatal(err)
	}
	deleted, err := repo.CleanupCompleted(tenantCtx, jobTestTime.AddDate(0, 3, 0).Add(time.Minute), 10)
	if err != nil || deleted != 1 {
		t.Errorf("cleanup deleted = %d, %v; want only unlinked completion", deleted, err)
	}
	var retained int
	if err := db.Pool.QueryRow(tenantCtx, `select count(*) from jobs where id=$1`, replayID).Scan(&retained); err != nil {
		t.Fatal(err)
	}
	if retained != 1 {
		t.Error("cleanup removed replay lineage")
	}
}

func insertJob(repo *Repository, ctx context.Context, job jobs.Job) (int64, error) {
	payload, err := json.Marshal(job)
	if err != nil {
		return 0, err
	}
	return repo.Insert(ctx, job.Type(), payload)
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

func TestUnsupportedStatsIncludesRunningAndFailedVersions(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	_, err := db.Pool.Exec(tenantCtx, `insert into jobs (tenant, task_type, payload, state, claim_token, lease_expires_at, failed_at, completed_at)
 values ('tadoku/prod', 'future.job.v2','{}','pending',null,null,null,null),
 ('tadoku/prod', 'future.job.v2','{}','running',gen_random_uuid(),clock_timestamp()-interval '1 second',null,null),
 ('tadoku/prod', 'future.job.v2','{}','failed',null,null,clock_timestamp(),null),
 ('tadoku/prod', 'future.job.v2','{}','completed',null,null,null,clock_timestamp()),
 ('tadoku/prod', 'leaderboard.invalidate_contest.v1','{}','pending',null,null,null,null)`)
	if err != nil {
		t.Fatal(err)
	}
	stats, err := repo.UnsupportedStats(tenantCtx, []jobs.Type{jobs.LeaderboardInvalidateContestV1})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Pending != 1 || stats.Running != 1 || stats.Failed != 1 || stats.OldestDueAt == nil {
		t.Errorf("unsupported stats = %+v; want pending=1 running=1 failed=1 and due time", stats)
	}
}

func TestReplayRequiresRegisteredVersion(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	var id int64
	if err := db.Pool.QueryRow(tenantCtx, `insert into jobs (tenant, task_type,payload,state,failed_at) values ('tadoku/prod', 'future.job.v2','{}','failed',clock_timestamp()) returning id`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Replay(tenantCtx, id, "operator", "repair", []jobs.Type{jobs.LeaderboardInvalidateContestV1}); !errors.Is(err, ErrNotFailed) {
		t.Fatalf("unregistered replay = %v", err)
	}
	replayID, err := repo.Replay(tenantCtx, id, "operator", "repair", []jobs.Type{"future.job.v2"})
	if err != nil || replayID <= id {
		t.Fatalf("registered replay = %d, %v", replayID, err)
	}
	claims, err := repo.Claim(tenantCtx, "future.job.v2", 1, time.Minute, 2)
	if err != nil || len(claims) != 1 || claims[0].ID != replayID {
		t.Errorf("generic replay claim = %+v, %v", claims, err)
	}
}

func TestCleanupCompletedRetainsThreeUTCCalendarMonths(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	cases := []struct {
		name   string
		now    time.Time
		cutoff time.Time
	}{
		{"month_end", time.Date(2026, 5, 31, 12, 30, 0, 0, time.UTC), time.Date(2026, 2, 28, 12, 30, 0, 0, time.UTC)},
		{"leap_year", time.Date(2024, 5, 31, 12, 30, 0, 0, time.UTC), time.Date(2024, 2, 29, 12, 30, 0, 0, time.UTC)},
		{"year_boundary", time.Date(2026, 1, 31, 12, 30, 0, 0, time.UTC), time.Date(2025, 10, 31, 12, 30, 0, 0, time.UTC)},
		{"utc_instant", time.Date(2026, 6, 1, 1, 30, 0, 0, time.FixedZone("UTC+13", 13*60*60)), time.Date(2026, 2, 28, 12, 30, 0, 0, time.UTC)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := jobTestDB(t)
			repo := NewRepository(db.Pool)
			_, err := db.Pool.Exec(tenantCtx, `insert into jobs (tenant, task_type,payload,state,completed_at)
    values ('tadoku/prod', 'retention.test.v1','{}','completed',$1), ('tadoku/prod', 'retention.test.v1','{}','completed',$2), ('tadoku/prod', 'retention.test.v1','{}','completed',$3)`, tc.cutoff.Add(-time.Microsecond), tc.cutoff, tc.cutoff.Add(time.Microsecond))
			if err != nil {
				t.Fatal(err)
			}
			err = postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
				executor, err := postgres.Executor(ctx, db.Pool)
				if err != nil {
					return err
				}
				if _, err := executor.Exec(ctx, `set local time zone 'Pacific/Auckland'`); err != nil {
					return err
				}
				deleted, err := repo.CleanupCompleted(ctx, tc.now, 10)
				if err != nil {
					return err
				}
				if deleted != 1 {
					t.Errorf("deleted %d completed jobs, want only the row strictly before %s", deleted, tc.cutoff)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			var retained int
			if err := db.Pool.QueryRow(tenantCtx, `select count(*) from jobs where completed_at >= $1`, tc.cutoff).Scan(&retained); err != nil {
				t.Fatal(err)
			}
			if retained != 2 {
				t.Errorf("retained %d boundary/newer completions, want 2", retained)
			}
		})
	}
}

func TestCleanupCompletedExpiresSuccessfulReplayAndPreservesOtherStates(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	old := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.Pool.Exec(tenantCtx, `insert into jobs (tenant, id,task_type,payload,state,created_at,next_attempt_at,failed_at,claim_token,lease_expires_at,completed_at,replay_of_id,replay_actor,replay_reason) values
  ('tadoku/prod', 1,'retention.test.v1','{}','failed',$1,$1,$1,null,null,null,null,null,null),
  ('tadoku/prod', 2,'retention.test.v1','{}','completed',$1,$1,null,null,null,$1,1,'operator','repaired'),
  ('tadoku/prod', 3,'retention.test.v1','{}','pending',$1,$1,null,null,null,null,null,null,null),
  ('tadoku/prod', 4,'retention.test.v1','{}','running',$1,$1,null,gen_random_uuid(),$1,null,null,null,null),
  ('tadoku/prod', 5,'retention.test.v1','{}','completed',$1,$1,null,null,null,$1,null,null,null),
  ('tadoku/prod', 6,'retention.test.v1','{}','completed',$1,$1,null,null,null,$1,null,null,null),
  ('tadoku/prod', 7,'retention.test.v1','{}','pending',$1,$1,null,null,null,null,6,'operator','linked')`, old)
	if err != nil {
		t.Fatal(err)
	}
	deleted, err := repo.CleanupCompleted(tenantCtx, now, 1)
	if err != nil || deleted != 1 {
		t.Fatalf("bounded cleanup = %d, %v", deleted, err)
	}
	var replayRetained int
	if err := db.Pool.QueryRow(tenantCtx, `select count(*) from jobs where id=2`).Scan(&replayRetained); err != nil {
		t.Fatal(err)
	}
	if replayRetained != 0 {
		t.Errorf("expired successful replay was retained")
	}
	deleted, err = repo.CleanupCompleted(tenantCtx, now, 10)
	if err != nil || deleted != 1 {
		t.Fatalf("remaining cleanup = %d, %v", deleted, err)
	}
	var retained int
	if err := db.Pool.QueryRow(tenantCtx, `select count(*) from jobs where id in (1,3,4,6,7)`).Scan(&retained); err != nil {
		t.Fatal(err)
	}
	if retained != 5 {
		t.Errorf("retained %d failed/active/referenced jobs, want 5", retained)
	}
}

func TestInsertFailureRollsBackBusinessWrite(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	if _, err := db.Pool.Exec(tenantCtx, `create table business_mutation (id integer primary key)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool.Exec(tenantCtx, `alter table jobs add constraint reject_official_for_test check (task_type <> 'leaderboard.invalidate_official.v1')`); err != nil {
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
