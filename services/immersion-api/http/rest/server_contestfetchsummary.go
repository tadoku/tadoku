package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the summary for a contest
// (GET /contests/{id}/summary)
func (s *Server) ContestFetchSummary(ctx echo.Context, id types.UUID) error {
	summary, err := s.queryService.FetchContestSummary(ctx.Request().Context(), &query.FetchContestSummaryRequest{
		ContestID: id,
	})
	if err != nil {
		if errors.Is(err, query.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
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
