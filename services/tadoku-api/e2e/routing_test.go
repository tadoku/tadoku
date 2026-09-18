package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestRouterWorksWithoutLegacyProxyRoutes(t *testing.T) {
	// Construct the same application router but do not attach legacy routes.
	// Normal scenarios continue to use the single suite-level router.
	handler, err := newTestRouter(t.Context(), api.db.Pool, keto.ReadURL())
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join("testdata", APITestName("ListActiveAnnouncements", http.StatusOK, "guest"))
	api.reset(t, dir)
	atFixtureInstant(func() { checkHTTPGolden(t, handler, dir, http.StatusOK) })

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

func TestUnclaimedMethodsRemainProxied(t *testing.T) {
	for _, route := range []struct {
		name    string
		path    string
		methods []string
	}{
		{
			name:    "active list",
			path:    "/content/announcements/main/active",
			methods: []string{http.MethodHead, http.MethodOptions, http.MethodPost, http.MethodPatch},
		},
		{
			name:    "admin list",
			path:    "/content/announcements/main",
			methods: []string{http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete, http.MethodPatch},
		},
		{
			name:    "by ID",
			path:    "/content/announcements/main/11111111-1111-4111-8111-111111111111",
			methods: []string{http.MethodHead, http.MethodOptions, http.MethodPost, http.MethodPatch},
		},
	} {
		t.Run(route.name, func(t *testing.T) {
			for _, method := range route.methods {
				t.Run(method, func(t *testing.T) {
					api.resetProxyCount()
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
