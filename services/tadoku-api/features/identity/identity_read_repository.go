package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/identity/reads"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type readRepository struct{ pool *pgxpool.Pool }

func (r *readRepository) find(ctx context.Context, id uuid.UUID) (Account, error) {
	executor, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return Account{}, err
	}
	row, err := queries.New(executor).FindAccount(ctx, postgres.UUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, errx.NewNotFoundError("identity not found")
	}
	if err != nil {
		return Account{}, fmt.Errorf("find account: %w", err)
	}
	return Account{
		ID:          uuid.UUID(row.UserID.Bytes),
		Email:       Email{value: row.Email},
		DisplayName: DisplayName{value: row.DisplayName},
		CreatedAt:   row.CreatedAt.Time,
	}, nil
}

func (r *readRepository) list(ctx context.Context, query string, after uuid.UUID, limit int) ([]Account, error) {
	executor, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).ListAccounts(ctx, queries.ListAccountsParams{
		AfterID:   postgres.UUID(after),
		Query:     query,
		PageLimit: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	accounts := make([]Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, Account{
			ID:          uuid.UUID(row.UserID.Bytes),
			Email:       Email{value: row.Email},
			DisplayName: DisplayName{value: row.DisplayName},
			CreatedAt:   row.CreatedAt.Time,
		})
	}
	return accounts, nil
}
