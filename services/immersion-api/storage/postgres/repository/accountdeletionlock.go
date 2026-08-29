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
	if err := r.q.LockAccountForDeletion(ctx, postgres.LockAccountForDeletionParams{
		ID:       req.UserID,
		LockedAt: sql.NullTime{Time: req.LockedAt(), Valid: true},
	}); err != nil {
		return fmt.Errorf("could not lock account for deletion: %w", err)
	}
	return nil
}
