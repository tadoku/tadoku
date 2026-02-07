package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/authz/roles"
	"github.com/tadoku/tadoku/services/common/domain"
)

// Fetches the role of the current user
// (GET /current-user/role)
func (s *Server) GetCurrentUserRole(ctx echo.Context) error {
	session := domain.ParseUserIdentity(ctx.Request().Context())
	if session == nil {
		return ctx.NoContent(http.StatusUnauthorized)
	}

	if session.Subject == "guest" {
		return ctx.JSON(http.StatusOK, map[string]string{"role": "guest"})
	}

	claims := roles.FromContext(ctx.Request().Context())
	if claims.Err != nil {
		return ctx.NoContent(http.StatusServiceUnavailable)
	}

	role := "user"
	if claims.Banned {
		role = "banned"
	} else if claims.Admin {
		role = "admin"
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"role": role,
	})
}
