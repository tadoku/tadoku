package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the global leaderboard
// (GET /leaderboard/global)
func (s *Server) FetchLeaderboardGlobal(ctx echo.Context, params openapi.FetchLeaderboardGlobalParams) error {
	req := &domain.LeaderboardGlobalRequest{
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

	leaderboard, err := s.leaderboardGlobal.Execute(ctx.Request().Context(), req)
	if err != nil {
		if handled, respErr := handleCommonDomainError(ctx, err); handled {
			return respErr
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, domainLeaderboardToAPI(*leaderboard))
}
