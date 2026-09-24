package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) AuthzRoleUpdate(
	ctx context.Context,
	request openapi.AuthzRoleUpdateRequestObject,
) (openapi.AuthzRoleUpdateResponseObject, error) {
	parameters := app.RoleUpdateParameters{
		UserID: request.Id,
		Role:   app.Role(request.Body.Role),
		Reason: request.Body.Reason,
	}

	if err := s.application.UpdateRole(ctx, parameters); err != nil {
		s.logOperationError(ctx, "update user role", err)
		return nil, err
	}

	return openapi.AuthzRoleUpdate200Response{}, nil
}
