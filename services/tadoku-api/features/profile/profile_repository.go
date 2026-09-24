package profile

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SynchronizeUser(
	ctx context.Context,
	userID uuid.UUID,
	displayName string,
	sessionCreatedAt time.Time,
	now time.Time,
) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = queries.New(executor).SynchronizeUser(ctx, queries.SynchronizeUserParams{
		ID:               postgres.UUID(userID),
		DisplayName:      displayName,
		SessionCreatedAt: pgtype.Timestamp{Time: sessionCreatedAt, Valid: true},
		CreatedAt:        postgres.Timestamp(now),
		UpdatedAt:        postgres.Timestamp(now),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAccountDeletionInProgress
	}
	if err != nil {
		return fmt.Errorf("synchronize local user: %w", err)
	}
	return nil
}

func (r *Repository) LockUser(ctx context.Context, userID uuid.UUID) (UserDeletionState, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return UserDeletionState{}, err
	}
	user, err := queries.New(executor).LockUser(ctx, postgres.UUID(userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return UserDeletionState{}, ErrLocalUserNotFound
	}
	if err != nil {
		return UserDeletionState{}, fmt.Errorf("lock local user: %w", err)
	}
	return UserDeletionState{
		DeletionLocked: user.DeletionLockedAt.Valid,
		Deleted:        user.DeletedAt.Valid,
	}, nil
}

func (r *Repository) DisplayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	values := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		values[i] = postgres.UUID(id)
	}
	rows, err := queries.New(executor).FindUserDisplayNames(ctx, values)
	if err != nil {
		return nil, fmt.Errorf("fetch user display names: %w", err)
	}
	names := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		names[uuid.UUID(row.ID.Bytes)] = row.DisplayName
	}
	return names, nil
}
