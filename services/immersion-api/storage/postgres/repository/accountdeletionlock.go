package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func lockUserForMutation(ctx context.Context, q *postgres.Queries, userID uuid.UUID) error {
	user, err := q.LockUserForMutation(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("could not lock user for mutation: %w", err)
	}
	if user.DeletionLockedAt.Valid || user.DeletedAt.Valid {
		return domain.ErrAccountDeletionInProgress
	}
	return nil
}

func (r *Repository) LockAccountForDeletion(ctx context.Context, req *domain.AccountDeletionLockRequest) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.q.WithTx(tx)
	if err := qtx.EnsureAccountDeletionTarget(ctx, req.UserID); err != nil {
		return fmt.Errorf("could not ensure account deletion target: %w", err)
	}
	target, err := qtx.LockAccountDeletionTarget(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("could not lock account deletion target: %w", err)
	}
	if target.DeletionLockedAt.Valid || target.DeletedAt.Valid {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("could not commit existing account deletion lock: %w", err)
		}
		return nil
	}

	availableAfter, err := findRunningOwnedContestAvailableAfter(ctx, qtx, req.UserID, req.LockedAt())
	if err != nil {
		return fmt.Errorf("could not check running contest ownership: %w", err)
	}
	if availableAfter != nil {
		return domain.ErrRunningContestOwned
	}

	if err := qtx.SetAccountDeletionLock(ctx, postgres.SetAccountDeletionLockParams{
		ID:       req.UserID,
		LockedAt: sql.NullTime{Time: req.LockedAt(), Valid: true},
	}); err != nil {
		return fmt.Errorf("could not set account deletion lock: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit account deletion lock: %w", err)
	}
	return nil
}
