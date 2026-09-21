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

func (r *Repository) YearlyActivity(ctx context.Context, userID uuid.UUID, year int16) ([]ActivityScore, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).YearlyActivityForUser(ctx, queries.YearlyActivityForUserParams{
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
		Year:   year,
	})
	if err != nil {
		return nil, fmt.Errorf("YearlyActivity: %w", err)
	}

	result := make([]ActivityScore, 0, len(rows))
	for _, row := range rows {
		result = append(result, ActivityScore{
			Date:    row.Date.Time,
			Score:   row.Score,
			Updates: int(row.UpdateCount),
		})
	}

	return result, nil
}

func (r *Repository) YearlyScores(ctx context.Context, userID uuid.UUID, year int16) ([]Score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).FetchScoresForProfile(ctx, queries.FetchScoresForProfileParams{
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
		Year:   year,
	})
	if err != nil {
		return nil, fmt.Errorf("YearlyScores: %w", err)
	}

	result := make([]Score, 0, len(rows))
	for _, row := range rows {
		result = append(result, Score{
			LanguageCode: row.LanguageCode,
			LanguageName: row.LanguageName,
			Score:        row.Score,
		})
	}

	return result, nil
}

func (r *Repository) YearlyActivitySplit(ctx context.Context, userID uuid.UUID, year int16) ([]ActivitySplitScore, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).YearlyActivitySplitForUser(ctx, queries.YearlyActivitySplitForUserParams{
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
		Year:   year,
	})
	if err != nil {
		return nil, fmt.Errorf("YearlyActivitySplit: %w", err)
	}

	result := make([]ActivitySplitScore, 0, len(rows))
	for _, row := range rows {
		result = append(result, ActivitySplitScore{
			ActivityID: int(row.LogActivityID),
			Score:      row.Score,
		})
	}

	return result, nil
}
