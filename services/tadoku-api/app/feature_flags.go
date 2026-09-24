package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type FeatureFlagDecisions = featureflags.PublicDecisions
type FeatureAccessState = featureflags.State

func (a *Application) FeatureFlagDecisions(ctx context.Context) FeatureFlagDecisions {
	user := identity.FromContext(ctx)
	if user == nil {
		return a.featureFlags.EvaluatePublic(ctx, "")
	}
	return a.featureFlags.EvaluatePublic(ctx, user.Subject)
}

func (a *Application) FeatureAccessGet(ctx context.Context, flagKey string, targetUserID uuid.UUID) (FeatureAccessState, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return featureflags.State{}, err
	}
	return a.featureFlags.GetNamedUserAccess(ctx, flagKey, targetUserID)
}

func (a *Application) FeatureAccessGrant(ctx context.Context, flagKey string, targetUserID uuid.UUID) (FeatureAccessState, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return featureflags.State{}, err
	}
	if err := a.featureFlags.ValidateRequest(flagKey, targetUserID); err != nil {
		return featureflags.State{}, err
	}
	actorID, err := identity.RequireCallerID(ctx)
	if err != nil {
		return featureflags.State{}, err
	}

	result, err := a.featureFlags.Grant(ctx, flagKey, targetUserID)
	if err != nil {
		return featureflags.State{}, err
	}

	if err := a.audit.Record(ctx, audit.Event{
		ActorID: actorID,
		Action:  "feature_access_grant",
		Metadata: map[string]any{
			"target_user_id":     targetUserID.String(),
			"flag_key":           flagKey,
			"environment":        result.Environment,
			"changed":            result.Changed,
			"resulting_revision": result.Revision,
		},
	}); err != nil {
		return featureflags.State{}, fmt.Errorf("create audit: %w", err)
	}

	return result, nil
}

func (a *Application) FeatureAccessRevoke(ctx context.Context, flagKey string, targetUserID uuid.UUID) (FeatureAccessState, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return featureflags.State{}, err
	}
	if err := a.featureFlags.ValidateRequest(flagKey, targetUserID); err != nil {
		return featureflags.State{}, err
	}
	actorID, err := identity.RequireCallerID(ctx)
	if err != nil {
		return featureflags.State{}, err
	}

	result, err := a.featureFlags.Revoke(ctx, flagKey, targetUserID)
	if err != nil {
		return featureflags.State{}, err
	}

	if err := a.audit.Record(ctx, audit.Event{
		ActorID: actorID,
		Action:  "feature_access_revoke",
		Metadata: map[string]any{
			"target_user_id":     targetUserID.String(),
			"flag_key":           flagKey,
			"environment":        result.Environment,
			"changed":            result.Changed,
			"resulting_revision": result.Revision,
		},
	}); err != nil {
		return featureflags.State{}, fmt.Errorf("create audit: %w", err)
	}

	return result, nil
}
