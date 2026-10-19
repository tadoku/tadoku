package e2e_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestIdentityV2IPThrottleAndFailOpen(t *testing.T) {
	resetIdentity(t)
	for i := 0; i < 51; i++ {
		body := fmt.Sprintf(`{"email":"unknown%d@example.test","password":"Independent-653!"}`, i)
		response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login", body, "",
			func(r *http.Request) {
				r.RemoteAddr = fmt.Sprintf("10.42.0.%d:12345", i+1)
				r.Header.Set("X-Real-IP", "198.51.100.30")
			})
		want := http.StatusUnauthorized
		if i == 50 {
			want = http.StatusTooManyRequests
		}
		requireIdentityStatus(t, response, want)
	}
	for i := 0; i < 51; i++ {
		body := fmt.Sprintf(`{"email":"unknown%d@example.test","password":"Independent-653!"}`, i)
		response := identityRequest(t, api.handler, identityBranchB, "POST", "/identity/login", body, "",
			func(r *http.Request) { r.Header.Set("X-Real-IP", fmt.Sprintf("198.51.100.%d", i+1)) })
		want := http.StatusUnauthorized
		if i == 50 {
			want = http.StatusTooManyRequests
		}
		requireIdentityStatus(t, response, want)
	}

	closed, err := newClosedLeaderboardValkeyClient()
	if err != nil {
		t.Fatal(err)
	}
	handler, _, _, err := newTestRouterWithLogger(
		t.Context(), api.db.AppPool, api.db.AppPool, api.keto, api.kratos,
		slog.New(slog.NewTextHandler(io.Discard, nil)), withIdentityValkey(closed),
	)
	if err != nil {
		t.Fatal(err)
	}
	response := identityRequest(t, handler, identityBranchA, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
}

func TestIdentityV1ProviderSessionAndLogout(t *testing.T) {
	resetIdentity(t)
	if err := flipt.SetIdentityV2(tenant.Production(), false); err != nil {
		t.Fatal(err)
	}
	active, err := os.ReadFile("testdata/IdentityV1/session.json")
	if err != nil {
		t.Fatal(err)
	}
	unauthorized, err := os.ReadFile("testdata/IdentityV1/unauthorized.json")
	if err != nil {
		t.Fatal(err)
	}
	var recorded struct {
		ID       string
		Identity struct{ ID string }
	}
	if err := json.Unmarshal(active, &recorded); err != nil {
		t.Fatal(err)
	}
	const cookie = "ory_kratos_session=recorded-session"
	var revoked atomic.Bool
	var writes atomic.Int64
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/sessions/whoami":
			if r.Header.Get("Cookie") != cookie {
				t.Errorf("legacy session cookie changed")
			}
			if revoked.Load() {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write(unauthorized)
			} else {
				_, _ = w.Write(active)
			}
		case r.Method == http.MethodDelete && r.URL.Path == "/admin/sessions/"+recorded.ID:
			revoked.Store(true)
			writes.Add(1)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected legacy request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer provider.Close()
	handler, _, _, err := newTestRouterWithLogger(
		t.Context(), api.db.AppPool, api.db.AppPool, api.keto, api.kratos,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		func(e *routerEndpoints) {
			e.legacySessions = identity.NewLegacySessions(provider.URL, provider.URL, provider.Client())
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	response := identityRequest(t, handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
	view := identityView(t, response)
	if view.AccountApp != "v1" || view.Account == nil || view.Account.ID != recorded.Identity.ID ||
		view.Account.Email != "legacy@example.test" || view.Account.DisplayName != "Legacy Reader" {
		t.Fatal("v1 session did not come from recorded Kratos")
	}
	callbackRequest := httptest.NewRequest("GET", "/identity/internal/v1/session", nil)
	callbackRequest.Header.Set("Authorization", "Bearer "+callbackToken)
	callbackRequest.Header.Set("Cookie", cookie)
	callback := httptest.NewRecorder()
	atFixtureInstant(func() { handler.ServeHTTP(callback, callbackRequest) })
	requireIdentityStatus(t, callback, http.StatusOK)
	var callbackView struct {
		Identity struct{ ID string } `json:"identity"`
	}
	if err := json.Unmarshal(callback.Body.Bytes(), &callbackView); err != nil {
		t.Fatal(err)
	}
	if callbackView.Identity.ID != recorded.Identity.ID {
		t.Error("v1 callback did not resolve the recorded Kratos session")
	}

	branchKey, err := tenant.Parse(identityBranchB)
	if err != nil {
		t.Fatal(err)
	}
	if err := flipt.SetIdentityV2(branchKey, false); err != nil {
		t.Fatal(err)
	}
	response = identityRequest(t, handler, identityBranchB, "POST", "/identity/logout", "", cookie, nil)
	requireIdentityStatus(t, response, http.StatusNoContent)
	cleared := response.Result().Cookies()
	if len(cleared) != 1 || cleared[0].Name != "ory_kratos_session" || cleared[0].MaxAge != -1 {
		t.Error("v1 test-tenant logout did not clear the browser cookie")
	}
	response = identityRequest(t, handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
	if remaining := identityView(t, response); remaining.Account == nil || remaining.Account.ID != recorded.Identity.ID {
		t.Error("v1 test-tenant logout changed the shared provider session")
	}

	response = identityRequest(t, handler, identityBranchA, "GET", "/identity/session", "", cookie, nil)
	if identityView(t, response).Account != nil {
		t.Error("v2 tenant accepted a Kratos session")
	}
	response = identityRequest(t, handler, "tadoku/prod", "POST", "/identity/logout", "", cookie, nil)
	requireIdentityStatus(t, response, http.StatusNoContent)
	response = identityRequest(t, handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
	if identityView(t, response).Account != nil || writes.Load() != 1 {
		t.Error("v1 logout did not revoke exactly the selected provider session")
	}
}
