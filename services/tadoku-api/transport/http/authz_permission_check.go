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
	parameters := app.PermissionCheckParameters{}
	if request.Body != nil {
		parameters = app.PermissionCheckParameters{
			Namespace: request.Body.Namespace,
			Object:    request.Body.Object,
			Relation:  request.Body.Relation,
		}
	}

	allowed, err := s.application.CheckPermission(ctx, parameters)
	if err != nil {
		s.logOperationError(ctx, "check permission", err)
		return nil, err
	}

	return openapi.AuthzPermissionCheck200JSONResponse{Allowed: allowed}, nil
}
