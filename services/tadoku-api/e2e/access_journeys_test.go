package e2e_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	transport "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

// Test safety: the audit service reports an unrecorded change through the
// default logger, so these journeys swap it for a buffer and restore it.
func captureDefaultLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return &logs
}

func auditUnavailableHandler(t *testing.T) *transport.Router {
	t.Helper()

	closedPool := openClosedPool(t, api.db.AppPool.Config().ConnString())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router, _, err := newTestRouterWithLogger(
		t.Context(),
		api.db.AppPool,
		closedPool,
		logger,
	)
	if err != nil {
		t.Fatal(err)
	}

	return router
}

// Test safety: the trigger changes the suite database's schema, so the cleanup
// drops it before the next test writes to the table.
func failInserts(t *testing.T, table string) {
	t.Helper()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := api.db.Pool.Exec(ctx, `drop trigger fail_insert on `+table+`;
			drop function fail_insert()`)
		if err != nil {
			t.Error(err)
		}
	})

	_, err := api.db.Pool.Exec(t.Context(), `create function fail_insert() returns trigger language plpgsql as $$
		begin
			raise exception 'injected insert failure' using errcode = 'disk_full';
		end
		$$;
		create trigger fail_insert before insert on `+table+` for each row execute function fail_insert()`)
	if err != nil {
		t.Fatal(err)
	}
}

func requireLogged(t *testing.T, logs *bytes.Buffer, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(logs.String(), fragment) {
			t.Errorf("logs do not contain %q:\n%s", fragment, logs.String())
		}
	}
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
			request: "ban_admin_rejected",
			as:      admin,
			want:    http.StatusForbidden,
			at:      unbannedAt,
		},
		{
			verify: "moderation_audited",
			at:     unbannedAt,
		},
	})
}

func TestAdminAccountDeletionJourney(t *testing.T) {
	useContestEventUUIDs(t)

	runJourney(t, api, "AdminAccountDeletion", []step{
		{
			request: "finished_contest_leaderboard",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "running_contest_leaderboard",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "yearly_leaderboard",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "global_leaderboard",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "file_deletion",
			as:      admin,
			want:    http.StatusAccepted,
			others: cast{
				guest:  http.StatusUnauthorized,
				user:   http.StatusForbidden,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "refile_refused",
			as:      admin,
			want:    http.StatusConflict,
		},
		{
			verify: "deletion_filed",
		},
		{
			job: "run_worker",
		},
		{
			request: "deleted_user_not_found",
			as:      admin,
			want:    http.StatusNotFound,
		},
		{
			verify: "account_deleted",
		},
		{
			verify: "deletion_recorded",
		},
		{
			verify: "others_untouched",
		},
		{
			verify: "finished_history_kept",
		},
		{
			request: "finished_contest_keeps_scores",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "running_contest_drops_user",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "yearly_leaderboard_drops_user",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "global_leaderboard_drops_user",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "organizer_shown_as_deleted",
			as:      user2,
			want:    http.StatusOK,
		},
		{
			request: "running_contest_withdrawals",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "finished_contest_updates",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "users_after_deletion",
			as:      admin,
			want:    http.StatusOK,
		},
	})
}

func TestModerationAuditFailureJourney(t *testing.T) {
	afterReset := func(t *testing.T) { failInserts(t, "moderation_audit_log") }

	runJourneyWithSetup(t, api, api.handler, "ModerationAuditFailure", afterReset, []step{
		{
			request: "ban_user_audit_unavailable",
			as:      admin,
			want:    http.StatusServiceUnavailable,
		},
		{
			request: "role_unchanged",
			as:      user,
			want:    http.StatusOK,
		},
		{
			request: "user_admitted",
			as:      user,
			want:    http.StatusOK,
		},
		{
			verify: "audit_empty",
		},
	})
}

func TestModerationBanWriteFailureJourney(t *testing.T) {
	afterReset := func(t *testing.T) { failInserts(t, "permission_relationships") }

	runJourneyWithSetup(t, api, api.handler, "ModerationBanWriteFailure", afterReset, []step{
		{
			request: "ban_user_write_unavailable",
			as:      admin,
			want:    http.StatusServiceUnavailable,
		},
		{
			request: "role_unchanged",
			as:      user,
			want:    http.StatusOK,
		},
		{
			verify: "audit_empty",
		},
	})
}
