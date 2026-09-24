package http

import (
	"context"

	callbackopenapi "github.com/tadoku/tadoku/services/tadoku-api/generated/openapi/callback"
)

func (s *server) AuthzProxyProxyAdminCheck(
	ctx context.Context,
	request callbackopenapi.AuthzProxyProxyAdminCheckRequestObject,
) (callbackopenapi.AuthzProxyProxyAdminCheckResponseObject, error) {
	allowed, err := s.application.ProxyAdminCheck(ctx, request.Body.Subject)
	if err != nil {
		s.logOperationError(ctx, "check proxy administrator", err)
		return nil, err
	}
	if !allowed {
		return callbackopenapi.AuthzProxyProxyAdminCheck403Response{}, nil
	}

	return callbackopenapi.AuthzProxyProxyAdminCheck200Response{}, nil
}
