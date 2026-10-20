package e2e_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

const identityBranchA = "tadoku/identity-a-0123abcd"
const identityBranchB = "tadoku/identity-b-0123abcd"
const identityFixturePassword = "Fixture-login-482!"

type identitySessionView struct {
	AccountApp string `json:"account_app"`
	Account    *struct {
		ID              string    `json:"id"`
		Email           string    `json:"email"`
		DisplayName     string    `json:"display_name"`
		CreatedAt       time.Time `json:"created_at"`
		AuthenticatedAt time.Time `json:"authenticated_at"`
	} `json:"account"`
}

func resetIdentity(t *testing.T) {
	t.Helper()
	api.reset(t, "")
	if err := api.db.Reset(t.Context(), "testdata/IdentityV2/setup.sql"); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"tadoku/prod", identityBranchA, identityBranchB} {
		key, err := tenant.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		if err := flipt.SetIdentityV2(key, true); err != nil {
			t.Fatal(err)
		}
	}
}

func identityRequest(
	t *testing.T,
	handler http.Handler,
	tenantKey, method, path, body, cookie string,
	edit func(*http.Request),
) *httptest.ResponseRecorder {
	t.Helper()
	contents, err := os.ReadFile("testdata/IdentityV2/tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	var tokens map[string]string
	if err := json.Unmarshal(contents, &tokens); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(method, "https://tadoku.dev.lab"+path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+tokens[tenantKey])
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://tadoku.dev.lab")
	request.Header.Set("Cookie", cookie)
	request.RemoteAddr = "198.51.100.20:12345"
	if edit != nil {
		edit(request)
	}
	response := httptest.NewRecorder()
	atFixtureInstant(func() { handler.ServeHTTP(response, request) })
	return response
}

func requireIdentityStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status=%d body=%s, want %d", response.Code, response.Body, want)
	}
}

func identityView(t *testing.T, response *httptest.ResponseRecorder) identitySessionView {
	t.Helper()
	requireIdentityStatus(t, response, http.StatusOK)
	var view identitySessionView
	if err := json.Unmarshal(response.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if response.Header().Get("Cache-Control") != "private, no-store" {
		t.Error("session response may be cached")
	}
	return view
}

func TestIdentityV2RegistrationValidation(t *testing.T) {
	cases := []struct {
		name, email, displayName, password string
		want                               int
	}{
		{"one character name", "anton@tadoku.app", "A", "antonreads99", 400},
		{"33 character name", "anton@tadoku.app", strings.Repeat("A", 33), "antonreads99", 400},
		{"seven character password", "anton@tadoku.app", "Anton", "abc123!", 400},
		{"over 72 bytes", "anton@tadoku.app", "Anton", strings.Repeat("界", 25), 400},
		{"similar domain", "anton@tadoku.app", "Anton", "tadoku.app!", 400},
		{"similar local part", "anton@tadoku.app", "Anton", "anton123", 400},
		{"allowed similarity", "anton@tadoku.app", "Anton", "antonreads99", 201},
		{"case insensitive duplicate", "READER@EXAMPLE.TEST", "Reader", "Independent-653!", 409},
		{"invalid email", "reader", "Reader", "Independent-653!", 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetIdentity(t)
			body, _ := json.Marshal(map[string]string{
				"email":        tc.email,
				"display_name": tc.displayName,
				"password":     tc.password,
			})
			response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/registration", string(body), "", nil)
			requireIdentityStatus(t, response, tc.want)
			if tc.want != 201 && response.Header().Get("Set-Cookie") != "" {
				t.Error("rejected registration created a cookie")
			}
		})
	}
}

func TestIdentityV2RequestSecurity(t *testing.T) {
	for _, tc := range []struct {
		name, path, body, origin, contentType string
		want                                  int
	}{
		{
			"foreign origin", "/identity/login",
			`{"email":"reader@example.test","password":"wrong-password"}`,
			"https://evil.test", "application/json", 403,
		},
		{"missing origin", "/identity/logout", "", "", "application/json", 403},
		{"origin suffix attack", "/identity/logout", "", "https://tadoku.dev.lab.evil.test", "application/json", 403},
		{
			"non JSON", "/identity/login",
			`{"email":"reader@example.test","password":"wrong-password"}`,
			"https://tadoku.dev.lab", "text/plain", 400,
		},
		{
			"foreign return", "/identity/recovery",
			`{"email":"reader@example.test","return_to":"https://evil.test/settings"}`,
			"https://tadoku.dev.lab", "application/json", 400,
		},
		{
			"return credentials", "/identity/recovery",
			`{"email":"reader@example.test","return_to":"https://evil.test@tadoku.app/settings"}`,
			"https://tadoku.dev.lab", "application/json", 400,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetIdentity(t)
			response := identityRequest(t, api.handler, identityBranchA, "POST", tc.path, tc.body, "", func(r *http.Request) {
				r.Header.Set("Origin", tc.origin)
				r.Header.Set("Content-Type", tc.contentType)
			})
			requireIdentityStatus(t, response, tc.want)
		})
	}
}

func TestIdentityV2LoginTimingAndThrottle(t *testing.T) {
	resetIdentity(t)
	var known, unknown []time.Duration
	for i := 0; i < 8; i++ {
		for _, email := range []string{"reader@example.test", "unknown@example.test"} {
			start := time.Now()
			body := fmt.Sprintf(`{"email":%q,"password":"wrong-password"}`, email)
			response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login", body, "", nil)
			duration := time.Since(start)
			requireIdentityStatus(t, response, http.StatusUnauthorized)
			if response.Body.String() != "{\"code\":\"wrong_password\",\"message\":\"invalid email or password\"}\n" {
				t.Errorf("credential error differs: %s", response.Body)
			}
			if email == "reader@example.test" {
				known = append(known, duration)
			} else {
				unknown = append(unknown, duration)
			}
		}
	}
	for i := range known {
		t.Logf("bcrypt sample %d known_ns=%d unknown_ns=%d", i, known[i].Nanoseconds(), unknown[i].Nanoseconds())
		if unknown[i] < 50*time.Millisecond {
			t.Errorf("unknown-email request omitted the expensive password comparison: %s", unknown[i])
		}
	}
	for i := 8; i < 10; i++ {
		response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login",
			`{"email":"reader@example.test","password":"wrong-password"}`, "", nil)
		requireIdentityStatus(t, response, http.StatusUnauthorized)
	}
	response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusTooManyRequests)
	if response.Header().Get("Set-Cookie") != "" {
		t.Error("throttled login issued a session")
	}
	response = identityRequest(t, api.handler, identityBranchB, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
}

func TestIdentityV2SessionJourney(t *testing.T) {
	resetIdentity(t)
	response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/registration",
		`{"email":"new@example.test","display_name":"New Reader","password":"Independent-653!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusCreated)
	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("registration must set one session cookie")
	}
	first := cookies[0]
	secret, err := base64.RawURLEncoding.DecodeString(first.Value)
	if err != nil || len(secret) != 32 {
		t.Fatal("session token does not contain a 256-bit secret")
	}
	digest := sha256.Sum256([]byte(first.Value))
	var storedHash, passwordHash string
	err = api.db.Pool.QueryRow(t.Context(), `select s.token_hash, a.password_hash
		from sessions s
		join accounts a on a.tenant = s.tenant and a.user_id = s.user_id
		where s.tenant = $1 and a.email_normalized = 'new@example.test'`, identityBranchA,
	).Scan(&storedHash, &passwordHash)
	if err != nil {
		t.Fatal(err)
	}
	if storedHash != hex.EncodeToString(digest[:]) || !strings.HasPrefix(passwordHash, "$2a$12$") {
		t.Error("credentials were not stored as SHA-256/bcrypt cost 12 hashes")
	}
	if first.Name != "tadoku_session_tadoku_identity-a-0123abcd" || first.Domain != "tadoku.dev.lab" ||
		!first.HttpOnly || !first.Secure || first.SameSite != http.SameSiteLaxMode ||
		first.Path != "/" || first.MaxAge != 7*24*60*60 {
		t.Error("registration cookie violates tenant security policy")
	}
	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", first.String(), nil)
	view := identityView(t, response)
	if view.AccountApp != "v2" || view.Account == nil || view.Account.Email != "new@example.test" ||
		view.Account.DisplayName != "New Reader" || !view.Account.CreatedAt.Equal(fixtureInstant) ||
		!view.Account.AuthenticatedAt.Equal(fixtureInstant) {
		t.Fatal("registration did not persist the complete session account")
	}
	response = identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login",
		`{"email":"NEW@EXAMPLE.TEST","password":"Independent-653!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
	second := response.Result().Cookies()[0]
	if first.Value == second.Value {
		t.Error("login reused a session token")
	}
	response = identityRequest(t, api.handler, identityBranchA, "POST", "/identity/logout", "", second.String(), nil)
	requireIdentityStatus(t, response, http.StatusNoContent)
	cleared := response.Result().Cookies()[0]
	if cleared.Name != second.Name || cleared.MaxAge != -1 || cleared.Domain != second.Domain {
		t.Error("logout failed to clear the selected tenant cookie")
	}
	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", second.String(), nil)
	if identityView(t, response).Account != nil {
		t.Error("revoked session still authenticates")
	}
	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", first.String(), nil)
	if identityView(t, response).Account == nil {
		t.Error("logout revoked another valid login")
	}
}
