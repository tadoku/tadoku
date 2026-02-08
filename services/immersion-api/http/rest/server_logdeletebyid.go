package rest

import (
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
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}

		ctx.Echo().Logger.Errorf("could not delete log: %v", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
