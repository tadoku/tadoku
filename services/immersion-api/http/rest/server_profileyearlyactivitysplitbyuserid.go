package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches a activity split summary of a user for a given year
// (GET /users/{userId}/activity-split/{year})
func (s *Server) ProfileYearlyActivitySplitByUserID(ctx echo.Context, userId types.UUID, year int) error {
	summary, err := s.queryService.YearlyActivitySplitForUser(ctx.Request().Context(), &query.YearlyActivitySplitForUserRequest{
		UserID: userId,
		Year:   year,
	})
	if err != nil {
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	scores := make([]openapi.ActivitySplitScore, len(summary.Activities))
	for i, it := range summary.Activities {
		scores[i] = openapi.ActivitySplitScore{
			ActivityId:   it.ActivityID,
			ActivityName: it.ActivityName,
			Score:        it.Score,
		}
	}

	res := &openapi.ActivitySplit{
		Activities: scores,
	}

	return ctx.JSON(http.StatusOK, res)
}
