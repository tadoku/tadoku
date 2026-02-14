package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockContestModerationDetachLogRepository struct {
	contest        *domain.ContestView
	findContestErr error
	log            *domain.Log
	findLogErr     error
	detachErr      error
	detachCalled   bool
	detachUserID   uuid.UUID
}

func (m *mockContestModerationDetachLogRepository) FindContestByID(ctx context.Context, req *domain.ContestFindRequest) (*domain.ContestView, error) {
	return m.contest, m.findContestErr
}

func (m *mockContestModerationDetachLogRepository) FindLogByID(ctx context.Context, req *domain.LogFindRequest) (*domain.Log, error) {
	return m.log, m.findLogErr
}

func (m *mockContestModerationDetachLogRepository) DetachLogFromContest(ctx context.Context, req *domain.ContestModerationDetachLogRequest, userID uuid.UUID) error {
	m.detachCalled = true
	m.detachUserID = userID
	return m.detachErr
}

func TestContestModerationDetachLog_Execute(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	contestID := uuid.New()
	logID := uuid.New()
	now := time.Now()

	contest := &domain.ContestView{
		ID:          contestID,
		OwnerUserID: userID,
		Title:       "Test Contest",
	}

	log := &domain.Log{
		ID:        logID,
		UserID:    otherUserID,
		CreatedAt: now,
	}

	t.Run("returns unauthorized for guest", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewContestModerationDetachLog(repo, store, rebuildRepo)

		ctx := ctxWithGuest()

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.detachCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewContestModerationDetachLog(repo, store, rebuildRepo)

		err := svc.Execute(context.Background(), &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.detachCalled)
	})

	t.Run("returns error when contest not found", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{
			findContestErr: domain.ErrNotFound,
		}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewContestModerationDetachLog(repo, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		assert.Error(t, err)
		assert.False(t, repo.detachCalled)
	})

	t.Run("returns forbidden for non-owner non-admin", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{
			contest: contest,
		}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewContestModerationDetachLog(repo, store, rebuildRepo)

		ctx := ctxWithUserSubject(otherUserID.String()) // Not the owner

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		assert.ErrorIs(t, err, domain.ErrForbidden)
		assert.False(t, repo.detachCalled)
	})

	t.Run("returns error when log not found", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{
			contest:    contest,
			findLogErr: domain.ErrNotFound,
		}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewContestModerationDetachLog(repo, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		assert.Error(t, err)
		assert.False(t, repo.detachCalled)
	})

	t.Run("allows contest owner to detach log", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{
			contest: contest,
			log:     log,
		}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewContestModerationDetachLog(repo, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String()) // Contest owner

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		require.NoError(t, err)
		assert.True(t, repo.detachCalled)
		assert.Equal(t, userID, repo.detachUserID)
	})

	t.Run("allows admin to detach log from any contest", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{
			contest: contest,
			log:     log,
		}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewContestModerationDetachLog(repo, store, rebuildRepo)

		adminID := uuid.New()
		ctx := ctxWithAdminSubject(adminID.String())

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "admin action",
		})

		require.NoError(t, err)
		assert.True(t, repo.detachCalled)
		assert.Equal(t, adminID, repo.detachUserID)
	})

	t.Run("returns error when detach fails", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{
			contest:   contest,
			log:       log,
			detachErr: errors.New("database error"),
		}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewContestModerationDetachLog(repo, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		assert.Error(t, err)
		assert.True(t, repo.detachCalled)
	})
}

func TestContestModerationDetachLog_LeaderboardUpdates(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	contestID := uuid.New()
	logID := uuid.New()
	now := time.Now()

	contest := &domain.ContestView{
		ID:          contestID,
		OwnerUserID: userID,
		Title:       "Test Contest",
	}

	log := &domain.Log{
		ID:        logID,
		UserID:    otherUserID,
		CreatedAt: now,
	}

	t.Run("rebuilds contest leaderboard after successful detach", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{
			contest: contest,
			log:     log,
		}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewContestModerationDetachLog(repo, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		require.NoError(t, err)
		assert.True(t, repo.detachCalled)
		require.Len(t, store.rebuildContestCalls, 1)
		assert.Equal(t, contestID, store.rebuildContestCalls[0].ContestID)
	})

	t.Run("leaderboard rebuild error does not fail the detach", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{
			contest: contest,
			log:     log,
		}
		store := &mockLeaderboardStore{
			rebuildContestErr: errors.New("redis unavailable"),
		}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewContestModerationDetachLog(repo, store, rebuildRepo)

		ctx := ctxWithAdminSubject(uuid.New().String())

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "admin action",
		})

		require.NoError(t, err)
		assert.True(t, repo.detachCalled)
	})

	t.Run("leaderboard fetch scores error does not fail the detach", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{
			contest: contest,
			log:     log,
		}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{
			contestErr: errors.New("database timeout"),
		}
		svc := domain.NewContestModerationDetachLog(repo, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		require.NoError(t, err)
		assert.True(t, repo.detachCalled)
		assert.Empty(t, store.rebuildContestCalls)
	})
}
