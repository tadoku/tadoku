package e2e_test

import (
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Every important user journey in the application lives in this file. Replay a
// request as other cast members only where that adds information: rejected
// identities on mutating steps, and a second user only to observe limited
// visibility of a resource.

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
