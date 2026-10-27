package identity

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (r *Repository) accountHasEmail(ctx context.Context, id uuid.UUID, emailNormalized string) (bool, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return false, err
	}
	return queries.New(db).AccountHasEmail(ctx, queries.AccountHasEmailParams{
		UserID:          postgres.UUID(id),
		EmailNormalized: emailNormalized,
	})
}

func (r *Repository) deleteAccount(ctx context.Context, id uuid.UUID) (Email, bool, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return Email{}, false, err
	}
	stored, err := queries.New(db).DeleteAccount(ctx, postgres.UUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Email{}, false, nil
	}
	if err != nil {
		return Email{}, false, err
	}
	return Email{stored}, true, nil
}
