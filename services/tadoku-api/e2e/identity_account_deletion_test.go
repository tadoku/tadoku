package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const accountDeletionPath = "/identity/settings/account-deletion"

func TestIdentityAccountDeletionRefusals(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		login  string
		body   string
		stale  bool
		edit   func(*http.Request)
		want   int
		reason string
	}{
		{
			name:   "guest",
			userID: readerUserID,
			body:   `{"email":"reader@example.test"}`,
			want:   http.StatusUnauthorized,
		},
		{
			name:   "stale session",
			userID: readerUserID,
			login:  "reader@example.test",
			body:   `{"email":"reader@example.test"}`,
			stale:  true,
			want:   http.StatusUnauthorized,
			reason: `"code":"stale_session"`,
		},
		{
			name:   "missing email",
			userID: readerUserID,
			login:  "reader@example.test",
			body:   `{}`,
			want:   http.StatusBadRequest,
			reason: `"code":"invalid_email"`,
		},
		{
			name:   "another address",
			userID: readerUserID,
			login:  "reader@example.test",
			body:   `{"email":"user2@example.test"}`,
			want:   http.StatusBadRequest,
			reason: `"code":"confirmation_mismatch"`,
		},
		{
			name:   "foreign origin",
			userID: readerUserID,
			login:  "reader@example.test",
			body:   `{"email":"reader@example.test"}`,
			edit:   func(r *http.Request) { r.Header.Set("Origin", "https://tadoku.example") },
			want:   http.StatusForbidden,
			reason: `"code":"invalid_origin"`,
		},
		{
			name:   "non-JSON body",
			userID: readerUserID,
			login:  "reader@example.test",
			body:   `{"email":"reader@example.test"}`,
			edit:   func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") },
			want:   http.StatusBadRequest,
			reason: "application/json",
		},
		{
			name:   "administrator",
			userID: adminUserID,
			login:  "admin@example.test",
			body:   `{"email":"admin@example.test"}`,
			want:   http.StatusConflict,
			reason: `"code":"admin_account"`,
		},
		{
			name:   "banned user",
			userID: bannedUserID,
			login:  "banned@example.test",
			body:   `{"email":"banned@example.test"}`,
			want:   http.StatusConflict,
			reason: `"code":"banned_account"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetAccountDeletion(t)
			cookie := ""
			if tc.login != "" {
				cookie = productionLoginCookie(t, tc.login)
			}
			if tc.stale {
				ageProductionSessions(t, tc.userID)
			}
			before := snapshot(t, targetSnapshotSQL, tc.userID)

			response := identityRequest(t, api.handler, "tadoku/prod", "POST", accountDeletionPath, tc.body, cookie,
				tc.edit)

			requireErrorBody(t, response, tc.want, tc.reason)
			if after := snapshot(t, targetSnapshotSQL, tc.userID); after != before {
				t.Errorf("refused request changed state\nbefore=%s\nafter=%s", before, after)
			}
			if jobs := countJobs(t); jobs != 0 {
				t.Errorf("refused request enqueued %d jobs", jobs)
			}
		})
	}
}

func TestIdentityAccountDeletionQueuedBehindFilingFails(t *testing.T) {
	resetAccountDeletion(t)
	cookie := productionLoginCookie(t, "reader@example.test")
	filing := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	_, err := filing.Exec(t.Context(), `update users set deletion_locked_at = '2026-09-12 12:00:00'
		where id = $1`, readerUserID)
	if err != nil {
		t.Fatal(err)
	}
	request := newIdentityRequest(t, "tadoku/prod", "POST", accountDeletionPath, `{"email":"reader@example.test"}`,
		cookie, nil)

	response := httptest.NewRecorder()
	served := make(chan error, 1)
	go func() {
		atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })
		served <- nil
	}()
	waitForLockWaiters(t, 1, served)
	if err := filing.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}

	<-served
	if response.Code != http.StatusUnauthorized {
		t.Errorf("request queued behind a filing: status=%d body=%s, want 401", response.Code, response.Body)
	}
	if jobs := countJobs(t); jobs != 0 {
		t.Errorf("request queued behind a filing enqueued %d jobs", jobs)
	}
	if events := snapshot(t, "select count(*)::text from moderation_audit_log"); events != "0" {
		t.Errorf("request queued behind a filing recorded %s audit events", events)
	}
}

func ageProductionSessions(t *testing.T, userID string) {
	t.Helper()
	_, err := api.db.Pool.Exec(t.Context(), `update sessions
		set authenticated_at = $2::timestamptz - interval '1 hour 1 second'
		where tenant = 'tadoku/prod' and user_id = $1`, userID, fixtureInstant)
	if err != nil {
		t.Fatal(err)
	}
}
