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
