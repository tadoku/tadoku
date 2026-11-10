package logs

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *LogsRepository) ContestProfileDays(
	ctx context.Context,
	userID, contestID uuid.UUID,
	filters queryFilters,
) ([]ContestDayScore, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ContestProfileSummaryDays(ctx, queries.ContestProfileSummaryDaysParams{
		ContestID:     postgres.UUID(contestID),
		UserID:        postgres.UUID(userID),
		LanguageCodes: filters.languageCodes,
		ActivityIds:   filters.activityIDs,
		Tags:          filters.tags,
	})
	if err != nil {
		return nil, fmt.Errorf("ContestProfileDays: %w", err)
	}

	result := make([]ContestDayScore, 0, len(rows))
	for _, row := range rows {
		result = append(result, ContestDayScore{
			Date:       row.Date.Time,
			ActivityID: int(row.ActivityID),
			Score:      row.Score,
			logCount:   int(row.LogCount),
		})
	}
	return result, nil
}

func (r *LogsRepository) ContestProfileLanguages(
	ctx context.Context,
	userID, contestID uuid.UUID,
	filters queryFilters,
) ([]Score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ContestProfileSummaryLanguages(ctx, queries.ContestProfileSummaryLanguagesParams{
		ContestID:   postgres.UUID(contestID),
		UserID:      postgres.UUID(userID),
		ActivityIds: filters.activityIDs,
		Tags:        filters.tags,
	})
	if err != nil {
		return nil, fmt.Errorf("ContestProfileLanguages: %w", err)
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

func (r *LogsRepository) ContestProfileActivities(
	ctx context.Context,
	userID, contestID uuid.UUID,
	filters queryFilters,
) ([]ActivitySplitScore, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ContestProfileSummaryActivities(ctx, queries.ContestProfileSummaryActivitiesParams{
		ContestID:     postgres.UUID(contestID),
		UserID:        postgres.UUID(userID),
		LanguageCodes: filters.languageCodes,
		Tags:          filters.tags,
	})
	if err != nil {
		return nil, fmt.Errorf("ContestProfileActivities: %w", err)
	}

	result := make([]ActivitySplitScore, 0, len(rows))
	for _, row := range rows {
		result = append(result, ActivitySplitScore{
			ActivityID: int(row.ActivityID),
			Score:      row.TotalScore,
		})
	}
	return result, nil
}

func (r *LogsRepository) ContestProfileTags(
	ctx context.Context,
	userID, contestID uuid.UUID,
	filters queryFilters,
) ([]TagScore, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ContestProfileSummaryTags(ctx, queries.ContestProfileSummaryTagsParams{
		ContestID:     postgres.UUID(contestID),
		UserID:        postgres.UUID(userID),
		LanguageCodes: filters.languageCodes,
		ActivityIds:   filters.activityIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("ContestProfileTags: %w", err)
	}

	result := make([]TagScore, 0, len(rows))
	for _, row := range rows {
		result = append(result, TagScore{
			Tag:      row.Tag,
			Score:    row.TotalScore,
			LogCount: int(row.LogCount),
		})
	}
	return result, nil
}
