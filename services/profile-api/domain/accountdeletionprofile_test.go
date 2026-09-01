package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/profile-api/domain"
)

type accountDeletionProfileRepositoryMock struct {
	identityID uuid.UUID
	err        error
}

func (m *accountDeletionProfileRepositoryMock) DeleteProfile(_ context.Context, identityID uuid.UUID) error {
	m.identityID = identityID
	return m.err
}

type accountDeletionProfileCacheMock struct {
	identityIDs []uuid.UUID
}

func (m *accountDeletionProfileCacheMock) SuppressAndEvict(identityID uuid.UUID) {
	m.identityIDs = append(m.identityIDs, identityID)
}

func TestAccountDeletionProfileDeletesThenSuppresses(t *testing.T) {
	repo := &accountDeletionProfileRepositoryMock{}
	userCache := &accountDeletionProfileCacheMock{}
	service := domain.NewAccountDeletionProfile(repo, userCache)
	identityID := uuid.New()

	err := service.Execute(context.Background(), identityID)

	require.NoError(t, err)
	assert.Equal(t, identityID, repo.identityID)
	assert.Equal(t, []uuid.UUID{identityID}, userCache.identityIDs)
}

func TestAccountDeletionProfileIsIdempotent(t *testing.T) {
	repo := &accountDeletionProfileRepositoryMock{}
	userCache := &accountDeletionProfileCacheMock{}
	service := domain.NewAccountDeletionProfile(repo, userCache)
	identityID := uuid.New()

	require.NoError(t, service.Execute(context.Background(), identityID))
	require.NoError(t, service.Execute(context.Background(), identityID))
	assert.Equal(t, []uuid.UUID{identityID, identityID}, userCache.identityIDs)
}

func TestAccountDeletionProfileValidatesIdentity(t *testing.T) {
	repo := &accountDeletionProfileRepositoryMock{}
	userCache := &accountDeletionProfileCacheMock{}
	service := domain.NewAccountDeletionProfile(repo, userCache)

	err := service.Execute(context.Background(), uuid.Nil)

	assert.ErrorIs(t, err, domain.ErrRequestInvalid)
	assert.Equal(t, uuid.Nil, repo.identityID)
	assert.Empty(t, userCache.identityIDs)
}

func TestAccountDeletionProfileDoesNotEvictWhenDeletionFails(t *testing.T) {
	repoErr := errors.New("database unavailable")
	repo := &accountDeletionProfileRepositoryMock{err: repoErr}
	userCache := &accountDeletionProfileCacheMock{}
	service := domain.NewAccountDeletionProfile(repo, userCache)

	err := service.Execute(context.Background(), uuid.New())

	assert.ErrorIs(t, err, repoErr)
	assert.Empty(t, userCache.identityIDs)
}
