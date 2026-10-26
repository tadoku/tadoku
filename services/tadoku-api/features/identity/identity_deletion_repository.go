package identity

import (
	"context"

	"github.com/google/uuid"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *Repository) lockAccountForRequest(ctx context.Context, id uuid.UUID) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).LockAccountForRequest(ctx, postgres.UUID(id))
}

func (r *Repository) lockAccountForDeletion(ctx context.Context, id uuid.UUID) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).LockAccountForDeletion(ctx, postgres.UUID(id))
}

func (r *Repository) deleteAccount(ctx context.Context, id uuid.UUID) ([]Email, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	stored, err := queries.New(db).DeleteAccount(ctx, postgres.UUID(id))
	if err != nil {
		return nil, err
	}

	emails := make([]Email, 0, len(stored))
	for _, value := range stored {
		emails = append(emails, Email{value})
	}
	return emails, nil
}
