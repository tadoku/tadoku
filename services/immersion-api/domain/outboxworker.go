package domain

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

// OutboxWorkerRepository provides transactional access to the outbox table.
// ProcessOutboxBatch handles the full transaction lifecycle internally:
// it begins a transaction, fetches and locks events, calls the provided
// callback, marks processed IDs, and commits (or rolls back on error).
type OutboxWorkerRepository interface {
	ProcessOutboxBatch(ctx context.Context, batchSize int32, fn func(events []postgres.FetchAndLockOutboxEventsRow) []int64) error
	CleanupProcessedOutboxEvents(ctx context.Context, before time.Time) error
}

// OutboxWorkerLeaderboardUpdater is the narrow interface the outbox worker
// needs from LeaderboardUpdater.
type OutboxWorkerLeaderboardUpdater interface {
	UpdateUserContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID)
	UpdateUserOfficialScores(ctx context.Context, year int, userID uuid.UUID)
}

// OutboxWorker polls the leaderboard_outbox table and processes events
// by calling the LeaderboardUpdater. It uses FOR UPDATE SKIP LOCKED to
// allow safe concurrent processing across multiple API instances.
type OutboxWorker struct {
	repo     OutboxWorkerRepository
	updater  OutboxWorkerLeaderboardUpdater
	interval time.Duration
}

func NewOutboxWorker(
	repo OutboxWorkerRepository,
	updater OutboxWorkerLeaderboardUpdater,
	interval time.Duration,
) *OutboxWorker {
	return &OutboxWorker{
		repo:     repo,
		updater:  updater,
		interval: interval,
	}
}

// Run polls the outbox table at the configured interval until the context
// is cancelled. It also periodically cleans up old processed events.
func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Clean up old processed events every hour
	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		case <-cleanupTicker.C:
			w.cleanup(ctx)
		}
	}
}

// ProcessBatchForTest exposes processBatch for unit testing.
func (w *OutboxWorker) ProcessBatchForTest(ctx context.Context) {
	w.processBatch(ctx)
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	err := w.repo.ProcessOutboxBatch(ctx, 100, func(events []postgres.FetchAndLockOutboxEventsRow) []int64 {
		if len(events) == 0 {
			return nil
		}

		// Deduplicate: only process one event per unique (event_type, user_id, contest_id/year)
		type dedupeKey struct {
			eventType string
			userID    uuid.UUID
			contestID uuid.UUID // zero for official scores
			year      int16     // zero for contest scores
		}
		seen := map[dedupeKey]struct{}{}
		var allIDs []int64

		for _, event := range events {
			allIDs = append(allIDs, event.ID)

			key := dedupeKey{
				eventType: event.EventType,
				userID:    event.UserID,
			}
			if event.ContestID.Valid {
				key.contestID = event.ContestID.UUID
			}
			if event.Year.Valid {
				key.year = event.Year.Int16
			}

			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}

			w.processEvent(ctx, event)
		}

		return allIDs
	})
	if err != nil {
		slog.ErrorContext(ctx, "outbox worker: batch processing failed", "error", err)
	}
}

func (w *OutboxWorker) processEvent(ctx context.Context, event postgres.FetchAndLockOutboxEventsRow) {
	switch event.EventType {
	case "refresh_contest_score":
		if !event.ContestID.Valid {
			slog.ErrorContext(ctx, "outbox worker: refresh_contest_score event missing contest_id", "event_id", event.ID)
			return
		}
		w.updater.UpdateUserContestScore(ctx, event.ContestID.UUID, event.UserID)

	case "refresh_official_scores":
		if !event.Year.Valid {
			slog.ErrorContext(ctx, "outbox worker: refresh_official_scores event missing year", "event_id", event.ID)
			return
		}
		w.updater.UpdateUserOfficialScores(ctx, int(event.Year.Int16), event.UserID)

	default:
		slog.ErrorContext(ctx, fmt.Sprintf("outbox worker: unknown event type: %s", event.EventType), "event_id", event.ID)
	}
}

func (w *OutboxWorker) cleanup(ctx context.Context) {
	before := time.Now().Add(-24 * time.Hour)
	if err := w.repo.CleanupProcessedOutboxEvents(ctx, before); err != nil {
		slog.ErrorContext(ctx, "outbox worker: could not cleanup old events", "error", err)
	}
}
