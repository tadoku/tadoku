package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

// ProcessOutboxBatch runs the full outbox processing cycle in a single transaction:
// BEGIN → fetch+lock events → call fn → mark processed → COMMIT.
// If fn returns nil (no IDs to mark), the transaction is rolled back as a no-op.
func (r *Repository) ProcessOutboxBatch(ctx context.Context, batchSize int32, fn func(events []postgres.FetchAndLockOutboxEventsRow) []int64) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	qtx := r.q.WithTx(tx)

	events, err := qtx.FetchAndLockOutboxEvents(ctx, batchSize)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not fetch outbox events: %w", err)
	}

	processedIDs := fn(events)

	if len(processedIDs) == 0 {
		_ = tx.Rollback()
		return nil
	}

	if err := qtx.MarkOutboxEventsProcessed(ctx, processedIDs); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not mark outbox events as processed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}

// CleanupProcessedOutboxEvents deletes outbox events that were processed before the given time.
func (r *Repository) CleanupProcessedOutboxEvents(ctx context.Context, before time.Time) error {
	if err := r.q.CleanupProcessedOutboxEvents(ctx, before); err != nil {
		return fmt.Errorf("could not cleanup outbox events: %w", err)
	}
	return nil
}
