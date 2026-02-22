package rest

import (
	"errors"
	"net/http"

	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Updates an existing log
// (PUT /logs/{id})
func (s *Server) LogUpdate(ctx echo.Context, id openapi_types.UUID) error {
	var req openapi.LogUpdateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	log, err := s.logUpdate.Execute(ctx.Request().Context(), &domain.LogUpdateRequest{
		LogID:           id,
		UnitID:          req.UnitId,
		Amount:          req.Amount,
		DurationSeconds: int32PtrFromIntPtr(req.DurationSeconds),
		Tags:            req.Tags,
		Description:     req.Description,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		if errors.Is(err, domain.ErrInvalidLog) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, logToAPI(log))
}
