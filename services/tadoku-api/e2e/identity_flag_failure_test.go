package e2e_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

const recordedKratosCookie = "ory_kratos_session=recorded-session"

// countingLegacyProvider stands in for shared Kratos: it answers the recorded session and counts every request.
type countingLegacyProvider struct {
	calls    atomic.Int64
	identity string
	server   *httptest.Server
}

func newCountingLegacyProvider(t *testing.T) *countingLegacyProvider {
	t.Helper()
	active, err := os.ReadFile("testdata/IdentityV1/session.json")
	if err != nil {
		t.Fatal(err)
	}
	var recorded struct {
		Identity struct{ ID string }
	}
	if err := json.Unmarshal(active, &recorded); err != nil {
		t.Fatal(err)
	}

	provider := &countingLegacyProvider{identity: recorded.Identity.ID}
	provider.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provider.calls.Add(1)
		if r.Method == http.MethodGet && r.URL.Path == "/sessions/whoami" && r.Header.Get("Cookie") == recordedKratosCookie {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(active)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(provider.server.Close)
	return provider
}

func (p *countingLegacyProvider) router(t *testing.T, options ...routerOption) http.Handler {
	t.Helper()
	options = append(options, withLegacyIdentity(
		identity.NewLegacySessions(p.server.URL, p.server.URL, p.server.Client()),
		identity.NewLegacyReader(p.server.URL, identity.WithHTTPClient(p.server.Client())),
	))
	handler, _, _, err := newTestRouterWithLogger(
		t.Context(), api.db.AppPool, api.db.AppPool, api.keto,
		slog.New(slog.NewTextHandler(io.Discard, nil)), options...,
	)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func identityCallback(t *testing.T, handler http.Handler, ctx context.Context, branch, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequestWithContext(ctx, "GET", "/identity/internal/v1/session", nil)
	request.Header.Set("Authorization", "Bearer "+callbackToken)
	request.Header.Set("Cookie", cookie)
	if branch != "" {
		request.Header.Set("X-Dev-Branch", branch)
	}
	response := httptest.NewRecorder()
	atFixtureInstant(func() { handler.ServeHTTP(response, request) })
	return response
}

func callbackIdentity(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()
	requireIdentityStatus(t, response, http.StatusOK)
	var view struct {
		Identity struct{ ID string } `json:"identity"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	return view.Identity.ID
}

func requireIdentityUnavailable(t *testing.T, label string, response *httptest.ResponseRecorder) {
	t.Helper()
	if response.Code != http.StatusServiceUnavailable {
		t.Errorf("%s during flag failure: status=%d body=%s, want 503", label, response.Code, response.Body)
		return
	}
	if cookies := response.Result().Cookies(); len(cookies) != 0 {
		t.Errorf("%s during flag failure changed cookies: %v", label, cookies)
	}
}

func nativeLogin(t *testing.T, handler http.Handler, key string) string {
	t.Helper()
	response := identityRequest(t, handler, key, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"`+identityFixturePassword+`"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
	cookie := response.Result().Cookies()[0]
	return cookie.Name + "=" + cookie.Value
}

func TestIdentityFlagFailureLogoutNeverClaimsRevocation(t *testing.T) {
	resetIdentity(t)
	t.Cleanup(func() { flipt.SetUnavailable(false) })
	legacy := newCountingLegacyProvider(t)
	handler := legacy.router(t)
	branch := strings.TrimPrefix(identityBranchA, "tadoku/")
	native := nativeLogin(t, handler, identityBranchA)

	flipt.SetUnavailable(true)
	logout := identityRequest(t, handler, identityBranchA, "POST", "/identity/logout", "", native, nil)
	requireIdentityUnavailable(t, "bodyless logout", logout)
	if logout.Header().Get("Cache-Control") != "private, no-store" {
		t.Error("failed logout may be cached")
	}
	requireIdentityUnavailable(t, "session",
		identityRequest(t, handler, identityBranchA, "GET", "/identity/session", "", native, nil))
	requireIdentityUnavailable(t, "callback", identityCallback(t, handler, t.Context(), branch, native))

	flipt.SetUnavailable(false)
	survived := identityView(t, identityRequest(t, handler, identityBranchA, "GET", "/identity/session", "", native, nil))
	if logout.Code == http.StatusNoContent && survived.Account != nil {
		t.Error("logout reported success while the native session stayed valid")
	}
	if survived.AccountApp != "v2" || survived.Account == nil {
		t.Fatal("native session did not survive the refused logout")
	}

	retry := identityRequest(t, handler, identityBranchA, "POST", "/identity/logout", "", native, nil)
	requireIdentityStatus(t, retry, http.StatusNoContent)
	cleared := retry.Result().Cookies()
	if len(cleared) != 1 || cleared[0].Name != strings.SplitN(native, "=", 2)[0] || cleared[0].MaxAge != -1 {
		t.Errorf("retried logout did not clear the native cookie: %v", cleared)
	}
	if identityView(t, identityRequest(t, handler, identityBranchA, "GET", "/identity/session", "", native, nil)).Account != nil {
		t.Error("original native cookie authenticates after the retried logout")
	}
	if id := callbackIdentity(t, identityCallback(t, handler, t.Context(), branch, native)); id != "guest" {
		t.Error("callback accepted the revoked native cookie")
	}
	if calls := legacy.calls.Load(); calls != 0 {
		t.Errorf("native logout called the shared provider %d times", calls)
	}
}

func TestIdentityFlagFailureNeverSelectsSharedSessions(t *testing.T) {
	resetIdentity(t)
	t.Cleanup(func() { flipt.SetUnavailable(false) })
	legacy := newCountingLegacyProvider(t)
	handler := legacy.router(t)
	native := nativeLogin(t, handler, identityBranchA)

	flipt.SetUnavailable(true)
	for _, cookie := range []string{"", native, recordedKratosCookie, native + "; " + recordedKratosCookie} {
		for _, key := range []string{identityBranchA, "tadoku/prod"} {
			response := identityRequest(t, handler, key, "GET", "/identity/session", "", cookie, nil)
			requireIdentityUnavailable(t, key+" session", response)
			response = identityRequest(t, handler, key, "POST", "/identity/logout", "", cookie, nil)
			requireIdentityUnavailable(t, key+" logout", response)
		}
		for _, branch := range []string{"", strings.TrimPrefix(identityBranchA, "tadoku/")} {
			response := identityCallback(t, handler, t.Context(), branch, cookie)
			if response.Code == http.StatusOK && strings.Contains(response.Body.String(), legacy.identity) {
				t.Errorf("callback for branch %q admitted the shared Kratos identity", branch)
			}
			requireIdentityUnavailable(t, "callback", response)
		}
	}
	if calls := legacy.calls.Load(); calls != 0 {
		t.Errorf("flag failure reached the shared provider %d times", calls)
	}

	flipt.SetUnavailable(false)
	if err := flipt.SetIdentityV2(tenant.Production(), false); err != nil {
		t.Fatal(err)
	}
	view := identityView(t, identityRequest(t, handler, "tadoku/prod", "GET", "/identity/session", "", recordedKratosCookie, nil))
	if view.AccountApp != "v1" || view.Account == nil || view.Account.ID != legacy.identity {
		t.Error("explicit v1 decision no longer selects shared Kratos")
	}
	view = identityView(t, identityRequest(t, handler, identityBranchA, "GET", "/identity/session", "", recordedKratosCookie, nil))
	if view.AccountApp != "v2" || view.Account != nil {
		t.Error("v2 tenant accepted a shared Kratos session after recovery")
	}
	if calls := legacy.calls.Load(); calls != 1 {
		t.Errorf("recovered v1 and v2 lookups called the shared provider %d times, want 1", calls)
	}
}

func TestIdentityFlagFailureNeverReadsSharedAccounts(t *testing.T) {
	seedIdentityReads(t)
	t.Cleanup(func() { flipt.SetUnavailable(false) })
	legacy := newCountingLegacyProvider(t)
	handler := legacy.router(t)
	profile := readHTTPRequest(t, filepath.Join("testdata", "ImmersionProfileFindByUserID", "200_user_populated"))
	admin := readHTTPRequest(t, filepath.Join("testdata", "ProfileUsersList", "200_admin"))

	flipt.SetUnavailable(true)
	for _, test := range []struct {
		name, method, path, body string
		header                   http.Header
	}{
		{"profile", http.MethodGet, profile.URL.String(), "", profile.Header},
		{"admin list", http.MethodGet, "/profile/users", "", admin.Header},
		{
			"ban target",
			http.MethodPut,
			"/authz/users/55555555-5555-4555-8555-555555555555/role",
			`{"role":"banned","reason":"target lookup during flag failure"}`,
			admin.Header,
		},
	} {
		request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		request.Header = test.header.Clone()
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		atFixtureInstant(func() { handler.ServeHTTP(response, request) })
		if response.Code != http.StatusServiceUnavailable {
			t.Errorf("%s during flag failure: status=%d body=%s, want 503", test.name, response.Code, response.Body)
		}
	}
	if calls := legacy.calls.Load(); calls != 0 {
		t.Errorf("reader flag failure reached the shared provider %d times", calls)
	}
}

// flagProviderFunc stands in for Flipt when a test controls one evaluation outcome.
type flagProviderFunc func(context.Context, featureflags.EvaluationRequest) (featureflags.ProviderResult, error)

func (f flagProviderFunc) EvaluateBoolean(
	ctx context.Context,
	request featureflags.EvaluationRequest,
) (featureflags.ProviderResult, error) {
	return f(ctx, request)
}

func TestIdentityFlagFailureStatusMapping(t *testing.T) {
	resetIdentity(t)
	legacy := newCountingLegacyProvider(t)
	cookie := recordedKratosCookie + "; " + nativeLogin(t, api.handler, "tadoku/prod")

	invalid := legacy.router(t, withFlagProvider(flagProviderFunc(
		func(context.Context, featureflags.EvaluationRequest) (featureflags.ProviderResult, error) {
			return featureflags.ProviderResult{}, featureflags.ErrInvalidResponse
		},
	)))
	requireIdentityUnavailable(t, "invalid response callback", identityCallback(t, invalid, t.Context(), "", cookie))

	var cancel context.CancelFunc
	canceling := legacy.router(t, withFlagProvider(flagProviderFunc(
		func(ctx context.Context, _ featureflags.EvaluationRequest) (featureflags.ProviderResult, error) {
			cancel()
			return featureflags.ProviderResult{}, ctx.Err()
		},
	)))
	ctx, cancel := context.WithCancel(t.Context())
	if response := identityCallback(t, canceling, ctx, "", cookie); response.Code != 499 {
		t.Errorf("canceled callback status=%d body=%s, want 499", response.Code, response.Body)
	}

	waiting := legacy.router(t, withFlagProvider(flagProviderFunc(
		func(ctx context.Context, _ featureflags.EvaluationRequest) (featureflags.ProviderResult, error) {
			<-ctx.Done()
			return featureflags.ProviderResult{}, ctx.Err()
		},
	)))
	ctx, stop := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer stop()
	if response := identityCallback(t, waiting, ctx, "", cookie); response.Code != http.StatusGatewayTimeout {
		t.Errorf("timed out callback status=%d body=%s, want 504", response.Code, response.Body)
	}
	if calls := legacy.calls.Load(); calls != 0 {
		t.Errorf("failed decisions reached the shared provider %d times", calls)
	}
}
