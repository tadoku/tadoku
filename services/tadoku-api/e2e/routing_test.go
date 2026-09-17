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
