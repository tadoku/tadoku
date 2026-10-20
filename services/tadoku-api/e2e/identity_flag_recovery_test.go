package e2e_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	fliptclient "github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	fliptsdk "go.flipt.io/flipt-client"
)

// snapshotGate is a cookie jar that holds the first snapshot response after the Flipt client's transport has
// received it and before the SDK can apply it. Call release to let the response through.
type snapshotGate struct {
	once    sync.Once
	held    chan struct{}
	opened  chan struct{}
	release func()
}

func newSnapshotGate() *snapshotGate {
	opened := make(chan struct{})
	return &snapshotGate{
		held:    make(chan struct{}),
		opened:  opened,
		release: sync.OnceFunc(func() { close(opened) }),
	}
}

func (g *snapshotGate) Cookies(*url.URL) []*http.Cookie { return nil }

func (g *snapshotGate) SetCookies(*url.URL, []*http.Cookie) {
	g.once.Do(func() {
		close(g.held)
		<-g.opened
	})
}

func TestIdentityFliptCanonicalRecoveryWaitsForAppliedSnapshot(t *testing.T) {
	seedIdentityReads(t)
	accounts, err := os.ReadFile("testdata/IdentityV2/setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := api.db.Pool.Exec(t.Context(), string(accounts)); err != nil {
		t.Fatal(err)
	}
	native := nativeLogin(t, api.handler, "tadoku/prod")
	legacy := newCountingLegacyProvider(t)

	flags := newSnapshotServer(t, map[string]identityFlag{"default": identityFlagTrue})
	flags.setOutage(true)
	gate := newSnapshotGate()
	handler := legacy.router(t, withFlagProvider(
		newSnapshotProvider(t, flags.server.URL, 2*time.Second, &http.Client{Jar: gate}),
	))
	t.Cleanup(gate.release)
	requireIdentityUnavailable(t, "canonical session before the first snapshot",
		identityRequest(t, handler, "tadoku/prod", "GET", "/identity/session", "", native+"; "+recordedKratosCookie, nil))

	flags.setOutage(false)
	select {
	case <-gate.held:
	case <-time.After(5 * time.Second):
		t.Fatal("the Flipt client never received the recovered snapshot")
	}
	requireCanonicalIdentityUnavailable(t, handler, native, legacy, "before the snapshot is applied")

	gate.release()
	requireCanonicalRecoveryRevokes(t, handler, native, legacy)
}

// requireCanonicalIdentityUnavailable asserts that every canonical identity path returns 503, without a cookie change
// or a shared-provider call, for the native cookie, the recorded Kratos cookie and both.
func requireCanonicalIdentityUnavailable(
	t *testing.T,
	handler http.Handler,
	native string,
	legacy *countingLegacyProvider,
	when string,
) {
	t.Helper()
	for _, cookie := range []string{native, recordedKratosCookie, native + "; " + recordedKratosCookie} {
		requireIdentityUnavailable(t, "session "+when,
			identityRequest(t, handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil))
		requireIdentityUnavailable(t, "callback "+when,
			identityCallback(t, handler, t.Context(), "", cookie))
	}
	requireIdentityUnavailable(t, "native logout "+when,
		identityRequest(t, handler, "tadoku/prod", "POST", "/identity/logout", "", native, nil))
	for _, fixture := range []string{"ImmersionProfileFindByUserID/200_user_populated", "ProfileUsersList/200_admin"} {
		request := readHTTPRequest(t, filepath.Join("testdata", fixture))
		response := httptest.NewRecorder()
		atFixtureInstant(func() { handler.ServeHTTP(response, request) })
		if response.Code != http.StatusServiceUnavailable {
			t.Errorf("%s %s: status=%d body=%s, want 503", fixture, when, response.Code, response.Body)
		}
	}
	if calls := legacy.calls.Load(); calls != 0 {
		t.Fatalf("canonical identity %s selected the shared provider %d times", when, calls)
	}
}

// requireCanonicalRecoveryRevokes waits for the canonical v2 decision, then asserts that logout revokes the native
// session and that no request reached the shared provider.
func requireCanonicalRecoveryRevokes(t *testing.T, handler http.Handler, native string, legacy *countingLegacyProvider) {
	t.Helper()
	eventually(t, "the applied v2 decision", func() bool {
		app, signedIn := sessionAccountApp(t, handler, "tadoku/prod", native)
		return app == "v2" && signedIn
	})
	logout := identityRequest(t, handler, "tadoku/prod", "POST", "/identity/logout", "", native, nil)
	requireIdentityStatus(t, logout, http.StatusNoContent)
	if cleared := logout.Result().Cookies(); len(cleared) != 1 || cleared[0].Name != strings.SplitN(native, "=", 2)[0] {
		t.Errorf("logout after recovery did not clear the native cookie: %v", cleared)
	}
	if _, signedIn := sessionAccountApp(t, handler, "tadoku/prod", native); signedIn {
		t.Error("native session survived logout after recovery")
	}
	if id := callbackIdentity(t, identityCallback(t, handler, t.Context(), "", native)); id != "guest" {
		t.Error("callback accepted the revoked native cookie")
	}
	if calls := legacy.calls.Load(); calls != 0 {
		t.Errorf("canonical recovery called the shared provider %d times", calls)
	}
}

type initializationRecorder struct {
	ready atomic.Bool
}

func (r *initializationRecorder) ObserveInitialization(status featureflags.InitializationStatus) {
	if status == featureflags.InitializationStatusReady {
		r.ready.Store(true)
	}
}

func (r *initializationRecorder) ObserveConfigRefresh() {}

func (r *initializationRecorder) ObserveProviderError(string) {}

func TestIdentityFliptCanonicalFailedStartIsNeverInstalled(t *testing.T) {
	seedIdentityReads(t)
	accounts, err := os.ReadFile("testdata/IdentityV2/setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := api.db.Pool.Exec(t.Context(), string(accounts)); err != nil {
		t.Fatal(err)
	}
	native := nativeLogin(t, api.handler, "tadoku/prod")
	legacy := newCountingLegacyProvider(t)

	flags := newSnapshotServer(t, map[string]identityFlag{"default": identityFlagTrue})
	flags.setOutage(true)
	var opens atomic.Int32
	held := make(chan *fliptsdk.Client, 1)
	released := make(chan struct{})
	release := sync.OnceFunc(func() { close(released) })
	t.Cleanup(release)
	targets, err := fliptclient.NewTargets("local", "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	observer := &initializationRecorder{}
	provider, err := fliptclient.NewProvider(t.Context(), fliptclient.Config{
		URL:            flags.server.URL,
		UpdateInterval: time.Second,
		RequestTimeout: time.Second,
		StartupTimeout: 200 * time.Millisecond,
		OpenSDK: func(ctx context.Context, options ...fliptsdk.Option) (*fliptsdk.Client, error) {
			sdk, err := fliptsdk.NewClient(ctx, options...)
			if opens.Add(1) == 1 && err == nil {
				held <- sdk
				<-released
			}
			return sdk, err
		},
	}, targets, observer)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = provider.Close(context.Background()) })
	handler := legacy.router(t, withFlagProvider(provider))

	var first *fliptsdk.Client
	select {
	case first = <-held:
	case <-time.After(5 * time.Second):
		t.Fatal("the canonical SDK client was never constructed")
	}
	if first.Err() == nil {
		t.Fatal("the canonical SDK client did not start from a failed fetch")
	}
	flags.setRejecting(true)
	flags.setOutage(false)
	eventually(t, "a rejected snapshot to clear the failed start's error", func() bool {
		return flags.rejected() && first.Err() == nil
	})
	release()
	eventually(t, "the client to settle the failed-start SDK", func() bool {
		return observer.ready.Load() || opens.Load() > 1
	})
	requireCanonicalIdentityUnavailable(t, handler, native, legacy, "after a failed start cleared its error")

	flags.setRejecting(false)
	requireCanonicalRecoveryRevokes(t, handler, native, legacy)
}

func TestIdentityFliptCanonicalKeepsItsDecisionWhenAnUpdateFailsValidation(t *testing.T) {
	seedIdentityReads(t)
	accounts, err := os.ReadFile("testdata/IdentityV2/setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := api.db.Pool.Exec(t.Context(), string(accounts)); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct{ name, body string }{
		{"another namespace", `{"namespace":{"key":"other"},"flags":[],"digest":"rev-2"}`},
		{"no digest", `{"namespace":{"key":"default"},"flags":[]}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			legacy := newCountingLegacyProvider(t)
			flags := newSnapshotServer(t, map[string]identityFlag{"default": identityFlagTrue})
			handler := legacy.router(t, withFlagProvider(newSnapshotProvider(t, flags.server.URL, 2*time.Second, nil)))
			native := nativeLogin(t, handler, "tadoku/prod")
			if app, signedIn := sessionAccountApp(t, handler, "tadoku/prod", native); app != "v2" || !signedIn {
				t.Fatalf("warm canonical decision: app=%q signed-in=%t, want a v2 session", app, signedIn)
			}

			flags.setReplacement(test.body)
			flags.awaitReplacementHandled(t)
			view := identityView(t,
				identityRequest(t, handler, "tadoku/prod", "GET", "/identity/session", "", native+"; "+recordedKratosCookie, nil))
			if view.AccountApp != "v2" || view.Account == nil || view.Account.ID == legacy.identity {
				t.Errorf("session after a rejected update: app=%q account=%v, want the native v2 account",
					view.AccountApp, view.Account)
			}
			if id := callbackIdentity(t, identityCallback(t, handler, t.Context(), "", recordedKratosCookie)); id != "guest" {
				t.Errorf("callback after a rejected update admitted %q from the Kratos cookie", id)
			}
			for _, fixture := range []string{"ImmersionProfileFindByUserID/200_user_populated", "ProfileUsersList/200_admin"} {
				request := readHTTPRequest(t, filepath.Join("testdata", fixture))
				response := httptest.NewRecorder()
				atFixtureInstant(func() { handler.ServeHTTP(response, request) })
				if response.Code != http.StatusOK {
					t.Errorf("%s after a rejected update: status=%d body=%s, want 200", fixture, response.Code, response.Body)
				}
			}
			logout := identityRequest(t, handler, "tadoku/prod", "POST", "/identity/logout", "", native, nil)
			cleared := logout.Result().Cookies()
			if logout.Code != http.StatusNoContent || len(cleared) != 1 || cleared[0].Name != strings.SplitN(native, "=", 2)[0] {
				t.Errorf("logout after a rejected update: status=%d cleared=%v, want 204 clearing the native cookie",
					logout.Code, cleared)
			}

			flags.setReplacement("")
			eventually(t, "the restored v2 decision", func() bool {
				app, _ := sessionAccountApp(t, handler, "tadoku/prod", native)
				return app == "v2"
			})
			if _, signedIn := sessionAccountApp(t, handler, "tadoku/prod", native); signedIn {
				t.Error("logout after a rejected update reported success while the native session stayed valid")
			}
			if id := callbackIdentity(t, identityCallback(t, handler, t.Context(), "", native)); id != "guest" {
				t.Errorf("callback accepted the logged-out native cookie as %q", id)
			}
			if calls := legacy.calls.Load(); calls != 0 {
				t.Errorf("canonical identity after a rejected update called the shared provider %d times", calls)
			}
		})
	}
}
