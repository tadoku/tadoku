package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockRegistrationListOngoingRepository struct {
	result *domain.ContestRegistrations
	err    error
}

func (m *mockRegistrationListOngoingRepository) FetchOngoingContestRegistrations(ctx context.Context, req *domain.RegistrationListOngoingRequest) (*domain.ContestRegistrations, error) {
	return m.result, m.err
}

type mockClock struct {
	now time.Time
}

func (m *mockClock) Now() time.Time {
	return m.now
}

func TestRegistrationListOngoing_Execute(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := &mockClock{now: now}

	t.Run("returns registrations for authenticated user", func(t *testing.T) {
		repo := &mockRegistrationListOngoingRepository{
			result: &domain.ContestRegistrations{
				Registrations: []domain.ContestRegistration{
					{ID: uuid.New(), UserID: userID},
				},
				TotalSize:     1,
				NextPageToken: "",
			},
		}
		svc := domain.NewRegistrationListOngoing(repo, clock)

		ctx := context.WithValue(context.Background(), commondomain.CtxSessionKey, &commondomain.SessionToken{
			Subject: userID.String(),
			Role:    commondomain.RoleUser,
		})

		result, err := svc.Execute(ctx)

		require.NoError(t, err)
		assert.Len(t, result.Registrations, 1)
	})

	t.Run("returns unauthorized for guests", func(t *testing.T) {
		repo := &mockRegistrationListOngoingRepository{}
		svc := domain.NewRegistrationListOngoing(repo, clock)

		ctx := context.WithValue(context.Background(), commondomain.CtxSessionKey, &commondomain.SessionToken{
			Subject: "",
			Role:    commondomain.RoleGuest,
		})

		result, err := svc.Execute(ctx)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})
}
