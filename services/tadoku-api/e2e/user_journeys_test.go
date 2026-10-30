package e2e_test

import (
	"bytes"
	"context"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestTenantIsolationJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	const key = "e2e/alpha-0000000a"
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := api.db.Pool.Exec(ctx, "delete from tenants where key = $1 and kind = 'test'", key)
		if err != nil {
			t.Error(err)
		}
	})

	runJourneyWithSetup(t, api, scoringEnabledHandler, "TenantIsolation", func(t *testing.T) {
		parsed, err := tenant.Parse(key)
		if err != nil {
			t.Fatal(err)
		}
		if err := api.flipt.SeedTenant(parsed); err != nil {
			t.Fatal(err)
		}
	}, []step{
		{
			request: "test_guest_admitted",
			as:      alphaGuest,
			want:    http.StatusOK,
			others:  cast{alphaBanned: http.StatusForbidden},
		},
		{request: "test_member_admitted", as: alphaUser, want: http.StatusOK},
		{
			request: "test_tenant_admin",
			as:      alphaAdmin,
			want:    http.StatusOK,
			others:  cast{alphaBanned: http.StatusOK},
		},
		{request: "test_explicit_tester", as: alphaTester, want: http.StatusOK},
		{request: "create_test_log", as: alphaTester, want: http.StatusOK},
		{request: "create_production_log", as: user, want: http.StatusOK},
		{request: "production_leaderboard", as: user, want: http.StatusOK},
		{request: "test_leaderboard", as: alphaTester, want: http.StatusOK},
		{valkey: "separate_cache_keys"},
		{job: "run_worker"},
		{verify: "production_complete_test_paused"},
		{valkey: "base_keeps_test_cache"},
		{job: "run_branch_worker", forTenant: key},
		{verify: "both_tenants_complete"},
		{request: "ban_test_tester", as: alphaAdmin, want: http.StatusOK},
		{request: "test_ban_enforced", as: alphaTester, want: http.StatusForbidden},
		{request: "production_tester_unaffected", as: user2, want: http.StatusOK},
		{request: "production_user_logs", as: user, want: http.StatusOK},
		{request: "production_admin_contest_logs", as: admin, want: http.StatusOK},
		{verify: "separate_log_rows"},
	})
}

func TestLanguageLifecycleJourney(t *testing.T) {
	runJourney(t, api, "LanguageCreateList", []step{
		{
			request: "list_languages",
			as:      admin,
			want:    http.StatusOK,
		},
		{
			request: "create_language",
			as:      admin,
			want:    http.StatusOK,
			others: cast{
				none:   http.StatusBadRequest,
				guest:  http.StatusUnauthorized,
				user:   http.StatusForbidden,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "list_languages",
			as:      admin,
			want:    http.StatusOK,
		},
		{
			request: "rename_language",
			as:      admin,
			want:    http.StatusOK,
			others: cast{
				none:   http.StatusBadRequest,
				guest:  http.StatusUnauthorized,
				user:   http.StatusForbidden,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "list_renamed",
			as:      admin,
			want:    http.StatusOK,
		},
		{
			request: "repeat_update",
			as:      admin,
			want:    http.StatusOK,
		},
		{
			request: "duplicate_create",
			as:      admin,
			want:    http.StatusConflict,
		},
		{
			request: "rejected_update",
			as:      user,
			want:    http.StatusForbidden,
			others: cast{
				guest:  http.StatusUnauthorized,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "list_unchanged",
			as:      admin,
			want:    http.StatusOK,
		},
	})
}

func TestContestCreationJourney(t *testing.T) {
	ids := append(bytes.Repeat([]byte{0x11}, 16), bytes.Repeat([]byte{0x22}, 16)...)
	uuid.SetRand(bytes.NewReader(ids))
	defer uuid.SetRand(nil)

	runJourney(t, api, "ContestCreation", []step{
		{
			request: "reject_invalid_contest",
			as:      user,
			want:    http.StatusBadRequest,
			others: cast{
				none:   http.StatusBadRequest,
				guest:  http.StatusUnauthorized,
				banned: http.StatusForbidden,
			},
		},
		{verify: "rejection_keeps_user_upsert"},
		{
			request: "create_public_contest",
			as:      user,
			want:    http.StatusOK,
			others: cast{
				none:   http.StatusBadRequest,
				guest:  http.StatusUnauthorized,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "other_user_finds_contest",
			as:      user2,
			want:    http.StatusOK,
		},
		{
			request: "other_user_lists_contest",
			as:      user2,
			want:    http.StatusOK,
		},
		{
			request: "admin_creates_official_contest",
			as:      admin,
			want:    http.StatusOK,
			others: cast{
				none:   http.StatusBadRequest,
				guest:  http.StatusUnauthorized,
				user:   http.StatusForbidden,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "official_contest_is_latest",
			as:      user2,
			want:    http.StatusOK,
		},
	})
}

func TestScoringRuleSetDraftJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	runJourney(t, api, "ScoringRuleSetDraft", []step{
		{request: "create_platform_draft", as: admin, want: http.StatusOK, others: cast{user: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "platform_draft_visible_to_admin", as: admin, want: http.StatusOK},
		{request: "platform_draft_hidden_from_member", as: user, want: http.StatusOK},
		{request: "create_contest_draft", as: user, want: http.StatusOK, others: cast{user2: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "contest_draft_visible_to_owner", as: user, want: http.StatusOK, others: cast{user2: http.StatusForbidden}},
	})
}

func TestScoringRuleSetDraftAtomicFailureJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	runJourney(t, api, "ScoringRuleSetDraftAtomicFailure", []step{
		{request: "create_two_rule_draft_with_collision", as: admin, want: http.StatusInternalServerError},
		{request: "failed_draft_left_no_version", as: admin, want: http.StatusOK},
	})
}

func TestPlatformScoringRuleSetLifecycleJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	runJourney(t, api, "PlatformScoringRuleSetLifecycle", []step{
		{request: "create_draft", as: admin, want: http.StatusOK, others: cast{user: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "publish_draft", as: admin, want: http.StatusOK, others: cast{user: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "published_not_active", as: admin, want: http.StatusOK},
		{request: "activate_published", as: admin, want: http.StatusNoContent, others: cast{user: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "active_scores_preview", as: user, want: http.StatusOK},
	})
}

func TestContestScoringRuleSetLifecycleJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	start := fixtureInstant.Add(12 * time.Hour)
	runJourney(t, api, "ContestScoringRuleSetLifecycle", []step{
		{request: "create_override_draft", as: user, want: http.StatusOK, others: cast{user2: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "publish_draft", as: user, want: http.StatusOK, others: cast{user2: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "activate_published", as: user, want: http.StatusNoContent, others: cast{user2: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "active_rule_set_visible", as: user, want: http.StatusOK},
		{request: "change_rejected_at_start", as: user, want: http.StatusConflict, others: cast{admin: http.StatusConflict, user2: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}, at: start},
		{request: "activate_new_platform", as: admin, want: http.StatusNoContent, at: start},
		{request: "pinned_fallback_preview", as: user, want: http.StatusOK, at: start},
	})
}

func TestContestModerationDetachLogJourney(t *testing.T) {
	mutationTime := time.Date(2027, 1, 2, 12, 0, 0, 0, time.UTC)
	runJourney(t, api, "ContestModerationDetachLog", []step{
		{request: "detach", as: user2, want: http.StatusOK, others: cast{user: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}, at: mutationTime},
		{request: "owner_reads_detached", as: user, want: http.StatusOK, at: mutationTime},
		{request: "repeat_absent_detach", as: user2, want: http.StatusOK, at: mutationTime},
		{request: "owner_reads_still_detached", as: user, want: http.StatusOK, at: mutationTime},
		{verify: "audit_and_jobs"},
	})
}

func TestContestRegistrationJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	runJourney(t, api, "ContestRegistration", []step{
		{
			request: "create_ongoing_contest",
			as:      user,
			want:    http.StatusOK,
		},
		{
			request: "registration_initially_missing",
			as:      user,
			want:    http.StatusNoContent,
		},
		{
			request: "register",
			as:      user,
			want:    http.StatusOK,
			others:  cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden},
		},
		{
			request: "registration_visible",
			as:      user,
			want:    http.StatusOK,
		},
		{
			request: "second_user_registers",
			as:      user2,
			want:    http.StatusOK,
		},
		{
			request: "second_user_sees_own_registration",
			as:      user2,
			want:    http.StatusOK,
		},
		{
			request: "ongoing_registration",
			as:      user,
			want:    http.StatusOK,
		},
		{
			request: "update_languages",
			as:      user,
			want:    http.StatusOK,
		},
		{
			request: "updated_registration_visible",
			as:      user,
			want:    http.StatusOK,
		},
		{
			request: "registration_no_longer_ongoing",
			as:      user,
			want:    http.StatusOK,
			at:      fixtureInstant.Add(24 * time.Hour),
		},
		{
			request: "registrants_appear_on_leaderboard",
			as:      user2,
			want:    http.StatusOK,
			at:      fixtureInstant.Add(24 * time.Hour),
		},
	})
}

func TestYearlyContestRegistrationJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	runJourney(t, api, "YearlyContestRegistration", []step{
		{request: "create_private_contest", as: user, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "register_private_contest", as: user, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "owner_reads_private_history", as: user, want: http.StatusOK},
		{request: "other_user_sees_empty_history", as: user2, want: http.StatusOK},
		{request: "create_public_contest", as: user, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "register_public_contest", as: user, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}, at: fixtureInstant.Add(time.Second)},
		{request: "owner_reads_mixed_history", as: user, want: http.StatusOK},
		{request: "other_user_reads_public_history", as: user2, want: http.StatusOK},
		{request: "history_remains_after_contests_end", as: user, want: http.StatusOK, at: fixtureInstant.Add(24 * time.Hour)},
	})
}

func TestContestRegistrationDetachJourney(t *testing.T) {
	runJourney(t, api, "ContestRegistrationDetach", []step{
		{
			request: "existing_registration",
			as:      user,
			want:    http.StatusOK,
		},
		{
			request: "summary_before_detach",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "remove_language",
			as:      user,
			want:    http.StatusOK,
			others:  cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden},
		},
		{
			request: "registration_updated",
			as:      user,
			want:    http.StatusOK,
		},
		{
			request: "summary_after_detach",
			as:      guest,
			want:    http.StatusOK,
		},
		{verify: "logs_detached_and_refreshes_enqueued"},
	})
}

func TestContestRegistrationFrozenAfterEndJourney(t *testing.T) {
	dayAfterEnd := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	runJourney(t, api, "ContestRegistrationFrozenAfterEnd", []step{
		{request: "remove_language_after_end", as: user, want: http.StatusForbidden, at: dayAfterEnd},
		{request: "registration_unchanged", as: user, want: http.StatusOK, at: dayAfterEnd},
		{verify: "logs_and_registration_kept"},
	})
}

func TestYearlyScoresLanguageNameChangeJourney(t *testing.T) {
	runJourney(t, api, "YearlyScoresLanguageNameChange", []step{
		{request: "scores_before", as: user, want: http.StatusOK},
		{request: "update_language_name", as: admin, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden}},
		{request: "scores_after", as: user2, want: http.StatusOK},
	})
}

func TestContestParticipantDetachJourney(t *testing.T) {
	runJourney(t, api, "ContestParticipantDetach", []step{
		{request: "scores_before", as: user2, want: http.StatusOK},
		{request: "activity_before", as: guest, want: http.StatusOK},
		{request: "remove_language", as: user, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "scores_after", as: user2, want: http.StatusOK},
		{request: "activity_after", as: guest, want: http.StatusOK},
	})
}

func TestLogRegistrationDetachJourney(t *testing.T) {
	runJourney(t, api, "LogRegistrationDetach", []step{
		{request: "log_before", as: user, want: http.StatusOK},
		{request: "contest_logs_before", as: user2, want: http.StatusOK},
		{request: "remove_language", as: user, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "log_after", as: user, want: http.StatusOK},
		{request: "contest_logs_after", as: user2, want: http.StatusOK},
		{request: "profile_logs_unchanged", as: user, want: http.StatusOK},
	})
}
