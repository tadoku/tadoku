package leaderboard

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) contestExists(ctx context.Context, id uuid.UUID) (bool, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return false, err
	}
	exists, err := queries.New(executor).ContestExists(ctx, postgres.UUID(id))
	if err != nil {
		return false, fmt.Errorf("check contest: %w", err)
	}
	return exists, nil
}

func (r *Repository) contestScores(ctx context.Context, request ContestRequest) ([]score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).ContestLeaderboard(ctx, queries.ContestLeaderboardParams{
		ContestID:    postgres.UUID(request.ContestID),
		LanguageCode: postgres.NullableNonEmptyText(request.LanguageCode),
		ActivityID:   postgres.NullableInt4(request.ActivityID),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch contest leaderboard: %w", err)
	}

	scores := make([]score, len(rows))
	for i, row := range rows {
		scores[i] = score{userID: uuid.UUID(row.UserID.Bytes), value: float64(row.Score)}
	}
	return scores, nil
}

// officialScores reads the yearly board for year, or the global board when year is nil.
func (r *Repository) officialScores(ctx context.Context, year *int16, request Request) ([]score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).OfficialLeaderboard(ctx, queries.OfficialLeaderboardParams{
		Year:         postgres.NullableInt2(year),
		LanguageCode: postgres.NullableNonEmptyText(request.LanguageCode),
		ActivityID:   postgres.NullableInt4(request.ActivityID),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch official leaderboard: %w", err)
	}

	scores := make([]score, len(rows))
	for i, row := range rows {
		scores[i] = score{userID: uuid.UUID(row.UserID.Bytes), value: float64(row.Score)}
	}
	return scores, nil
}
