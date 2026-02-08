package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the scores of a user profile in a contest
// (GET /contests/{id}/profile/{user_id}/scores)
func (s *Server) ContestProfileFetchScores(ctx echo.Context, id types.UUID, userId types.UUID) error {
	profile, err := s.profileContest.Execute(ctx.Request().Context(), &domain.ProfileContestRequest{
		UserID:    userId,
		ContestID: id,
	})
	if err != nil {
		if handled, respErr := noContentForCommonDomainError(ctx, err); handled {
			return respErr
		}
		ctx.Logger().Errorf("could not fetch profile: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	scores := make([]openapi.Score, len(profile.Scores))
	for i, it := range profile.Scores {
		scores[i] = openapi.Score{
			LanguageCode: it.LanguageCode,
			Score:        it.Score,
		}
	}

	return ctx.JSON(http.StatusOK, &openapi.ContestProfileScores{
		OverallScore: profile.OverallScore,
		Registration: *contestRegistrationToAPI(profile.Registration),
		Scores:       scores,
	})
}
