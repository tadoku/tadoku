package profile

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) ListAccountDeletionSuppressedIdentityIDs(ctx context.Context) ([]string, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	identityIDs, err := queries.New(executor).ListAccountDeletionSuppressedIdentityIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list account deletion suppressions: %w", err)
	}
	return identityIDs, nil
}
