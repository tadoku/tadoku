package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionFeatureAccessGet(
	ctx context.Context,
	request openapi.ImmersionFeatureAccessGetRequestObject,
) (openapi.ImmersionFeatureAccessGetResponseObject, error) {
	result, err := s.application.FeatureAccessGet(ctx, string(request.FlagKey), request.UserId)
	if err != nil {
		s.logOperationError(ctx, "get feature access", err)
		return nil, err
	}
	return openapi.ImmersionFeatureAccessGet200JSONResponse{
		Body:    featureAccessResponse(result),
		Headers: openapi.ImmersionFeatureAccessGet200ResponseHeaders{CacheControl: pointer(privateNoStore)},
	}, nil
}

func (s *server) ImmersionFeatureAccessGrant(
	ctx context.Context,
	request openapi.ImmersionFeatureAccessGrantRequestObject,
) (openapi.ImmersionFeatureAccessGrantResponseObject, error) {
	result, err := s.application.FeatureAccessGrant(ctx, string(request.FlagKey), request.UserId)
	if err != nil {
		s.logOperationError(ctx, "grant feature access", err)
		return nil, err
	}
	return openapi.ImmersionFeatureAccessGrant200JSONResponse{
		Body:    featureAccessResponse(result),
		Headers: openapi.ImmersionFeatureAccessGrant200ResponseHeaders{CacheControl: pointer(privateNoStore)},
	}, nil
}

func (s *server) ImmersionFeatureAccessRevoke(
	ctx context.Context,
	request openapi.ImmersionFeatureAccessRevokeRequestObject,
) (openapi.ImmersionFeatureAccessRevokeResponseObject, error) {
	result, err := s.application.FeatureAccessRevoke(ctx, string(request.FlagKey), request.UserId)
	if err != nil {
		s.logOperationError(ctx, "revoke feature access", err)
		return nil, err
	}
	return openapi.ImmersionFeatureAccessRevoke200JSONResponse{
		Body:    featureAccessResponse(result),
		Headers: openapi.ImmersionFeatureAccessRevoke200ResponseHeaders{CacheControl: pointer(privateNoStore)},
	}, nil
}

func featureAccessResponse(result app.FeatureAccessState) openapi.ImmersionFeatureAccessResponse {
	return openapi.ImmersionFeatureAccessResponse{
		Enabled:     result.Enabled,
		Changed:     result.Changed,
		Environment: openapi.ImmersionFeatureAccessResponseEnvironment(result.Environment),
		Revision:    result.Revision,
	}
}
