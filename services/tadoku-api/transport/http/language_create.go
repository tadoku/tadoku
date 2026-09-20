package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionLanguageCreate(
	ctx context.Context,
	request openapi.ImmersionLanguageCreateRequestObject,
) (openapi.ImmersionLanguageCreateResponseObject, error) {
	if request.Body == nil {
		return openapi.ImmersionLanguageCreate400Response{}, nil
	}

	err := s.application.CreateLanguage(ctx, app.CreateLanguageParameters{
		Code: request.Body.Code,
		Name: request.Body.Name,
	})
	if err != nil {
		s.logOperationError(ctx, "create language", err)
		return nil, err
	}
	return openapi.ImmersionLanguageCreate200Response{}, nil
}
