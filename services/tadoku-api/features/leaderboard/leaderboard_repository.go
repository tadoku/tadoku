package leaderboard

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	exists, err := queries.New(executor).ContestExists(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return false, fmt.Errorf("check contest: %w", err)
	}
	return exists, nil
}

func (r *Repository) contest(ctx context.Context, request ContestRequest) (*Leaderboard, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).LeaderboardForContest(ctx, queries.LeaderboardForContestParams{
		ContestID:    pgtype.UUID{Bytes: request.ContestID, Valid: true},
		LanguageCode: postgres.NullableText(sqlLanguageCode(request.LanguageCode)),
		ActivityID:   postgres.NullableInt4(request.ActivityID),
		StartFrom:    int32(request.Page * request.PageSize),
		PageSize:     int32(request.PageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch contest leaderboard: %w", err)
	}
	entries := make([]Entry, len(rows))
	for i, row := range rows {
		entries[i] = Entry{
			Rank:            int(row.Rank),
			UserID:          uuid.UUID(row.UserID.Bytes),
			UserDisplayName: row.UserDisplayName,
			Score:           row.Score,
			IsTie:           row.IsTie,
		}
	}
	total := 0
	if len(rows) > 0 {
		total = int(rows[0].TotalSize)
	}
	return result(entries, total, request.Page, request.PageSize), nil
}

func (r *Repository) yearly(ctx context.Context, request YearlyRequest) (*Leaderboard, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).YearlyLeaderboard(ctx, queries.YearlyLeaderboardParams{
		Year:         int16(request.Year),
		LanguageCode: postgres.NullableText(sqlLanguageCode(request.LanguageCode)),
		ActivityID:   postgres.NullableInt4(request.ActivityID),
		StartFrom:    int32(request.Page * request.PageSize),
		PageSize:     int32(request.PageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch yearly leaderboard: %w", err)
	}
	entries := make([]Entry, len(rows))
	for i, row := range rows {
		entries[i] = Entry{
			Rank:            int(row.Rank),
			UserID:          uuid.UUID(row.UserID.Bytes),
			UserDisplayName: row.UserDisplayName,
			Score:           row.Score,
			IsTie:           row.IsTie,
		}
	}
	total := 0
	if len(rows) > 0 {
		total = int(rows[0].TotalSize)
	}
	return result(entries, total, request.Page, request.PageSize), nil
}

func (r *Repository) global(ctx context.Context, request Request) (*Leaderboard, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).GlobalLeaderboard(ctx, queries.GlobalLeaderboardParams{
		LanguageCode: postgres.NullableText(sqlLanguageCode(request.LanguageCode)),
		ActivityID:   postgres.NullableInt4(request.ActivityID),
		StartFrom:    int32(request.Page * request.PageSize),
		PageSize:     int32(request.PageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch global leaderboard: %w", err)
	}
	entries := make([]Entry, len(rows))
	for i, row := range rows {
		entries[i] = Entry{
			Rank:            int(row.Rank),
			UserID:          uuid.UUID(row.UserID.Bytes),
			UserDisplayName: row.UserDisplayName,
			Score:           row.Score,
			IsTie:           row.IsTie,
		}
	}
	total := 0
	if len(rows) > 0 {
		total = int(rows[0].TotalSize)
	}
	return result(entries, total, request.Page, request.PageSize), nil
}

func (r *Repository) allContestScores(ctx context.Context, id uuid.UUID) ([]score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).ContestLeaderboardAllScores(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("fetch all contest leaderboard scores: %w", err)
	}
	res := make([]score, len(rows))
	for i, row := range rows {
		res[i] = score{userID: uuid.UUID(row.UserID.Bytes), value: float64(row.Score)}
	}
	return res, nil
}

func (r *Repository) allYearlyScores(ctx context.Context, year int) ([]score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).YearlyLeaderboardAllScores(ctx, int16(year))
	if err != nil {
		return nil, fmt.Errorf("fetch all yearly leaderboard scores: %w", err)
	}
	res := make([]score, len(rows))
	for i, row := range rows {
		res[i] = score{userID: uuid.UUID(row.UserID.Bytes), value: float64(row.Score)}
	}
	return res, nil
}

func (r *Repository) allGlobalScores(ctx context.Context) ([]score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).GlobalLeaderboardAllScores(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch all global leaderboard scores: %w", err)
	}
	res := make([]score, len(rows))
	for i, row := range rows {
		res[i] = score{userID: uuid.UUID(row.UserID.Bytes), value: float64(row.Score)}
	}
	return res, nil
}

func sqlLanguageCode(value *string) *string {
	if value != nil && *value == "" {
		return nil
	}
	return value
}
