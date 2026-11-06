package e2e_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

func contestEventUUIDs() io.Reader {
	randomBytes := make([]byte, 16*128)
	for index := range 128 {
		randomBytes[index*16+15] = byte(index + 1)
	}
	return bytes.NewReader(randomBytes)
}

func TestContestEventsJourney(t *testing.T) {
	steps := []step{
		{request: "empty_now", as: guest, want: http.StatusOK},
		{request: "create_a", as: user, want: http.StatusOK},
		{request: "update_a_before_submission", as: user, want: http.StatusOK},
		{request: "unsubmitted_changes_absent", as: guest, want: http.StatusOK},
		{request: "submit_a", as: user, want: http.StatusOK},
		{request: "preview_contest_a", as: guest, want: http.StatusOK},
		{request: "first_submission", as: guest, want: http.StatusOK},
		{request: "start_at_now", as: guest, want: http.StatusOK},
		{request: "update_submitted_a", as: user, want: http.StatusOK},
		{request: "updated_after_checkpoint", as: guest, want: http.StatusOK},
		{request: "create_and_submit_b", as: user, want: http.StatusOK},
		{request: "withdraw_b", as: user, want: http.StatusOK},
		{request: "update_b_after_withdrawal", as: user, want: http.StatusOK},
		{request: "exact_four_events", as: guest, want: http.StatusOK},
		{request: "first_page", as: guest, want: http.StatusOK},
		{request: "second_page", as: guest, want: http.StatusOK},
		{request: "replayed_second_page", as: guest, want: http.StatusOK},
		{request: "caught_up_cursor", as: guest, want: http.StatusOK},
		{request: "withdrawn_preview", as: guest, want: http.StatusNotFound},
		{request: "unchanged_update", as: user, want: http.StatusOK},
		{request: "unchanged_submission", as: user, want: http.StatusOK},
		{request: "no_op_keeps_cursor", as: guest, want: http.StatusOK},
		{request: "current_a_preview", as: guest, want: http.StatusOK},
	}
	for index := range steps {
		steps[index].at = fixtureInstant.Add(time.Duration(index) * time.Second)
	}

	runScenarioJourney(t, scoringEnabledHandler, "ContestEvents", contestEventUUIDs(), steps)
}

func TestContestEventWithdrawalsJourney(t *testing.T) {
	steps := []step{
		{request: "existing_logs_have_no_synthetic_history", as: guest, want: http.StatusOK},
		{request: "unchanged_log", as: user, want: http.StatusOK},
		{request: "unchanged_attachments", as: user, want: http.StatusOK},
		{request: "unchanged_feed", as: guest, want: http.StatusOK},
		{request: "moderator_withdraws", as: admin, want: http.StatusOK},
		{request: "moderated_preview_missing", as: guest, want: http.StatusNotFound},
		{request: "other_contest_preserved", as: guest, want: http.StatusOK},
		{request: "resubmit", as: user, want: http.StatusOK},
		{request: "resubmitted_preview", as: guest, want: http.StatusOK},
		{request: "remove_registered_language", as: user, want: http.StatusOK},
		{request: "removed_language_preview_missing", as: guest, want: http.StatusNotFound},
		{request: "other_language_preserved", as: guest, want: http.StatusOK},
		{request: "other_user_preserved", as: guest, want: http.StatusOK},
		{request: "repeat_moderation_is_no_op", as: admin, want: http.StatusOK},
		{request: "delete_attached_log", as: user, want: http.StatusOK},
		{request: "deleted_preview_missing", as: guest, want: http.StatusNotFound},
		{request: "retained_withdrawal_history", as: guest, want: http.StatusOK},
		{request: "delete_remaining_log", as: user, want: http.StatusOK},
		{request: "other_contest_withdrawal", as: guest, want: http.StatusOK},
		{request: "unattached_delete_does_not_repeat_withdrawal", as: guest, want: http.StatusOK},
	}
	for index := range steps {
		steps[index].at = fixtureInstant.Add(time.Duration(index) * time.Second)
	}

	runScenarioJourney(t, scoringEnabledHandler, "ContestEventWithdrawals", contestEventUUIDs(), steps)
}

func TestContestEventFrozenScoresJourney(t *testing.T) {
	steps := []step{
		{request: "personal_tracking_changes", as: admin, want: http.StatusOK},
		{request: "closed_contest_silent", as: guest, want: http.StatusOK},
		{request: "open_contest_updated", as: guest, want: http.StatusOK},
		{request: "closed_contest_tracking_frozen", as: guest, want: http.StatusOK},
		{request: "shared_content_changes", as: admin, want: http.StatusOK},
		{request: "closed_contest_content_event", as: guest, want: http.StatusOK},
		{request: "closed_preview_new_content_frozen_score", as: guest, want: http.StatusOK},
		{request: "open_preview_current_score", as: guest, want: http.StatusOK},
	}
	for index := range steps {
		steps[index].at = fixtureInstant.Add(time.Duration(index) * time.Second)
	}

	runScenarioJourney(t, scoringEnabledHandler, "ContestEventFrozenScores", contestEventUUIDs(), steps)
}

func TestContestEventRollbackJourney(t *testing.T) {
	runJourneyWithSetup(t, api, scoringEnabledHandler, "ContestEventRollback", func(t *testing.T) {
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := api.db.Pool.Exec(ctx, `drop trigger contest_event_rollback_fault on contest_events;
				drop function contest_event_rollback_fault()`)
			if err != nil {
				t.Error(err)
			}
		})
		// Test safety: a serial journey owns the database, so the fault cannot fail another test's commit.
		_, err := api.db.Pool.Exec(t.Context(), `create function contest_event_rollback_fault() returns trigger
				language plpgsql as $$ begin raise exception 'contest event rollback fault'; end; $$;
			create constraint trigger contest_event_rollback_fault after insert on contest_events
				deferrable initially deferred for each row execute function contest_event_rollback_fault()`)
		if err != nil {
			t.Fatal(err)
		}
	}, contestEventUUIDs(), []step{
		{request: "create_fails_after_event_append", as: user, want: http.StatusInternalServerError},
		{request: "rolled_back_feed_empty", as: guest, want: http.StatusOK},
		{request: "rolled_back_log_missing", as: guest, want: http.StatusNotFound},
	})
}
