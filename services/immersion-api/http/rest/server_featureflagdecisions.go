package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// FeatureFlagDecisions returns the allowlisted public decisions for the
// request identity.
// (GET /feature-flags)
func (s *Server) FeatureFlagDecisions(ctx echo.Context) error {
	requestContext := ctx.Request().Context()
	decisions := s.featureFlagDecisions.EvaluatePublic(
		requestContext,
		commondomain.ParseUserIdentity(requestContext),
	)

	ctx.Response().Header().Set(echo.HeaderCacheControl, "private, no-store")
	return ctx.JSON(http.StatusOK, openapi.FeatureFlagDecisionsResponse{
		Decisions: openapi.FeatureFlagDecisions{
			ReleaseLogEntryV2: decisions.ReleaseLogEntryV2,
		},
	})
}
