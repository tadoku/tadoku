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
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewLogDelete(repo, clock, updater)

		ctx := ctxWithGuest()

		err := svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.deleteCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		repo := &mockLogDeleteRepository{}
		clock := commondomain.NewMockClock(now)
		updater := &mockLeaderboardUpdater{}
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
		updater := &mockLeaderboardUpdater{}
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
		updater := &mockLeaderboardUpdater{}
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
		updater := &mockLeaderboardUpdater{}
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
		updater := &mockLeaderboardUpdater{}
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
		updater := &mockLeaderboardUpdater{}
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

	t.Run("updates user contest score for each registration", func(t *testing.T) {
		updater := &mockLeaderboardUpdater{}
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
		require.Len(t, updater.updateContestCalls, 2)
		assert.Equal(t, contestID, updater.updateContestCalls[0].ContestID)
		assert.Equal(t, userID, updater.updateContestCalls[0].UserID)
		assert.Equal(t, contestID2, updater.updateContestCalls[1].ContestID)
		assert.Equal(t, userID, updater.updateContestCalls[1].UserID)
	})

	t.Run("updates official scores when EligibleOfficialLeaderboard is true", func(t *testing.T) {
		updater := &mockLeaderboardUpdater{}
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
		require.Len(t, updater.updateContestCalls, 1)
		require.Len(t, updater.updateOfficialCalls, 1)
		assert.Equal(t, 2026, updater.updateOfficialCalls[0].Year)
		assert.Equal(t, userID, updater.updateOfficialCalls[0].UserID)
	})

	t.Run("does not update official scores when EligibleOfficialLeaderboard is false", func(t *testing.T) {
		updater := &mockLeaderboardUpdater{}
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
		require.Len(t, updater.updateContestCalls, 1)
		assert.Empty(t, updater.updateOfficialCalls)
	})

	t.Run("does not update leaderboards when delete fails", func(t *testing.T) {
		updater := &mockLeaderboardUpdater{}
		repo := &mockLogDeleteRepository{
			log: &domain.Log{
				ID:     logID,
				UserID: userID,
				Score:  50,
				Registrations: []domain.ContestRegistrationReference{
					{RegistrationID: uuid.New(), ContestID: contestID},
				},
				CreatedAt: now,
			},
			deleteErr: errors.New("database error"),
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogDelete(repo, clock, updater)

		ctx := ctxWithUserSubject(userID.String())
		_ = svc.Execute(ctx, &domain.LogDeleteRequest{LogID: logID})

		assert.Empty(t, updater.updateContestCalls)
		assert.Empty(t, updater.updateOfficialCalls)
	})
}
