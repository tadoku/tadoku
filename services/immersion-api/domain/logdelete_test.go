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

type mockLogDeleteRepository struct {
	log          *domain.Log
	findErr      error
	deleteErr    error
	deleteCalled bool
}

func (m *mockLogDeleteRepository) FindLogByID(ctx context.Context, req *domain.LogFindRequest) (*domain.Log, error) {
	return m.log, m.findErr
}

func (m *mockLogDeleteRepository) DeleteLog(ctx context.Context, req *domain.LogDeleteRequest) error {
	m.deleteCalled = true
	return m.deleteErr
}

func TestLogDelete_Execute(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	logID := uuid.New()
	now := time.Now()

	t.Run("returns unauthorized for guest", func(t *testing.T) {
		repo := &mockLogDeleteRepository{}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock)

		ctx := ctxWithToken(&commondomain.UserIdentity{
			Role:    commondomain.RoleGuest,
			Subject: userID.String(),
		})

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.deleteCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		repo := &mockLogDeleteRepository{}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock)

		err := svc.Execute(context.Background(), &domain.LogDeleteRequest{LogID: logID})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.deleteCalled)
	})

	t.Run("allows owner to delete their own log", func(t *testing.T) {
		repo := &mockLogDeleteRepository{
			log: &domain.Log{ID: logID, UserID: userID},
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock)

		ctx := ctxWithToken(&commondomain.UserIdentity{
			Role:    commondomain.RoleUser,
			Subject: userID.String(),
		})

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)
	})

	t.Run("returns forbidden for non-owner non-admin", func(t *testing.T) {
		repo := &mockLogDeleteRepository{
			log: &domain.Log{ID: logID, UserID: otherUserID},
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock)

		ctx := ctxWithToken(&commondomain.UserIdentity{
			Role:    commondomain.RoleUser,
			Subject: userID.String(),
		})

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		assert.ErrorIs(t, err, domain.ErrForbidden)
		assert.False(t, repo.deleteCalled)
	})

	t.Run("allows admin to delete any log", func(t *testing.T) {
		repo := &mockLogDeleteRepository{
			log: &domain.Log{ID: logID, UserID: otherUserID},
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock)

		ctx := ctxWithToken(&commondomain.UserIdentity{
			Role:    commondomain.RoleAdmin,
			Subject: userID.String(),
		})

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)
	})

	t.Run("returns error when log not found", func(t *testing.T) {
		repo := &mockLogDeleteRepository{
			findErr: domain.ErrNotFound,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock)

		ctx := ctxWithToken(&commondomain.UserIdentity{
			Role:    commondomain.RoleUser,
			Subject: userID.String(),
		})

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		assert.Error(t, err)
		assert.False(t, repo.deleteCalled)
	})

	t.Run("returns error when delete fails", func(t *testing.T) {
		repo := &mockLogDeleteRepository{
			log:       &domain.Log{ID: logID, UserID: userID},
			deleteErr: errors.New("database error"),
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock)

		ctx := ctxWithToken(&commondomain.UserIdentity{
			Role:    commondomain.RoleUser,
			Subject: userID.String(),
		})

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		assert.Error(t, err)
		assert.True(t, repo.deleteCalled)
	})
}
