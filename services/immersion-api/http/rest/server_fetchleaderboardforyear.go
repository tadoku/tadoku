package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the leaderboard for a given year
// (GET /leaderboard/yearly/{year})
func (s *Server) FetchLeaderboardForYear(ctx echo.Context, year int, params openapi.FetchLeaderboardForYearParams) error {
	req := &domain.LeaderboardYearlyRequest{
		LanguageCode: params.LanguageCode,
		Year:         int32(year),
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

	leaderboard, err := s.leaderboardYearly.Execute(ctx.Request().Context(), req)
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, domainLeaderboardToAPI(*leaderboard))
}
