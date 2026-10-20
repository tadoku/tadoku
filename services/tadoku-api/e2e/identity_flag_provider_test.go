package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	fliptclient "github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
)

const (
	identityNamespaceA = "tadoku_identity-a-0123abcd"
	identityNamespaceB = "tadoku_identity-b-0123abcd"
)

type identityFlag uint8

const (
	identityFlagAbsent identityFlag = iota
	identityFlagFalse
	identityFlagTrue
)

// snapshotServer serves Flipt evaluation snapshots, so the real Flipt client and provider evaluate them.
type snapshotServer struct {
	mu         sync.Mutex
	namespaces map[string]identityFlag
	outage     bool
	misses     map[string]int
	server     *httptest.Server
}

func newSnapshotServer(t *testing.T, namespaces map[string]identityFlag) *snapshotServer {
	t.Helper()
	s := &snapshotServer{namespaces: namespaces, misses: make(map[string]int)}
	s.server = httptest.NewServer(http.HandlerFunc(s.serveHTTP))
	t.Cleanup(s.server.Close)
	return s
}

func (s *snapshotServer) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	namespace, ok := strings.CutPrefix(r.URL.Path, "/internal/v1/evaluation/snapshot/namespace/")
	if s.outage {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
		return
	}
	flag, exists := s.namespaces[namespace]
	if !ok || !exists {
		s.misses[namespace]++
		http.NotFound(w, r)
		return
	}

	flags := []map[string]any{snapshotFlag("release-log-entry-v2", false)}
	if flag != identityFlagAbsent {
		flags = append(flags, snapshotFlag("identity-v2", flag == identityFlagTrue))
	}
	encoded, _ := json.Marshal(flags)
	digest := sha256.Sum256(encoded)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"namespace": map[string]string{"key": namespace},
		"flags":     flags,
		"digest":    hex.EncodeToString(digest[:20]),
	})
}

func snapshotFlag(key string, enabled bool) map[string]any {
	return map[string]any{
		"key":         key,
		"name":        key,
		"description": "",
		"enabled":     enabled,
		"type":        "BOOLEAN_FLAG_TYPE",
		"createdAt":   "2026-08-24T06:01:42Z",
		"updatedAt":   "2026-08-24T06:01:42Z",
		"rules":       []any{},
		"rollouts":    []any{},
	}
}

func (s *snapshotServer) set(namespace string, flag identityFlag) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.namespaces[namespace] = flag
}

func (s *snapshotServer) remove(namespace string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.namespaces, namespace)
	s.misses[namespace] = 0
}

func (s *snapshotServer) missed(namespace string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.misses[namespace] > 0
}

func (s *snapshotServer) setOutage(outage bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outage = outage
}

func newSnapshotProvider(t *testing.T, url string, startup time.Duration) *fliptclient.Provider {
	t.Helper()
	targets, err := fliptclient.NewTargets("local", "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	provider, err := fliptclient.NewProvider(t.Context(), fliptclient.Config{
		URL:            url,
		UpdateInterval: time.Second,
		RequestTimeout: time.Second,
		StartupTimeout: startup,
	}, targets, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = provider.Close(context.Background()) })
	return provider
}

func eventually(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func sessionAccountApp(t *testing.T, handler http.Handler, key, cookie string) (string, bool) {
	t.Helper()
	response := identityRequest(t, handler, key, "GET", "/identity/session", "", cookie, nil)
	if response.Code != http.StatusOK {
		return "", false
	}
	view := identityView(t, response)
	return view.AccountApp, view.Account != nil
}

func TestIdentityFliptProviderSelection(t *testing.T) {
	resetIdentity(t)
	legacy := newCountingLegacyProvider(t)
	flags := newSnapshotServer(t, map[string]identityFlag{
		"default":          identityFlagAbsent,
		identityNamespaceA: identityFlagTrue,
		identityNamespaceB: identityFlagAbsent,
	})
	warm := legacy.router(t, withFlagProvider(newSnapshotProvider(t, flags.server.URL, 2*time.Second)))

	view := identityView(t, identityRequest(t, warm, "tadoku/prod", "GET", "/identity/session", "", recordedKratosCookie, nil))
	if view.AccountApp != "v1" || view.Account == nil || view.Account.ID != legacy.identity {
		t.Error("canonical tenant without an identity-v2 flag no longer uses v1")
	}
	requireIdentityUnavailable(t, "branch without identity-v2",
		identityRequest(t, warm, identityBranchB, "GET", "/identity/session", "", recordedKratosCookie, nil))
	native := nativeLogin(t, warm, identityBranchA)

	flags.remove(identityNamespaceA)
	eventually(t, "the branch namespace to disappear", func() bool { return flags.missed(identityNamespaceA) })
	if app, signedIn := sessionAccountApp(t, warm, identityBranchA, native+"; "+recordedKratosCookie); app != "v2" || !signedIn {
		t.Errorf("warm stale decision lost the native session: app=%q signed-in=%t", app, signedIn)
	}

	cold := legacy.router(t, withFlagProvider(newSnapshotProvider(t, flags.server.URL, 2*time.Second)))
	requireIdentityUnavailable(t, "cold missing namespace session",
		identityRequest(t, cold, identityBranchA, "GET", "/identity/session", "", native+"; "+recordedKratosCookie, nil))
	requireIdentityUnavailable(t, "cold missing namespace logout",
		identityRequest(t, cold, identityBranchA, "POST", "/identity/logout", "", native, nil))

	flags.set(identityNamespaceA, identityFlagFalse)
	eventually(t, "the explicit v1 decision", func() bool {
		app, _ := sessionAccountApp(t, warm, identityBranchA, native)
		return app == "v1"
	})
	flags.set(identityNamespaceA, identityFlagTrue)
	eventually(t, "the explicit v2 decision", func() bool {
		app, signedIn := sessionAccountApp(t, warm, identityBranchA, native)
		return app == "v2" && signedIn
	})

	flags.setOutage(true)
	outage := legacy.router(t, withFlagProvider(newSnapshotProvider(t, flags.server.URL, 2*time.Second)))
	requireIdentityUnavailable(t, "cold outage canonical session",
		identityRequest(t, outage, "tadoku/prod", "GET", "/identity/session", "", recordedKratosCookie, nil))
	requireIdentityUnavailable(t, "cold outage branch session",
		identityRequest(t, outage, identityBranchA, "GET", "/identity/session", "", native, nil))
	if app, signedIn := sessionAccountApp(t, warm, identityBranchA, native); app != "v2" || !signedIn {
		t.Errorf("warm decision during the outage: app=%q signed-in=%t", app, signedIn)
	}
	flags.setOutage(false)

	release := make(chan struct{})
	silent := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	t.Cleanup(silent.Close)
	t.Cleanup(func() { close(release) })
	uninitialized := legacy.router(t, withFlagProvider(newSnapshotProvider(t, silent.URL, 10*time.Millisecond)))
	requireIdentityUnavailable(t, "uninitialized provider callback",
		identityCallback(t, uninitialized, t.Context(), "", recordedKratosCookie))

	logout := identityRequest(t, warm, identityBranchA, "POST", "/identity/logout", "", native, nil)
	requireIdentityStatus(t, logout, http.StatusNoContent)
	if _, signedIn := sessionAccountApp(t, warm, identityBranchA, native); signedIn {
		t.Error("native logout through the recovered provider left the session valid")
	}
	if calls := legacy.calls.Load(); calls != 1 {
		t.Errorf("shared provider calls=%d, want only the canonical v1 lookup", calls)
	}
}
