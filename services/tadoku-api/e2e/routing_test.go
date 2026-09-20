package e2e_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/getkin/kin-openapi/openapi3"
)

func TestRouterWorksWithoutLegacyProxyRoutes(t *testing.T) {
	// Construct the same application router but do not attach legacy routes.
	// Normal scenarios continue to use the single suite-level router.
	handler, _, _, err := newTestRouter(t.Context(), api.db.Pool, keto, api.kratos)
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join("testdata", APITestName("ListActiveAnnouncements", http.StatusOK, "guest"))
	api.reset(t, dir)
	atFixtureInstant(func() { checkHTTPGolden(t, handler, dir, http.StatusOK, *updateGoldens) })

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
			path:   "/content/flags/example",
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

func TestCallbackCredentialDoesNotAuthenticateBusinessRoutes(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/content/announcements/main/active", nil)
	request.Header.Set("Authorization", "Bearer "+callbackToken)
	response := httptest.NewRecorder()

	api.handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("status=%d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestContractRouteOwnership(t *testing.T) {
	specPath, err := bazel.Runfile("services/tadoku-api/spec/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	contract, err := openapi3.NewLoader().LoadFromFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	pathParameters := regexp.MustCompile(`\{[^}]+\}`)
	for path, pathItem := range contract.Paths {
		for method, operation := range pathItem.Operations() {
			ownerJSON, ok := operation.Extensions["x-tadoku-owner"].(json.RawMessage)
			var owner string
			if !ok || json.Unmarshal(ownerJSON, &owner) != nil || (owner != "legacy" && owner != "native") {
				t.Fatalf("%s %s has invalid x-tadoku-owner %s", method, path, ownerJSON)
			}
			requestPath := strings.NewReplacer("{year}", "2026", "{flagKey}", "release-log-entry-v2").Replace(path)
			requestPath = pathParameters.ReplaceAllString(requestPath, "11111111-1111-4111-8111-111111111111")
			t.Run(method+" "+requestPath, func(t *testing.T) {
				response := httptest.NewRecorder()
				api.handler.ServeHTTP(response, httptest.NewRequest(method, requestPath, nil))
				proxied := response.Header().Get("X-Proxied") == "yes"
				if want := owner == "legacy"; proxied != want {
					t.Errorf("%s %s (x-tadoku-owner: %s): proxied = %t, want %t", method, requestPath, owner, proxied, want)
				}
			})
		}
	}
}

func TestRetiredAuthzRoutesAreNotForwarded(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodGet, path: "/authz/ping", want: http.StatusNotFound},
		{method: http.MethodHead, path: "/authz/ping", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/authz/internal/v1/ping", want: http.StatusNotFound},
		{method: http.MethodPost, path: "/authz/internal/v1/permission/check", want: http.StatusNotFound},
		{method: http.MethodPost, path: "/authz/internal/v1/relationships", want: http.StatusNotFound},
		{method: http.MethodDelete, path: "/authz/internal/v1/relationships", want: http.StatusNotFound},
		{method: http.MethodHead, path: "/authz/internal/v1/proxy/admin-check", want: http.StatusMethodNotAllowed},
		{method: http.MethodPost, path: "/authz/internal/v1/proxy/admin-check/extra", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/authz/unknown", want: http.StatusNotFound},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			api.resetProxyCount()
			response := httptest.NewRecorder()
			api.handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))

			if response.Code != test.want {
				t.Errorf("status=%d, want %d", response.Code, test.want)
			}
			if got := api.proxied.Load(); got != 0 {
				t.Errorf("retired route made %d upstream requests", got)
			}
		})
	}
}
