package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

// mockLeaderboardStore implements domain.LeaderboardStore for testing.
type mockLeaderboardStore struct {
	updateContestCalls   []storeUpdateContestCall
	updateOfficialCalls  []storeUpdateOfficialCall
	rebuildContestCalls  []storeRebuildContestCall
	rebuildOfficialCalls []storeRebuildOfficialCall
	removeContestCalls   []mockLeaderboardOutboxContestCall
	removeOfficialCalls  []mockLeaderboardOutboxOfficialCall

	// Control behavior
	updateContestExists  bool
	updateOfficialYearly bool
	updateOfficialGlobal bool
	updateContestErr     error
	updateOfficialErr    error
	rebuildContestErr    error
	rebuildOfficialErr   error
	removeContestErr     error
	removeOfficialErr    error
}

type storeUpdateContestCall struct {
	ContestID uuid.UUID
	UserID    uuid.UUID
	Score     float64
}

type storeUpdateOfficialCall struct {
	Year        int
	UserID      uuid.UUID
	YearlyScore float64
	GlobalScore float64
}

type storeRebuildContestCall struct {
	ContestID uuid.UUID
	Scores    []domain.LeaderboardScore
}

type storeRebuildOfficialCall struct {
	Year         int
	YearlyScores []domain.LeaderboardScore
	GlobalScores []domain.LeaderboardScore
}

func (m *mockLeaderboardStore) UpdateContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID, score float64) (bool, error) {
	m.updateContestCalls = append(m.updateContestCalls, storeUpdateContestCall{
		ContestID: contestID, UserID: userID, Score: score,
	})
	return m.updateContestExists, m.updateContestErr
}

func (m *mockLeaderboardStore) UpdateOfficialScores(ctx context.Context, year int, userID uuid.UUID, yearlyScore float64, globalScore float64) (bool, bool, error) {
	m.updateOfficialCalls = append(m.updateOfficialCalls, storeUpdateOfficialCall{
		Year: year, UserID: userID, YearlyScore: yearlyScore, GlobalScore: globalScore,
	})
	return m.updateOfficialYearly, m.updateOfficialGlobal, m.updateOfficialErr
}

func (m *mockLeaderboardStore) RemoveContestScore(_ context.Context, contestID uuid.UUID, userID uuid.UUID) error {
	m.removeContestCalls = append(m.removeContestCalls, mockLeaderboardOutboxContestCall{ContestID: contestID, UserID: userID})
	return m.removeContestErr
}

func (m *mockLeaderboardStore) RemoveOfficialScores(_ context.Context, year int, userID uuid.UUID) error {
	m.removeOfficialCalls = append(m.removeOfficialCalls, mockLeaderboardOutboxOfficialCall{Year: year, UserID: userID})
	return m.removeOfficialErr
}

func (m *mockLeaderboardStore) RebuildContestLeaderboard(ctx context.Context, contestID uuid.UUID, scores []domain.LeaderboardScore) error {
	m.rebuildContestCalls = append(m.rebuildContestCalls, storeRebuildContestCall{
		ContestID: contestID, Scores: scores,
	})
	return m.rebuildContestErr
}

func (m *mockLeaderboardStore) RebuildOfficialLeaderboards(ctx context.Context, year int, yearlyScores []domain.LeaderboardScore, globalScores []domain.LeaderboardScore) error {
	m.rebuildOfficialCalls = append(m.rebuildOfficialCalls, storeRebuildOfficialCall{
		Year: year, YearlyScores: yearlyScores, GlobalScores: globalScores,
	})
	return m.rebuildOfficialErr
}

// mockLeaderboardRepo implements domain.LeaderboardRepository for testing.
type mockLeaderboardRepo struct {
	contestScores    []domain.LeaderboardScore
	yearlyScores     []domain.LeaderboardScore
	globalScores     []domain.LeaderboardScore
	userContestScore float64
	userYearlyScore  float64
	userGlobalScore  float64
	contestErr       error
	yearlyErr        error
	globalErr        error
	userContestErr   error
	userYearlyErr    error
	userGlobalErr    error
}

func (m *mockLeaderboardRepo) FetchAllContestLeaderboardScores(ctx context.Context, contestID uuid.UUID) ([]domain.LeaderboardScore, error) {
	return m.contestScores, m.contestErr
}

func (m *mockLeaderboardRepo) FetchAllYearlyLeaderboardScores(ctx context.Context, year int) ([]domain.LeaderboardScore, error) {
	return m.yearlyScores, m.yearlyErr
}

func (m *mockLeaderboardRepo) FetchAllGlobalLeaderboardScores(ctx context.Context) ([]domain.LeaderboardScore, error) {
	return m.globalScores, m.globalErr
}

func (m *mockLeaderboardRepo) FetchUserContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID) (float64, error) {
	return m.userContestScore, m.userContestErr
}

func (m *mockLeaderboardRepo) FetchUserYearlyScore(ctx context.Context, year int, userID uuid.UUID) (float64, error) {
	return m.userYearlyScore, m.userYearlyErr
}

func (m *mockLeaderboardRepo) FetchUserGlobalScore(ctx context.Context, userID uuid.UUID) (float64, error) {
	return m.userGlobalScore, m.userGlobalErr
}

func TestLeaderboardUpdater_UpdateUserContestScore(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	contestID := uuid.New()

	t.Run("sets score when leaderboard exists in store", func(t *testing.T) {
		store := &mockLeaderboardStore{updateContestExists: true}
		repo := &mockLeaderboardRepo{userContestScore: 142.5}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserContestScore(ctx, contestID, userID)

		require.NoError(t, err)
		require.Len(t, store.updateContestCalls, 1)
		assert.Equal(t, contestID, store.updateContestCalls[0].ContestID)
		assert.Equal(t, userID, store.updateContestCalls[0].UserID)
		assert.InDelta(t, 142.5, store.updateContestCalls[0].Score, 0.01)
		assert.Empty(t, store.rebuildContestCalls)
	})

	t.Run("rebuilds leaderboard when not in store", func(t *testing.T) {
		dbScores := []domain.LeaderboardScore{
			{UserID: userID, Score: 42.5},
			{UserID: uuid.New(), Score: 100},
		}
		store := &mockLeaderboardStore{updateContestExists: false}
		repo := &mockLeaderboardRepo{userContestScore: 42.5, contestScores: dbScores}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserContestScore(ctx, contestID, userID)

		require.NoError(t, err)
		require.Len(t, store.rebuildContestCalls, 1)
		assert.Equal(t, contestID, store.rebuildContestCalls[0].ContestID)
		assert.Equal(t, dbScores, store.rebuildContestCalls[0].Scores)
	})

	t.Run("returns fetch user score error", func(t *testing.T) {
		fetchErr := errors.New("database error")
		store := &mockLeaderboardStore{}
		repo := &mockLeaderboardRepo{userContestErr: fetchErr}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserContestScore(ctx, contestID, userID)

		assert.ErrorIs(t, err, fetchErr)
		assert.Empty(t, store.updateContestCalls)
		assert.Empty(t, store.rebuildContestCalls)
	})

	t.Run("returns store update error", func(t *testing.T) {
		updateErr := errors.New("redis error")
		store := &mockLeaderboardStore{updateContestErr: updateErr}
		repo := &mockLeaderboardRepo{userContestScore: 50}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserContestScore(ctx, contestID, userID)

		assert.ErrorIs(t, err, updateErr)
		require.Len(t, store.updateContestCalls, 1)
		assert.Empty(t, store.rebuildContestCalls)
	})

	t.Run("returns rebuild fetch error", func(t *testing.T) {
		fetchErr := errors.New("database timeout")
		store := &mockLeaderboardStore{updateContestExists: false}
		repo := &mockLeaderboardRepo{
			userContestScore: 50,
			contestErr:       fetchErr,
		}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserContestScore(ctx, contestID, userID)

		assert.ErrorIs(t, err, fetchErr)
		assert.Empty(t, store.rebuildContestCalls)
	})

	t.Run("returns rebuild store error", func(t *testing.T) {
		rebuildErr := errors.New("redis error")
		store := &mockLeaderboardStore{updateContestExists: false, rebuildContestErr: rebuildErr}
		repo := &mockLeaderboardRepo{userContestScore: 50}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserContestScore(ctx, contestID, userID)

		assert.ErrorIs(t, err, rebuildErr)
		require.Len(t, store.rebuildContestCalls, 1)
	})
}

func TestLeaderboardUpdater_UpdateUserOfficialScores(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	year := 2026

	t.Run("sets scores when both leaderboards exist in store", func(t *testing.T) {
		store := &mockLeaderboardStore{
			updateOfficialYearly: true,
			updateOfficialGlobal: true,
		}
		repo := &mockLeaderboardRepo{
			userYearlyScore: 200,
			userGlobalScore: 500,
		}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserOfficialScores(ctx, year, userID)

		require.NoError(t, err)
		require.Len(t, store.updateOfficialCalls, 1)
		assert.Equal(t, year, store.updateOfficialCalls[0].Year)
		assert.Equal(t, userID, store.updateOfficialCalls[0].UserID)
		assert.InDelta(t, 200, store.updateOfficialCalls[0].YearlyScore, 0.01)
		assert.InDelta(t, 500, store.updateOfficialCalls[0].GlobalScore, 0.01)
		assert.Empty(t, store.rebuildOfficialCalls)
	})

	t.Run("rebuilds when yearly leaderboard missing from store", func(t *testing.T) {
		yearlyScores := []domain.LeaderboardScore{{UserID: userID, Score: 200}}
		globalScores := []domain.LeaderboardScore{{UserID: userID, Score: 500}}
		store := &mockLeaderboardStore{
			updateOfficialYearly: false,
			updateOfficialGlobal: true,
		}
		repo := &mockLeaderboardRepo{
			userYearlyScore: 200,
			userGlobalScore: 500,
			yearlyScores:    yearlyScores,
			globalScores:    globalScores,
		}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserOfficialScores(ctx, year, userID)

		require.NoError(t, err)
		require.Len(t, store.rebuildOfficialCalls, 1)
		assert.Equal(t, year, store.rebuildOfficialCalls[0].Year)
		assert.Equal(t, yearlyScores, store.rebuildOfficialCalls[0].YearlyScores)
		assert.Equal(t, globalScores, store.rebuildOfficialCalls[0].GlobalScores)
	})

	t.Run("rebuilds when global leaderboard missing from store", func(t *testing.T) {
		yearlyScores := []domain.LeaderboardScore{{UserID: userID, Score: 200}}
		globalScores := []domain.LeaderboardScore{{UserID: userID, Score: 500}}
		store := &mockLeaderboardStore{
			updateOfficialYearly: true,
			updateOfficialGlobal: false,
		}
		repo := &mockLeaderboardRepo{
			userYearlyScore: 200,
			userGlobalScore: 500,
			yearlyScores:    yearlyScores,
			globalScores:    globalScores,
		}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserOfficialScores(ctx, year, userID)

		require.NoError(t, err)
		require.Len(t, store.rebuildOfficialCalls, 1)
	})

	t.Run("returns fetch yearly score error", func(t *testing.T) {
		fetchErr := errors.New("database error")
		store := &mockLeaderboardStore{}
		repo := &mockLeaderboardRepo{userYearlyErr: fetchErr}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserOfficialScores(ctx, year, userID)

		assert.ErrorIs(t, err, fetchErr)
		assert.Empty(t, store.updateOfficialCalls)
	})

	t.Run("returns fetch global score error", func(t *testing.T) {
		fetchErr := errors.New("database error")
		store := &mockLeaderboardStore{}
		repo := &mockLeaderboardRepo{
			userYearlyScore: 200,
			userGlobalErr:   fetchErr,
		}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserOfficialScores(ctx, year, userID)

		assert.ErrorIs(t, err, fetchErr)
		assert.Empty(t, store.updateOfficialCalls)
	})

	t.Run("returns store update error", func(t *testing.T) {
		updateErr := errors.New("redis error")
		store := &mockLeaderboardStore{updateOfficialErr: updateErr}
		repo := &mockLeaderboardRepo{
			userYearlyScore: 200,
			userGlobalScore: 500,
		}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserOfficialScores(ctx, year, userID)

		assert.ErrorIs(t, err, updateErr)
		require.Len(t, store.updateOfficialCalls, 1)
		assert.Empty(t, store.rebuildOfficialCalls)
	})

	t.Run("returns rebuild yearly fetch error", func(t *testing.T) {
		fetchErr := errors.New("database error")
		store := &mockLeaderboardStore{updateOfficialYearly: false, updateOfficialGlobal: true}
		repo := &mockLeaderboardRepo{
			userYearlyScore: 200,
			userGlobalScore: 500,
			yearlyErr:       fetchErr,
		}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserOfficialScores(ctx, year, userID)

		assert.ErrorIs(t, err, fetchErr)
		assert.Empty(t, store.rebuildOfficialCalls)
	})

	t.Run("returns rebuild global fetch error", func(t *testing.T) {
		fetchErr := errors.New("database error")
		store := &mockLeaderboardStore{updateOfficialYearly: true, updateOfficialGlobal: false}
		repo := &mockLeaderboardRepo{
			userYearlyScore: 200,
			userGlobalScore: 500,
			globalErr:       fetchErr,
		}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserOfficialScores(ctx, year, userID)

		assert.ErrorIs(t, err, fetchErr)
		assert.Empty(t, store.rebuildOfficialCalls)
	})

	t.Run("returns rebuild store error", func(t *testing.T) {
		rebuildErr := errors.New("redis error")
		store := &mockLeaderboardStore{
			updateOfficialYearly: false,
			updateOfficialGlobal: true,
			rebuildOfficialErr:   rebuildErr,
		}
		repo := &mockLeaderboardRepo{userYearlyScore: 200, userGlobalScore: 500}
		updater := domain.NewLeaderboardUpdater(store, repo)

		err := updater.UpdateUserOfficialScores(ctx, year, userID)

		assert.ErrorIs(t, err, rebuildErr)
		require.Len(t, store.rebuildOfficialCalls, 1)
	})
}

func TestLeaderboardUpdater_RemoveUserScores(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	contestID := uuid.New()
	store := &mockLeaderboardStore{}
	updater := domain.NewLeaderboardUpdater(store, &mockLeaderboardRepo{})

	require.NoError(t, updater.RemoveUserContestScore(ctx, contestID, userID))
	require.NoError(t, updater.RemoveUserOfficialScores(ctx, 2026, userID))
	require.Len(t, store.removeContestCalls, 1)
	assert.Equal(t, contestID, store.removeContestCalls[0].ContestID)
	assert.Equal(t, userID, store.removeContestCalls[0].UserID)
	require.Len(t, store.removeOfficialCalls, 1)
	assert.Equal(t, 2026, store.removeOfficialCalls[0].Year)
	assert.Equal(t, userID, store.removeOfficialCalls[0].UserID)
}
