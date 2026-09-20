package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionLanguageUpdate(
	ctx context.Context,
	request openapi.ImmersionLanguageUpdateRequestObject,
) (openapi.ImmersionLanguageUpdateResponseObject, error) {
	if request.Body == nil {
		return openapi.ImmersionLanguageUpdate400Response{}, nil
	}

	err := s.application.UpdateLanguage(ctx, app.UpdateLanguageParameters{
		Code: request.Code,
		Name: request.Body.Name,
	})
	if err != nil {
		s.logOperationError(ctx, "update language", err)
		return nil, err
	}
	return openapi.ImmersionLanguageUpdate200Response{}, nil
}
