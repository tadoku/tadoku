package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the scores of a user for a given year
// (GET /users/{userId}/scores/{year})
func (s *Server) ProfileYearlyScoresByUserID(ctx echo.Context, userId types.UUID, year int) error {
	summary, err := s.queryService.YearlyScoresForUser(ctx.Request().Context(), &query.YearlyScoresForUserRequest{
		UserID: userId,
		Year:   year,
	})
	if err != nil {
		ctx.Logger().Errorf("could not fetch scores: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	scores := make([]openapi.Score, len(summary.Scores))
	for i, it := range summary.Scores {
		scores[i] = openapi.Score{
			LanguageCode: it.LanguageCode,
			Score:        it.Score,
			LanguageName: it.LanguageName,
		}
	}

	return ctx.JSON(http.StatusOK, &openapi.ProfileScores{
		OverallScore: summary.OverallScore,
		Scores:       scores,
	})
}
