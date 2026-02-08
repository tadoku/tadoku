package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

// Fetches a log by id
// (GET /logs/{id})
func (s *Server) LogFindByID(ctx echo.Context, id types.UUID) error {
	log, err := s.logFind.Execute(ctx.Request().Context(), &domain.LogFindRequest{
		ID: id,
	})
	if err != nil {
		if handled, respErr := noContentForCommonDomainError(ctx, err); handled {
			return respErr
		}
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, logToAPI(log))
}
