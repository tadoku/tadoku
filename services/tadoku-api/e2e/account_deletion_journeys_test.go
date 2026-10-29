package e2e_test

import (
	"net/http"
	"testing"
)

func TestAccountDeletionSelfServiceJourney(t *testing.T) {
	runJourney(t, api, "AccountDeletion", []step{
		{request: "running_leaderboard_before", as: guest, want: http.StatusOK},
		{request: "delete_account", as: guest, want: http.StatusAccepted},
		{request: "session_after_request", as: guest, want: http.StatusOK},
		{request: "repeat_request", as: guest, want: http.StatusUnauthorized},
		{verify: "filed"},
		{job: "run_worker"},
		{verify: "deleted"},
		{verify: "jobs"},
		{verify: "contest_events"},
		{verify: "audit"},
		{verify: "emails"},
		{request: "running_leaderboard_after", as: guest, want: http.StatusOK},
		{request: "finished_leaderboard_after", as: guest, want: http.StatusOK},
	})
}

func TestAccountDeletionRefusalsJourney(t *testing.T) {
	runJourney(t, api, "AccountDeletionRefusals", []step{
		{request: "guest", as: guest, want: http.StatusUnauthorized},
		{request: "guest_malformed_address", as: guest, want: http.StatusUnauthorized},
		{request: "stale_session", as: guest, want: http.StatusUnauthorized},
		{request: "missing_email", as: guest, want: http.StatusBadRequest},
		{request: "malformed_email", as: guest, want: http.StatusBadRequest},
		{request: "display_address", as: guest, want: http.StatusBadRequest},
		{request: "another_address", as: guest, want: http.StatusBadRequest},
		{request: "foreign_origin", as: guest, want: http.StatusForbidden},
		{request: "non_json", as: guest, want: http.StatusBadRequest},
		{request: "administrator", as: guest, want: http.StatusConflict},
		{request: "banned_user", as: guest, want: http.StatusConflict},
		{verify: "unchanged"},
	})
}
