package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

const privateNoStore = "private, no-store"

func (s *server) ImmersionFeatureFlagDecisions(
	ctx context.Context,
	_ openapi.ImmersionFeatureFlagDecisionsRequestObject,
) (openapi.ImmersionFeatureFlagDecisionsResponseObject, error) {
	decisions := s.application.FeatureFlagDecisions(ctx)
	return openapi.ImmersionFeatureFlagDecisions200JSONResponse{
		Body: openapi.ImmersionFeatureFlagDecisionsResponse{
			Decisions: openapi.ImmersionFeatureFlagDecisions{
				ReleaseWebV3: decisions.ReleaseWebV3,
			},
		},
		Headers: openapi.ImmersionFeatureFlagDecisions200ResponseHeaders{
			CacheControl: pointer(privateNoStore),
		},
	}, nil
}

func pointer[T any](value T) *T {
	return &value
}
