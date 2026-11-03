package e2e_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

const registrationAdmissionBudget = 10

func registerFrom(
	t *testing.T,
	handler http.Handler,
	tenantKey, email, peer, forwarded string,
) *httptest.ResponseRecorder {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"display_name":"Budget Reader","password":"Independent-653!"}`, email)
	return identityRequest(t, handler, tenantKey, "POST", "/identity/registration", body, "", func(r *http.Request) {
		r.RemoteAddr = peer + ":12345"
		if forwarded != "" {
			r.Header.Set("X-Real-IP", forwarded)
		}
	})
}

func requireRegistrationThrottled(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	requireIdentityStatus(t, response, http.StatusTooManyRequests)
	var failure struct{ Code string }
	if err := json.Unmarshal(response.Body.Bytes(), &failure); err != nil {
		t.Fatal(err)
	}
	if failure.Code != "throttled" || response.Header().Get("Set-Cookie") != "" {
		t.Errorf("rejected registration returned code %q or issued a session", failure.Code)
	}
}

func identityUserCount(t *testing.T, tenantKey string) int {
	t.Helper()
	var count int
	err := api.db.Pool.QueryRow(t.Context(), "select count(*) from users where tenant = $1", tenantKey).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func TestIdentityV2RegistrationAdmissionBudget(t *testing.T) {
	t.Parallel()
	s := newPairedIdentityScenario(t)
	branchA := s.tenant(t, "identity-a").String()

	const peer = "198.51.100.20"

	for i := 0; i < 10; i++ {
		response := s.identityRequest(t, api.handler, "identity-a", "POST", "/identity/login",
			`{"email":"reader@example.test","password":"wrong-password"}`, "", nil)
		requireIdentityStatus(t, response, http.StatusUnauthorized)
	}
	for i := 0; i < 3; i++ {
		response := s.identityRequest(t, api.handler, "identity-a", "POST", "/identity/registration",
			`{"email":"invalid@example.test","display_name":"Budget Reader","password":"short"}`, "", nil)
		requireIdentityStatus(t, response, http.StatusBadRequest)
	}

	for i := 0; i < registrationAdmissionBudget/2; i++ {
		response := s.registerFrom(t, api.handler, "identity-a", "READER@example.test", peer, "")
		requireIdentityStatus(t, response, http.StatusConflict)
	}
	for i := 0; i < registrationAdmissionBudget/2; i++ {
		response := s.registerFrom(t, api.handler, "identity-a", fmt.Sprintf("budget%d@example.test", i), peer, "")
		requireIdentityStatus(t, response, http.StatusCreated)
	}
	before := identityUserCount(t, branchA)

	requireRegistrationThrottled(t, s.registerFrom(t, api.handler, "identity-a", "unique@example.test", peer, ""))
	requireRegistrationThrottled(t, s.registerFrom(t, api.handler, "identity-a", "reader@example.test", peer, ""))
	requireRegistrationThrottled(t,
		s.registerFrom(t, api.handler, "identity-a", "spoofed@example.test", peer, "198.51.100.77"))
	requireRegistrationThrottled(t,
		s.registerFrom(t, api.handler, "identity-a", "proxied@example.test", "10.42.0.9", peer))
	if after := identityUserCount(t, branchA); after != before {
		t.Errorf("throttled registrations changed users: before=%d after=%d", before, after)
	}

	response := s.identityRequest(t, api.handler, "identity-a", "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusTooManyRequests)
	response = s.identityRequest(t, api.handler, "identity-a", "POST", "/identity/login",
		`{"email":"budget0@example.test","password":"Independent-653!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)

	response = s.registerFrom(t, api.handler, "identity-b", "unique@example.test", peer, "")
	requireIdentityStatus(t, response, http.StatusCreated)
	response = s.registerFrom(t, api.handler, "identity-a", "other-peer@example.test", "203.0.113.5", peer)
	requireIdentityStatus(t, response, http.StatusCreated)
	response = s.registerFrom(t, api.handler, "identity-a", "forwarded@example.test", "10.42.0.9", "198.51.100.41")
	requireIdentityStatus(t, response, http.StatusCreated)
}

func TestIdentityV2RegistrationAdmissionUnavailable(t *testing.T) {
	t.Parallel()
	s := newIdentityScenario(t)
	branchA := s.tenant(t, "identity-a").String()

	closed, err := newClosedLeaderboardValkeyClient()
	if err != nil {
		t.Fatal(err)
	}
	handler, _, err := newTestRouterWithLogger(
		t.Context(), api.db.AppPool, api.db.AppPool,
		slog.New(slog.NewTextHandler(io.Discard, nil)), withIdentityValkey(closed),
	)
	if err != nil {
		t.Fatal(err)
	}
	before := identityUserCount(t, branchA)

	response := s.registerFrom(t, handler, "identity-a", "unavailable@example.test", "198.51.100.20", "")
	requireIdentityStatus(t, response, http.StatusServiceUnavailable)
	if response.Header().Get("Set-Cookie") != "" || identityUserCount(t, branchA) != before {
		t.Error("registration without admission issued a session or created a user")
	}
	response = s.registerFrom(t, handler, "identity-a", "reader@example.test", "198.51.100.20", "")
	requireIdentityStatus(t, response, http.StatusServiceUnavailable)

	response = s.identityRequest(t, handler, "identity-a", "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
}

func TestIdentityLoginAndRegistrationThrottleWindows(t *testing.T) {
	t.Parallel()
	s := newIdentityScenario(t)
	branchA := s.tenant(t, "identity-a").String()

	const peer = "198.51.100.25"
	login := func() {
		response := s.identityRequest(t, api.handler, "identity-a", "POST", "/identity/login",
			`{"email":"reader@example.test","password":"wrong-password"}`, "", nil)
		requireIdentityStatus(t, response, http.StatusUnauthorized)
	}

	login()
	requireIdentityStatus(t, s.registerFrom(t, api.handler, "identity-a", "window@example.test", peer, ""),
		http.StatusCreated)
	first := map[string]int64{}
	for operation, counters := range map[string]int{"login": 2, "registration": 1} {
		keys := identityThrottleKeys(t, branchA, operation)
		if len(keys) != counters {
			t.Fatalf("%s throttle keys = %v, want %d counters", operation, keys, counters)
		}
		for key, ttl := range keys {
			if ttl <= 890_000 || ttl > 900_000 {
				t.Errorf("%s throttle key %q has window %dms, want 900000ms", operation, key, ttl)
			}
			first[key] = ttl
		}
	}

	login()
	requireIdentityStatus(t, s.registerFrom(t, api.handler, "identity-a", "window2@example.test", peer, ""),
		http.StatusCreated)
	for _, operation := range []string{"login", "registration"} {
		for key, ttl := range identityThrottleKeys(t, branchA, operation) {
			if ttl > first[key] {
				t.Errorf("later %s attempt extended %q from %dms to %dms", operation, key, first[key], ttl)
			}
		}
	}
}
