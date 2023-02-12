package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

// Check if user has permission to create a new contest
// (GET /contests/create-permissions)
func (s *Server) ContestCreatePermissionCheck(ctx echo.Context) error {
	err := s.queryService.CreateContestPermissionCheck(ctx.Request().Context())
	if err != nil {
		if errors.Is(err, query.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		ctx.Echo().Logger.Errorf("could not fetch create permission check: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
