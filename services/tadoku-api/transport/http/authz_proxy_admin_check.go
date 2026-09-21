package http

import (
	"context"

	"github.com/google/uuid"
	callbackopenapi "github.com/tadoku/tadoku/services/tadoku-api/generated/openapi/callback"
)

func (s *server) AuthzProxyProxyAdminCheck(
	ctx context.Context,
	request callbackopenapi.AuthzProxyProxyAdminCheckRequestObject,
) (callbackopenapi.AuthzProxyProxyAdminCheckResponseObject, error) {
	subject := uuid.Nil
	if request.Body != nil {
		subject = request.Body.Subject
	}

	allowed, err := s.application.ProxyAdminCheck(ctx, subject)
	if err != nil {
		s.logOperationError(ctx, "check proxy administrator", err)
		return nil, err
	}
	if !allowed {
		return callbackopenapi.AuthzProxyProxyAdminCheck403Response{}, nil
	}

	return callbackopenapi.AuthzProxyProxyAdminCheck200Response{}, nil
}
