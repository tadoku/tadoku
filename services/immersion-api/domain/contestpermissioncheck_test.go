package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockContestPermissionCheckRepository struct {
	count int32
	err   error
}

func (m *mockContestPermissionCheckRepository) GetContestsByUserCountForYear(ctx context.Context, now time.Time, userID uuid.UUID) (int32, error) {
	return m.count, m.err
}

type mockContestPermissionCheckKratos struct {
	traits *domain.UserTraits
	err    error
}

func (m *mockContestPermissionCheckKratos) FetchIdentity(ctx context.Context, id uuid.UUID) (*domain.UserTraits, error) {
	return m.traits, m.err
}

func TestContestPermissionCheck_Execute(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	clock := &mockClock{now: now}

	t.Run("allows admin", func(t *testing.T) {
		repo := &mockContestPermissionCheckRepository{}
		kratos := &mockContestPermissionCheckKratos{}
		svc := domain.NewContestPermissionCheck(repo, kratos, clock)

		ctx := ctxWithAdminSubject(userID.String())

		err := svc.Execute(ctx)

		assert.NoError(t, err)
	})

	t.Run("allows user with old account and under limit", func(t *testing.T) {
		repo := &mockContestPermissionCheckRepository{count: 5}
		kratos := &mockContestPermissionCheckKratos{
			traits: &domain.UserTraits{
				CreatedAt: now.AddDate(0, -2, 0), // 2 months ago
			},
		}
		svc := domain.NewContestPermissionCheck(repo, kratos, clock)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx)

		assert.NoError(t, err)
	})

	t.Run("rejects new account", func(t *testing.T) {
		repo := &mockContestPermissionCheckRepository{count: 0}
		kratos := &mockContestPermissionCheckKratos{
			traits: &domain.UserTraits{
				CreatedAt: now.AddDate(0, 0, -15), // 15 days ago
			},
		}
		svc := domain.NewContestPermissionCheck(repo, kratos, clock)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account too young")
	})

	t.Run("rejects user at yearly limit", func(t *testing.T) {
		repo := &mockContestPermissionCheckRepository{count: 12}
		kratos := &mockContestPermissionCheckKratos{
			traits: &domain.UserTraits{
				CreatedAt: now.AddDate(-1, 0, 0), // 1 year ago
			},
		}
		svc := domain.NewContestPermissionCheck(repo, kratos, clock)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx)

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("returns error when kratos fails", func(t *testing.T) {
		repo := &mockContestPermissionCheckRepository{}
		kratos := &mockContestPermissionCheckKratos{
			err: errors.New("identity not found"),
		}
		svc := domain.NewContestPermissionCheck(repo, kratos, clock)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not check permission")
	})
}
