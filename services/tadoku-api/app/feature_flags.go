package app

import (
	"context"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type FeatureFlagDecisions = featureflags.PublicDecisions

func (a *Application) FeatureFlagDecisions(ctx context.Context) FeatureFlagDecisions {
	user := identity.FromContext(ctx)
	if user == nil {
		return a.featureFlags.EvaluatePublicForSubject(ctx, "")
	}
	return a.featureFlags.EvaluatePublicForSubject(ctx, user.Subject)
}
