package jobqueue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

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
	_, err = repo.Replay(tenantCtx, id, "operator", "repair", []jobs.Type{jobs.LeaderboardInvalidateOfficialV1})
	if !errors.Is(err, ErrNotFailed) || errx.KindOf(err) != errx.Conflict {
		t.Errorf("replay pending = %v, kind=%v", err, errx.KindOf(err))
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

	if _, err := db.Pool.Exec(tenantCtx, `
		insert into jobs (tenant, task_type, payload, state, completed_at, created_at)
		values ('tadoku/prod', 'leaderboard.invalidate_official.v1', '{}', 'completed', $1, $1)`, jobTestTime); err != nil {
		t.Fatal(err)
	}
	var deleted int64
	at(t, jobTestTime.AddDate(0, 3, 0).Add(time.Minute), func() {
		deleted, err = NewService(repo).CleanupCompleted(tenantCtx, 10)
	})
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

func TestUnsupportedStatsIncludesRunningAndFailedVersions(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db := jobTestDB(t)
	repo := NewRepository(db.Pool)

	_, err := db.Pool.Exec(tenantCtx, `
		insert into jobs (
			tenant, task_type, payload, state, claim_token, lease_expires_at, failed_at, completed_at
		)
		values
			('tadoku/prod', 'future.job.v2', '{}', 'pending', null, null, null, null),
			('tadoku/prod', 'future.job.v2', '{}', 'running',
			 gen_random_uuid(), clock_timestamp()-interval '1 second', null, null),
			('tadoku/prod', 'future.job.v2', '{}', 'failed', null, null, clock_timestamp(), null),
			('tadoku/prod', 'future.job.v2', '{}', 'completed', null, null, null, clock_timestamp()),
			('tadoku/prod', 'leaderboard.invalidate_contest.v1', '{}', 'pending', null, null, null, null)`)
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
	if err := db.Pool.QueryRow(tenantCtx, `
		insert into jobs (tenant, task_type, payload, state, failed_at)
		values ('tadoku/prod', 'future.job.v2', '{}', 'failed', clock_timestamp())
		returning id`).Scan(&id); err != nil {
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

			_, err := db.Pool.Exec(tenantCtx, `
				insert into jobs (tenant, task_type, payload, state, completed_at)
				values
					('tadoku/prod', 'retention.test.v1', '{}', 'completed', $1),
					('tadoku/prod', 'retention.test.v1', '{}', 'completed', $2),
					('tadoku/prod', 'retention.test.v1', '{}', 'completed', $3)`,
				tc.cutoff.Add(-time.Microsecond), tc.cutoff, tc.cutoff.Add(time.Microsecond),
			)
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

				var deleted int64
				at(t, tc.now, func() { deleted, err = NewService(repo).CleanupCompleted(ctx, 10) })
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

	_, err := db.Pool.Exec(tenantCtx, `
		insert into jobs (
			tenant, id, task_type, payload, state, created_at, next_attempt_at, failed_at, claim_token,
			lease_expires_at, completed_at, replay_of_id, replay_actor, replay_reason
		)
		values
			('tadoku/prod', 1, 'retention.test.v1', '{}', 'failed', $1, $1, $1, null,
			 null, null, null, null, null),
			('tadoku/prod', 2, 'retention.test.v1', '{}', 'completed', $1, $1, null, null,
			 null, $1, 1, 'operator', 'repaired'),
			('tadoku/prod', 3, 'retention.test.v1', '{}', 'pending', $1, $1, null, null,
			 null, null, null, null, null),
			('tadoku/prod', 4, 'retention.test.v1', '{}', 'running', $1, $1, null, gen_random_uuid(),
			 $1, null, null, null, null),
			('tadoku/prod', 5, 'retention.test.v1', '{}', 'completed', $1, $1, null, null,
			 null, $1, null, null, null),
			('tadoku/prod', 6, 'retention.test.v1', '{}', 'completed', $1, $1, null, null,
			 null, $1, null, null, null),
			('tadoku/prod', 7, 'retention.test.v1', '{}', 'pending', $1, $1, null, null,
			 null, null, 6, 'operator', 'linked')`, old)
	if err != nil {
		t.Fatal(err)
	}

	var deleted int64
	at(t, now, func() { deleted, err = NewService(repo).CleanupCompleted(tenantCtx, 1) })
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

	at(t, now, func() { deleted, err = NewService(repo).CleanupCompleted(tenantCtx, 10) })
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
