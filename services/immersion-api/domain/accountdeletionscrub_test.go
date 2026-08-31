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

type accountDeletionScrubRepositoryMock struct {
	req *domain.AccountDeletionScrubRequest
	err error
}

func (m *accountDeletionScrubRepositoryMock) ScrubAccount(_ context.Context, req *domain.AccountDeletionScrubRequest) error {
	m.req = req
	return m.err
}

func accountDeletionServiceContext(name string) context.Context {
	return context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.ServiceIdentity{Name: name})
}

func TestAccountDeletionScrubExecutesWithDomainTime(t *testing.T) {
	clockTime := time.Date(2026, 8, 31, 14, 30, 0, 0, time.FixedZone("CEST", 2*60*60))
	now := clockTime.UTC()
	repo := &accountDeletionScrubRepositoryMock{}
	service := domain.NewAccountDeletionScrub(repo, commondomain.NewMockClock(clockTime))
	req := &domain.AccountDeletionScrubRequest{UserID: uuid.New(), RequestID: uuid.New()}

	require.NoError(t, service.Execute(accountDeletionServiceContext("profile-api"), req))
	require.Same(t, req, repo.req)
	assert.Equal(t, now, req.DeletedAt())
}

func TestAccountDeletionScrubValidatesCallerAndRequest(t *testing.T) {
	valid := &domain.AccountDeletionScrubRequest{UserID: uuid.New(), RequestID: uuid.New()}
	for name, tc := range map[string]struct {
		ctx context.Context
		req *domain.AccountDeletionScrubRequest
		err error
	}{
		"user caller":     {context.Background(), valid, domain.ErrUnauthorized},
		"wrong service":   {accountDeletionServiceContext("content-api"), valid, domain.ErrForbidden},
		"nil request":     {accountDeletionServiceContext("profile-api"), nil, domain.ErrRequestInvalid},
		"missing user":    {accountDeletionServiceContext("profile-api"), &domain.AccountDeletionScrubRequest{RequestID: uuid.New()}, domain.ErrRequestInvalid},
		"missing request": {accountDeletionServiceContext("profile-api"), &domain.AccountDeletionScrubRequest{UserID: uuid.New()}, domain.ErrRequestInvalid},
	} {
		t.Run(name, func(t *testing.T) {
			service := domain.NewAccountDeletionScrub(&accountDeletionScrubRepositoryMock{}, commondomain.NewMockClock(time.Now()))
			assert.ErrorIs(t, service.Execute(tc.ctx, tc.req), tc.err)
		})
	}
}

func TestAccountDeletionScrubWrapsRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	service := domain.NewAccountDeletionScrub(&accountDeletionScrubRepositoryMock{err: repoErr}, commondomain.NewMockClock(time.Now()))
	err := service.Execute(accountDeletionServiceContext("profile-api"), &domain.AccountDeletionScrubRequest{UserID: uuid.New(), RequestID: uuid.New()})
	assert.ErrorIs(t, err, repoErr)
}
