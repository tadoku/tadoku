package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type accountDeletionLockRepositoryMock struct {
	req *domain.AccountDeletionLockRequest
	err error
}

func (m *accountDeletionLockRepositoryMock) LockAccountForDeletion(_ context.Context, req *domain.AccountDeletionLockRequest) error {
	m.req = req
	return m.err
}

func serviceContext(name string) context.Context {
	return context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.ServiceIdentity{
		Subject:   "system:serviceaccount:tadoku:" + name,
		Name:      name,
		Namespace: "tadoku",
		Audience:  []string{"immersion-api"},
	})
}

func TestAccountDeletionLockExecute(t *testing.T) {
	lockedAt := time.Date(2026, 8, 29, 17, 0, 0, 0, time.UTC)
	repo := &accountDeletionLockRepositoryMock{}
	service := domain.NewAccountDeletionLock(repo, commondomain.NewMockClock(lockedAt))
	req := &domain.AccountDeletionLockRequest{UserID: uuid.New(), RequestID: uuid.New()}

	err := service.Execute(serviceContext("profile-api"), req)

	require.NoError(t, err)
	assert.Same(t, req, repo.req)
	assert.Equal(t, lockedAt, repo.req.LockedAt())
}

func TestAccountDeletionLockRejectsUserAndUnexpectedService(t *testing.T) {
	repo := &accountDeletionLockRepositoryMock{}
	service := domain.NewAccountDeletionLock(repo, commondomain.NewMockClock(time.Now()))
	req := &domain.AccountDeletionLockRequest{UserID: uuid.New(), RequestID: uuid.New()}

	assert.ErrorIs(t, service.Execute(context.Background(), req), domain.ErrUnauthorized)
	assert.ErrorIs(t, service.Execute(serviceContext("content-api"), req), domain.ErrForbidden)
	assert.Nil(t, repo.req)
}

func TestAccountDeletionLockValidatesIdentifiers(t *testing.T) {
	repo := &accountDeletionLockRepositoryMock{}
	service := domain.NewAccountDeletionLock(repo, commondomain.NewMockClock(time.Now()))

	for name, req := range map[string]*domain.AccountDeletionLockRequest{
		"missing request":    nil,
		"missing user id":    {RequestID: uuid.New()},
		"missing request id": {UserID: uuid.New()},
	} {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, service.Execute(serviceContext("profile-api"), req), domain.ErrRequestInvalid)
		})
	}
	assert.Nil(t, repo.req)
}

func TestAccountDeletionLockWrapsRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	service := domain.NewAccountDeletionLock(
		&accountDeletionLockRepositoryMock{err: repoErr},
		commondomain.NewMockClock(time.Now()),
	)

	err := service.Execute(serviceContext("profile-api"), &domain.AccountDeletionLockRequest{
		UserID:    uuid.New(),
		RequestID: uuid.New(),
	})

	assert.ErrorIs(t, err, repoErr)
}
