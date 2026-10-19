package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestIdentityRouteAuthentication(t *testing.T) {
	api.reset(t, "")
	guest := readHTTPRequest(t, filepath.Join("testdata", APITestName("Authentication", http.StatusOK, "guest")))
	routes := []struct{ method, path, body string }{
		{
			"POST", "/identity/registration",
			`{"email":"reader@example.test","display_name":"Reader","password":"password123"}`,
		},
		{"POST", "/identity/login", `{"email":"reader@example.test","password":"password123"}`},
		{"POST", "/identity/logout", ""},
		{"GET", "/identity/session", ""},
		{"POST", "/identity/recovery", `{"email":"reader@example.test","return_to":"https://tadoku.app/settings"}`},
		{"GET", "/identity/recovery/recovery-token", ""},
		{"PATCH", "/identity/settings/profile", `{"display_name":"Reader"}`},
		{"PUT", "/identity/settings/password", `{"password":"password123"}`},
	}
	for _, route := range routes {
		t.Run(route.method+route.path, func(t *testing.T) {
			for _, auth := range []struct {
				name, authorization string
				status              int
			}{
				{"guest", guest.Header.Get("Authorization"), http.StatusNotImplemented},
				{"missing", "", http.StatusBadRequest},
				{"invalid", "Bearer invalid", http.StatusUnauthorized},
				{"access token outside email", "Bearer tdk_at_e2e_skeleton_secret", http.StatusUnauthorized},
			} {
				if strings.HasPrefix(route.path, "/emails") && auth.name == "access token outside email" {
					continue
				}
				request := httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("Origin", "https://tadoku.dev.lab")
				request.Header.Set("Authorization", auth.authorization)
				response := httptest.NewRecorder()

				atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })
				if auth.name == "guest" {
					if route.path == "/identity/logout" {
						auth.status = http.StatusNoContent
					}
					if route.path == "/identity/session" {
						auth.status = http.StatusOK
					}
				}

				if response.Code != auth.status {
					t.Errorf("%s status=%d body=%s, want %d", auth.name, response.Code, response.Body, auth.status)
				}
				if auth.status == http.StatusNotImplemented && response.Body.String() != "{\"message\":\"not implemented\"}\n" {
					t.Errorf("stub error=%s", response.Body)
				}
			}
		})
	}
}

func TestIdentitySessionCallbackAuthentication(t *testing.T) {
	for _, auth := range []struct {
		authorization string
		status        int
	}{
		{"", http.StatusUnauthorized},
		{"Bearer wrong", http.StatusUnauthorized},
		{"Bearer " + callbackToken, http.StatusOK},
	} {
		request := httptest.NewRequest("GET", "/identity/internal/v1/session", nil)
		request.Header.Set("Authorization", auth.authorization)
		response := httptest.NewRecorder()

		api.handler.ServeHTTP(response, request)

		if response.Code != auth.status {
			t.Errorf("status=%d body=%s, want %d", response.Code, response.Body, auth.status)
		}
		if auth.status == http.StatusNotImplemented && response.Body.String() != "{\"message\":\"not implemented\"}\n" {
			t.Errorf("stub error=%s", response.Body)
		}
	}
}

func TestIdentityEmailTokenSkeleton(t *testing.T) {
	for _, path := range []string{"/emails", "/emails/11111111-1111-4111-8111-111111111111"} {
		request := httptest.NewRequest("GET", path, nil)
		request.Header.Set("Authorization", "Bearer tdk_at_e2e_skeleton_secret")
		response := httptest.NewRecorder()

		api.handler.ServeHTTP(response, request)

		if response.Code != http.StatusNotImplemented || response.Body.String() != "{\"message\":\"not implemented\"}\n" {
			t.Errorf("%s token verifier stub status=%d body=%s", path, response.Code, response.Body)
		}
	}
}

func TestIdentitySessionCallbackGolden(t *testing.T) {
	name := APITestName("IdentityProxySessionCheck", http.StatusUnauthorized, "missing", "credential")
	runCase(t, api, name, http.StatusUnauthorized, implementation{name: "tadoku-api", handler: api.handler})
}
