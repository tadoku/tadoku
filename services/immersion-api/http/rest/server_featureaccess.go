package rest

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

func (s *Server) FeatureAccessGet(ctx echo.Context, flagKey openapi.ManagedFeatureFlagKey, userID uuid.UUID) error {
	result, err := s.featureAccess.Get(ctx.Request().Context(), string(flagKey), userID)
	if err != nil {
		return handleFeatureAccessError(ctx, err)
	}
	return featureAccessResponse(ctx, result)
}

func (s *Server) FeatureAccessGrant(ctx echo.Context, flagKey openapi.ManagedFeatureFlagKey, userID uuid.UUID) error {
	result, err := s.featureAccess.Grant(ctx.Request().Context(), string(flagKey), userID)
	if err != nil {
		return handleFeatureAccessError(ctx, err)
	}
	return featureAccessResponse(ctx, result)
}

func (s *Server) FeatureAccessRevoke(ctx echo.Context, flagKey openapi.ManagedFeatureFlagKey, userID uuid.UUID) error {
	result, err := s.featureAccess.Revoke(ctx.Request().Context(), string(flagKey), userID)
	if err != nil {
		return handleFeatureAccessError(ctx, err)
	}
	return featureAccessResponse(ctx, result)
}

func featureAccessResponse(ctx echo.Context, result domain.FeatureAccessState) error {
	ctx.Response().Header().Set(echo.HeaderCacheControl, "private, no-store")
	return ctx.JSON(http.StatusOK, openapi.FeatureAccessResponse{
		Enabled:     result.Enabled,
		Changed:     result.Changed,
		Environment: openapi.FeatureAccessResponseEnvironment(result.Environment),
		Revision:    result.Revision,
	})
}

func handleFeatureAccessError(ctx echo.Context, err error) error {
	if handled, responseErr := handleCommonErrors(ctx, err); handled {
		return responseErr
	}
	ctx.Echo().Logger.Error(err)
	if errors.Is(err, domain.ErrFeatureAccessUnavailable) {
		return ctx.NoContent(http.StatusBadGateway)
	}
	return ctx.NoContent(http.StatusInternalServerError)
}
