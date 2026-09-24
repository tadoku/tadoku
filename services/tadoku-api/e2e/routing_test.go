package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/getkin/kin-openapi/openapi3"
)

func TestRouterServesBusinessRoutesAndProbes(t *testing.T) {
	dir := filepath.Join("testdata", APITestName("ListActiveAnnouncements", http.StatusOK, "guest"))
	api.reset(t, dir)
	atFixtureInstant(func() { checkHTTPGolden(t, api.handler, dir, http.StatusOK, *updateGoldens) })

	for _, test := range []struct {
		name   string
		method string
		path   string
		status int
	}{
		{
			name:   "liveness",
			method: http.MethodGet,
			path:   "/livez",
			status: http.StatusOK,
		},
		{
			name:   "readiness",
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
			api.handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
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

func TestContentHeadUsesGetRoute(t *testing.T) {
	dir := filepath.Join("testdata", APITestName("ListActiveAnnouncements", http.StatusOK, "guest"))
	api.reset(t, dir)
	request := readHTTPRequest(t, dir)
	defer request.Body.Close()
	request.Method = http.MethodHead
	response := httptest.NewRecorder()

	atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })

	if response.Code != http.StatusOK {
		t.Errorf("status=%d, want %d", response.Code, http.StatusOK)
	}
}

func TestProfileHeadUsesGetRoute(t *testing.T) {
	dir := filepath.Join("testdata", APITestName("ProfileUsersList", http.StatusOK, "admin"))
	api.reset(t, dir)
	request := readHTTPRequest(t, dir)
	defer request.Body.Close()
	request.Method = http.MethodHead
	response := httptest.NewRecorder()

	atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })

	if response.Code != http.StatusOK {
		t.Errorf("status=%d, want %d", response.Code, http.StatusOK)
	}
}

func TestImmersionHeadUsesGetRoute(t *testing.T) {
	dir := filepath.Join("testdata", APITestName("ListLanguages", http.StatusOK, "admin", "ordered"))
	api.reset(t, dir)
	request := readHTTPRequest(t, dir)
	defer request.Body.Close()
	request.Method = http.MethodHead
	response := httptest.NewRecorder()

	atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })

	if response.Code != http.StatusOK {
		t.Errorf("status=%d, want %d", response.Code, http.StatusOK)
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
	for path, pathItem := range contract.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			requestPath := strings.NewReplacer("{year}", "2026", "{flagKey}", "release-log-entry-v2").Replace(path)
			requestPath = pathParameters.ReplaceAllString(requestPath, "11111111-1111-4111-8111-111111111111")
			t.Run(method+" "+requestPath, func(t *testing.T) {
				response := httptest.NewRecorder()
				api.handler.ServeHTTP(response, httptest.NewRequest(method, requestPath, nil))
				if response.Code == http.StatusNotFound || response.Code == http.StatusMethodNotAllowed {
					t.Errorf("%s %s is not registered: status=%d", method, requestPath, response.Code)
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
			response := httptest.NewRecorder()
			api.handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))

			if response.Code != test.want {
				t.Errorf("status=%d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestRetiredContentProxyRoutesAreNotForwarded(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodGet, path: "/content/ping", want: http.StatusNotFound},
		{method: http.MethodHead, path: "/content/ping", want: http.StatusNotFound},
		{method: http.MethodPost, path: "/content/announcements/main/active", want: http.StatusMethodNotAllowed},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			api.handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))

			if response.Code != test.want {
				t.Errorf("status=%d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestRetiredProfileProxyRoutesAreNotForwarded(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodGet, path: "/profile/ping", want: http.StatusNotFound},
		{method: http.MethodHead, path: "/profile/ping", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/profile/internal/v1/ping", want: http.StatusNotFound},
		{method: http.MethodHead, path: "/profile/internal/v1/ping", want: http.StatusNotFound},
		{method: http.MethodPost, path: "/profile/users", want: http.StatusMethodNotAllowed},
		{method: http.MethodGet, path: "/profile/users/unknown", want: http.StatusNotFound},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			api.handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))

			if response.Code != test.want {
				t.Errorf("status=%d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestRetiredImmersionRoutesAreNotServed(t *testing.T) {
	for _, test := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/immersion/ping"},
		{http.MethodHead, "/immersion/ping"},
		{http.MethodGet, "/immersion/unknown"},
		{http.MethodPost, "/immersion/internal/v1/account-deletion-eligibility"},
		{http.MethodPost, "/immersion/internal/v1/account-deletion-locks"},
		{http.MethodPost, "/immersion/internal/v1/account-deletion-scrubs"},
	} {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			api.handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
			if response.Code != http.StatusNotFound {
				t.Errorf("status=%d, want %d", response.Code, http.StatusNotFound)
			}
		})
	}
}

func TestUnsupportedImmersionMethodIsRejected(t *testing.T) {
	response := httptest.NewRecorder()
	api.handler.ServeHTTP(response, httptest.NewRequest(http.MethodPatch, "/immersion/languages", nil))
	if response.Code != http.StatusMethodNotAllowed {
		t.Errorf("status=%d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}
