package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionLanguageList(
	ctx context.Context,
	_ openapi.ImmersionLanguageListRequestObject,
) (openapi.ImmersionLanguageListResponseObject, error) {
	items, err := s.application.ListLanguages(ctx)
	if err != nil {
		s.logOperationError(ctx, "list languages", err)
		return nil, err
	}

	response := openapi.ImmersionLanguageList200JSONResponse{
		Languages: make([]openapi.ImmersionLanguage, 0, len(items)),
	}
	for _, item := range items {
		response.Languages = append(response.Languages, openapi.ImmersionLanguage{
			Code: item.Code,
			Name: item.Name,
		})
	}
	return response, nil
}
