package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the summary for a contest
// (GET /contests/{id}/summary)
func (s *Server) ContestFetchSummary(ctx echo.Context, id types.UUID) error {
	summary, err := s.contestSummaryFetch.Execute(ctx.Request().Context(), &domain.ContestSummaryFetchRequest{
		ContestID: id,
	})
	if err != nil {
		if handled, respErr := noContentForCommonDomainError(ctx, err); handled {
			return respErr
		}
		ctx.Echo().Logger.Errorf("could not fetch summary: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, &openapi.ContestSummary{
		ParticipantCount: summary.ParticipantCount,
		LanguageCount:    summary.LanguageCount,
		TotalScore:       summary.TotalScore,
	})
}
