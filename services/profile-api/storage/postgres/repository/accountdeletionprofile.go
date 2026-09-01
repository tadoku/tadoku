package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) DeleteProfile(ctx context.Context, identityID uuid.UUID) error {
	if err := r.q.DeleteProfileForAccountDeletion(ctx, identityID); err != nil {
		return fmt.Errorf("could not delete profile for account deletion: %w", err)
	}
	return nil
}

func (r *Repository) ListAccountDeletionSuppressedIdentityIDs(ctx context.Context) ([]uuid.UUID, error) {
	identityIDs, err := r.q.ListAccountDeletionSuppressedIdentityIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list account deletion suppressions: %w", err)
	}
	return identityIDs, nil
}
