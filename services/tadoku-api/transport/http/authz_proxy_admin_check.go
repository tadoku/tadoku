package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) AuthzProxyProxyAdminCheck(
	ctx context.Context,
	request openapi.AuthzProxyProxyAdminCheckRequestObject,
) (openapi.AuthzProxyProxyAdminCheckResponseObject, error) {
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
		return openapi.AuthzProxyProxyAdminCheck403Response{}, nil
	}

	return openapi.AuthzProxyProxyAdminCheck200Response{}, nil
}
