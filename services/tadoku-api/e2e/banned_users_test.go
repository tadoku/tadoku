package e2e_test

import (
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestBannedUsers(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest"}, want: http.StatusOK},
		{description: []string{"without", "tuple"}, want: http.StatusOK},
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"banned"}, want: http.StatusForbidden},
		{description: []string{"admin", "and", "banned"}, want: http.StatusForbidden},
		{description: []string{"other", "user", "banned"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("BannedUsers", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("testdata", name)
			for _, implementation := range []struct {
				name    string
				handler http.Handler
			}{
				{name: "tadoku-api", handler: api.authenticatedHandler},
				{name: "legacy", handler: legacyBannedUsers},
			} {
				t.Run(implementation.name, func(t *testing.T) {
					resetCase(t, path)

					previous := jwt.TimeFunc
					jwt.TimeFunc = timex.Now
					defer func() { jwt.TimeFunc = previous }()
					timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
						checkHTTPGolden(t, implementation.handler, path, test.want)
					})
				})
			}
		})
	}
}

func bannedUsersSuccess(w http.ResponseWriter, r *http.Request) {
	if user := identity.FromContext(r.Context()); user != nil {
		writeIdentityHeaders(w.Header(), user.Subject, user.DisplayName, user.Email, user.CreatedAt)
	}
	writeAuthenticationSuccess(w)
}

func newLegacyBannedUsersHandler(jwksURL, ketoReadURL string) http.Handler {
	roles := commonroles.NewKetoService(ketoclient.NewReadClient(ketoReadURL), "app", "tadoku")
	router := echo.New()
	router.Logger.SetOutput(io.Discard)
	router.GET("/test/banned", func(c echo.Context) error {
		if user := commondomain.ParseUserIdentity(c.Request().Context()); user != nil {
			writeIdentityHeaders(c.Response().Header(), user.Subject, user.DisplayName, user.Email, user.CreatedAt)
		}
		writeAuthenticationSuccess(c.Response())
		return nil
	},
		middleware.VerifyJWT(jwksURL),
		middleware.Identity(),
		middleware.RolesFromKeto(roles),
		middleware.RejectBannedUsers(),
	)
	return router
}
