package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) AuthzPermissionCheck(
	ctx context.Context,
	request openapi.AuthzPermissionCheckRequestObject,
) (openapi.AuthzPermissionCheckResponseObject, error) {
	allowed, err := s.application.CheckPermission(ctx, app.PermissionCheckParameters{
		Namespace: request.Body.Namespace,
		Object:    request.Body.Object,
		Relation:  request.Body.Relation,
	})
	if err != nil {
		s.logOperationError(ctx, "check permission", err)
		return nil, err
	}

	return openapi.AuthzPermissionCheck200JSONResponse{Allowed: allowed}, nil
}
