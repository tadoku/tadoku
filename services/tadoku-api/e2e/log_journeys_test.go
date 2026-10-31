package e2e_test

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestLogMutationJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	end := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	runJourneyWithHandler(t, api, scoringEnabledHandler, "LogMutation", []step{
		{request: "create", as: user, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "read_created", as: user, want: http.StatusOK},
		{request: "update_at_contest_end", as: user, want: http.StatusOK, others: cast{user2: http.StatusForbidden, banned: http.StatusForbidden}, at: end},
		{request: "read_exact_end_update", as: user, want: http.StatusOK, at: end},
		{request: "update_after_contest_end", as: admin, want: http.StatusOK, at: end.AddDate(0, 0, 1)},
		{request: "read_after_end", as: user, want: http.StatusOK, at: end.AddDate(0, 0, 1)},
		{verify: "provenance_and_jobs"},
	})
}

func TestStandaloneWorkerLeaderboardJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	runJourneyWithSetup(t, api, scoringEnabledHandler, "LeaderboardJobs", func(t *testing.T) {
		seedLeaderboardCache(t, "leaderboard:global", "hit")
	}, []step{
		{request: "create_log", as: user, want: http.StatusOK},
		{request: "before_worker_stale_cache", as: guest, want: http.StatusOK},
		{job: "run_worker"},
		{request: "after_worker_cache_miss", as: guest, want: http.StatusOK},
	})
	if t.Failed() {
		return
	}
	if score := cachedLeaderboardScore(t, "leaderboard:global"); score != 10 {
		t.Errorf("rebuilt leaderboard cache score = %v, want 10", score)
	}
}

func TestLogCreateAtomicFailureJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	runJourneyWithSetup(t, api, scoringEnabledHandler, "LogCreateAtomicFailure", func(t *testing.T) {
		t.Cleanup(func() {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := api.db.Pool.Exec(cleanupCtx, `select setval(pg_get_serial_sequence('jobs', 'id'), greatest(coalesce((select max(id) from jobs), 0), 1), exists(select 1 from jobs))`); err != nil {
				t.Error(err)
			}
		})
		if _, err := api.db.Pool.Exec(t.Context(), `select setval(pg_get_serial_sequence('jobs', 'id'), 9223372036854775807, true)`); err != nil {
			t.Fatal(err)
		}
		var id int64
		err := api.db.Pool.QueryRow(t.Context(), `select nextval(pg_get_serial_sequence('jobs', 'id'))`).Scan(&id)
		var postgresError *pgconn.PgError
		if !errors.As(err, &postgresError) || postgresError.Code != "2200H" {
			t.Fatalf("job sequence did not reject next ID: %v", err)
		}
	}, []step{
		{request: "create_duplicate_registration", as: user, want: http.StatusInternalServerError, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "failed_log_missing", as: user, want: http.StatusNotFound},
		{request: "async_enqueue_failure", as: user, want: http.StatusInternalServerError},
		{verify: "rollback_keeps_user_sync"},
	})
}

func TestLogCreateValidationSyncJourney(t *testing.T) {
	runJourneyWithHandler(t, api, scoringEnabledHandler, "LogCreateValidationSync", []step{
		{request: "reject_invalid", as: user, want: http.StatusBadRequest, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{verify: "user_was_synchronized"},
		{request: "unknown_language_commits_before_readback", as: user, want: http.StatusNotFound},
		{verify: "unknown_language_log_was_committed"},
	})
}

func TestLogAttachmentAndDeleteJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	runJourneyWithHandler(t, api, scoringEnabledHandler, "LogAttachmentAndDelete", []step{
		{request: "create", as: user, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "attach", as: user, want: http.StatusOK, others: cast{user2: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "read_attached", as: user, want: http.StatusOK},
		{request: "detach", as: user, want: http.StatusOK, others: cast{user2: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "read_detached", as: user, want: http.StatusOK},
		{request: "delete", as: user, want: http.StatusOK, others: cast{user2: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "deleted_missing", as: user, want: http.StatusNotFound},
		{verify: "effects"},
	})
}

func TestFrozenLogRefusalsJourney(t *testing.T) {
	runJourneyWithHandler(t, api, scoringEnabledHandler, "FrozenLogRefusals", []step{
		{request: "update_refused", as: user, want: http.StatusConflict},
		{request: "registrations_refused", as: user, want: http.StatusConflict},
		{request: "delete_refused", as: user, want: http.StatusConflict},
		{request: "read_unchanged", as: user, want: http.StatusOK},
		{verify: "no_effects"},
	})
}

func TestLogAttachmentAtomicFailureJourney(t *testing.T) {
	runJourneyWithHandler(t, api, scoringEnabledHandler, "LogAttachmentAtomicFailure", []step{
		{request: "replace_with_duplicate", as: user, want: http.StatusInternalServerError},
		{request: "original_attachment_remains", as: user, want: http.StatusOK},
		{verify: "no_partial_effects"},
	})
}
