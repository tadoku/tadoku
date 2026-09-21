package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionContestCreatePermissionCheck(
	ctx context.Context,
	_ openapi.ImmersionContestCreatePermissionCheckRequestObject,
) (openapi.ImmersionContestCreatePermissionCheckResponseObject, error) {
	err := s.application.CheckContestCreatePermission(ctx)
	if err != nil {
		s.logOperationError(ctx, "check contest create permission", err)
		return nil, err
	}
	return openapi.ImmersionContestCreatePermissionCheck200Response{}, nil
}
