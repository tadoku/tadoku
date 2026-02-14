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
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewLogDelete(repo, clock, store, rebuildRepo)

		ctx := ctxWithGuest()

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.deleteCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		repo := &mockLogDeleteRepository{}
		clock := commondomain.NewMockClock(now)
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewLogDelete(repo, clock, store, rebuildRepo)

		err := svc.Execute(context.Background(), &domain.LogDeleteRequest{LogID: logID})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.deleteCalled)
	})

	t.Run("allows owner to delete their own log", func(t *testing.T) {
		repo := &mockLogDeleteRepository{
			log: &domain.Log{ID: logID, UserID: userID},
		}
		clock := commondomain.NewMockClock(now)
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewLogDelete(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)
	})

	t.Run("returns forbidden for non-owner non-admin", func(t *testing.T) {
		repo := &mockLogDeleteRepository{
			log: &domain.Log{ID: logID, UserID: otherUserID},
		}
		clock := commondomain.NewMockClock(now)
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewLogDelete(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		assert.ErrorIs(t, err, domain.ErrForbidden)
		assert.False(t, repo.deleteCalled)
	})

	t.Run("allows admin to delete any log", func(t *testing.T) {
		repo := &mockLogDeleteRepository{
			log: &domain.Log{ID: logID, UserID: otherUserID},
		}
		clock := commondomain.NewMockClock(now)
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewLogDelete(repo, clock, store, rebuildRepo)

		ctx := ctxWithAdminSubject(uuid.New().String())

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)
	})

	t.Run("returns error when log not found", func(t *testing.T) {
		repo := &mockLogDeleteRepository{
			findErr: domain.ErrNotFound,
		}
		clock := commondomain.NewMockClock(now)
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewLogDelete(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())

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
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		svc := domain.NewLogDelete(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		assert.Error(t, err)
		assert.True(t, repo.deleteCalled)
	})
}

func TestLogDelete_LeaderboardUpdates(t *testing.T) {
	userID := uuid.New()
	logID := uuid.New()
	contestID := uuid.New()
	contestID2 := uuid.New()
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)

	t.Run("rebuilds contest leaderboards after delete with registrations", func(t *testing.T) {
		dbScores := []domain.LeaderboardScore{
			{UserID: uuid.New(), Score: 100},
		}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{contestScores: dbScores}
		repo := &mockLogDeleteRepository{
			log: &domain.Log{
				ID:     logID,
				UserID: userID,
				Score:  42.5,
				Registrations: []domain.ContestRegistrationReference{
					{RegistrationID: uuid.New(), ContestID: contestID},
					{RegistrationID: uuid.New(), ContestID: contestID2},
				},
				CreatedAt: now,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())
		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)
		require.Len(t, store.rebuildContestCalls, 2)
		assert.Equal(t, contestID, store.rebuildContestCalls[0].ContestID)
		assert.Equal(t, contestID2, store.rebuildContestCalls[1].ContestID)
		assert.Equal(t, dbScores, store.rebuildContestCalls[0].Scores)
	})

	t.Run("rebuilds yearly and global leaderboards when EligibleOfficialLeaderboard is true", func(t *testing.T) {
		yearlyScores := []domain.LeaderboardScore{
			{UserID: userID, Score: 200},
		}
		globalScores := []domain.LeaderboardScore{
			{UserID: userID, Score: 500},
		}
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{
			contestScores: []domain.LeaderboardScore{},
			yearlyScores:  yearlyScores,
			globalScores:  globalScores,
		}
		repo := &mockLogDeleteRepository{
			log: &domain.Log{
				ID:                          logID,
				UserID:                      userID,
				Score:                       50,
				EligibleOfficialLeaderboard: true,
				Registrations: []domain.ContestRegistrationReference{
					{RegistrationID: uuid.New(), ContestID: contestID},
				},
				CreatedAt: now,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())
		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)

		// Contest leaderboard rebuilt
		require.Len(t, store.rebuildContestCalls, 1)

		// Yearly leaderboard rebuilt
		require.Len(t, store.rebuildYearlyCalls, 1)
		assert.Equal(t, 2026, store.rebuildYearlyCalls[0].Year)
		assert.Equal(t, yearlyScores, store.rebuildYearlyCalls[0].Scores)

		// Global leaderboard rebuilt
		assert.Equal(t, 1, store.rebuildGlobalCalls)
	})

	t.Run("does not rebuild yearly or global when EligibleOfficialLeaderboard is false", func(t *testing.T) {
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{
			contestScores: []domain.LeaderboardScore{},
		}
		repo := &mockLogDeleteRepository{
			log: &domain.Log{
				ID:                          logID,
				UserID:                      userID,
				Score:                       50,
				EligibleOfficialLeaderboard: false,
				Registrations: []domain.ContestRegistrationReference{
					{RegistrationID: uuid.New(), ContestID: contestID},
				},
				CreatedAt: now,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())
		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)

		// Contest leaderboard rebuilt
		require.Len(t, store.rebuildContestCalls, 1)

		// Yearly and global NOT rebuilt
		assert.Empty(t, store.rebuildYearlyCalls)
		assert.Equal(t, 0, store.rebuildGlobalCalls)
	})

	t.Run("leaderboard errors do not fail deletion", func(t *testing.T) {
		store := &mockLeaderboardStore{
			rebuildContestErr: errors.New("redis connection refused"),
			rebuildYearlyErr:  errors.New("redis timeout"),
			rebuildGlobalErr:  errors.New("redis unavailable"),
		}
		rebuildRepo := &mockLeaderboardRebuildRepo{
			contestScores: []domain.LeaderboardScore{{UserID: userID, Score: 10}},
			yearlyScores:  []domain.LeaderboardScore{{UserID: userID, Score: 20}},
			globalScores:  []domain.LeaderboardScore{{UserID: userID, Score: 30}},
		}
		repo := &mockLogDeleteRepository{
			log: &domain.Log{
				ID:                          logID,
				UserID:                      userID,
				Score:                       50,
				EligibleOfficialLeaderboard: true,
				Registrations: []domain.ContestRegistrationReference{
					{RegistrationID: uuid.New(), ContestID: contestID},
				},
				CreatedAt: now,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())
		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)
	})
}
