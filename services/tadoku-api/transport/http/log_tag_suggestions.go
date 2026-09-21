package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionLogTagSuggestions(ctx context.Context, request openapi.ImmersionLogTagSuggestionsRequestObject) (openapi.ImmersionLogTagSuggestionsResponseObject, error) {
	query := ""
	if request.Params.Query != nil {
		query = *request.Params.Query
	}

	suggestions, err := s.application.LogTagSuggestions(ctx, query)
	if err != nil {
		s.logOperationError(ctx, "get log tag suggestions", err)
		return nil, err
	}

	response := openapi.ImmersionLogTagSuggestions200JSONResponse{Suggestions: make([]openapi.ImmersionTagSuggestion, 0, len(suggestions))}
	for _, suggestion := range suggestions {
		response.Suggestions = append(response.Suggestions, openapi.ImmersionTagSuggestion{Tag: suggestion.Tag, Count: suggestion.Count})
	}
	return response, nil
}
