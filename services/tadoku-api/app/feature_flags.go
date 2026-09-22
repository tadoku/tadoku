package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/featureaccess"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type FeatureFlagDecisions = featureaccess.PublicDecisions
type FeatureAccessState = featureaccess.State

var ErrFeatureAccessUnavailable = featureaccess.ErrUnavailable

func (a *Application) FeatureFlagDecisions(ctx context.Context) FeatureFlagDecisions {
	user := identity.FromContext(ctx)
	if user == nil {
		return a.featureFlags.EvaluatePublic(ctx, "")
	}
	return a.featureFlags.EvaluatePublic(ctx, user.Subject)
}

func (a *Application) FeatureAccessGet(ctx context.Context, flagKey string, targetUserID uuid.UUID) (FeatureAccessState, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return featureaccess.State{}, err
	}
	return a.featureFlags.GetNamedUserAccess(ctx, flagKey, targetUserID)
}

func (a *Application) FeatureAccessGrant(ctx context.Context, flagKey string, targetUserID uuid.UUID) (FeatureAccessState, error) {
	return a.setFeatureAccess(ctx, flagKey, targetUserID, true)
}

func (a *Application) FeatureAccessRevoke(ctx context.Context, flagKey string, targetUserID uuid.UUID) (FeatureAccessState, error) {
	return a.setFeatureAccess(ctx, flagKey, targetUserID, false)
}

func (a *Application) setFeatureAccess(ctx context.Context, flagKey string, targetUserID uuid.UUID, enabled bool) (FeatureAccessState, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return featureaccess.State{}, err
	}
	if err := a.featureFlags.ValidateRequest(flagKey, targetUserID); err != nil {
		return featureaccess.State{}, err
	}
	actorID, err := uuid.Parse(identity.FromContext(ctx).Subject)
	if err != nil || actorID == uuid.Nil {
		return featureaccess.State{}, errx.NewUnauthorizedError("unauthorized")
	}

	result, err := a.featureFlags.SetNamedUserAccess(ctx, flagKey, targetUserID, enabled)
	if err != nil {
		return featureaccess.State{}, err
	}

	action := "feature_access_revoke"
	if enabled {
		action = "feature_access_grant"
	}
	if err := a.audit.Record(ctx, audit.Event{
		ActorID: actorID,
		Action:  action,
		Metadata: map[string]any{
			"target_user_id":     targetUserID.String(),
			"flag_key":           flagKey,
			"environment":        result.Environment,
			"changed":            result.Changed,
			"resulting_revision": result.Revision,
		},
	}); err != nil {
		return featureaccess.State{}, fmt.Errorf("create audit: %w", err)
	}

	return result, nil
}
