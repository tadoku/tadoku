package e2e_test

import (
	"net/http"
	"testing"
	"time"
)

func TestLogMutationJourney(t *testing.T) {
	end := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	runScenarioJourney(t, scoringEnabledHandler, "LogMutation", nil, []step{
		{request: "create", as: user, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "read_created", as: user, want: http.StatusOK},
		{request: "update_at_contest_end", as: user, want: http.StatusOK, others: cast{user2: http.StatusForbidden, banned: http.StatusForbidden}, at: end},
		{request: "read_exact_end_update", as: user, want: http.StatusOK, at: end},
		{request: "update_after_contest_end", as: admin, want: http.StatusOK, at: end.AddDate(0, 0, 1)},
		{request: "read_after_end", as: user, want: http.StatusOK, at: end.AddDate(0, 0, 1)},
		{verify: "provenance_and_jobs"},
	})
}

func TestLeaderboardShowsACreatedLogOnTheNextReadJourney(t *testing.T) {
	var tenantKey string
	runScenarioJourneyWithSetup(t, scoringEnabledHandler, "LeaderboardNextRead", func(t *testing.T, s *fixtureScenario) {
		tenantKey = s.tenant(t, primaryRole).String()
		seedLeaderboardCache(t, tenantKey, "leaderboard:global", "hit")
	}, nil, []step{
		{request: "create_log", as: user, want: http.StatusOK},
		{request: "next_read_shows_the_log", as: guest, want: http.StatusOK},
	})
	if t.Failed() {
		return
	}
	if score := cachedLeaderboardScore(t, tenantKey, "leaderboard:global"); score != 10 {
		t.Errorf("rebuilt leaderboard cache score = %v, want 10", score)
	}
}

func TestLogCreateAtomicFailureJourney(t *testing.T) {
	runJourneyWithHandler(t, api, scoringEnabledHandler, "LogCreateAtomicFailure", []step{
		{request: "create_duplicate_registration", as: user, want: http.StatusInternalServerError, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "failed_log_missing", as: user, want: http.StatusNotFound},
		{verify: "rollback_keeps_user_sync"},
	})
}

func TestLogCreateValidationSyncJourney(t *testing.T) {
	runScenarioJourney(t, scoringEnabledHandler, "LogCreateValidationSync", nil, []step{
		{request: "reject_invalid", as: user, want: http.StatusBadRequest, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{verify: "user_was_synchronized"},
		{request: "unknown_language_commits_before_readback", as: user, want: http.StatusNotFound},
		{verify: "unknown_language_log_was_committed"},
	})
}

func TestLogAttachmentAndDeleteJourney(t *testing.T) {
	runScenarioJourney(t, scoringEnabledHandler, "LogAttachmentAndDelete", nil, []step{
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
	runScenarioJourney(t, scoringEnabledHandler, "FrozenLogRefusals", nil, []step{
		{request: "update_refused", as: user, want: http.StatusConflict},
		{request: "registrations_refused", as: user, want: http.StatusConflict},
		{request: "delete_refused", as: user, want: http.StatusConflict},
		{request: "read_unchanged", as: user, want: http.StatusOK},
		{verify: "no_effects"},
	})
}

func TestLogOfficialEligibilityJourney(t *testing.T) {
	runScenarioJourney(t, api.handler, "LogOfficialEligibility", nil, []step{
		{request: "create_unofficial", as: user, want: http.StatusOK},
		{request: "create_official", as: user, want: http.StatusOK},
		{verify: "created"},
		{request: "attach_official", as: user, want: http.StatusOK},
		{verify: "attached"},
		{request: "detach_official", as: user, want: http.StatusOK},
		{verify: "detached"},
		{request: "moderate_detach_official", as: admin, want: http.StatusOK},
		{verify: "moderated"},
		{request: "reattach_official", as: user, want: http.StatusOK},
		{verify: "reattached"},
		{request: "remove_official_language", as: user, want: http.StatusOK},
		{verify: "language_removed"},
	})
}

func TestLogAttachmentAtomicFailureJourney(t *testing.T) {
	runScenarioJourney(t, scoringEnabledHandler, "LogAttachmentAtomicFailure", nil, []step{
		{request: "replace_with_duplicate", as: user, want: http.StatusInternalServerError},
		{request: "original_attachment_remains", as: user, want: http.StatusOK},
		{verify: "no_partial_effects"},
	})
}
