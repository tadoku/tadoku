package e2e_test

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commonhttperr "github.com/tadoku/tadoku/services/common/http/httperr"
	"github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"io"
	"net/http"
	"path/filepath"
	"testing"
)

func TestRequireAdmin(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"admin"}, want: http.StatusNoContent},
		{description: []string{"admin", "and", "banned"}, want: http.StatusForbidden},
		{description: []string{"other", "subject", "admin"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("RequireAdmin", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			checkPermissionsGolden(t, filepath.Join("testdata", name), test.want)
		})
	}
}

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest"}, want: http.StatusOK},
		{description: []string{"non", "admin"}, want: http.StatusOK},
		{description: []string{"admin"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("IsAdmin", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			checkPermissionsGolden(t, filepath.Join("testdata", name), test.want)
		})
	}
}

func checkPermissionsGolden(t *testing.T, path string, want int) {
	t.Helper()

	for _, implementation := range []struct {
		name    string
		handler http.Handler
	}{
		{name: "tadoku-api", handler: api.handler},
		{name: "legacy", handler: legacyPermissions},
	} {
		t.Run(implementation.name, func(t *testing.T) {
			checkCaseGolden(t, implementation.handler, path, want)
		})
	}
}

func requireAdmin(checker *permissions.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := checker.RequireAdmin(r.Context()); err != nil {
			writePermissionError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func checkAdmin(checker *permissions.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, err := checker.IsAdmin(r.Context())
		if err != nil {
			writePermissionError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_ = json.NewEncoder(w).Encode(struct {
			Admin bool `json:"admin"`
		}{Admin: admin})
	}
}

func writePermissionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, permissions.ErrUnauthorized):
		w.WriteHeader(http.StatusUnauthorized)
	case errors.Is(err, permissions.ErrForbidden):
		w.WriteHeader(http.StatusForbidden)
	case errors.Is(err, permissions.ErrUnavailable):
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func newLegacyPermissionsHandler(jwksURL, ketoReadURL string) http.Handler {
	roles := commonroles.NewKetoService(ketoclient.NewReadClient(ketoReadURL), "app", "tadoku")
	router := echo.New()
	router.Logger.SetOutput(io.Discard)
	router.GET("/test/permissions/admin", func(c echo.Context) error {
		if err := commonroles.RequireAdmin(c.Request().Context()); err != nil {
			status, ok := commonhttperr.StatusCode(err)
			if !ok {
				status = http.StatusInternalServerError
			}
			return c.NoContent(status)
		}
		return c.NoContent(http.StatusNoContent)
	},
		middleware.VerifyJWT(jwksURL),
		middleware.Identity(),
		middleware.RolesFromKeto(roles),
		middleware.RejectBannedUsers(),
	)
	router.GET("/test/permissions/check", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct {
			Admin bool `json:"admin"`
		}{Admin: commonroles.IsAdmin(c.Request().Context())})
	},
		middleware.VerifyJWT(jwksURL),
		middleware.Identity(),
		middleware.RolesFromKeto(roles),
		middleware.RejectBannedUsers(),
	)
	return router
}
