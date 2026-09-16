package e2e_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/middleware"
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
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "legacy", handler: legacyBannedUsers},
			)
		})
	}
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
