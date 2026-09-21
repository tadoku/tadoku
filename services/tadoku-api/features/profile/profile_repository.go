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

func (r *Repository) SynchronizeUser(ctx context.Context, user SignedUser, now time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = queries.New(executor).SynchronizeUser(ctx, queries.SynchronizeUserParams{
		ID:               pgtype.UUID{Bytes: user.ID, Valid: true},
		DisplayName:      user.DisplayName,
		SessionCreatedAt: pgtype.Timestamp{Time: user.SessionCreatedAt(), Valid: true},
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

func (r *Repository) LockUser(ctx context.Context, userID uuid.UUID) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	user, err := queries.New(executor).LockUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrLocalUserNotFound
	}
	if err != nil {
		return fmt.Errorf("lock local user: %w", err)
	}
	if user.DeletionLockedAt.Valid || user.DeletedAt.Valid {
		return ErrAccountDeletionInProgress
	}
	return nil
}
