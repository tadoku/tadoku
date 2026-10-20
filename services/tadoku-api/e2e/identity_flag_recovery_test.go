package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
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
	for _, cookie := range []string{native, recordedKratosCookie, native + "; " + recordedKratosCookie} {
		requireIdentityUnavailable(t, "session before the snapshot is applied",
			identityRequest(t, handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil))
		requireIdentityUnavailable(t, "callback before the snapshot is applied",
			identityCallback(t, handler, t.Context(), "", cookie))
	}
	requireIdentityUnavailable(t, "native logout before the snapshot is applied",
		identityRequest(t, handler, "tadoku/prod", "POST", "/identity/logout", "", native, nil))
	for _, fixture := range []string{"ImmersionProfileFindByUserID/200_user_populated", "ProfileUsersList/200_admin"} {
		request := readHTTPRequest(t, filepath.Join("testdata", fixture))
		response := httptest.NewRecorder()
		atFixtureInstant(func() { handler.ServeHTTP(response, request) })
		if response.Code != http.StatusServiceUnavailable {
			t.Errorf("%s before the snapshot is applied: status=%d body=%s, want 503", fixture, response.Code, response.Body)
		}
	}
	if calls := legacy.calls.Load(); calls != 0 {
		t.Fatalf("an unapplied snapshot selected the shared provider %d times", calls)
	}

	gate.release()
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
