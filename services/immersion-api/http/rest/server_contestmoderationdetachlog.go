package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
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

	err := s.commandService.DetachLogFromContest(ctx.Request().Context(), &command.DetachLogFromContestRequest{
		ContestID: id,
		LogID:     logId,
		Reason:    req.Reason,
	})
	if err != nil {
		if errors.Is(err, query.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}
		if errors.Is(err, command.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, command.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
