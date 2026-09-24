package e2e_test

import (
	"bytes"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
)

// Every important user journey in the application lives in this file. Replay a
// request as other cast members only where that adds information: rejected
// identities on mutating steps, and a second user only to observe limited
// visibility of a resource.

func TestFeatureAccessJourney(t *testing.T) {
	runJourney(t, api, "FeatureAccess", []step{
		{request: "target_initially_disabled", as: admin, want: http.StatusOK},
		{request: "grant_target", as: admin, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden}},
		{request: "target_enabled", as: admin, want: http.StatusOK},
		{request: "repeat_grant", as: admin, want: http.StatusOK},
		{request: "target_decision_enabled", as: user2, want: http.StatusOK},
		{request: "revoke_target", as: admin, want: http.StatusOK, others: cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden}},
		{request: "target_disabled_again", as: admin, want: http.StatusOK},
		{request: "target_decision_disabled", as: user2, want: http.StatusOK},
		{verify: "changes_audited"},
	})
}

func TestFeatureAccessAuditFailureJourney(t *testing.T) {
	failureAPI := &suite{
		db:      api.db,
		keto:    api.keto,
		kratos:  api.kratos,
		flipt:   api.flipt,
		handler: auditUnavailableRoleUpdateHandler(t),
	}
	runJourney(t, failureAPI, "FeatureAccessAuditFailure", []step{
		{request: "grant_audit_unavailable", as: admin, want: http.StatusInternalServerError},
		{request: "provider_change_persisted", as: admin, want: http.StatusOK},
		{verify: "audit_empty"},
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
		{verify: "provenance_and_outbox"},
	})
}

func TestLogWriteOutboxLeaderboardJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	ready := &leaderboardReadyWriter{ready: make(chan struct{})}
	logger := slog.New(slog.NewTextHandler(ready, nil))
	leaderboardService := leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), leaderboardValkey.client, time.Second, "")
	handler, profileService, roleService, err := newTestRouterWithLeaderboardService(t.Context(), api.db.Pool, api.db.Pool, api.keto, api.kratos, logger, true, leaderboardService)
	if err != nil {
		t.Fatal(err)
	}
	worker := leaderboard.NewWorker(leaderboardService, logger)
	journey := &suite{
		db:          api.db,
		keto:        api.keto,
		kratos:      api.kratos,
		flipt:       api.flipt,
		handler:     handler,
		profile:     profileService,
		leaderboard: leaderboardService,
		outbox:      worker,
		outboxReady: ready.ready,
		roles:       roleService,
	}
	runJourneyWithSetup(t, journey, handler, "LeaderboardOutbox", func(t *testing.T) {
		seedLeaderboardCache(t, "leaderboard:global", "hit")
	}, []step{
		{request: "create_log", as: user, want: http.StatusOK},
		{request: "before_worker", as: guest, want: http.StatusOK},
		{job: "run_leaderboard_outbox"},
		{request: "after_worker_cache_miss", as: guest, want: http.StatusOK},
		{request: "update_log", as: user, want: http.StatusOK},
		{job: "wait_for_leaderboard_outbox_poll"},
		{request: "after_poll_cache_miss", as: guest, want: http.StatusOK},
		{request: "after_poll_cache_hit", as: guest, want: http.StatusOK},
		{verify: "outbox_processed"},
	})
	marker, err := leaderboardValkey.client.Do(t.Context(), leaderboardValkey.client.B().Get().Key("leaderboard:global:last_updated").Build()).ToString()
	if err != nil || !strings.HasPrefix(marker, "native:") {
		t.Errorf("leaderboard cache marker after HTTP reads = %q, err = %v", marker, err)
	}
}

type leaderboardReadyWriter struct {
	ready chan struct{}
	once  sync.Once
}

func (writer *leaderboardReadyWriter) Write(message []byte) (int, error) {
	if bytes.Contains(message, []byte("leaderboard outbox ready")) {
		writer.once.Do(func() { close(writer.ready) })
	}
	return len(message), nil
}

func TestLogCreateAtomicFailureJourney(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	runJourneyWithHandler(t, api, scoringEnabledHandler, "LogCreateAtomicFailure", []step{
		{request: "create_duplicate_registration", as: user, want: http.StatusInternalServerError, others: cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden}},
		{request: "failed_log_missing", as: user, want: http.StatusNotFound},
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

func TestLogAttachmentAtomicFailureJourney(t *testing.T) {
	runJourneyWithHandler(t, api, scoringEnabledHandler, "LogAttachmentAtomicFailure", []step{
		{request: "replace_with_duplicate", as: user, want: http.StatusInternalServerError},
		{request: "original_attachment_remains", as: user, want: http.StatusOK},
		{verify: "no_partial_effects"},
	})
}

func TestContestModerationDetachLogJourney(t *testing.T) {
	mutationTime := time.Date(2027, 1, 2, 12, 0, 0, 0, time.UTC)
	runJourney(t, api, "ContestModerationDetachLog", []step{
		{request: "detach", as: user2, want: http.StatusOK, others: cast{user: http.StatusForbidden, guest: http.StatusUnauthorized, banned: http.StatusForbidden}, at: mutationTime},
		{request: "owner_reads_detached", as: user, want: http.StatusOK, at: mutationTime},
		{request: "repeat_absent_detach", as: user2, want: http.StatusOK, at: mutationTime},
		{request: "owner_reads_still_detached", as: user, want: http.StatusOK, at: mutationTime},
		{verify: "audit_and_outbox"},
	})
}

func TestContestRegistrationJourney(t *testing.T) {
	// Keep API-created contest and registration IDs stable in the HTTP fixtures.
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
	// Detachment requires an existing registration and linked logs, so this
	// journey first exposes the seeded registration through the API.
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

func TestAuthorizationVisibilityJourney(t *testing.T) {
	runJourney(t, api, "AuthorizationVisibility", []step{
		{
			request: "guest_role",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "user_role",
			as:      user,
			want:    http.StatusOK,
		},
		{
			request: "admin_role",
			as:      admin,
			want:    http.StatusOK,
		},
		{
			request: "banned_role_visible",
			as:      banned,
			want:    http.StatusOK,
		},
		{
			request: "permission_denied",
			as:      user,
			want:    http.StatusForbidden,
			others: cast{
				guest:  http.StatusUnauthorized,
				banned: http.StatusForbidden,
			},
		},
	})
}

func TestUserListingJourney(t *testing.T) {
	runJourney(t, api, "UserListing", []step{
		{
			request: "list_users",
			as:      admin,
			want:    http.StatusOK,
			others: cast{
				guest:  http.StatusUnauthorized,
				user:   http.StatusForbidden,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "search_and_paginate",
			as:      admin,
			want:    http.StatusOK,
		},
	})
}

func TestUserModerationJourney(t *testing.T) {
	unbannedAt := fixtureInstant.Add(time.Minute)

	runJourney(t, api, "UserModeration", []step{
		{
			request: "users_before_ban",
			as:      admin,
			want:    http.StatusOK,
		},
		{
			request: "ban_user",
			as:      admin,
			want:    http.StatusOK,
			others: cast{
				guest:  http.StatusUnauthorized,
				user:   http.StatusForbidden,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "users_after_ban",
			as:      admin,
			want:    http.StatusOK,
		},
		{
			request: "banned_role_visible",
			as:      user,
			want:    http.StatusOK,
		},
		{
			request: "banned_user_rejected",
			as:      user,
			want:    http.StatusForbidden,
		},
		{
			request: "unban_user",
			as:      admin,
			want:    http.StatusOK,
			others: cast{
				guest:  http.StatusUnauthorized,
				user:   http.StatusForbidden,
				banned: http.StatusForbidden,
			},
			at: unbannedAt,
		},
		{
			request: "users_after_unban",
			as:      admin,
			want:    http.StatusOK,
			at:      unbannedAt,
		},
		{
			request: "unbanned_user_restored",
			as:      user,
			want:    http.StatusOK,
			at:      unbannedAt,
		},
		{
			verify: "moderation_audited",
			at:     unbannedAt,
		},
	})
}

func TestModerationAuditFailureJourney(t *testing.T) {
	handler := auditUnavailableRoleUpdateHandler(t)
	failureAPI := &suite{
		db:      api.db,
		keto:    api.keto,
		kratos:  api.kratos,
		handler: handler,
	}

	runJourney(t, failureAPI, "ModerationAuditFailure", []step{
		{
			request: "ban_user_audit_unavailable",
			as:      admin,
			want:    http.StatusInternalServerError,
		},
		{
			request: "banned_role_visible",
			as:      user,
			want:    http.StatusOK,
		},
		{
			request: "banned_user_rejected",
			as:      user,
			want:    http.StatusForbidden,
		},
		{
			verify: "audit_empty",
		},
	})
}

func TestAnnouncementLifecycleJourney(t *testing.T) {
	afterExpiry := fixtureInstant.Add(8 * 24 * time.Hour)

	runJourney(t, api, "AnnouncementLifecycle", []step{
		{
			request: "nothing_announced",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "create_announcement",
			as:      admin,
			want:    http.StatusCreated,
			others: cast{
				none:   http.StatusBadRequest,
				guest:  http.StatusUnauthorized,
				user:   http.StatusForbidden,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "announced",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "expired",
			as:      guest,
			want:    http.StatusOK,
			at:      afterExpiry,
		},
		{
			request: "listed_for_admin",
			as:      admin,
			want:    http.StatusOK,
			at:      afterExpiry,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden},
		},
		{
			request: "delete_announcement",
			as:      admin,
			want:    http.StatusNoContent,
			at:      afterExpiry,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "gone",
			as:      admin,
			want:    http.StatusNotFound,
			at:      afterExpiry,
		},
		{verify: "soft_deleted"},
	})
}

func TestPageLifecycleJourney(t *testing.T) {
	// Keep API-created revision IDs stable in the HTTP fixtures. This journey
	// and its steps must stay sequential while the UUID source is overridden.
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	scheduled := fixtureInstant.Add(time.Hour)
	published := fixtureInstant.Add(2 * time.Hour)
	retitled := fixtureInstant.Add(3 * time.Hour)
	revised := fixtureInstant.Add(4 * time.Hour)
	saved := fixtureInstant.Add(5 * time.Hour)
	deleted := fixtureInstant.Add(6 * time.Hour)

	runJourney(t, api, "PageLifecycle", []step{
		{
			request: "create_draft",
			as:      admin,
			want:    http.StatusCreated,
			others: cast{
				none:   http.StatusBadRequest,
				guest:  http.StatusUnauthorized,
				user:   http.StatusForbidden,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "draft_for_admin",
			as:      admin,
			want:    http.StatusOK,
			others:  cast{user: http.StatusForbidden},
		},
		{
			request: "draft_hidden",
			as:      guest,
			want:    http.StatusNotFound,
		},
		{
			request: "draft_listed",
			as:      admin,
			want:    http.StatusOK,
		},
		{
			request: "draft_excluded",
			as:      admin,
			want:    http.StatusOK,
		},
		{
			request: "initial_version",
			as:      admin,
			want:    http.StatusOK,
		},
		{
			request: "schedule_and_rename",
			as:      admin,
			want:    http.StatusOK,
			at:      scheduled,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "scheduled_hidden",
			as:      guest,
			want:    http.StatusNotFound,
			at:      scheduled,
		},
		{
			request: "metadata_keeps_version",
			as:      admin,
			want:    http.StatusOK,
			at:      scheduled,
		},
		{
			request: "published",
			as:      guest,
			want:    http.StatusOK,
			at:      published,
		},
		{
			request: "old_slug_gone",
			as:      guest,
			want:    http.StatusNotFound,
			at:      published,
		},
		{
			request: "change_title",
			as:      admin,
			want:    http.StatusOK,
			at:      retitled,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "change_html",
			as:      admin,
			want:    http.StatusOK,
			at:      revised,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "latest_visible",
			as:      guest,
			want:    http.StatusOK,
			at:      revised,
		},
		{
			request: "version_history",
			as:      admin,
			want:    http.StatusOK,
			at:      revised,
			others:  cast{user: http.StatusForbidden},
		},
		{
			request: "original_version",
			as:      admin,
			want:    http.StatusOK,
			at:      revised,
			others:  cast{user: http.StatusForbidden},
		},
		{
			request: "title_version",
			as:      admin,
			want:    http.StatusOK,
			at:      revised,
		},
		{
			request: "latest_version",
			as:      admin,
			want:    http.StatusOK,
			at:      revised,
		},
		{
			request: "save_unchanged",
			as:      admin,
			want:    http.StatusOK,
			at:      saved,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "unchanged_history",
			as:      admin,
			want:    http.StatusOK,
			at:      saved,
		},
		{
			request: "delete_page",
			as:      admin,
			want:    http.StatusNoContent,
			at:      deleted,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "deleted_page_hidden",
			as:      guest,
			want:    http.StatusNotFound,
			at:      deleted,
		},
		{
			request: "deleted_admin_page",
			as:      admin,
			want:    http.StatusNotFound,
			at:      deleted,
		},
		{
			request: "deleted_from_list",
			as:      admin,
			want:    http.StatusOK,
			at:      deleted,
		},
		{
			request: "deleted_history_empty",
			as:      admin,
			want:    http.StatusOK,
			at:      deleted,
		},
		{
			request: "deleted_version_missing",
			as:      admin,
			want:    http.StatusNotFound,
			at:      deleted,
		},
		{verify: "soft_deleted", at: deleted},
	})
}

func TestPostLifecycleJourney(t *testing.T) {
	// Keep API-created revision IDs stable in the HTTP fixtures. This journey
	// and its steps must stay sequential while the UUID source is overridden.
	uuid.SetRand(rand.New(rand.NewSource(1)))
	defer uuid.SetRand(nil)

	scheduled := fixtureInstant.Add(time.Hour)
	published := fixtureInstant.Add(2 * time.Hour)
	retitled := fixtureInstant.Add(3 * time.Hour)
	revised := fixtureInstant.Add(4 * time.Hour)
	saved := fixtureInstant.Add(5 * time.Hour)
	deleted := fixtureInstant.Add(6 * time.Hour)

	runJourney(t, api, "PostLifecycle", []step{
		{
			request: "create_draft",
			as:      admin,
			want:    http.StatusCreated,
			others: cast{
				none:   http.StatusBadRequest,
				guest:  http.StatusUnauthorized,
				user:   http.StatusForbidden,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "draft_for_admin",
			as:      admin,
			want:    http.StatusOK,
			others:  cast{user: http.StatusForbidden},
		},
		{
			request: "draft_hidden",
			as:      guest,
			want:    http.StatusNotFound,
		},
		{
			request: "draft_listed",
			as:      admin,
			want:    http.StatusOK,
			others:  cast{guest: http.StatusForbidden, user: http.StatusForbidden},
		},
		{
			request: "draft_excluded",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "initial_version",
			as:      admin,
			want:    http.StatusOK,
		},
		{
			request: "schedule_and_rename",
			as:      admin,
			want:    http.StatusOK,
			at:      scheduled,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "scheduled_hidden",
			as:      guest,
			want:    http.StatusNotFound,
			at:      scheduled,
		},
		{
			request: "scheduled_excluded",
			as:      guest,
			want:    http.StatusOK,
			at:      scheduled,
		},
		{
			request: "metadata_keeps_version",
			as:      admin,
			want:    http.StatusOK,
			at:      scheduled,
		},
		{
			request: "published",
			as:      guest,
			want:    http.StatusOK,
			at:      published,
		},
		{
			request: "published_listed",
			as:      guest,
			want:    http.StatusOK,
			at:      published,
		},
		{
			request: "old_slug_gone",
			as:      guest,
			want:    http.StatusNotFound,
			at:      published,
		},
		{
			request: "change_title",
			as:      admin,
			want:    http.StatusOK,
			at:      retitled,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "change_content",
			as:      admin,
			want:    http.StatusOK,
			at:      revised,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "latest_visible",
			as:      guest,
			want:    http.StatusOK,
			at:      revised,
		},
		{
			request: "latest_listed",
			as:      guest,
			want:    http.StatusOK,
			at:      revised,
		},
		{
			request: "version_history",
			as:      admin,
			want:    http.StatusOK,
			at:      revised,
			others:  cast{user: http.StatusForbidden},
		},
		{
			request: "original_version",
			as:      admin,
			want:    http.StatusOK,
			at:      revised,
			others:  cast{user: http.StatusForbidden},
		},
		{
			request: "title_version",
			as:      admin,
			want:    http.StatusOK,
			at:      revised,
		},
		{
			request: "latest_version",
			as:      admin,
			want:    http.StatusOK,
			at:      revised,
		},
		{
			request: "save_unchanged",
			as:      admin,
			want:    http.StatusOK,
			at:      saved,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "unchanged_history",
			as:      admin,
			want:    http.StatusOK,
			at:      saved,
		},
		{
			request: "delete_post",
			as:      admin,
			want:    http.StatusNoContent,
			at:      deleted,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "deleted_post_hidden",
			as:      guest,
			want:    http.StatusNotFound,
			at:      deleted,
		},
		{
			request: "deleted_admin_post",
			as:      admin,
			want:    http.StatusNotFound,
			at:      deleted,
		},
		{
			request: "deleted_from_list",
			as:      guest,
			want:    http.StatusOK,
			at:      deleted,
		},
		{
			request: "deleted_admin_list",
			as:      admin,
			want:    http.StatusOK,
			at:      deleted,
		},
		{
			request: "deleted_history_empty",
			as:      admin,
			want:    http.StatusOK,
			at:      deleted,
		},
		{
			request: "deleted_version_missing",
			as:      admin,
			want:    http.StatusNotFound,
			at:      deleted,
		},
		{verify: "soft_deleted", at: deleted},
	})
}

// An administrator updates the language catalog display name; another user sees
// that name on existing yearly scores without changing any logs.
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
