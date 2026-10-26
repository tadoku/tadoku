package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app/api"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionLanguageUpdate(
	ctx context.Context,
	request openapi.ImmersionLanguageUpdateRequestObject,
) (openapi.ImmersionLanguageUpdateResponseObject, error) {
	err := s.application.UpdateLanguage(ctx, api.UpdateLanguageParameters{
		Code: request.Code,
		Name: request.Body.Name,
	})
	if err != nil {
		s.logOperationError(ctx, "update language", err)
		return nil, err
	}
	return openapi.ImmersionLanguageUpdate200Response{}, nil
}
