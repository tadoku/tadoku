package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the leaderboard for a contest
// (GET /contests/{id}/leaderboard)
func (s *Server) ContestFetchLeaderboard(ctx echo.Context, id types.UUID, params openapi.ContestFetchLeaderboardParams) error {
	req := &domain.ContestLeaderboardFetchRequest{
		ContestID:    id,
		LanguageCode: params.LanguageCode,
	}

	if params.PageSize != nil {
		req.PageSize = *params.PageSize
	}
	if params.Page != nil {
		req.Page = *params.Page
	}
	if params.ActivityId != nil {
		activityID := int32(*params.ActivityId)
		req.ActivityID = &activityID
	}

	leaderboard, err := s.contestLeaderboardFetch.Execute(ctx.Request().Context(), req)
	if err != nil {
		if handled, respErr := handleCommonDomainError(ctx, err); handled {
			return respErr
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, domainLeaderboardToAPI(*leaderboard))
}
