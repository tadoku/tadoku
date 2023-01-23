package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
)

// Deletes a log by id
// (DELETE /logs/{id})
func (s *Server) LogDeleteByID(ctx echo.Context, id types.UUID) error {
	if err := s.commandService.DeleteLog(ctx.Request().Context(), &command.DeleteLogRequest{
		LogID: id,
	}); err != nil {
		if errors.Is(err, command.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, command.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}

		ctx.Echo().Logger.Errorf("could not delete log: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
