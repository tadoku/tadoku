package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches a activity summary of a user for a given year
// (GET /users/{userId}/activity/{year})
func (s *Server) ProfileYearlyActivityByUserID(ctx echo.Context, userId types.UUID, year int) error {
	summary, err := s.queryService.YearlyActivityForUser(ctx.Request().Context(), &query.YearlyActivityForUserRequest{
		UserID: userId,
		Year:   year,
	})
	if err != nil {
		ctx.Echo().Logger.Errorf("could not fetch activity summary: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	scores := make([]openapi.UserActivityScore, len(summary.Scores))
	for i, it := range summary.Scores {
		scores[i] = openapi.UserActivityScore{
			Date:  types.Date{Time: it.Date},
			Score: it.Score,
		}
	}

	res := &openapi.UserActivity{
		TotalUpdates: summary.TotalUpdates,
		Scores:       scores,
	}

	return ctx.JSON(http.StatusOK, res)
}
