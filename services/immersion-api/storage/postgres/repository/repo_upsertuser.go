package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) UpsertUser(ctx context.Context, req *domain.UserUpsertRequest) error {
	if err := r.q.UpsertUser(ctx, postgres.UpsertUserParams{
		ID:               req.ID,
		DisplayName:      req.DisplayName,
		SessionCreatedAt: req.SessionCreatedAt,
	}); err != nil {
		return fmt.Errorf("could not update user: %w", err)
	}

	return nil
}
