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
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewContestModerationDetachLog(repo, updater)

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
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewContestModerationDetachLog(repo, updater)

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
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewContestModerationDetachLog(repo, updater)

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
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewContestModerationDetachLog(repo, updater)

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
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewContestModerationDetachLog(repo, updater)

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
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewContestModerationDetachLog(repo, updater)

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
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewContestModerationDetachLog(repo, updater)

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
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewContestModerationDetachLog(repo, updater)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		assert.Error(t, err)
		assert.True(t, repo.detachCalled)
	})

	t.Run("updates user contest score for the log owner after detach", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{
			contest: contest,
			log:     log,
		}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewContestModerationDetachLog(repo, updater)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		require.NoError(t, err)
		require.Len(t, updater.updateContestCalls, 1)
		assert.Equal(t, contestID, updater.updateContestCalls[0].ContestID)
		assert.Equal(t, otherUserID, updater.updateContestCalls[0].UserID)
	})

	t.Run("does not update leaderboard when detach fails", func(t *testing.T) {
		repo := &mockContestModerationDetachLogRepository{
			contest:   contest,
			log:       log,
			detachErr: errors.New("database error"),
		}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewContestModerationDetachLog(repo, updater)

		ctx := ctxWithUserSubject(userID.String())

		_ = svc.Execute(ctx, &domain.ContestModerationDetachLogRequest{
			ContestID: contestID,
			LogID:     logID,
			Reason:    "test",
		})

		assert.Empty(t, updater.updateContestCalls)
	})
}
