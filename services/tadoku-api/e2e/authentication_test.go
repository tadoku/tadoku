package e2e_test

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestAuthentication(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  bool
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
			want:        http.StatusOK,
		},
		{
			description: []string{"old", "iat"},
			want:        http.StatusOK,
		},
		{
			description: []string{"other", "issuer", "audience"},
			want:        http.StatusOK,
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
			// Legacy Identity panics when iat is missing.
			description: []string{"missing", "iat"},
			want:        http.StatusUnauthorized,
			skipParity:  true,
		},
		{
			// Service identities are intentionally unsupported by Tadoku API.
			description: []string{"service", "token"},
			want:        http.StatusUnauthorized,
			skipParity:  true,
		},
	}
	for _, test := range tests {
		name := APITestName("Authentication", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("testdata", name)
			for _, implementation := range []struct {
				name    string
				handler http.Handler
			}{
				{name: "tadoku-api", handler: api.authentication},
				{name: "legacy", handler: legacyAuthentication},
			} {
				t.Run(implementation.name, func(t *testing.T) {
					if test.skipParity && implementation.name == "legacy" {
						t.Skip("intentional authentication difference")
					}
					checkAuthenticationGolden(t, implementation.handler, path, test.want)
				})
			}
		})
	}
}

func checkAuthenticationGolden(t *testing.T, handler http.Handler, path string, want int) {
	t.Helper()
	reset(t, filepath.Join(path, "setup.sql"))

	// Both parsers validate RegisteredClaims using jwt/v4's clock. Keep this
	// override test-only, scoped and sequential, just like the SQL clock binding.
	previous := jwt.TimeFunc
	jwt.TimeFunc = timex.Now
	defer func() { jwt.TimeFunc = previous }()

	timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
		checkHTTPGolden(t, handler, path, want)
	})
}

// End the middleware chain here so authentication goldens do not exercise a
// business endpoint. Both success handlers observe the real downstream context.
func authenticationSuccess(w http.ResponseWriter, r *http.Request) {
	if user := identity.FromContext(r.Context()); user != nil {
		writeIdentityHeaders(w.Header(), user.Subject, user.DisplayName, user.Email, user.CreatedAt)
	}
	writeAuthenticationSuccess(w)
}

func newLegacyAuthenticationHandler(jwksURL string) http.Handler {
	router := echo.New()
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

func TestAuthenticationWiring(t *testing.T) {
	input, err := os.ReadFile("testdata/Authentication/200_user/request.http")
	if err != nil {
		t.Fatal(err)
	}
	validRequest, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(input)))
	if err != nil {
		t.Fatal(err)
	}
	defer validRequest.Body.Close()

	previous := jwt.TimeFunc
	jwt.TimeFunc = timex.Now
	defer func() { jwt.TimeFunc = previous }()

	timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
		for _, test := range []struct {
			name          string
			authorization string
			want          int
		}{
			{name: "missing", want: http.StatusBadRequest},
			{name: "invalid", authorization: "Bearer invalid-token", want: http.StatusUnauthorized},
			{name: "valid", authorization: validRequest.Header.Get("Authorization"), want: http.StatusOK},
		} {
			t.Run(test.name, func(t *testing.T) {
				reset(t)
				request := httptest.NewRequest(http.MethodGet, "/content/announcements/main/active", nil)
				if test.authorization != "" {
					request.Header.Set("Authorization", test.authorization)
				}
				response := httptest.NewRecorder()
				api.authenticatedHandler.ServeHTTP(response, request)
				if response.Code != test.want {
					t.Errorf("status=%d, want %d", response.Code, test.want)
				}
				if api.proxied.Load() != 0 {
					t.Error("protected route contacted an upstream")
				}
			})
		}
	})
}

func TestAuthenticationDoesNotChangeProbesOrProxyRoutes(t *testing.T) {
	for _, test := range []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodGet, path: "/livez", want: http.StatusOK},
		{method: http.MethodGet, path: "/readyz", want: http.StatusOK},
		{method: http.MethodGet, path: "/authz/ping", want: http.StatusNoContent},
		{method: http.MethodGet, path: "/content/ping", want: http.StatusNoContent},
		{method: http.MethodGet, path: "/immersion/ping", want: http.StatusNoContent},
		{method: http.MethodGet, path: "/profile/ping", want: http.StatusNoContent},
		{method: http.MethodHead, path: "/content/announcements/main/active", want: http.StatusNoContent},
		{method: http.MethodPost, path: "/content/announcements/main/active", want: http.StatusNoContent},
	} {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			for _, authorization := range []string{"", "Bearer invalid-token"} {
				reset(t)
				request := httptest.NewRequest(test.method, test.path, nil)
				if authorization != "" {
					request.Header.Set("Authorization", authorization)
				}
				response := httptest.NewRecorder()
				api.authenticatedHandler.ServeHTTP(response, request)
				if response.Code != test.want {
					t.Errorf("authorization=%q status=%d, want %d", authorization, response.Code, test.want)
				}
				if got := api.proxied.Load(); (test.want == http.StatusNoContent) != (got == 1) {
					t.Errorf("unexpected upstream request count: %d", got)
				}
			}
		})
	}
}
