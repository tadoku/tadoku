package app

import (
	"context"

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

func (a *Application) FeatureAccessGrant(
	ctx context.Context,
	flagKey string,
	targetUserID uuid.UUID,
) (state FeatureAccessState, err error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return featureflags.State{}, err
	}
	actorID, err := identity.RequireActorID(ctx)
	if err != nil {
		return featureflags.State{}, err
	}

	defer func() {
		event := featureAccessEvent("feature_access_grant", actorID, flagKey, targetUserID, state, err)
		a.audit.RecordOutcome(ctx, event, err)
	}()

	return a.featureFlags.Grant(ctx, flagKey, targetUserID)
}

func (a *Application) FeatureAccessRevoke(
	ctx context.Context,
	flagKey string,
	targetUserID uuid.UUID,
) (state FeatureAccessState, err error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return featureflags.State{}, err
	}
	actorID, err := identity.RequireActorID(ctx)
	if err != nil {
		return featureflags.State{}, err
	}

	defer func() {
		event := featureAccessEvent("feature_access_revoke", actorID, flagKey, targetUserID, state, err)
		a.audit.RecordOutcome(ctx, event, err)
	}()

	return a.featureFlags.Revoke(ctx, flagKey, targetUserID)
}

func featureAccessEvent(
	action string,
	actorID uuid.UUID,
	flagKey string,
	targetUserID uuid.UUID,
	state FeatureAccessState,
	err error,
) audit.Event {
	metadata := map[string]any{
		"target_user_id": targetUserID.String(),
		"flag_key":       flagKey,
	}
	if err == nil {
		metadata["environment"] = state.Environment
		metadata["changed"] = state.Changed
		metadata["resulting_revision"] = state.Revision
	}

	return audit.Event{
		ActorID:  actorID,
		Action:   action,
		Metadata: metadata,
	}
}
