package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

func TestAuthentication(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{
			description: []string{"user"},
			want:        http.StatusOK,
		},
		{
			description: []string{"legacy", "type"},
			want:        http.StatusOK,
		},
		{
			description: []string{"guest"},
			want:        http.StatusOK,
		},
		{
			description: []string{"without", "exp"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"old", "iat"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"wrong", "issuer"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"lowercase", "bearer"},
			want:        http.StatusOK,
		},
		{
			description: []string{"repeated", "authorization"},
			want:        http.StatusOK,
		},
		{
			description: []string{"malformed", "then", "valid", "header"},
			want:        http.StatusOK,
		},
		{
			description: []string{"without", "credentials"},
			want:        http.StatusBadRequest,
		},
		{
			description: []string{"wrong", "scheme"},
			want:        http.StatusBadRequest,
		},
		{
			description: []string{"empty", "bearer"},
			want:        http.StatusBadRequest,
		},
		{
			description: []string{"malformed", "token"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"bad", "signature"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"unknown", "kid"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"missing", "kid"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"wrong", "algorithm"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"unsigned", "token"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"expired"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"expiration", "boundary"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"future", "nbf"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"future", "iat"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"invalid", "token", "after", "malformed", "header"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"missing", "iat"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"service", "token"},
			want:        http.StatusUnauthorized,
		},
		{
			description: []string{"non", "uuid", "subject"},
			want:        http.StatusUnauthorized,
		},
	}
	for _, test := range tests {
		name := APITestName("Authentication", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}

func observeIdentity(w http.ResponseWriter, r *http.Request) {
	if user := identity.FromContext(r.Context()); user != nil {
		writeIdentityHeaders(w.Header(), user.Subject, user.DisplayName, user.Email, user.CreatedAt)
	}
	writeAuthenticationSuccess(w)
}

func writeAuthenticationSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, _ = w.Write([]byte("{\"status\":\"success\"}\n"))
}

func writeIdentityHeaders(header http.Header, subject, displayName, email string, createdAt time.Time) {
	header.Set("X-Test-Identity-Subject", subject)
	header.Set("X-Test-Identity-Display-Name", displayName)
	header.Set("X-Test-Identity-Email", email)
	header.Set("X-Test-Identity-Created-At", createdAt.UTC().Format(time.RFC3339))
}

func TestAuthenticationDoesNotChangeProbesOrUnknownRoutes(t *testing.T) {
	for _, test := range []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodGet, path: "/livez", want: http.StatusOK},
		{method: http.MethodGet, path: "/readyz", want: http.StatusOK},
		{method: http.MethodGet, path: "/authz/ping", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/immersion/ping", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/immersion/unknown", want: http.StatusNotFound},
	} {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			for _, authorization := range []string{"", "Bearer invalid-token"} {
				request := httptest.NewRequest(test.method, test.path, nil)
				if authorization != "" {
					request.Header.Set("Authorization", authorization)
				}
				response := httptest.NewRecorder()
				api.handler.ServeHTTP(response, request)
				if response.Code != test.want {
					t.Errorf("authorization=%q status=%d, want %d", authorization, response.Code, test.want)
				}
			}
		})
	}
}
