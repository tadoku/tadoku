package e2e_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const recoverySettingsURL = "https://identity-a-0123abcd.authv2.tadoku.dev.lab/?from=recovery#password"

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
	resetIdentity(t)
	var durations [2][]time.Duration
	for n := 0; n < 4; n++ {
		for i, email := range []string{"READER@EXAMPLE.TEST", "unknown@example.test"} {
			start := time.Now()
			response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/recovery",
				fmt.Sprintf(`{"email":%q,"return_to":%q}`, email, recoverySettingsURL), "", nil)
			durations[i] = append(durations[i], time.Since(start))
			requireIdentityStatus(t, response, http.StatusAccepted)
			if response.Body.Len() != 0 || response.Header().Get("Set-Cookie") != "" ||
				response.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatal("recovery disclosed existence or created a session")
			}
		}
	}
	var means [2]time.Duration
	for i, samples := range durations {
		for _, duration := range samples {
			means[i] += duration / time.Duration(len(samples))
		}
	}
	if means[0] < 200*time.Millisecond || means[1] < 200*time.Millisecond ||
		means[0] > means[1]+100*time.Millisecond || means[1] > means[0]+100*time.Millisecond {
		t.Errorf("recovery timing known=%s unknown=%s", means[0], means[1])
	}
	t.Logf("recovery means known=%s unknown=%s; samples=%v", means[0], means[1], durations)

	var tokens, emails, jobs int
	err := api.db.Pool.QueryRow(t.Context(), `select
		(select count(*) from recovery_tokens where tenant = $1),
		(select count(*) from emails where tenant = $1),
		(select count(*) from jobs where tenant = $1 and task_type = 'email.send.v1'
		 and (payload::jsonb - 'email_id') = '{}'::jsonb)`, identityBranchA).Scan(&tokens, &emails, &jobs)
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
		where e.tenant = $1 and r.tenant = e.tenant limit 1`, identityBranchA).
		Scan(&rawURL, &recipient, &storedHash, &expires)
	if err != nil {
		t.Fatal(err)
	}
	link, err := url.Parse(rawURL)
	if err != nil || link.Host != "identity-a-0123abcd.tadoku.dev.lab" ||
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
	for _, tc := range []struct {
		name, tenant string
		expires      time.Time
		used         *time.Time
	}{
		{"expired", identityBranchA, fixtureInstant, nil},
		{"used", identityBranchA, fixtureInstant.Add(time.Hour), &fixtureInstant},
		{"other tenant", identityBranchB, fixtureInstant.Add(time.Hour), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetIdentity(t)
			token := seedRecoveryToken(t, 'a', tc.expires, tc.used)
			response := identityRequest(t, api.handler, tc.tenant, "GET", "/identity/recovery/"+token, "", "", nil)
			requireIdentityStatus(t, response, http.StatusBadRequest)
			if !strings.Contains(response.Body.String(), `"code":"invalid_recovery_token"`) ||
				response.Header().Get("Set-Cookie") != "" {
				t.Fatal("invalid token changed a session or disclosed a different error")
			}
		})
	}
}

func TestIdentityRecoveryConsumeRevokesAndSetsPassword(t *testing.T) {
	resetIdentity(t)
	old := identityLoginCookie(t)
	first := seedRecoveryToken(t, 'a', fixtureInstant.Add(time.Hour), nil)
	second := seedRecoveryToken(t, 'b', fixtureInstant.Add(time.Hour), nil)
	response := identityRequest(t, api.handler, identityBranchA, "GET",
		"/identity/recovery/"+first+"?return_to="+url.QueryEscape(recoverySettingsURL), "", "", nil)
	requireIdentityStatus(t, response, http.StatusSeeOther)
	if response.Header().Get("Location") != "https://identity-a-0123abcd.authv2.tadoku.dev.lab/?return_to="+
		url.QueryEscape(recoverySettingsURL) {
		t.Fatal("recovery lost its validated return destination")
	}
	cookie := response.Result().Cookies()[0]
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatal("recovery cookie weakened session security")
	}
	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", old, nil)
	if identityView(t, response).Account != nil {
		t.Fatal("recovery left an old session active")
	}
	for _, token := range []string{first, second} {
		response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/recovery/"+token, "", "", nil)
		requireIdentityStatus(t, response, http.StatusBadRequest)
	}
	response = identityRequest(t, api.handler, identityBranchA, "PUT", "/identity/settings/password",
		`{"password":"Replacement-839!"}`, cookie.String(), nil)
	requireIdentityStatus(t, response, http.StatusNoContent)
	response = identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Replacement-839!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
	response = identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusUnauthorized)
}

func TestIdentityRecoveryConcurrentConsumption(t *testing.T) {
	for _, sameToken := range []bool{true, false} {
		t.Run(fmt.Sprintf("sameToken=%t", sameToken), func(t *testing.T) {
			resetIdentity(t)
			first := seedRecoveryToken(t, 'a', fixtureInstant.Add(time.Hour), nil)
			second := first
			if !sameToken {
				second = seedRecoveryToken(t, 'b', fixtureInstant.Add(time.Hour), nil)
			}
			contents, err := os.ReadFile("testdata/IdentityV2/tokens.json")
			if err != nil {
				t.Fatal(err)
			}
			var tokens map[string]string
			if err := json.Unmarshal(contents, &tokens); err != nil {
				t.Fatal(err)
			}
			responses := make([]*httptest.ResponseRecorder, 2)
			atFixtureInstant(func() {
				var group sync.WaitGroup
				start := make(chan struct{})
				for i, token := range []string{first, second} {
					group.Add(1)
					go func() {
						defer group.Done()
						request := httptest.NewRequest("GET", "/identity/recovery/"+token, nil)
						request.Header.Set("Authorization", "Bearer "+tokens[identityBranchA])
						responses[i] = httptest.NewRecorder()
						<-start
						api.handler.ServeHTTP(responses[i], request)
					}()
				}
				close(start)
				group.Wait()
			})
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
	resetIdentity(t)
	token := seedRecoveryToken(t, 'a', fixtureInstant.Add(time.Hour), nil)
	response := identityRequest(t, api.handler, identityBranchA, "GET",
		"/identity/recovery/"+token+"?return_to=https%3A%2F%2Fevil.test", "", "", nil)
	requireIdentityStatus(t, response, http.StatusBadRequest)
	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/recovery/"+token, "", "", nil)
	requireIdentityStatus(t, response, http.StatusSeeOther)
	if response.Header().Get("Location") != "https://identity-a-0123abcd.authv2.tadoku.dev.lab/" {
		t.Fatal("default Settings redirect escaped the token tenant")
	}
}

func TestIdentityRecoveryRejectsUnrepresentableReturnTo(t *testing.T) {
	prefix := "https://authv2.tadoku.dev.lab/?q="
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
			resetIdentity(t)
			destination := prefix + strings.Repeat(tc.character, 2048-len(prefix))
			body, err := json.Marshal(map[string]string{"email": tc.email, "return_to": destination})
			if err != nil {
				t.Fatal(err)
			}
			response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/recovery", string(body), "", nil)
			requireIdentityStatus(t, response, tc.want)
		})
	}
}
