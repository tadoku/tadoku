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
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

var outboxTestTime = time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

func outboxTestDB(t *testing.T) *testpostgres.Database {
	t.Helper()
	db, err := testpostgres.New(t.Context())
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

func TestEnqueueSharesBusinessTransaction(t *testing.T) {
	db := outboxTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()}
	var err error
	if _, err := db.Pool.Exec(t.Context(), `create table business_mutation (id integer primary key)`); err != nil {
		t.Fatal(err)
	}
	rollback := errors.New("roll back business write")

	at(t, outboxTestTime, func() {
		err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
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
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from business_mutation`).Scan(&business); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from async_outbox`).Scan(&queued); err != nil {
		t.Fatal(err)
	}
	if business != 0 || queued != 0 {
		t.Errorf("rollback left business=%d queued=%d", business, queued)
	}
}

func TestClaimLimitsKnownTypesAndFencesExpiredLeases(t *testing.T) {
	db := outboxTestDB(t)
	repo := NewRepository(db.Pool)
	contest := jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()}
	official := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	at(t, outboxTestTime, func() {
		for _, task := range []jobs.Job{contest, contest, official} {
			if _, err := insertJob(repo, t.Context(), task); err != nil {
				t.Fatal(err)
			}
		}
	})
	if _, err := db.Pool.Exec(t.Context(), `insert into async_outbox (task_type,payload,created_at,next_attempt_at)
		values ('future.task.v1','{}',$1,$1)`, outboxTestTime); err != nil {
		t.Fatal(err)
	}

	var first, second, third ClaimedJob
	at(t, outboxTestTime, func() {
		claims, err := repo.Claim(t.Context(), jobs.InvalidateContest, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("first claim = %v, %v", claims, err)
		}
		first = claims[0]
		if first.Attempts != 1 || first.Type != jobs.InvalidateContest || first.Reclaimed || first.LeaseExpiresAt.IsZero() {
			t.Errorf("first claim = %+v", first)
		}
		claims, err = repo.Claim(t.Context(), jobs.InvalidateContest, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("second claim = %v, %v", claims, err)
		}
		second = claims[0]
		if second.ID == first.ID {
			t.Error("claim reused active row")
		}
		claims, err = repo.Claim(t.Context(), jobs.InvalidateOfficial, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("other type claim = %v, %v", claims, err)
		}
		third = claims[0]
		if expires, ok, err := repo.Renew(t.Context(), third, 2*time.Minute); err != nil || !ok || !expires.After(third.LeaseExpiresAt) {
			t.Errorf("renew = %v, %v, %v", expires, ok, err)
		}
		if claims, err := repo.Claim(t.Context(), jobs.Type(""), 1, time.Minute, 2); err == nil || len(claims) != 0 {
			t.Errorf("unsupported claim = %v, %v", claims, err)
		}
	})
	if _, err := db.Pool.Exec(t.Context(), `update async_outbox
		set lease_expires_at=clock_timestamp()-interval '1 second' where id=$1`, first.ID); err != nil {
		t.Fatal(err)
	}

	at(t, outboxTestTime.Add(time.Minute), func() {
		if ok, err := repo.Complete(t.Context(), first); err != nil || ok {
			t.Errorf("expired complete = %v, %v", ok, err)
		}
		if expires, ok, err := repo.Renew(t.Context(), first, time.Minute); err != nil || ok || !expires.IsZero() {
			t.Errorf("expired renew = %v, %v, %v", expires, ok, err)
		}
		if ok, err := repo.Retry(t.Context(), first, outboxTestTime.Add(2*time.Minute), "temporary", 2); err != nil || ok {
			t.Errorf("expired retry = %v, %v", ok, err)
		}
		claims, err := repo.Claim(t.Context(), jobs.InvalidateContest, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("reclaim = %v, %v", claims, err)
		}
		if claims[0].ID != first.ID || claims[0].Token == first.Token || claims[0].Attempts != 2 || !claims[0].Reclaimed {
			t.Errorf("reclaim = %+v, previous %+v", claims[0], first)
		}
		if ok, err := repo.Complete(t.Context(), first); err != nil || ok {
			t.Errorf("stale token complete = %v, %v", ok, err)
		}
	})
	if _, err := db.Pool.Exec(t.Context(), `update async_outbox
		set lease_expires_at=clock_timestamp()-interval '1 second' where state='running'`); err != nil {
		t.Fatal(err)
	}

	at(t, outboxTestTime.Add(2*time.Minute), func() {
		if claims, err := repo.Claim(t.Context(), jobs.InvalidateContest, 2, time.Minute, 2); err != nil || len(claims) != 1 || claims[0].ID != second.ID || claims[0].Attempts != 2 {
			t.Errorf("exhausted crash claims = %v, %v; want only second task", claims, err)
		}
		if ok, err := repo.Complete(t.Context(), third); err != nil || ok {
			t.Errorf("other expired claim completed = %v, %v", ok, err)
		}
	})
	var failed, unknown string
	if err := db.Pool.QueryRow(t.Context(), `select state from async_outbox where id=$1`, first.ID).Scan(&failed); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(t.Context(), `select state from async_outbox where task_type='future.task.v1'`).Scan(&unknown); err != nil {
		t.Fatal(err)
	}
	if failed != "failed" || unknown != "pending" {
		t.Errorf("states failed=%q unknown=%q", failed, unknown)
	}
	stats, err := repo.UnsupportedStats(t.Context(), []jobs.Type{jobs.InvalidateContest, jobs.InvalidateOfficial})
	if err != nil || stats.Pending != 1 || stats.OldestDueAt == nil || !stats.OldestDueAt.Equal(outboxTestTime) {
		t.Errorf("unsupported stats = %+v, %v", stats, err)
	}
}

func TestReducedAttemptLimitFailsDuePendingTask(t *testing.T) {
	db := outboxTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	at(t, outboxTestTime, func() {
		if _, err := insertJob(repo, t.Context(), task); err != nil {
			t.Fatal(err)
		}
		claims, err := repo.Claim(t.Context(), jobs.InvalidateOfficial, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("claim = %v, %v", claims, err)
		}
		if ok, err := repo.Retry(t.Context(), claims[0], outboxTestTime, "temporary", 2); err != nil || !ok {
			t.Fatalf("retry = %v, %v", ok, err)
		}
		claims, err = repo.Claim(t.Context(), jobs.InvalidateOfficial, 1, time.Minute, 1)
		if err != nil || len(claims) != 0 {
			t.Errorf("reduced limit claim = %v, %v", claims, err)
		}
	})
	var state, code string
	if err := db.Pool.QueryRow(t.Context(), `select state,last_error from async_outbox`).Scan(&state, &code); err != nil {
		t.Fatal(err)
	}
	if state != "failed" || code != "attempts_exhausted" {
		t.Errorf("reduced limit left state=%q code=%q", state, code)
	}
}

func TestTransitionsRejectDatabaseExpiredLease(t *testing.T) {
	db := outboxTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var err error
	var claims []ClaimedJob
	at(t, outboxTestTime, func() {
		for range 4 {
			if _, err := insertJob(repo, t.Context(), task); err != nil {
				t.Fatal(err)
			}
		}
		claims, err = repo.Claim(t.Context(), jobs.InvalidateOfficial, 4, time.Minute, 2)
		if err != nil || len(claims) != 4 {
			t.Fatalf("claim = %v, %v", claims, err)
		}
	})
	if _, err := db.Pool.Exec(t.Context(), `update async_outbox
		set lease_expires_at=clock_timestamp()-interval '1 second' where state='running'`); err != nil {
		t.Fatal(err)
	}
	at(t, outboxTestTime, func() {
		if ok, err := repo.Complete(t.Context(), claims[0]); err != nil || ok {
			t.Errorf("complete expired = %v, %v", ok, err)
		}
		if _, ok, err := repo.Renew(t.Context(), claims[1], time.Minute); err != nil || ok {
			t.Errorf("renew expired = %v, %v", ok, err)
		}
		if ok, err := repo.Retry(t.Context(), claims[2], outboxTestTime.Add(time.Minute), "temporary", 2); err != nil || ok {
			t.Errorf("retry expired = %v, %v", ok, err)
		}
		if ok, err := repo.Fail(t.Context(), claims[3], "invalid_payload"); err != nil || ok {
			t.Errorf("fail expired = %v, %v", ok, err)
		}
	})
}

func TestCompleteSkipsLockedClaim(t *testing.T) {
	db := outboxTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var err error
	var claim ClaimedJob
	at(t, outboxTestTime, func() {
		if _, err := insertJob(repo, t.Context(), task); err != nil {
			t.Fatal(err)
		}
		claims, err := repo.Claim(t.Context(), jobs.InvalidateOfficial, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("claim = %v, %v", claims, err)
		}
		claim = claims[0]
	})
	tx, err := db.Pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(t.Context())
	if _, err := tx.Exec(t.Context(), `select id from async_outbox where id=$1 for update`, claim.ID); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if ok, err := repo.Complete(ctx, claim); err != nil || ok {
		t.Errorf("complete locked claim = %v, %v", ok, err)
	}
}

func TestClaimSkipsLockedRows(t *testing.T) {
	db := outboxTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()}
	var err error
	var lockedID int64
	at(t, outboxTestTime, func() {
		lockedID, err = insertJob(repo, t.Context(), task)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := insertJob(repo, t.Context(), task); err != nil {
			t.Fatal(err)
		}
	})
	tx, err := db.Pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(t.Context())
	if _, err := tx.Exec(t.Context(), `select id from async_outbox where id=$1 for update`, lockedID); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	at(t, outboxTestTime, func() {
		claims, err := repo.Claim(ctx, jobs.InvalidateContest, 2, time.Minute, 2)
		if err != nil || len(claims) != 1 || claims[0].ID == lockedID {
			t.Errorf("skip locked claim = %v, %v", claims, err)
		}
	})
}

func TestRetryReplayAndRetention(t *testing.T) {
	db := outboxTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var err error
	var id int64
	at(t, outboxTestTime, func() { id, err = insertJob(repo, t.Context(), task) })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Replay(t.Context(), id, "operator", "repair", []jobs.Type{jobs.InvalidateOfficial}); !errors.Is(err, ErrNotFailed) {
		t.Errorf("replay pending = %v", err)
	}
	var claim ClaimedJob
	at(t, outboxTestTime, func() {
		claims, err := repo.Claim(t.Context(), jobs.InvalidateOfficial, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("claim = %v, %v", claims, err)
		}
		claim = claims[0]
		if ok, err := repo.Retry(t.Context(), claim, outboxTestTime.Add(time.Minute), "temporary", 2); err != nil || !ok {
			t.Fatalf("retry = %v, %v", ok, err)
		}
		if count, err := repo.Outstanding(t.Context(), jobs.InvalidateOfficial); err != nil || count != 1 {
			t.Errorf("outstanding delayed = %d, %v", count, err)
		}
		if claims, err := repo.Claim(t.Context(), jobs.InvalidateOfficial, 1, time.Minute, 2); err != nil || len(claims) != 0 {
			t.Errorf("early claim = %v, %v", claims, err)
		}
	})
	at(t, outboxTestTime.Add(time.Minute), func() {
		claims, err := repo.Claim(t.Context(), jobs.InvalidateOfficial, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 || claims[0].Attempts != 2 {
			t.Fatalf("retry claim = %v, %v", claims, err)
		}
		claim = claims[0]
		if ok, err := repo.Retry(t.Context(), claim, outboxTestTime.Add(2*time.Minute), "temporary", 2); err != nil || !ok {
			t.Fatalf("exhausted retry = %v, %v", ok, err)
		}
	})
	var replayID int64
	at(t, outboxTestTime.Add(2*time.Minute), func() {
		replayID, err = repo.Replay(t.Context(), id, "operator", "repair", []jobs.Type{jobs.InvalidateOfficial})
	})
	if err != nil || replayID == 0 {
		t.Fatalf("replay = %d, %v", replayID, err)
	}
	var replayOf int64
	if err := db.Pool.QueryRow(t.Context(), `select replay_of_id from async_outbox where id=$1`, replayID).Scan(&replayOf); err != nil {
		t.Fatal(err)
	}
	if replayOf != id {
		t.Errorf("replay source = %d, want %d", replayOf, id)
	}
	at(t, outboxTestTime.Add(2*time.Minute), func() {
		claims, err := repo.Claim(t.Context(), jobs.InvalidateOfficial, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("replay claim = %v, %v", claims, err)
		}
		if ok, err := repo.Complete(t.Context(), claims[0]); err != nil || !ok {
			t.Fatalf("complete replay = %v, %v", ok, err)
		}
	})
	if _, err := db.Pool.Exec(t.Context(), `insert into async_outbox (task_type,payload,state,completed_at,created_at) values ('leaderboard.invalidate_official.v1','{}','completed',$1,$1)`, outboxTestTime); err != nil {
		t.Fatal(err)
	}
	deleted, err := repo.CleanupCompleted(t.Context(), outboxTestTime.Add(3*time.Minute), 10)
	if err != nil || deleted != 1 {
		t.Errorf("cleanup deleted = %d, %v; want only unlinked completion", deleted, err)
	}
	var retained int
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from async_outbox where id=$1`, replayID).Scan(&retained); err != nil {
		t.Fatal(err)
	}
	if retained != 1 {
		t.Error("cleanup removed replay lineage")
	}
}

// insertJob seeds repository tests through the same transaction boundary used by applications.
func insertJob(repo *Repository, ctx context.Context, job jobs.Job) (int64, error) {
	payload, err := json.Marshal(job)
	if err != nil {
		return 0, err
	}
	var id int64
	insert := func(ctx context.Context) error {
		id, err = repo.Insert(ctx, job.Type(), payload)
		return err
	}
	if _, txErr := postgres.TransactionExecutor(ctx, repo.db); errors.Is(txErr, postgres.ErrTransactionRequired) {
		err = postgres.RunInTransaction(ctx, repo.db, insert)
	} else {
		err = insert(ctx)
	}
	return id, err
}

func TestEnqueueRequiresApplicationTransaction(t *testing.T) {
	db := outboxTestDB(t)
	queue := NewService(NewRepository(db.Pool))
	job := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	for _, batch := range [][]jobs.Job{nil, {job}} {
		if err := queue.Enqueue(t.Context(), batch...); !errors.Is(err, postgres.ErrTransactionRequired) {
			t.Errorf("enqueue without transaction = %v", err)
		}
	}
	var ended context.Context
	if err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		ended = ctx
		return queue.Enqueue(ctx)
	}); err != nil {
		t.Fatal(err)
	}
	if err := queue.Enqueue(ended, job); err == nil {
		t.Fatal("accepted ended transaction")
	}
	other := outboxTestDB(t)
	err := postgres.RunInTransaction(t.Context(), other.Pool, func(ctx context.Context) error { return queue.Enqueue(ctx, job) })
	if !errors.Is(err, postgres.ErrWrongDatabase) {
		t.Errorf("wrong database = %v", err)
	}
}

func TestEnqueueRejectsInvalidBatchBeforeInsertion(t *testing.T) {
	db := outboxTestDB(t)
	queue := NewService(NewRepository(db.Pool))
	valid := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var typedNil *jobs.InvalidateOfficialLeaderboardV1
	for _, invalid := range []jobs.Job{nil, typedNil, jobs.InvalidateOfficialLeaderboardV1{}, jobs.InvalidateContestLeaderboardV1{}} {
		err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
			if err := queue.Enqueue(ctx, valid, invalid); err == nil {
				t.Fatal("accepted invalid batch")
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from async_outbox`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("invalid batches persisted %d jobs", count)
	}
}

func TestEnqueueFailureRollsBackBusinessWrite(t *testing.T) {
	db := outboxTestDB(t)
	queue := NewService(NewRepository(db.Pool))
	if _, err := db.Pool.Exec(t.Context(), `create table business_mutation (id integer primary key)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool.Exec(t.Context(), `alter table async_outbox add constraint reject_official_for_test check (task_type <> 'leaderboard.invalidate_official.v1')`); err != nil {
		t.Fatal(err)
	}
	err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		executor, err := postgres.Executor(ctx, db.Pool)
		if err != nil {
			return err
		}
		if _, err := executor.Exec(ctx, `insert into business_mutation (id) values (1)`); err != nil {
			return err
		}
		return queue.Enqueue(ctx, jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()}, jobs.InvalidateOfficialLeaderboardV1{Year: 2026})
	})
	if err == nil {
		t.Fatal("expected insertion failure")
	}
	var business, queued int
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from business_mutation`).Scan(&business); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from async_outbox`).Scan(&queued); err != nil {
		t.Fatal(err)
	}
	if business != 0 || queued != 0 {
		t.Errorf("failed enqueue left business=%d queued=%d", business, queued)
	}
}

func TestUnsupportedStatsIncludesRunningAndFailedVersions(t *testing.T) {
	db := outboxTestDB(t)
	repo := NewRepository(db.Pool)
	_, err := db.Pool.Exec(t.Context(), `insert into async_outbox (task_type, payload, state, claim_token, lease_expires_at, failed_at, completed_at)
 values ('future.job.v2','{}','pending',null,null,null,null),
 ('future.job.v2','{}','running',gen_random_uuid(),clock_timestamp()-interval '1 second',null,null),
 ('future.job.v2','{}','failed',null,null,clock_timestamp(),null),
 ('future.job.v2','{}','completed',null,null,null,clock_timestamp()),
 ('leaderboard.invalidate_contest.v1','{}','pending',null,null,null,null)`)
	if err != nil {
		t.Fatal(err)
	}
	stats, err := repo.UnsupportedStats(t.Context(), []jobs.Type{jobs.InvalidateContest})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Pending != 1 || stats.Running != 1 || stats.Failed != 1 || stats.OldestDueAt == nil {
		t.Errorf("unsupported stats = %+v; want pending=1 running=1 failed=1 and due time", stats)
	}
}

func TestReplayRequiresRegisteredVersion(t *testing.T) {
	db := outboxTestDB(t)
	repo := NewRepository(db.Pool)
	var id int64
	if err := db.Pool.QueryRow(t.Context(), `insert into async_outbox (task_type,payload,state,failed_at) values ('future.job.v2','{}','failed',clock_timestamp()) returning id`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Replay(t.Context(), id, "operator", "repair", []jobs.Type{jobs.InvalidateContest}); !errors.Is(err, ErrNotFailed) {
		t.Fatalf("unregistered replay = %v", err)
	}
	replayID, err := repo.Replay(t.Context(), id, "operator", "repair", []jobs.Type{"future.job.v2"})
	if err != nil || replayID <= id {
		t.Fatalf("registered replay = %d, %v", replayID, err)
	}
	claims, err := repo.Claim(t.Context(), "future.job.v2", 1, time.Minute, 2)
	if err != nil || len(claims) != 1 || claims[0].ID != replayID {
		t.Errorf("generic replay claim = %+v, %v", claims, err)
	}
}
