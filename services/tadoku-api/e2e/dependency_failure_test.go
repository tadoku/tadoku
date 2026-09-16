package e2e_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/testketo"
)

func TestDatabaseUnavailable(t *testing.T) {
	reset(t)
	// Only destructive dependency tests construct another API: closing its pool
	// must not break the suite-level router used by normal HTTP scenarios.
	isolated, err := newTestAPI(t.Context(), keto.ReadURL())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := isolated.db.Close(); err != nil {
			t.Error(err)
		}
	})
	isolated.db.Pool.Close()

	t.Run("permission denial precedes database access", func(t *testing.T) {
		path := filepath.Join("testdata", APITestName("ListAnnouncements", http.StatusForbidden, "non", "admin"))
		checkCaseGolden(t, isolated.handler, path, http.StatusForbidden)
	})

	t.Run("reads fail without proxy fallback", func(t *testing.T) {
		for _, operation := range []string{"ListActiveAnnouncements", "ListAnnouncements"} {
			t.Run(operation, func(t *testing.T) {
				path := filepath.Join("testdata", APITestName(operation, http.StatusInternalServerError, "read", "failure"))
				checkCaseGolden(t, isolated.handler, path, http.StatusInternalServerError)
			})
		}
	})
	if isolated.proxied.Load() != 0 {
		t.Error("failed request fell back to the proxy")
	}

	t.Run("liveness survives but readiness fails", func(t *testing.T) {
		for path, status := range map[string]int{
			"/livez":  http.StatusOK,
			"/readyz": http.StatusServiceUnavailable,
		} {
			response := httptest.NewRecorder()
			isolated.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != status {
				t.Errorf("%s status=%d, want %d", path, response.Code, status)
			}
		}
	})
}

func TestListAnnouncementsRejectsFailedBanLookup(t *testing.T) {
	path := filepath.Join("testdata", APITestName("ListAnnouncements", http.StatusServiceUnavailable, "failed", "ban", "lookup"))
	resetCase(t, path)
	// Stop a separate real Keto process so the production middleware records
	// the provider failure. Never stop the shared suite's Keto dependency.
	provider, err := testketo.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := provider.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := provider.Reset(t.Context(), filepath.Join(path, "relationships.json")); err != nil {
		t.Fatal(err)
	}
	isolated, err := newTestAPI(t.Context(), provider.ReadURL())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := isolated.db.Close(); err != nil {
			t.Error(err)
		}
	})
	legacy, err := newLegacyContentAPI(t.Context(), isolated.db.DSN, authenticationJWKS.URL, provider.ReadURL())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := legacy.db.Close(); err != nil {
			t.Error(err)
		}
	})
	// First prove the seeded admin can access both APIs with this provider.
	adminPath := filepath.Join("testdata", APITestName("ListAnnouncements", http.StatusOK, "admin"))
	checkHTTPGolden(t, isolated.handler, adminPath, http.StatusOK)
	checkHTTPGolden(t, legacy.handler, adminPath, http.StatusOK)
	if err := provider.Close(); err != nil {
		t.Fatal(err)
	}
	for _, implementation := range []struct {
		name    string
		handler http.Handler
	}{
		{name: "tadoku-api", handler: isolated.handler},
		{name: "content-api", handler: legacy.handler},
	} {
		t.Run(implementation.name, func(t *testing.T) {
			if err := isolated.db.Reset(t.Context()); err != nil {
				t.Fatal(err)
			}
			checkHTTPGolden(t, implementation.handler, path, http.StatusServiceUnavailable)
		})
	}
	if isolated.proxied.Load() != 0 {
		t.Error("failed permission check fell back to the proxy")
	}
}

func TestCanceledReadDoesNotFallBack(t *testing.T) {
	path := filepath.Join("testdata", APITestName("ListActiveAnnouncements", http.StatusInternalServerError, "read", "failure"))
	resetCase(t, path)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		cancel()
		api.handler.ServeHTTP(w, r.WithContext(ctx))
	})
	checkHTTPGolden(t, handler, path, http.StatusInternalServerError)
	if api.proxied.Load() != 0 {
		t.Error("canceled read fell back to the proxy")
	}
}
