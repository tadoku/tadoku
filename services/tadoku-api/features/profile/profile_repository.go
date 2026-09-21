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
		ID:               pgtype.UUID{Bytes: userID, Valid: true},
		DisplayName:      displayName,
		SessionCreatedAt: pgtype.Timestamp{Time: sessionCreatedAt, Valid: true},
		CreatedAt:        pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:        pgtype.Timestamp{Time: now, Valid: true},
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
	user, err := queries.New(executor).LockUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
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
