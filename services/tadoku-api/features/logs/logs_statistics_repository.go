package logs

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *LogsRepository) YearlyActivity(ctx context.Context, userID uuid.UUID, year int16) ([]ActivityScore, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).YearlyActivityForUser(ctx, queries.YearlyActivityForUserParams{
		UserID: postgres.UUID(userID),
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

func (r *LogsRepository) YearlyScores(ctx context.Context, userID uuid.UUID, year int16) ([]Score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).FetchScoresForProfile(ctx, queries.FetchScoresForProfileParams{
		UserID: postgres.UUID(userID),
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

func (r *LogsRepository) YearlyActivitySplit(ctx context.Context, userID uuid.UUID, year int16) ([]ActivitySplitScore, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).YearlyActivitySplitForUser(ctx, queries.YearlyActivitySplitForUserParams{
		UserID: postgres.UUID(userID),
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

func (r *LogsRepository) ContestScores(ctx context.Context, userID, contestID uuid.UUID) ([]Score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).FetchScoresForContestProfile(ctx, queries.FetchScoresForContestProfileParams{
		UserID:    postgres.UUID(userID),
		ContestID: postgres.UUID(contestID),
	})
	if err != nil {
		return nil, fmt.Errorf("ContestScores: %w", err)
	}

	result := make([]Score, 0, len(rows))
	for _, row := range rows {
		result = append(result, Score{
			LanguageCode: row.LanguageCode,
			Score:        row.Score,
		})
	}
	return result, nil
}

func (r *LogsRepository) ContestActivity(ctx context.Context, userID, contestID uuid.UUID) ([]ContestActivity, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ActivityPerLanguageForContestProfile(ctx, queries.ActivityPerLanguageForContestProfileParams{
		UserID:    postgres.UUID(userID),
		ContestID: postgres.UUID(contestID),
	})
	if err != nil {
		return nil, fmt.Errorf("ContestActivity: %w", err)
	}

	result := make([]ContestActivity, 0, len(rows))
	for _, row := range rows {
		result = append(result, ContestActivity{
			Date:         row.Date.Time,
			LanguageCode: row.LanguageCode,
			Score:        row.Score,
		})
	}
	return result, nil
}
