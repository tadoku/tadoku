package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type FeatureFlagKey string

const FeatureFlagReleaseLogEntryV2 FeatureFlagKey = "release-log-entry-v2"

type FeatureAccessState struct {
	Enabled     bool
	Changed     bool
	Environment string
	Revision    string
}

type FeatureAccessStore interface {
	GetNamedUserAccess(ctx context.Context, flagKey FeatureFlagKey, targetUserID uuid.UUID) (FeatureAccessState, error)
	SetNamedUserAccess(ctx context.Context, flagKey FeatureFlagKey, targetUserID uuid.UUID, enabled bool) (FeatureAccessState, error)
}

type FeatureAccessAuditRepository interface {
	CreateModerationAuditLog(ctx context.Context, req *ModerationAuditLogCreateRequest) error
}

type FeatureAccess struct {
	store FeatureAccessStore
	audit FeatureAccessAuditRepository
}

func NewFeatureAccess(store FeatureAccessStore, audit FeatureAccessAuditRepository) *FeatureAccess {
	return &FeatureAccess{store: store, audit: audit}
}

func (s *FeatureAccess) Get(ctx context.Context, flagKey string, targetUserID uuid.UUID) (FeatureAccessState, error) {
	key, err := validateFeatureAccessRequest(ctx, flagKey, targetUserID)
	if err != nil {
		return FeatureAccessState{}, err
	}

	result, err := s.store.GetNamedUserAccess(ctx, key, targetUserID)
	if err != nil {
		return FeatureAccessState{}, fmt.Errorf("could not fetch feature access: %w", err)
	}
	return result, nil
}

func (s *FeatureAccess) Grant(ctx context.Context, flagKey string, targetUserID uuid.UUID) (FeatureAccessState, error) {
	return s.set(ctx, flagKey, targetUserID, true)
}

func (s *FeatureAccess) Revoke(ctx context.Context, flagKey string, targetUserID uuid.UUID) (FeatureAccessState, error) {
	return s.set(ctx, flagKey, targetUserID, false)
}

func (s *FeatureAccess) set(ctx context.Context, flagKey string, targetUserID uuid.UUID, enabled bool) (FeatureAccessState, error) {
	key, err := validateFeatureAccessRequest(ctx, flagKey, targetUserID)
	if err != nil {
		return FeatureAccessState{}, err
	}

	actor := commondomain.ParseUserIdentity(ctx)
	if actor == nil {
		return FeatureAccessState{}, ErrUnauthorized
	}
	actorID, err := uuid.Parse(actor.Subject)
	if err != nil || actorID == uuid.Nil {
		return FeatureAccessState{}, ErrUnauthorized
	}

	result, err := s.store.SetNamedUserAccess(ctx, key, targetUserID, enabled)
	if err != nil {
		return FeatureAccessState{}, fmt.Errorf("could not update feature access: %w", err)
	}

	// Flipt and Postgres cannot share a transaction. Audit immediately after
	// Flipt confirms its Git revision; an audit failure is returned to the
	// caller even though that external revision may already be committed.
	action := "feature_access_revoke"
	if enabled {
		action = "feature_access_grant"
	}
	if err := s.audit.CreateModerationAuditLog(ctx, &ModerationAuditLogCreateRequest{
		ModeratorUserID: actorID,
		Action:          action,
		Metadata: map[string]any{
			"target_user_id":     targetUserID.String(),
			"flag_key":           string(key),
			"environment":        result.Environment,
			"changed":            result.Changed,
			"resulting_revision": result.Revision,
		},
	}); err != nil {
		return FeatureAccessState{}, fmt.Errorf("could not persist feature access audit: %w", err)
	}

	return result, nil
}

func validateFeatureAccessRequest(ctx context.Context, flagKey string, targetUserID uuid.UUID) (FeatureFlagKey, error) {
	if err := requireAdmin(ctx); err != nil {
		return "", err
	}
	if targetUserID == uuid.Nil {
		return "", fmt.Errorf("%w: target user ID must not be nil", ErrRequestInvalid)
	}
	key := FeatureFlagKey(flagKey)
	if key != FeatureFlagReleaseLogEntryV2 {
		return "", fmt.Errorf("%w: feature flag is not supported", ErrRequestInvalid)
	}
	return key, nil
}
