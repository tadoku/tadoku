package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) AuthzRoleGet(
	ctx context.Context,
	_ openapi.AuthzRoleGetRequestObject,
) (openapi.AuthzRoleGetResponseObject, error) {
	role, err := s.application.CurrentUserRole(ctx)
	if err != nil {
		s.logOperationError(ctx, "get current user role", err)
		return nil, err
	}

	return openapi.AuthzRoleGet200JSONResponse{
		Role: openapi.AuthzUserRoleRole(role),
	}, nil
}
