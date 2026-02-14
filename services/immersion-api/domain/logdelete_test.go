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
		lbRepo := &mockLeaderboardRepo{}
		updater := domain.NewLeaderboardUpdater(store, lbRepo)
		svc := domain.NewLogDelete(repo, clock, updater)

		ctx := ctxWithGuest()

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.deleteCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		repo := &mockLogDeleteRepository{}
		clock := commondomain.NewMockClock(now)
		store := &mockLeaderboardStore{}
		lbRepo := &mockLeaderboardRepo{}
		updater := domain.NewLeaderboardUpdater(store, lbRepo)
		svc := domain.NewLogDelete(repo, clock, updater)

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
		lbRepo := &mockLeaderboardRepo{}
		updater := domain.NewLeaderboardUpdater(store, lbRepo)
		svc := domain.NewLogDelete(repo, clock, updater)

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
		lbRepo := &mockLeaderboardRepo{}
		updater := domain.NewLeaderboardUpdater(store, lbRepo)
		svc := domain.NewLogDelete(repo, clock, updater)

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
		lbRepo := &mockLeaderboardRepo{}
		updater := domain.NewLeaderboardUpdater(store, lbRepo)
		svc := domain.NewLogDelete(repo, clock, updater)

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
		lbRepo := &mockLeaderboardRepo{}
		updater := domain.NewLeaderboardUpdater(store, lbRepo)
		svc := domain.NewLogDelete(repo, clock, updater)

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
		lbRepo := &mockLeaderboardRepo{}
		updater := domain.NewLeaderboardUpdater(store, lbRepo)
		svc := domain.NewLogDelete(repo, clock, updater)

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
		lbRepo := &mockLeaderboardRepo{contestScores: dbScores}
		updater := domain.NewLeaderboardUpdater(store, lbRepo)
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
		svc := domain.NewLogDelete(repo, clock, updater)

		ctx := ctxWithUserSubject(userID.String())
		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)
		require.Len(t, store.rebuildContestCalls, 2)
		assert.Equal(t, contestID, store.rebuildContestCalls[0].ContestID)
		assert.Equal(t, contestID2, store.rebuildContestCalls[1].ContestID)
		assert.Equal(t, dbScores, store.rebuildContestCalls[0].Scores)
	})

	t.Run("rebuilds official leaderboards when EligibleOfficialLeaderboard is true", func(t *testing.T) {
		yearlyScores := []domain.LeaderboardScore{
			{UserID: userID, Score: 200},
		}
		globalScores := []domain.LeaderboardScore{
			{UserID: userID, Score: 500},
		}
		store := &mockLeaderboardStore{}
		lbRepo := &mockLeaderboardRepo{
			contestScores: []domain.LeaderboardScore{},
			yearlyScores:  yearlyScores,
			globalScores:  globalScores,
		}
		updater := domain.NewLeaderboardUpdater(store, lbRepo)
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
		svc := domain.NewLogDelete(repo, clock, updater)

		ctx := ctxWithUserSubject(userID.String())
		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)

		// Contest leaderboard rebuilt
		require.Len(t, store.rebuildContestCalls, 1)

		// Official leaderboards rebuilt (pipelined)
		require.Len(t, store.rebuildOfficialCalls, 1)
		assert.Equal(t, 2026, store.rebuildOfficialCalls[0].Year)
		assert.Equal(t, yearlyScores, store.rebuildOfficialCalls[0].YearlyScores)
		assert.Equal(t, globalScores, store.rebuildOfficialCalls[0].GlobalScores)
	})

	t.Run("does not rebuild official leaderboards when EligibleOfficialLeaderboard is false", func(t *testing.T) {
		store := &mockLeaderboardStore{}
		lbRepo := &mockLeaderboardRepo{
			contestScores: []domain.LeaderboardScore{},
		}
		updater := domain.NewLeaderboardUpdater(store, lbRepo)
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
		svc := domain.NewLogDelete(repo, clock, updater)

		ctx := ctxWithUserSubject(userID.String())
		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)

		// Contest leaderboard rebuilt
		require.Len(t, store.rebuildContestCalls, 1)

		// Official leaderboards NOT rebuilt
		assert.Empty(t, store.rebuildOfficialCalls)
	})

	t.Run("leaderboard errors do not fail deletion", func(t *testing.T) {
		store := &mockLeaderboardStore{
			rebuildContestErr:  errors.New("redis connection refused"),
			rebuildOfficialErr: errors.New("redis timeout"),
		}
		lbRepo := &mockLeaderboardRepo{
			contestScores: []domain.LeaderboardScore{{UserID: userID, Score: 10}},
			yearlyScores:  []domain.LeaderboardScore{{UserID: userID, Score: 20}},
			globalScores:  []domain.LeaderboardScore{{UserID: userID, Score: 30}},
		}
		updater := domain.NewLeaderboardUpdater(store, lbRepo)
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
		svc := domain.NewLogDelete(repo, clock, updater)

		ctx := ctxWithUserSubject(userID.String())
		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)
	})
}
