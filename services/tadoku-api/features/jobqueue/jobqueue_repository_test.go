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

var jobTestTime = time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

func jobTestDB(t *testing.T) *testpostgres.Database {
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

func TestInsertSharesBusinessTransaction(t *testing.T) {
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()}
	var err error
	if _, err := db.Pool.Exec(t.Context(), `create table business_mutation (id integer primary key)`); err != nil {
		t.Fatal(err)
	}
	rollback := errors.New("roll back business write")

	at(t, jobTestTime, func() {
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
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from jobs`).Scan(&queued); err != nil {
		t.Fatal(err)
	}
	if business != 0 || queued != 0 {
		t.Errorf("rollback left business=%d queued=%d", business, queued)
	}
}

func TestClaimLimitsKnownTypesAndFencesExpiredLeases(t *testing.T) {
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	contest := jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()}
	official := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	at(t, jobTestTime, func() {
		for _, task := range []jobs.Job{contest, contest, official} {
			if _, err := insertJob(repo, t.Context(), task); err != nil {
				t.Fatal(err)
			}
		}
	})
	if _, err := db.Pool.Exec(t.Context(), `insert into jobs (task_type,payload,created_at,next_attempt_at)
		values ('future.task.v1','{}',$1,$1)`, jobTestTime); err != nil {
		t.Fatal(err)
	}

	var first, second, third ClaimedJob
	at(t, jobTestTime, func() {
		claims, err := repo.Claim(t.Context(), jobs.LeaderboardInvalidateContestV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("first claim = %v, %v", claims, err)
		}
		first = claims[0]
		if first.Attempts != 1 || first.Type != jobs.LeaderboardInvalidateContestV1 || first.Reclaimed || first.LeaseExpiresAt.IsZero() {
			t.Errorf("first claim = %+v", first)
		}
		claims, err = repo.Claim(t.Context(), jobs.LeaderboardInvalidateContestV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("second claim = %v, %v", claims, err)
		}
		second = claims[0]
		if second.ID == first.ID {
			t.Error("claim reused active row")
		}
		claims, err = repo.Claim(t.Context(), jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("other type claim = %v, %v", claims, err)
		}
		third = claims[0]
		if expires, ok, err := repo.Renew(t.Context(), third, 2*time.Minute); err != nil || !ok || !expires.After(third.LeaseExpiresAt) {
			t.Errorf("renew = %v, %v, %v", expires, ok, err)
		}
	})
	if _, err := db.Pool.Exec(t.Context(), `update jobs
		set lease_expires_at=clock_timestamp()-interval '1 second' where id=$1`, first.ID); err != nil {
		t.Fatal(err)
	}

	at(t, jobTestTime.Add(time.Minute), func() {
		if ok, err := repo.Complete(t.Context(), first); err != nil || ok {
			t.Errorf("expired complete = %v, %v", ok, err)
		}
		if expires, ok, err := repo.Renew(t.Context(), first, time.Minute); err != nil || ok || !expires.IsZero() {
			t.Errorf("expired renew = %v, %v, %v", expires, ok, err)
		}
		if ok, err := repo.Retry(t.Context(), first, jobTestTime.Add(2*time.Minute), "temporary", 2); err != nil || ok {
			t.Errorf("expired retry = %v, %v", ok, err)
		}
		claims, err := repo.Claim(t.Context(), jobs.LeaderboardInvalidateContestV1, 1, time.Minute, 2)
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
	if _, err := db.Pool.Exec(t.Context(), `update jobs
		set lease_expires_at=clock_timestamp()-interval '1 second' where state='running'`); err != nil {
		t.Fatal(err)
	}

	at(t, jobTestTime.Add(2*time.Minute), func() {
		if claims, err := repo.Claim(t.Context(), jobs.LeaderboardInvalidateContestV1, 2, time.Minute, 2); err != nil || len(claims) != 1 || claims[0].ID != second.ID || claims[0].Attempts != 2 {
			t.Errorf("exhausted crash claims = %v, %v; want only second task", claims, err)
		}
		if ok, err := repo.Complete(t.Context(), third); err != nil || ok {
			t.Errorf("other expired claim completed = %v, %v", ok, err)
		}
	})
	var failed, unknown string
	if err := db.Pool.QueryRow(t.Context(), `select state from jobs where id=$1`, first.ID).Scan(&failed); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(t.Context(), `select state from jobs where task_type='future.task.v1'`).Scan(&unknown); err != nil {
		t.Fatal(err)
	}
	if failed != "failed" || unknown != "pending" {
		t.Errorf("states failed=%q unknown=%q", failed, unknown)
	}
	stats, err := repo.UnsupportedStats(t.Context(), []jobs.Type{jobs.LeaderboardInvalidateContestV1, jobs.LeaderboardInvalidateOfficialV1})
	if err != nil || stats.Pending != 1 || stats.OldestDueAt == nil || !stats.OldestDueAt.Equal(jobTestTime) {
		t.Errorf("unsupported stats = %+v, %v", stats, err)
	}
}

func TestReducedAttemptLimitFailsDuePendingTask(t *testing.T) {
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	at(t, jobTestTime, func() {
		if _, err := insertJob(repo, t.Context(), task); err != nil {
			t.Fatal(err)
		}
		claims, err := repo.Claim(t.Context(), jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("claim = %v, %v", claims, err)
		}
		if ok, err := repo.Retry(t.Context(), claims[0], jobTestTime, "temporary", 2); err != nil || !ok {
			t.Fatalf("retry = %v, %v", ok, err)
		}
		claims, err = repo.Claim(t.Context(), jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 1)
		if err != nil || len(claims) != 0 {
			t.Errorf("reduced limit claim = %v, %v", claims, err)
		}
	})
	var state, code string
	if err := db.Pool.QueryRow(t.Context(), `select state,last_error from jobs`).Scan(&state, &code); err != nil {
		t.Fatal(err)
	}
	if state != "failed" || code != "attempts_exhausted" {
		t.Errorf("reduced limit left state=%q code=%q", state, code)
	}
}

func TestTransitionsRejectDatabaseExpiredLease(t *testing.T) {
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var err error
	var claims []ClaimedJob
	at(t, jobTestTime, func() {
		for range 4 {
			if _, err := insertJob(repo, t.Context(), task); err != nil {
				t.Fatal(err)
			}
		}
		claims, err = repo.Claim(t.Context(), jobs.LeaderboardInvalidateOfficialV1, 4, time.Minute, 2)
		if err != nil || len(claims) != 4 {
			t.Fatalf("claim = %v, %v", claims, err)
		}
	})
	if _, err := db.Pool.Exec(t.Context(), `update jobs
		set lease_expires_at=clock_timestamp()-interval '1 second' where state='running'`); err != nil {
		t.Fatal(err)
	}
	at(t, jobTestTime, func() {
		if ok, err := repo.Complete(t.Context(), claims[0]); err != nil || ok {
			t.Errorf("complete expired = %v, %v", ok, err)
		}
		if _, ok, err := repo.Renew(t.Context(), claims[1], time.Minute); err != nil || ok {
			t.Errorf("renew expired = %v, %v", ok, err)
		}
		if ok, err := repo.Retry(t.Context(), claims[2], jobTestTime.Add(time.Minute), "temporary", 2); err != nil || ok {
			t.Errorf("retry expired = %v, %v", ok, err)
		}
		if ok, err := repo.Fail(t.Context(), claims[3], "invalid_payload"); err != nil || ok {
			t.Errorf("fail expired = %v, %v", ok, err)
		}
	})
}

func TestCompleteSkipsLockedClaim(t *testing.T) {
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var err error
	var claim ClaimedJob
	at(t, jobTestTime, func() {
		if _, err := insertJob(repo, t.Context(), task); err != nil {
			t.Fatal(err)
		}
		claims, err := repo.Claim(t.Context(), jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
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
	if _, err := tx.Exec(t.Context(), `select id from jobs where id=$1 for update`, claim.ID); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if ok, err := repo.Complete(ctx, claim); err != nil || ok {
		t.Errorf("complete locked claim = %v, %v", ok, err)
	}
}

func TestClaimSkipsLockedRows(t *testing.T) {
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()}
	var err error
	var lockedID int64
	at(t, jobTestTime, func() {
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
	if _, err := tx.Exec(t.Context(), `select id from jobs where id=$1 for update`, lockedID); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	at(t, jobTestTime, func() {
		claims, err := repo.Claim(ctx, jobs.LeaderboardInvalidateContestV1, 2, time.Minute, 2)
		if err != nil || len(claims) != 1 || claims[0].ID == lockedID {
			t.Errorf("skip locked claim = %v, %v", claims, err)
		}
	})
}

func TestRetryReplayAndRetention(t *testing.T) {
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	task := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var err error
	var id int64
	at(t, jobTestTime, func() { id, err = insertJob(repo, t.Context(), task) })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Replay(t.Context(), id, "operator", "repair", []jobs.Type{jobs.LeaderboardInvalidateOfficialV1}); !errors.Is(err, ErrNotFailed) {
		t.Errorf("replay pending = %v", err)
	}
	var claim ClaimedJob
	at(t, jobTestTime, func() {
		claims, err := repo.Claim(t.Context(), jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("claim = %v, %v", claims, err)
		}
		claim = claims[0]
		if ok, err := repo.Retry(t.Context(), claim, jobTestTime.Add(time.Minute), "temporary", 2); err != nil || !ok {
			t.Fatalf("retry = %v, %v", ok, err)
		}
		if count, err := repo.Outstanding(t.Context(), jobs.LeaderboardInvalidateOfficialV1); err != nil || count != 1 {
			t.Errorf("outstanding delayed = %d, %v", count, err)
		}
		if claims, err := repo.Claim(t.Context(), jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2); err != nil || len(claims) != 0 {
			t.Errorf("early claim = %v, %v", claims, err)
		}
	})
	at(t, jobTestTime.Add(time.Minute), func() {
		claims, err := repo.Claim(t.Context(), jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 || claims[0].Attempts != 2 {
			t.Fatalf("retry claim = %v, %v", claims, err)
		}
		claim = claims[0]
		if ok, err := repo.Retry(t.Context(), claim, jobTestTime.Add(2*time.Minute), "temporary", 2); err != nil || !ok {
			t.Fatalf("exhausted retry = %v, %v", ok, err)
		}
	})
	var replayID int64
	at(t, jobTestTime.Add(2*time.Minute), func() {
		replayID, err = repo.Replay(t.Context(), id, "operator", "repair", []jobs.Type{jobs.LeaderboardInvalidateOfficialV1})
	})
	if err != nil || replayID == 0 {
		t.Fatalf("replay = %d, %v", replayID, err)
	}
	var replayOf int64
	if err := db.Pool.QueryRow(t.Context(), `select replay_of_id from jobs where id=$1`, replayID).Scan(&replayOf); err != nil {
		t.Fatal(err)
	}
	if replayOf != id {
		t.Errorf("replay source = %d, want %d", replayOf, id)
	}
	at(t, jobTestTime.Add(2*time.Minute), func() {
		claims, err := repo.Claim(t.Context(), jobs.LeaderboardInvalidateOfficialV1, 1, time.Minute, 2)
		if err != nil || len(claims) != 1 {
			t.Fatalf("replay claim = %v, %v", claims, err)
		}
		if ok, err := repo.Complete(t.Context(), claims[0]); err != nil || !ok {
			t.Fatalf("complete replay = %v, %v", ok, err)
		}
	})
	if _, err := db.Pool.Exec(t.Context(), `insert into jobs (task_type,payload,state,completed_at,created_at) values ('leaderboard.invalidate_official.v1','{}','completed',$1,$1)`, jobTestTime); err != nil {
		t.Fatal(err)
	}
	deleted, err := repo.CleanupCompleted(t.Context(), jobTestTime.AddDate(0, 3, 0).Add(time.Minute), 10)
	if err != nil || deleted != 1 {
		t.Errorf("cleanup deleted = %d, %v; want only unlinked completion", deleted, err)
	}
	var retained int
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from jobs where id=$1`, replayID).Scan(&retained); err != nil {
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
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	job := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	id, err := insertJob(repo, t.Context(), job)
	if err != nil {
		t.Fatal(err)
	}
	var typ string
	var payload []byte
	if err := db.Pool.QueryRow(t.Context(), `select task_type, payload from jobs where id=$1`, id).Scan(&typ, &payload); err != nil {
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
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	job := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var ended context.Context
	if err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		ended = ctx
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := insertJob(repo, ended, job); err == nil {
		t.Fatal("accepted ended transaction")
	}
	other := jobTestDB(t)
	err := postgres.RunInTransaction(t.Context(), other.Pool, func(ctx context.Context) error {
		_, err := insertJob(repo, ctx, job)
		return err
	})
	if !errors.Is(err, postgres.ErrWrongDatabase) {
		t.Errorf("wrong database = %v", err)
	}
}

func TestUnsupportedStatsIncludesRunningAndFailedVersions(t *testing.T) {
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	_, err := db.Pool.Exec(t.Context(), `insert into jobs (task_type, payload, state, claim_token, lease_expires_at, failed_at, completed_at)
 values ('future.job.v2','{}','pending',null,null,null,null),
 ('future.job.v2','{}','running',gen_random_uuid(),clock_timestamp()-interval '1 second',null,null),
 ('future.job.v2','{}','failed',null,null,clock_timestamp(),null),
 ('future.job.v2','{}','completed',null,null,null,clock_timestamp()),
 ('leaderboard.invalidate_contest.v1','{}','pending',null,null,null,null)`)
	if err != nil {
		t.Fatal(err)
	}
	stats, err := repo.UnsupportedStats(t.Context(), []jobs.Type{jobs.LeaderboardInvalidateContestV1})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Pending != 1 || stats.Running != 1 || stats.Failed != 1 || stats.OldestDueAt == nil {
		t.Errorf("unsupported stats = %+v; want pending=1 running=1 failed=1 and due time", stats)
	}
}

func TestReplayRequiresRegisteredVersion(t *testing.T) {
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	var id int64
	if err := db.Pool.QueryRow(t.Context(), `insert into jobs (task_type,payload,state,failed_at) values ('future.job.v2','{}','failed',clock_timestamp()) returning id`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Replay(t.Context(), id, "operator", "repair", []jobs.Type{jobs.LeaderboardInvalidateContestV1}); !errors.Is(err, ErrNotFailed) {
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

func TestCleanupCompletedRetainsThreeUTCCalendarMonths(t *testing.T) {
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
			_, err := db.Pool.Exec(t.Context(), `insert into jobs (task_type,payload,state,completed_at)
    values ('retention.test.v1','{}','completed',$1), ('retention.test.v1','{}','completed',$2), ('retention.test.v1','{}','completed',$3)`, tc.cutoff.Add(-time.Microsecond), tc.cutoff, tc.cutoff.Add(time.Microsecond))
			if err != nil {
				t.Fatal(err)
			}
			err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
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
			if err := db.Pool.QueryRow(t.Context(), `select count(*) from jobs where completed_at >= $1`, tc.cutoff).Scan(&retained); err != nil {
				t.Fatal(err)
			}
			if retained != 2 {
				t.Errorf("retained %d boundary/newer completions, want 2", retained)
			}
		})
	}
}

func TestCleanupCompletedExpiresSuccessfulReplayAndPreservesOtherStates(t *testing.T) {
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	old := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := db.Pool.Exec(t.Context(), `insert into jobs (id,task_type,payload,state,created_at,next_attempt_at,failed_at,claim_token,lease_expires_at,completed_at,replay_of_id,replay_actor,replay_reason) values
  (1,'retention.test.v1','{}','failed',$1,$1,$1,null,null,null,null,null,null),
  (2,'retention.test.v1','{}','completed',$1,$1,null,null,null,$1,1,'operator','repaired'),
  (3,'retention.test.v1','{}','pending',$1,$1,null,null,null,null,null,null,null),
  (4,'retention.test.v1','{}','running',$1,$1,null,gen_random_uuid(),$1,null,null,null,null),
  (5,'retention.test.v1','{}','completed',$1,$1,null,null,null,$1,null,null,null),
  (6,'retention.test.v1','{}','completed',$1,$1,null,null,null,$1,null,null,null),
  (7,'retention.test.v1','{}','pending',$1,$1,null,null,null,null,6,'operator','linked')`, old)
	if err != nil {
		t.Fatal(err)
	}
	deleted, err := repo.CleanupCompleted(t.Context(), now, 1)
	if err != nil || deleted != 1 {
		t.Fatalf("bounded cleanup = %d, %v", deleted, err)
	}
	var replayRetained int
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from jobs where id=2`).Scan(&replayRetained); err != nil {
		t.Fatal(err)
	}
	if replayRetained != 0 {
		t.Errorf("expired successful replay was retained")
	}
	deleted, err = repo.CleanupCompleted(t.Context(), now, 10)
	if err != nil || deleted != 1 {
		t.Fatalf("remaining cleanup = %d, %v", deleted, err)
	}
	var retained int
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from jobs where id in (1,3,4,6,7)`).Scan(&retained); err != nil {
		t.Fatal(err)
	}
	if retained != 5 {
		t.Errorf("retained %d failed/active/referenced jobs, want 5", retained)
	}
}

func TestInsertFailureRollsBackBusinessWrite(t *testing.T) {
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)
	if _, err := db.Pool.Exec(t.Context(), `create table business_mutation (id integer primary key)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool.Exec(t.Context(), `alter table jobs add constraint reject_official_for_test check (task_type <> 'leaderboard.invalidate_official.v1')`); err != nil {
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
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from business_mutation`).Scan(&business); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from jobs`).Scan(&queued); err != nil {
		t.Fatal(err)
	}
	if business != 0 || queued != 0 {
		t.Errorf("failed enqueue left business=%d queued=%d", business, queued)
	}
}
