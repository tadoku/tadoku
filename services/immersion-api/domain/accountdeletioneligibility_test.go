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

type accountDeletionEligibilityRepositoryMock struct {
	userID         uuid.UUID
	checkedAt      time.Time
	availableAfter *time.Time
	err            error
}

func (m *accountDeletionEligibilityRepositoryMock) FindRunningOwnedContestAvailableAfter(_ context.Context, userID uuid.UUID, checkedAt time.Time) (*time.Time, error) {
	m.userID = userID
	m.checkedAt = checkedAt
	return m.availableAfter, m.err
}

func TestAccountDeletionEligibilityReturnsRepositoryResult(t *testing.T) {
	clockTime := time.Date(2026, 9, 1, 14, 0, 0, 0, time.FixedZone("CEST", 2*60*60))
	availableAfter := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	repo := &accountDeletionEligibilityRepositoryMock{availableAfter: &availableAfter}
	service := domain.NewAccountDeletionEligibility(repo, commondomain.NewMockClock(clockTime))
	userID := uuid.New()
	req := &domain.AccountDeletionEligibilityRequest{UserID: userID}

	result, err := service.Execute(serviceContext("profile-api"), req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Eligible())
	assert.Equal(t, availableAfter, *result.AvailableAfter)
	assert.Equal(t, userID, repo.userID)
	assert.Equal(t, clockTime.UTC(), repo.checkedAt)
	assert.Equal(t, clockTime.UTC(), req.CheckedAt())
}

func TestAccountDeletionEligibilityReturnsEligibleWithoutBlocker(t *testing.T) {
	service := domain.NewAccountDeletionEligibility(&accountDeletionEligibilityRepositoryMock{}, commondomain.NewMockClock(time.Now()))
	result, err := service.Execute(serviceContext("profile-api"), &domain.AccountDeletionEligibilityRequest{UserID: uuid.New()})
	require.NoError(t, err)
	assert.True(t, result.Eligible())
}

func TestAccountDeletionEligibilityValidatesCallerAndRequest(t *testing.T) {
	valid := &domain.AccountDeletionEligibilityRequest{UserID: uuid.New()}
	for name, tc := range map[string]struct {
		ctx context.Context
		req *domain.AccountDeletionEligibilityRequest
		err error
	}{
		"user caller":   {context.Background(), valid, domain.ErrUnauthorized},
		"wrong service": {serviceContext("content-api"), valid, domain.ErrForbidden},
		"nil request":   {serviceContext("profile-api"), nil, domain.ErrRequestInvalid},
		"missing user":  {serviceContext("profile-api"), &domain.AccountDeletionEligibilityRequest{}, domain.ErrRequestInvalid},
	} {
		t.Run(name, func(t *testing.T) {
			service := domain.NewAccountDeletionEligibility(&accountDeletionEligibilityRepositoryMock{}, commondomain.NewMockClock(time.Now()))
			_, err := service.Execute(tc.ctx, tc.req)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestAccountDeletionEligibilityWrapsRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	service := domain.NewAccountDeletionEligibility(
		&accountDeletionEligibilityRepositoryMock{err: repoErr},
		commondomain.NewMockClock(time.Now()),
	)

	_, err := service.Execute(serviceContext("profile-api"), &domain.AccountDeletionEligibilityRequest{UserID: uuid.New()})
	assert.ErrorIs(t, err, repoErr)
}
