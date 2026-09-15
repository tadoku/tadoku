package e2e_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestDatabaseUnavailable(t *testing.T) {
	reset(t)
	// Only destructive dependency tests construct another API: closing its pool
	// must not break the suite-level router used by normal HTTP scenarios.
	isolated, err := newTestAPI(t.Context())
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
		previous := jwt.TimeFunc
		jwt.TimeFunc = timex.Now
		defer func() { jwt.TimeFunc = previous }()
		timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
			checkHTTPGolden(t, isolated.authenticatedHandler, path, http.StatusForbidden)
		})
	})

	t.Run("reads fail without proxy fallback", func(t *testing.T) {
		for _, operation := range []string{"ListActiveAnnouncements", "ListAnnouncements"} {
			t.Run(operation, func(t *testing.T) {
				path := filepath.Join("testdata", APITestName(operation, http.StatusInternalServerError, "read", "failure"))
				checkHTTPGolden(t, withContractAdminIdentity(isolated.handler), path, http.StatusInternalServerError)
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
