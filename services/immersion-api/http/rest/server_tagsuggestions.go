package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches tag suggestions for autocomplete
// (GET /logs/tag-suggestions)
func (s *Server) LogTagSuggestions(ctx echo.Context, params openapi.LogTagSuggestionsParams) error {
	query := ""
	if params.Query != nil {
		query = *params.Query
	}

	result, err := s.tagSuggestions.Execute(ctx.Request().Context(), &domain.TagSuggestionsRequest{
		Query: query,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.TagSuggestions{
		Suggestions: result.Suggestions,
	})
}
