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

type mockLeaderboardOutboxRepository struct {
	events     []domain.LeaderboardOutboxEvent
	batchErr   error
	cleanupErr error

	markedIDs []int64
	cleanedUp bool
}

func (m *mockLeaderboardOutboxRepository) ProcessOutboxBatch(ctx context.Context, batchSize int32, fn func(events []domain.LeaderboardOutboxEvent) []int64) error {
	if m.batchErr != nil {
		return m.batchErr
	}
	ids := fn(m.events)
	m.markedIDs = ids
	return nil
}

func (m *mockLeaderboardOutboxRepository) CleanupProcessedOutboxEvents(ctx context.Context, before time.Time) error {
	m.cleanedUp = true
	return m.cleanupErr
}

type mockLeaderboardOutboxUpdater struct {
	contestCalls         []mockLeaderboardOutboxContestCall
	officialCalls        []mockLeaderboardOutboxOfficialCall
	rebuildOfficialCalls []int
	removeContestCalls   []mockLeaderboardOutboxContestCall
	removeOfficialCalls  []mockLeaderboardOutboxOfficialCall
	contestErr           error
	officialErr          error
	rebuildOfficialErr   error
	removeContestErr     error
	removeOfficialErr    error
	rebuildCalled        chan int
}

type mockLeaderboardOutboxContestCall struct {
	ContestID uuid.UUID
	UserID    uuid.UUID
}

type mockLeaderboardOutboxOfficialCall struct {
	Year   int
	UserID uuid.UUID
}

func (m *mockLeaderboardOutboxUpdater) UpdateUserContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID) error {
	m.contestCalls = append(m.contestCalls, mockLeaderboardOutboxContestCall{ContestID: contestID, UserID: userID})
	return m.contestErr
}

func (m *mockLeaderboardOutboxUpdater) UpdateUserOfficialScores(ctx context.Context, year int, userID uuid.UUID) error {
	m.officialCalls = append(m.officialCalls, mockLeaderboardOutboxOfficialCall{Year: year, UserID: userID})
	return m.officialErr
}

func (m *mockLeaderboardOutboxUpdater) RemoveUserContestScore(_ context.Context, contestID uuid.UUID, userID uuid.UUID) error {
	m.removeContestCalls = append(m.removeContestCalls, mockLeaderboardOutboxContestCall{ContestID: contestID, UserID: userID})
	return m.removeContestErr
}

func (m *mockLeaderboardOutboxUpdater) RemoveUserOfficialScores(_ context.Context, year int, userID uuid.UUID) error {
	m.removeOfficialCalls = append(m.removeOfficialCalls, mockLeaderboardOutboxOfficialCall{Year: year, UserID: userID})
	return m.removeOfficialErr
}

func (m *mockLeaderboardOutboxUpdater) RebuildOfficialLeaderboards(ctx context.Context, year int) error {
	m.rebuildOfficialCalls = append(m.rebuildOfficialCalls, year)
	if m.rebuildCalled != nil {
		m.rebuildCalled <- year
	}
	return m.rebuildOfficialErr
}

func TestLeaderboardOutboxWorker_ProcessEvent(t *testing.T) {
	userID := uuid.New()
	contestID := uuid.New()
	year2026 := 2026
	now := time.Date(2026, time.August, 22, 12, 0, 0, 0, time.UTC)

	t.Run("processes refresh_contest_score events", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{}
		repo := &mockLeaderboardOutboxRepository{
			events: []domain.LeaderboardOutboxEvent{
				{
					ID:        1,
					EventType: "refresh_contest_score",
					UserID:    userID,
					ContestID: &contestID,
				},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		worker.ProcessBatchForTest(context.Background())

		require.Len(t, updater.contestCalls, 1)
		assert.Equal(t, contestID, updater.contestCalls[0].ContestID)
		assert.Equal(t, userID, updater.contestCalls[0].UserID)
		assert.Empty(t, updater.officialCalls)
		assert.Equal(t, []int64{1}, repo.markedIDs)
	})

	t.Run("processes refresh_official_scores events", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{}
		repo := &mockLeaderboardOutboxRepository{
			events: []domain.LeaderboardOutboxEvent{
				{
					ID:        2,
					EventType: "refresh_official_scores",
					UserID:    userID,
					Year:      &year2026,
				},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		worker.ProcessBatchForTest(context.Background())

		assert.Empty(t, updater.contestCalls)
		require.Len(t, updater.officialCalls, 1)
		assert.Equal(t, 2026, updater.officialCalls[0].Year)
		assert.Equal(t, userID, updater.officialCalls[0].UserID)
		assert.Equal(t, []int64{2}, repo.markedIDs)
	})

	t.Run("processes account deletion removal events", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{}
		repo := &mockLeaderboardOutboxRepository{events: []domain.LeaderboardOutboxEvent{
			{ID: 3, EventType: "remove_contest_score", UserID: userID, ContestID: &contestID},
			{ID: 4, EventType: "remove_official_scores", UserID: userID, Year: &year2026},
		}}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		worker.ProcessBatchForTest(context.Background())

		require.Len(t, updater.removeContestCalls, 1)
		assert.Equal(t, contestID, updater.removeContestCalls[0].ContestID)
		assert.Equal(t, userID, updater.removeContestCalls[0].UserID)
		require.Len(t, updater.removeOfficialCalls, 1)
		assert.Equal(t, year2026, updater.removeOfficialCalls[0].Year)
		assert.Equal(t, userID, updater.removeOfficialCalls[0].UserID)
		assert.Equal(t, []int64{3, 4}, repo.markedIDs)
	})

	t.Run("keeps deletion removals pending when valkey fails", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{removeContestErr: errors.New("valkey unavailable")}
		repo := &mockLeaderboardOutboxRepository{events: []domain.LeaderboardOutboxEvent{
			{ID: 3, EventType: "remove_contest_score", UserID: userID, ContestID: &contestID},
		}}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		worker.ProcessBatchForTest(context.Background())

		assert.Empty(t, repo.markedIDs)
	})

	t.Run("keeps refresh_official_scores events pending when the cache update fails", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{officialErr: errors.New("valkey unavailable")}
		repo := &mockLeaderboardOutboxRepository{
			events: []domain.LeaderboardOutboxEvent{
				{
					ID:        2,
					EventType: "refresh_official_scores",
					UserID:    userID,
					Year:      &year2026,
				},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		worker.ProcessBatchForTest(context.Background())

		require.Len(t, updater.officialCalls, 1)
		assert.Empty(t, repo.markedIDs)
	})

	t.Run("marks all duplicate events only after their update succeeds", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{officialErr: errors.New("valkey unavailable")}
		repo := &mockLeaderboardOutboxRepository{
			events: []domain.LeaderboardOutboxEvent{
				{ID: 1, EventType: "refresh_official_scores", UserID: userID, Year: &year2026},
				{ID: 2, EventType: "refresh_official_scores", UserID: userID, Year: &year2026},
				{ID: 3, EventType: "refresh_official_scores", UserID: userID, Year: &year2026},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		worker.ProcessBatchForTest(context.Background())

		require.Len(t, updater.officialCalls, 1)
		assert.Empty(t, repo.markedIDs)

		updater.officialErr = nil
		worker.ProcessBatchForTest(context.Background())

		require.Len(t, updater.officialCalls, 2)
		assert.Equal(t, []int64{1, 2, 3}, repo.markedIDs)
	})

	t.Run("marks successful groups while leaving failed groups pending", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{officialErr: errors.New("valkey unavailable")}
		repo := &mockLeaderboardOutboxRepository{
			events: []domain.LeaderboardOutboxEvent{
				{ID: 1, EventType: "refresh_official_scores", UserID: userID, Year: &year2026},
				{ID: 2, EventType: "refresh_contest_score", UserID: userID, ContestID: &contestID},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		worker.ProcessBatchForTest(context.Background())

		require.Len(t, updater.officialCalls, 1)
		require.Len(t, updater.contestCalls, 1)
		assert.Equal(t, []int64{2}, repo.markedIDs)
	})

	t.Run("deduplicates events with same key", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{}
		repo := &mockLeaderboardOutboxRepository{
			events: []domain.LeaderboardOutboxEvent{
				{
					ID:        1,
					EventType: "refresh_contest_score",
					UserID:    userID,
					ContestID: &contestID,
				},
				{
					ID:        2,
					EventType: "refresh_contest_score",
					UserID:    userID,
					ContestID: &contestID,
				},
				{
					ID:        3,
					EventType: "refresh_contest_score",
					UserID:    userID,
					ContestID: &contestID,
				},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		worker.ProcessBatchForTest(context.Background())

		// Only one actual update should happen despite 3 events
		require.Len(t, updater.contestCalls, 1)
		// But all 3 should be marked as processed
		assert.Equal(t, []int64{1, 2, 3}, repo.markedIDs)
	})

	t.Run("processes different events separately", func(t *testing.T) {
		contestID2 := uuid.New()
		userID2 := uuid.New()

		updater := &mockLeaderboardOutboxUpdater{}
		repo := &mockLeaderboardOutboxRepository{
			events: []domain.LeaderboardOutboxEvent{
				{
					ID:        1,
					EventType: "refresh_contest_score",
					UserID:    userID,
					ContestID: &contestID,
				},
				{
					ID:        2,
					EventType: "refresh_contest_score",
					UserID:    userID2,
					ContestID: &contestID2,
				},
				{
					ID:        3,
					EventType: "refresh_official_scores",
					UserID:    userID,
					Year:      &year2026,
				},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		worker.ProcessBatchForTest(context.Background())

		require.Len(t, updater.contestCalls, 2)
		require.Len(t, updater.officialCalls, 1)
		assert.Equal(t, []int64{1, 2, 3}, repo.markedIDs)
	})

	t.Run("no-op when no events", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{}
		repo := &mockLeaderboardOutboxRepository{
			events: []domain.LeaderboardOutboxEvent{},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		worker.ProcessBatchForTest(context.Background())

		assert.Empty(t, updater.contestCalls)
		assert.Empty(t, updater.officialCalls)
		assert.Nil(t, repo.markedIDs)
	})

	t.Run("handles batch processing error gracefully", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{}
		repo := &mockLeaderboardOutboxRepository{
			batchErr: errors.New("db connection lost"),
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		// Should not panic
		worker.ProcessBatchForTest(context.Background())

		assert.Empty(t, updater.contestCalls)
		assert.Empty(t, updater.officialCalls)
	})

	t.Run("marks permanently malformed events so they do not poison the queue", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{}
		repo := &mockLeaderboardOutboxRepository{
			events: []domain.LeaderboardOutboxEvent{
				{ID: 1, EventType: "refresh_contest_score", UserID: userID},
				{ID: 2, EventType: "refresh_official_scores", UserID: userID},
				{ID: 3, EventType: "not_a_real_event", UserID: userID},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Second)
		worker.ProcessBatchForTest(context.Background())

		assert.Empty(t, updater.contestCalls)
		assert.Empty(t, updater.officialCalls)
		assert.Equal(t, []int64{1, 2, 3}, repo.markedIDs)
	})
}

func TestLeaderboardOutboxWorker_RunReconcilesOfficialLeaderboardsAtStartup(t *testing.T) {
	now := time.Date(2026, time.August, 22, 12, 0, 0, 0, time.UTC)
	rebuildCalled := make(chan int, 1)
	updater := &mockLeaderboardOutboxUpdater{rebuildCalled: rebuildCalled}
	repo := &mockLeaderboardOutboxRepository{}
	worker := domain.NewLeaderboardOutboxWorker(repo, updater, &mockClock{now: now}, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		worker.Run(ctx)
	}()

	assert.Equal(t, 2026, <-rebuildCalled)
	cancel()
	<-done
	assert.Equal(t, []int{2026}, updater.rebuildOfficialCalls)
}
