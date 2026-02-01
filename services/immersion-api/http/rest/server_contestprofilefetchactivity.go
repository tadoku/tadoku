package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the activity of a user profile in a contest
// (GET /contests/{id}/profile/{user_id}/activity)
func (s *Server) ContestProfileFetchActivity(ctx echo.Context, id types.UUID, userId types.UUID) error {
	stats, err := s.profileContestActivity.Execute(ctx.Request().Context(), &domain.ProfileContestActivityRequest{
		UserID:    userId,
		ContestID: id,
	})
	if err != nil {
		ctx.Echo().Logger.Errorf("could not fetch activity: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	rows := make([]openapi.ContestProfileActivityRow, len(stats.Rows))
	for i, it := range stats.Rows {
		rows[i] = openapi.ContestProfileActivityRow{
			Date:         types.Date{Time: it.Date},
			LanguageCode: it.LanguageCode,
			Score:        it.Score,
		}
	}

	return ctx.JSON(http.StatusOK, &openapi.ContestProfileActivity{
		Rows: rows,
	})
}
