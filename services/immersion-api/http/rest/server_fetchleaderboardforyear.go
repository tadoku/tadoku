package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the leaderboard for a given year
// (GET /leaderboard/yearly/{year})
func (s *Server) FetchLeaderboardForYear(ctx echo.Context, year int, params openapi.FetchLeaderboardForYearParams) error {
	req := &query.FetchYearlyLeaderboardRequest{
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
		id := int32(*params.ActivityId)
		req.ActivityID = &id
	}

	leaderboard, err := s.queryService.FetchYearlyLeaderboard(ctx.Request().Context(), req)
	if err != nil {
		if errors.Is(err, query.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, queryLeaderboardToAPI(*leaderboard))
}
