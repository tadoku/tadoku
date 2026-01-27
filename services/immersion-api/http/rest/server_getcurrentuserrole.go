package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/domain"
)

// Fetches the role of the current user
// (GET /current-user/role)
func (s *Server) GetCurrentUserRole(ctx echo.Context) error {
	session := domain.ParseSession(ctx.Request().Context())
	if session == nil {
		return ctx.NoContent(http.StatusUnauthorized)
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"role": string(session.Role),
	})
}
