package e2e_test

import (
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
	repository := content.NewAnnouncementsRepository(api.db.Pool)
	service := content.NewService(repository)
	application := app.New(service, api.db.Pool, nil)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	handler, err := transport.NewHandler(application, api.db.Pool.Ping, time.Second, logger, skipAuthentication, skipBanCheck)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join("testdata", APITestName("ListActiveAnnouncements", http.StatusOK, "without", "auth"))
	resetCase(t, path)
	timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
		checkHTTPGolden(t, handler, path, http.StatusOK)
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

func TestMiddlewareFixtureRoutesUseApplicationRouter(t *testing.T) {
	for _, test := range []struct {
		name   string
		method string
		path   string
		want   int
	}{
		{name: "authentication method", method: http.MethodPost, path: "/test/authentication", want: http.StatusMethodNotAllowed},
		{name: "ban method", method: http.MethodPost, path: "/test/banned", want: http.StatusMethodNotAllowed},
		{name: "unknown path", method: http.MethodGet, path: "/test/missing", want: http.StatusNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			api.authenticatedHandler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))

			if response.Code != test.want {
				t.Errorf("status=%d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestUnclaimedMethodsRemainProxied(t *testing.T) {
	for _, route := range []struct {
		name string
		path string
	}{
		{name: "active list", path: "/content/announcements/main/active"},
		{name: "admin list", path: "/content/announcements/main"},
	} {
		t.Run(route.name, func(t *testing.T) {
			for _, method := range []string{http.MethodHead, http.MethodOptions, http.MethodPost, http.MethodDelete, http.MethodPatch} {
				t.Run(method, func(t *testing.T) {
					reset(t)
					request := httptest.NewRequest(method, route.path, nil)
					response := httptest.NewRecorder()
					api.handler.ServeHTTP(response, request)

					if response.Header().Get("X-Proxied") != "yes" {
						t.Errorf("%s was not proxied", method)
					}
				})
			}
		})
	}
}
