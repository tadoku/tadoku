package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

// Fetches a log by id
// (GET /logs/{id})
func (s *Server) LogFindByID(ctx echo.Context, id types.UUID) error {
	log, err := s.queryService.FindLogByID(ctx.Request().Context(), &query.FindLogByIDRequest{
		ID: id,
	})
	if err != nil {
		if errors.Is(err, query.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, logToAPI(log))
}
