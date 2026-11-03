package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

const recoverySettingsURL = "https://identity-a-0123abcd.account.tadoku.dev.lab/?from=recovery#password"

func seedRecoveryToken(t *testing.T, marker byte, expires time.Time, used *time.Time) string {
	t.Helper()
	raw := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat(string(marker), 32)))
	digest := sha256.Sum256([]byte(raw))
	_, err := api.db.Pool.Exec(t.Context(), `insert into recovery_tokens
		(tenant, token_hash, user_id, expires_at, used_at)
		values ($1, $2, 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', $3, $4)`,
		identityBranchA, hex.EncodeToString(digest[:]), expires, used)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func identityLoginCookie(t *testing.T) string {
	t.Helper()
	response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
	return response.Result().Cookies()[0].String()
}

func TestIdentityRecoveryEnumerationAndTransactionalEmail(t *testing.T) {
	t.Parallel()
	s := newIdentityScenario(t)
	tenantKey := s.tenant(t, "identity-a").String()
	jwks := publishedJWKS(t)

	synctest.Test(t, func(t *testing.T) {
		handler := recoveryTimingHandler(t, &recoveryWriteDelay{}, jwks)
		var durations [2][]time.Duration
		for n := 0; n < 4; n++ {
			for i, email := range []string{"READER@EXAMPLE.TEST", "unknown@example.test"} {
				request := newIdentityRequest(t, "", "POST", "/identity/recovery",
					fmt.Sprintf(`{"email":%q,"return_to":%q}`, email, recoverySettingsURL), "", nil)
				start := time.Now()
				response := s.serveWith(t, handler, "identity-a", guest, request)
				elapsed := time.Since(start)
				durations[i] = append(durations[i], elapsed)
				if elapsed != 300*time.Millisecond {
					t.Errorf("%s elapsed=%s, want 300ms", email, elapsed)
				}
				requireIdentityStatus(t, response, http.StatusAccepted)
				if response.Body.Len() != 0 || response.Header().Get("Set-Cookie") != "" ||
					response.Header().Get("Cache-Control") != "private, no-store" {
					t.Fatal("recovery disclosed existence or created a session")
				}
			}
		}
		t.Logf("recovery known/unknown samples=%v", durations)
	})

	var tokens, emails, jobs int
	err := api.db.Pool.QueryRow(t.Context(), `select
		(select count(*) from recovery_tokens where tenant = $1),
		(select count(*) from emails where tenant = $1),
		(select count(*) from jobs where tenant = $1 and task_type = 'email.send.v1'
		 and (payload::jsonb - 'email_id') = '{}'::jsonb)`, tenantKey).Scan(&tokens, &emails, &jobs)
	if err != nil ||
		tokens != recoveryRecipientBudget ||
		emails != recoveryRecipientBudget ||
		jobs != recoveryRecipientBudget {
		t.Fatalf("recovery tokens=%d emails=%d ID-only jobs=%d err=%v", tokens, emails, jobs, err)
	}
	var rawURL, recipient, storedHash string
	var expires time.Time
	err = api.db.Pool.QueryRow(t.Context(), `select e.params->>'recovery_url', e.recipient, r.token_hash, r.expires_at
		from emails e cross join recovery_tokens r
		where e.tenant = $1 and r.tenant = e.tenant limit 1`, tenantKey).
		Scan(&rawURL, &recipient, &storedHash, &expires)
	if err != nil {
		t.Fatal(err)
	}
	_, route, _ := strings.Cut(tenantKey, "/")
	link, err := url.Parse(rawURL)
	if err != nil || link.Host != route+".tadoku.dev.lab" ||
		!strings.HasPrefix(link.Path, "/api/internal/identity/recovery/") ||
		link.Query().Get("return_to") != recoverySettingsURL || recipient != "reader@example.test" ||
		!expires.Equal(fixtureInstant.Add(time.Hour)) {
		t.Fatalf("recovery link host/path/return/recipient/expiry invalid (parse error %v)", err)
	}
	raw := strings.TrimPrefix(link.Path, "/api/internal/identity/recovery/")
	secret, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(secret) != 32 || len(storedHash) != 64 || storedHash == raw {
		t.Fatal("recovery secret must contain 256 random bits and only its hash may be stored")
	}
}

func TestIdentityRecoveryInvalidTokens(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, role string
		expires    time.Time
		used       *time.Time
	}{
		{"expired", "identity-a", fixtureInstant, nil},
		{"used", "identity-a", fixtureInstant.Add(time.Hour), &fixtureInstant},
		{"other tenant", "identity-b", fixtureInstant.Add(time.Hour), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newPairedIdentityScenario(t)

			token := seedRecoveryTokenIn(t, s, "identity-a", 'a', tc.expires, tc.used)
			response := guestIdentityRequest(t, s, tc.role, "GET", "/identity/recovery/"+token, "", "", nil)
			requireIdentityStatus(t, response, http.StatusBadRequest)
			if !strings.Contains(response.Body.String(), `"code":"invalid_recovery_token"`) ||
				response.Header().Get("Set-Cookie") != "" {
				t.Fatal("invalid token changed a session or disclosed a different error")
			}
		})
	}
}

func TestIdentityRecoveryConsumeRevokesAndSetsPassword(t *testing.T) {
	t.Parallel()
	s := newIdentityScenario(t)
	old := loginCookieIn(t, s, "identity-a")
	first := seedRecoveryTokenIn(t, s, "identity-a", 'a', fixtureInstant.Add(time.Hour), nil)
	second := seedRecoveryTokenIn(t, s, "identity-a", 'b', fixtureInstant.Add(time.Hour), nil)

	response := guestIdentityRequest(t, s, "identity-a", "GET",
		"/identity/recovery/"+first+"?return_to="+url.QueryEscape(recoverySettingsURL), "", "", nil)
	requireIdentityStatus(t, response, http.StatusSeeOther)
	want := recoveryLocation(t, s, "identity-a") + "&return_to=" + url.QueryEscape(recoverySettingsURL)
	if response.Header().Get("Location") != want {
		t.Fatal("recovery must open the Password editor with its validated return destination")
	}
	cookie := response.Result().Cookies()[0]
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatal("recovery cookie weakened session security")
	}
	response = guestIdentityRequest(t, s, "identity-a", "GET", "/identity/session", "", old, nil)
	if identityView(t, response).Account != nil {
		t.Fatal("recovery left an old session active")
	}
	for _, token := range []string{first, second} {
		response = guestIdentityRequest(t, s, "identity-a", "GET", "/identity/recovery/"+token, "", "", nil)
		requireIdentityStatus(t, response, http.StatusBadRequest)
	}

	response = guestIdentityRequest(t, s, "identity-a", "PUT", "/identity/settings/password",
		`{"password":"Replacement-839!"}`, cookie.String(), nil)
	requireIdentityStatus(t, response, http.StatusNoContent)
	response = guestIdentityRequest(t, s, "identity-a", "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Replacement-839!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
	response = guestIdentityRequest(t, s, "identity-a", "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusUnauthorized)
}

func TestIdentityRecoveryConcurrentConsumption(t *testing.T) {
	t.Parallel()
	for _, sameToken := range []bool{true, false} {
		t.Run(fmt.Sprintf("sameToken=%t", sameToken), func(t *testing.T) {
			s := newIdentityScenario(t)
			first := seedRecoveryTokenIn(t, s, "identity-a", 'a', fixtureInstant.Add(time.Hour), nil)
			second := first
			if !sameToken {
				second = seedRecoveryTokenIn(t, s, "identity-a", 'b', fixtureInstant.Add(time.Hour), nil)
			}

			responses := make([]*httptest.ResponseRecorder, 2)
			var group sync.WaitGroup
			start := make(chan struct{})
			for i, token := range []string{first, second} {
				request := httptest.NewRequest("GET", "/identity/recovery/"+token, nil)
				group.Go(func() {
					<-start
					responses[i] = s.serve(t, "identity-a", guest, request)
				})
			}
			close(start)
			group.Wait()

			statuses := map[int]int{}
			for _, response := range responses {
				statuses[response.Code]++
			}
			if statuses[303] != 1 || statuses[400] != 1 {
				t.Fatalf("concurrent consume statuses=%v, want one 303 and one 400", statuses)
			}
		})
	}
}

func TestIdentityRecoveryRejectsRedirectBeforeConsuming(t *testing.T) {
	t.Parallel()
	s := newIdentityScenario(t)
	token := seedRecoveryTokenIn(t, s, "identity-a", 'a', fixtureInstant.Add(time.Hour), nil)

	response := guestIdentityRequest(t, s, "identity-a", "GET",
		"/identity/recovery/"+token+"?return_to=https%3A%2F%2Fevil.test", "", "", nil)
	requireIdentityStatus(t, response, http.StatusBadRequest)
	response = guestIdentityRequest(t, s, "identity-a", "GET", "/identity/recovery/"+token, "", "", nil)
	requireIdentityStatus(t, response, http.StatusSeeOther)
	if response.Header().Get("Location") != recoveryLocation(t, s, "identity-a") {
		t.Fatal("default Settings redirect escaped the token tenant or closed the Password editor")
	}
}

func TestIdentityRecoveryRejectsUnrepresentableReturnTo(t *testing.T) {
	t.Parallel()
	prefix := "https://account.tadoku.dev.lab/?q="
	for _, tc := range []struct {
		name, email, character string
		want                   int
	}{
		{"full destination known", "reader@example.test", "a", http.StatusBadRequest},
		{"full destination unknown", "unknown@example.test", "a", http.StatusBadRequest},
		{"expanded link known", "reader@example.test", "&", http.StatusBadRequest},
		{"expanded link unknown", "unknown@example.test", "&", http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newIdentityScenario(t)
			destination := prefix + strings.Repeat(tc.character, 2048-len(prefix))
			body, err := json.Marshal(map[string]string{"email": tc.email, "return_to": destination})
			if err != nil {
				t.Fatal(err)
			}

			response := guestIdentityRequest(t, s, "identity-a", "POST", "/identity/recovery", string(body), "", nil)
			requireIdentityStatus(t, response, tc.want)
		})
	}
}

func TestIdentityRecoveryRequiresUsableAccount(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, change string
		usable       bool
	}{
		{
			name:   "active",
			usable: true,
		},
		{
			name:   "inactive",
			change: "update accounts set state = 'inactive' where user_id = $1",
		},
		{
			name:   "deleted",
			change: "update users set deleted_at = '2026-01-05' where id = $1",
		},
		{
			name:   "deletion locked",
			change: "update users set deletion_locked_at = '2026-01-05' where id = $1",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := newIdentityScenario(t)
			tenantKey := s.tenant(t, "identity-a").String()
			if test.change != "" {
				writeTenantRows(t, tenantKey, test.change, scenarioCast[user].subject)
			}

			requested := recoverFromRole(t, s, "identity-a", "reader@example.test", "198.51.100.20", "")
			requireRecoveryAccepted(t, requested)
			written := 0
			if test.usable {
				written = 1
			}
			requireRecoveryWrites(t, tenantKey, written)

			token := seedRecoveryTokenIn(t, s, "identity-a", 'z', fixtureInstant.Add(time.Hour), nil)
			consumed := guestIdentityRequest(t, s, "identity-a", "GET", "/identity/recovery/"+token, "", "", nil)
			if test.usable {
				requireIdentityStatus(t, consumed, http.StatusSeeOther)
				return
			}
			requireRecoveryRejected(t, consumed, http.StatusBadRequest, "invalid_recovery_token")
		})
	}
}

func TestIdentityRecoveryDoesNotWaitForTheDeletionWorker(t *testing.T) {
	t.Parallel()
	for _, consume := range []bool{false, true} {
		t.Run(fmt.Sprintf("consume=%t", consume), func(t *testing.T) {
			s := newIdentityScenario(t)
			tenantKey := s.tenant(t, "identity-a").String()
			reader := scenarioCast[user].subject
			writeTenantRows(t, tenantKey, "update users set deletion_locked_at = '2026-01-05' where id = $1", reader)

			body := fmt.Sprintf(`{"email":"reader@example.test","return_to":%q}`, recoverySettingsURL)
			request := newIdentityRequest(t, "", "POST", "/identity/recovery", body, "", nil)
			if consume {
				token := seedRecoveryTokenIn(t, s, "identity-a", 'z', fixtureInstant.Add(time.Hour), nil)
				request = newIdentityRequest(t, "", "GET", "/identity/recovery/"+token, "", "", nil)
			}
			worker, err := api.db.Pool.Begin(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = worker.Rollback(context.Background()) })
			_, err = worker.Exec(
				t.Context(),
				"select 1 from accounts where tenant = $1 and user_id = $2 for update",
				tenantKey,
				reader,
			)
			if err != nil {
				t.Fatal(err)
			}

			served := serveAsyncIn(t, s, "identity-a", request)
			select {
			case response := <-served:
				if consume {
					requireRecoveryRejected(
						t,
						response,
						http.StatusBadRequest,
						"invalid_recovery_token",
					)
					return
				}
				requireRecoveryAccepted(t, response)
				requireRecoveryWrites(t, tenantKey, 0)
			case <-time.After(5 * time.Second):
				t.Error("recovery waits for the deletion worker holding the account row")
				_ = worker.Rollback(t.Context())
				<-served
			}
		})
	}
}
