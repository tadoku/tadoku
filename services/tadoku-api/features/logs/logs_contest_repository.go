package logs

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *LogsRepository) ContestLogViews(
	ctx context.Context,
	logID uuid.UUID,
	contestID *uuid.UUID,
) ([]contestLogView, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).ReadContestLogViews(ctx, queries.ReadContestLogViewsParams{
		LogID:     postgres.UUID(logID),
		ContestID: postgres.NullableUUID(contestID),
	})
	if err != nil {
		return nil, err
	}

	views := make([]contestLogView, 0, len(rows))
	for _, row := range rows {
		views = append(views, contestLogViewFromRow(row))
	}
	return views, nil
}

func contestLogViewFromRow(row queries.ReadContestLogViewsRow) contestLogView {
	return contestLogView{
		ContestID: uuid.UUID(row.ContestID.Bytes),
		Log: Log{
			ID:              uuid.UUID(row.ID.Bytes),
			UserID:          uuid.UUID(row.UserID.Bytes),
			UserDisplayName: &row.UserDisplayName,
			LanguageCode:    row.LanguageCode,
			LanguageName:    row.LanguageName,
			Activity:        activities.Activity{ID: int32(row.ActivityID)},
			UnitName:        row.UnitName,
			Description:     postgres.TextPointer(row.Description),
			Amount:          row.Amount,
			Modifier:        row.Modifier,
			DurationSeconds: postgres.Int4Pointer(row.DurationSeconds),
			Score:           row.Score,
			CreatedAt:       row.CreatedAt.Time,
			Tags:            row.Tags,
		},
	}
}
