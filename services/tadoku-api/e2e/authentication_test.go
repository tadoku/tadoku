package e2e_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

func TestAuthentication(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
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
			skipParity:  "Tadoku API requires exp but legacy accepts tokens without it",
		},
		{
			description: []string{"old", "iat"},
			want:        http.StatusUnauthorized,
			skipParity:  "Tadoku API rejects tokens older than the configured maximum age",
		},
		{
			description: []string{"wrong", "issuer"},
			want:        http.StatusUnauthorized,
			skipParity:  "Tadoku API enforces the configured issuer but legacy accepts other issuers",
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
			skipParity:  "legacy Identity panics when iat is missing",
		},
		{
			description: []string{"service", "token"},
			want:        http.StatusUnauthorized,
			skipParity:  "service identities are intentionally unsupported by Tadoku API",
		},
	}
	for _, test := range tests {
		name := APITestName("Authentication", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "legacy", handler: legacyAuthentication, skip: test.skipParity},
			)
		})
	}
}

// End the middleware chain here so authentication goldens do not exercise a
// business endpoint. Both success handlers observe the real downstream context.
func observeIdentity(w http.ResponseWriter, r *http.Request) {
	if user := identity.FromContext(r.Context()); user != nil {
		writeIdentityHeaders(w.Header(), user.Subject, user.DisplayName, user.Email, user.CreatedAt)
	}
	writeAuthenticationSuccess(w)
}

func newLegacyAuthenticationHandler(jwksURL string) http.Handler {
	router := echo.New()
	middleware.RestoreJSONCharset(router)
	router.Logger.SetOutput(io.Discard)
	router.GET("/test/authentication", func(c echo.Context) error {
		if user := commondomain.ParseUserIdentity(c.Request().Context()); user != nil {
			writeIdentityHeaders(c.Response().Header(), user.Subject, user.DisplayName, user.Email, user.CreatedAt)
		}
		writeAuthenticationSuccess(c.Response())
		return nil
	}, middleware.VerifyJWT(jwksURL), middleware.Identity())
	return router
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

func TestAuthenticationDoesNotChangeProbesOrRemainingProxyRoutes(t *testing.T) {
	for _, test := range []struct {
		method  string
		path    string
		want    int
		proxied bool
	}{
		{method: http.MethodGet, path: "/livez", want: http.StatusOK},
		{method: http.MethodGet, path: "/readyz", want: http.StatusOK},
		{method: http.MethodGet, path: "/authz/ping", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/immersion/ping", want: http.StatusNoContent, proxied: true},
		{method: http.MethodGet, path: "/profile/ping", want: http.StatusNoContent, proxied: true},
		{method: http.MethodHead, path: "/immersion/languages", want: http.StatusNoContent, proxied: true},
		{method: http.MethodPatch, path: "/immersion/languages", want: http.StatusNoContent, proxied: true},
	} {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			for _, authorization := range []string{"", "Bearer invalid-token"} {
				api.resetProxyCount()
				request := httptest.NewRequest(test.method, test.path, nil)
				if authorization != "" {
					request.Header.Set("Authorization", authorization)
				}
				response := httptest.NewRecorder()
				api.handler.ServeHTTP(response, request)
				if response.Code != test.want {
					t.Errorf("authorization=%q status=%d, want %d", authorization, response.Code, test.want)
				}
				if got := api.proxied.Load(); (got == 1) != test.proxied {
					t.Errorf("unexpected upstream request count: %d", got)
				}
			}
		})
	}
}
