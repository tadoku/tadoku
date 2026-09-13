package e2e_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	transport "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

func TestRouterWorksWithoutLegacyProxyRoutes(t *testing.T) {
	// Construct the same application router but do not attach legacy routes.
	// Normal scenarios continue to use the single suite-level router.
	repository := content.NewRepository(api.db.Pool)
	service := content.NewService(repository)
	application := app.New(service)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	handler, err := transport.NewHandler(application, api.db.Pool.Ping, time.Second, logger)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join("testdata", "list_active_announcements", "200_plain")
	reset(t, filepath.Join(path, "setup.sql"))
	timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
		checkHTTPGolden(t, handler, path)
	})

	for _, test := range []struct {
		name   string
		method string
		path   string
		status int
	}{
		{
			name:   "liveness does not need a proxy",
			method: http.MethodGet,
			path:   "/livez",
			status: http.StatusOK,
		},
		{
			name:   "readiness does not need a proxy",
			method: http.MethodGet,
			path:   "/readyz",
			status: http.StatusOK,
		},
		{
			name:   "unimplemented paths are not found",
			method: http.MethodGet,
			path:   "/content/pages/example",
			status: http.StatusNotFound,
		},
		{
			name:   "unsupported methods are rejected",
			method: http.MethodPost,
			path:   "/content/announcements/main/active",
			status: http.StatusMethodNotAllowed,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
			if response.Code != test.status {
				t.Errorf("status=%d, want %d", response.Code, test.status)
			}
		})
	}
}

func TestReadFailureDoesNotFallBack(t *testing.T) {
	// Closing the pool must not destroy the shared suite's database dependency.
	api, err := newTestAPI(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := api.db.Close(); err != nil {
			t.Error(err)
		}
	})
	api.db.Pool.Close()

	request := httptest.NewRequest(http.MethodGet, "/content/announcements/main/active", nil)
	response := httptest.NewRecorder()
	api.handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError || response.Body.Len() != 0 {
		t.Errorf("database failure: status=%d body=%s", response.Code, response.Body)
	}
	if api.proxied.Load() != 0 {
		t.Error("failed native read fell back to the proxy")
	}

	for path, status := range map[string]int{
		"/livez":  http.StatusOK,
		"/readyz": http.StatusServiceUnavailable,
	} {
		response := httptest.NewRecorder()
		api.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != status {
			t.Errorf("%s status=%d, want %d", path, response.Code, status)
		}
	}
}

func TestUnclaimedMethodsRemainProxied(t *testing.T) {
	for _, method := range []string{http.MethodHead, http.MethodOptions, http.MethodPost, http.MethodDelete, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			reset(t)
			request := httptest.NewRequest(method, "/content/announcements/main/active", nil)
			response := httptest.NewRecorder()
			api.handler.ServeHTTP(response, request)

			if response.Header().Get("X-Proxied") != "yes" {
				t.Errorf("%s was not proxied", method)
			}
		})
	}
}

func TestCanceledNativeReadDoesNotFallBack(t *testing.T) {
	reset(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	request := httptest.NewRequest(http.MethodGet, "/content/announcements/main/active", nil).WithContext(ctx)
	response := httptest.NewRecorder()
	api.handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Errorf("canceled read: status=%d", response.Code)
	}
	if api.proxied.Load() != 0 {
		t.Error("canceled native read fell back to the proxy")
	}
}
