package domain_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

type mockLeaderboardOutboxRepository struct {
	events     []postgres.FetchAndLockOutboxEventsRow
	batchErr   error
	cleanupErr error

	markedIDs []int64
	cleanedUp bool
}

func (m *mockLeaderboardOutboxRepository) ProcessOutboxBatch(ctx context.Context, batchSize int32, fn func(events []postgres.FetchAndLockOutboxEventsRow) []int64) error {
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
	contestCalls  []mockLeaderboardOutboxContestCall
	officialCalls []mockLeaderboardOutboxOfficialCall
}

type mockLeaderboardOutboxContestCall struct {
	ContestID uuid.UUID
	UserID    uuid.UUID
}

type mockLeaderboardOutboxOfficialCall struct {
	Year   int
	UserID uuid.UUID
}

func (m *mockLeaderboardOutboxUpdater) UpdateUserContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID) {
	m.contestCalls = append(m.contestCalls, mockLeaderboardOutboxContestCall{ContestID: contestID, UserID: userID})
}

func (m *mockLeaderboardOutboxUpdater) UpdateUserOfficialScores(ctx context.Context, year int, userID uuid.UUID) {
	m.officialCalls = append(m.officialCalls, mockLeaderboardOutboxOfficialCall{Year: year, UserID: userID})
}

func TestLeaderboardOutboxWorker_ProcessEvent(t *testing.T) {
	userID := uuid.New()
	contestID := uuid.New()

	t.Run("processes refresh_contest_score events", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{}
		repo := &mockLeaderboardOutboxRepository{
			events: []postgres.FetchAndLockOutboxEventsRow{
				{
					ID:        1,
					EventType: "refresh_contest_score",
					UserID:    userID,
					ContestID: uuid.NullUUID{UUID: contestID, Valid: true},
				},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, time.Second)
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
			events: []postgres.FetchAndLockOutboxEventsRow{
				{
					ID:        2,
					EventType: "refresh_official_scores",
					UserID:    userID,
					Year:      sql.NullInt16{Int16: 2026, Valid: true},
				},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, time.Second)
		worker.ProcessBatchForTest(context.Background())

		assert.Empty(t, updater.contestCalls)
		require.Len(t, updater.officialCalls, 1)
		assert.Equal(t, 2026, updater.officialCalls[0].Year)
		assert.Equal(t, userID, updater.officialCalls[0].UserID)
		assert.Equal(t, []int64{2}, repo.markedIDs)
	})

	t.Run("deduplicates events with same key", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{}
		repo := &mockLeaderboardOutboxRepository{
			events: []postgres.FetchAndLockOutboxEventsRow{
				{
					ID:        1,
					EventType: "refresh_contest_score",
					UserID:    userID,
					ContestID: uuid.NullUUID{UUID: contestID, Valid: true},
				},
				{
					ID:        2,
					EventType: "refresh_contest_score",
					UserID:    userID,
					ContestID: uuid.NullUUID{UUID: contestID, Valid: true},
				},
				{
					ID:        3,
					EventType: "refresh_contest_score",
					UserID:    userID,
					ContestID: uuid.NullUUID{UUID: contestID, Valid: true},
				},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, time.Second)
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
			events: []postgres.FetchAndLockOutboxEventsRow{
				{
					ID:        1,
					EventType: "refresh_contest_score",
					UserID:    userID,
					ContestID: uuid.NullUUID{UUID: contestID, Valid: true},
				},
				{
					ID:        2,
					EventType: "refresh_contest_score",
					UserID:    userID2,
					ContestID: uuid.NullUUID{UUID: contestID2, Valid: true},
				},
				{
					ID:        3,
					EventType: "refresh_official_scores",
					UserID:    userID,
					Year:      sql.NullInt16{Int16: 2026, Valid: true},
				},
			},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, time.Second)
		worker.ProcessBatchForTest(context.Background())

		require.Len(t, updater.contestCalls, 2)
		require.Len(t, updater.officialCalls, 1)
		assert.Equal(t, []int64{1, 2, 3}, repo.markedIDs)
	})

	t.Run("no-op when no events", func(t *testing.T) {
		updater := &mockLeaderboardOutboxUpdater{}
		repo := &mockLeaderboardOutboxRepository{
			events: []postgres.FetchAndLockOutboxEventsRow{},
		}

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, time.Second)
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

		worker := domain.NewLeaderboardOutboxWorker(repo, updater, time.Second)
		// Should not panic
		worker.ProcessBatchForTest(context.Background())

		assert.Empty(t, updater.contestCalls)
		assert.Empty(t, updater.officialCalls)
	})
}
