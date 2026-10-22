package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
	}
}

func TestIdentitySessionCallbackGolden(t *testing.T) {
	name := APITestName("IdentityProxySessionCheck", http.StatusUnauthorized, "missing", "credential")
	runCase(t, api, name, http.StatusUnauthorized, implementation{name: "tadoku-api", handler: api.handler})
}
