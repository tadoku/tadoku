package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Check if user has permission to create a new contest
// (GET /contests/create-permissions)
func (s *Server) ContestCreatePermissionCheck(ctx echo.Context) error {
	err := s.contestPermissionCheck.Execute(ctx.Request().Context())
	if err != nil {
		if handled, respErr := handleCommonDomainError(ctx, err); handled {
			return respErr
		}
		ctx.Echo().Logger.Errorf("could not fetch create permission check: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
