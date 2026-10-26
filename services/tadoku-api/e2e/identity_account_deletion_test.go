package e2e_test

import (
	"net/http"
	"testing"
)

const queuedSelfServiceDir = "testdata/AccountDeletion/queued_self_service_refused"

func TestIdentityAccountDeletionQueuedBehindFilingFails(t *testing.T) {
	resetAccountDeletion(t)
	request := selfServiceDeletion(t)
	filing := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	_, err := filing.Exec(t.Context(), `update users set deletion_locked_at = '2026-09-12 12:00:00'
		where id = $1`, readerUserID)
	if err != nil {
		t.Fatal(err)
	}

	atFixtureInstant(func() {
		served := serveAsync(request)
		waitForLockWaiters(t, 1, served)

		if err := filing.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		if response := <-served; response.Code != http.StatusUnauthorized {
			t.Errorf("request queued behind a filing: status=%d body=%s, want 401", response.Code, response.Body)
		}
	})

	checkVerifyGolden(t, api, queuedSelfServiceDir)
}

// selfServiceDeletion logs the reader in and returns their deletion request, so call it before holding their rows.
func selfServiceDeletion(t *testing.T) *http.Request {
	t.Helper()
	login := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login", readerLogin, "", nil)
	requireIdentityStatus(t, login, http.StatusOK)
	return newIdentityRequest(
		t,
		"tadoku/prod",
		"POST",
		"/identity/settings/account-deletion",
		`{"email":"reader@example.test"}`,
		login.Result().Cookies()[0].String(),
		nil,
	)
}
