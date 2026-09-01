package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type featureAccessStoreMock struct {
	state      domain.FeatureAccessState
	err        error
	setEnabled bool
	setCalls   int
	getCalls   int
}

func (m *featureAccessStoreMock) GetNamedUserAccess(context.Context, domain.FeatureFlagKey, uuid.UUID) (domain.FeatureAccessState, error) {
	m.getCalls++
	return m.state, m.err
}

func (m *featureAccessStoreMock) SetNamedUserAccess(_ context.Context, _ domain.FeatureFlagKey, _ uuid.UUID, enabled bool) (domain.FeatureAccessState, error) {
	m.setCalls++
	m.setEnabled = enabled
	return m.state, m.err
}

type featureAccessAuditMock struct {
	req *domain.ModerationAuditLogCreateRequest
	err error
}

func (m *featureAccessAuditMock) CreateModerationAuditLog(_ context.Context, req *domain.ModerationAuditLogCreateRequest) error {
	m.req = req
	return m.err
}

func TestFeatureAccessGetRequiresAdminAndValidatesAllowlist(t *testing.T) {
	targetID := uuid.New()
	store := &featureAccessStoreMock{state: domain.FeatureAccessState{
		Enabled: true, Environment: "production", Revision: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}}
	service := domain.NewFeatureAccess(store, &featureAccessAuditMock{})

	_, err := service.Get(ctxWithUserSubject(uuid.NewString()), "release-log-entry-v2", targetID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	assert.Zero(t, store.getCalls)

	_, err = service.Get(ctxWithAdminSubject(uuid.NewString()), "unknown", targetID)
	assert.ErrorIs(t, err, domain.ErrRequestInvalid)

	_, err = service.Get(ctxWithAdminSubject(uuid.NewString()), "release-log-entry-v2", uuid.Nil)
	assert.ErrorIs(t, err, domain.ErrRequestInvalid)

	result, err := service.Get(ctxWithAdminSubject(uuid.NewString()), "release-log-entry-v2", targetID)
	require.NoError(t, err)
	assert.True(t, result.Enabled)
	assert.Equal(t, "production", result.Environment)
}

func TestFeatureAccessGrantPersistsDurableAudit(t *testing.T) {
	actorID := uuid.New()
	targetID := uuid.New()
	store := &featureAccessStoreMock{state: domain.FeatureAccessState{
		Enabled: true, Changed: true, Environment: "production", Revision: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}}
	audit := &featureAccessAuditMock{}
	service := domain.NewFeatureAccess(store, audit)

	result, err := service.Grant(ctxWithAdminSubject(actorID.String()), "release-log-entry-v2", targetID)
	require.NoError(t, err)
	assert.True(t, result.Enabled)
	assert.True(t, store.setEnabled)
	require.NotNil(t, audit.req)
	assert.Equal(t, actorID, audit.req.ModeratorUserID)
	assert.Equal(t, "feature_access_grant", audit.req.Action)
	assert.Equal(t, targetID.String(), audit.req.Metadata["target_user_id"])
	assert.Equal(t, "release-log-entry-v2", audit.req.Metadata["flag_key"])
	assert.Equal(t, "production", audit.req.Metadata["environment"])
	assert.Equal(t, true, audit.req.Metadata["changed"])
	assert.Equal(t, store.state.Revision, audit.req.Metadata["resulting_revision"])
}

func TestFeatureAccessRevokeReportsAuditFailureAfterExternalMutation(t *testing.T) {
	store := &featureAccessStoreMock{state: domain.FeatureAccessState{
		Enabled: false, Environment: "local", Revision: "cccccccccccccccccccccccccccccccccccccccc",
	}}
	auditErr := errors.New("database unavailable")
	service := domain.NewFeatureAccess(store, &featureAccessAuditMock{err: auditErr})

	_, err := service.Revoke(ctxWithAdminSubject(uuid.NewString()), "release-log-entry-v2", uuid.New())
	assert.ErrorIs(t, err, auditErr)
	assert.False(t, store.setEnabled)
	// The Flipt update may already be committed when the independent Postgres audit fails.
	assert.Equal(t, 1, store.setCalls)
}

func TestFeatureAccessDoesNotAuditFailedExternalMutation(t *testing.T) {
	storeErr := errors.New("flipt unavailable")
	store := &featureAccessStoreMock{err: storeErr}
	audit := &featureAccessAuditMock{}
	service := domain.NewFeatureAccess(store, audit)

	_, err := service.Grant(ctxWithAdminSubject(uuid.NewString()), "release-log-entry-v2", uuid.New())
	assert.ErrorIs(t, err, storeErr)
	assert.Nil(t, audit.req)
}
