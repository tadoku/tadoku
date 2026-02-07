package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

// Deletes a log by id
// (DELETE /logs/{id})
func (s *Server) LogDeleteByID(ctx echo.Context, id types.UUID) error {
	if err := s.logDelete.Execute(ctx.Request().Context(), &domain.LogDeleteRequest{
		LogID: id,
	}); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrAuthzUnavailable) {
			return ctx.NoContent(http.StatusServiceUnavailable)
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}

		ctx.Echo().Logger.Errorf("could not delete log: %v", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
