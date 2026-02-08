package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Detaches a log from a contest (moderation action)
// (POST /contests/{id}/moderation/detach/{log_id})
func (s *Server) ContestModerationDetachLog(ctx echo.Context, id types.UUID, logId types.UUID) error {
	var req openapi.ContestModerationDetachLogJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	err := s.contestModerationDetachLog.Execute(ctx.Request().Context(), &domain.ContestModerationDetachLogRequest{
		ContestID: id,
		LogID:     logId,
		Reason:    req.Reason,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
