package logs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *LogsRepository) ListUnits(ctx context.Context) ([]Unit, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListUnits(ctx)
	if err != nil {
		return nil, fmt.Errorf("list log units: %w", err)
	}

	result := make([]Unit, 0, len(rows))
	for _, row := range rows {
		result = append(result, Unit{
			ID:            uuid.UUID(row.ID.Bytes),
			Key:           row.UnitKey,
			LogActivityID: int(row.LogActivityID),
			Name:          row.Name,
			Modifier:      row.Modifier,
			LanguageCode:  postgres.TextPointer(row.LanguageCode),
		})
	}
	return result, nil
}

func (r *LogsRepository) ListUserLanguageCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	codes, err := queries.New(executor).ListDistinctLanguageCodesForUser(ctx, postgres.UUID(userID))
	if err != nil {
		return nil, fmt.Errorf("list user log languages: %w", err)
	}
	return codes, nil
}

func (r *LogsRepository) ListTagSuggestions(ctx context.Context, userID uuid.UUID, query string) ([]TagSuggestion, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListTagSuggestionsForUser(ctx, queries.ListTagSuggestionsForUserParams{
		UserID: postgres.UUID(userID),
		Query:  pgtype.Text{String: query, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list tag suggestions: %w", err)
	}

	result := make([]TagSuggestion, 0, len(rows))
	for _, row := range rows {
		result = append(result, TagSuggestion{Tag: row.Tag, Count: int(row.UsageCount)})
	}
	return result, nil
}

func (r *LogsRepository) FindLog(ctx context.Context, id uuid.UUID, includeDeleted bool) (*Log, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindLogByID(ctx, queries.FindLogByIDParams{
		ID:             postgres.UUID(id),
		IncludeDeleted: includeDeleted,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrLogNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find log: %w", err)
	}

	tracking := logscore.Tracking{
		DurationSeconds: postgres.Int4Pointer(row.DurationSeconds),
		Score:           row.Score.Float32,
		RuleSetID:       postgres.UUIDPointer(row.ScoreRuleSetID),
		RuleIDs:         postgres.UUIDs(row.ScoreRuleIds),
		Rates:           row.ScoreRates,
	}
	if row.ScoreSource.Valid {
		tracking.Source = row.ScoreSource.String
	}
	if row.Amount.Valid && row.Modifier.Valid {
		tracking.UnitID = postgres.UUIDPointer(row.UnitID)
		tracking.UnitKey = row.UnitKey
		tracking.Amount = postgres.Float4Pointer(row.Amount)
		tracking.Modifier = postgres.Float4Pointer(row.Modifier)
	}

	return &Log{
		ID:              uuid.UUID(row.ID.Bytes),
		UserID:          uuid.UUID(row.UserID.Bytes),
		Description:     postgres.TextPointer(row.Description),
		LanguageCode:    row.LanguageCode,
		LanguageName:    row.LanguageName,
		Activity:        activities.Activity{ID: int32(row.ActivityID)},
		UnitID:          uuid.UUID(row.UnitID.Bytes),
		UnitKey:         row.UnitKey,
		UnitName:        row.UnitName,
		Tags:            row.Tags,
		Amount:          row.Amount.Float32,
		Modifier:        row.Modifier.Float32,
		Score:           row.Score.Float32,
		DurationSeconds: postgres.Int4Pointer(row.DurationSeconds),
		CreatedAt:       row.CreatedAt.Time,
		Deleted:         row.DeletedAt.Valid,
		UserDisplayName: &row.UserDisplayName,
		Tracking:        tracking,
	}, nil
}

func (r *LogsRepository) ListUserLogs(
	ctx context.Context,
	parameters ListParameters,
	createdOn *time.Time,
) (*LogList, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListLogsForUser(ctx, queries.ListLogsForUserParams{
		UserID:         postgres.NullableUUID(parameters.UserID),
		StartFrom:      parameters.offset(),
		PageSize:       int32(parameters.PageSize),
		IncludeDeleted: parameters.IncludeDeleted,
		CreatedOn:      postgres.NullableDate(createdOn),
	})
	if err != nil {
		return nil, fmt.Errorf("list logs: %w", err)
	}

	result := &LogList{Logs: make([]Log, 0, len(rows))}
	for _, row := range rows {
		result.TotalSize = int(row.TotalSize)
		if !row.ID.Valid {
			continue
		}
		result.Logs = append(result.Logs, Log{
			ID:              uuid.UUID(row.ID.Bytes),
			UserID:          uuid.UUID(row.UserID.Bytes),
			Description:     postgres.TextPointer(row.Description),
			LanguageCode:    row.LanguageCode.String,
			LanguageName:    row.LanguageName.String,
			Activity:        activities.Activity{ID: int32(row.ActivityID.Int16)},
			UnitID:          uuid.UUID(row.UnitID.Bytes),
			UnitKey:         row.UnitKey.String,
			UnitName:        row.UnitName.String,
			Tags:            row.Tags,
			Amount:          row.Amount.Float32,
			Modifier:        row.Modifier.Float32,
			Score:           row.Score.Float32,
			DurationSeconds: postgres.Int4Pointer(row.DurationSeconds),
			CreatedAt:       row.CreatedAt.Time,
			Deleted:         row.DeletedAt.Valid,
		})
	}
	if int64(parameters.offset())+int64(parameters.PageSize) < int64(result.TotalSize) {
		result.NextPageToken = fmt.Sprint(parameters.Page + 1)
	}
	return result, nil
}

func (r *LogsRepository) ListContestLogs(ctx context.Context, parameters ListParameters) (*LogList, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListLogsForContest(ctx, queries.ListLogsForContestParams{
		UserID:         postgres.NullableUUID(parameters.UserID),
		StartFrom:      parameters.offset(),
		PageSize:       int32(parameters.PageSize),
		IncludeDeleted: parameters.IncludeDeleted,
		ContestID:      postgres.UUID(parameters.ContestID),
	})
	if err != nil {
		return nil, fmt.Errorf("list logs: %w", err)
	}

	result := &LogList{Logs: make([]Log, 0, len(rows))}
	for _, row := range rows {
		result.TotalSize = int(row.TotalSize)
		if !row.ID.Valid {
			continue
		}
		result.Logs = append(result.Logs, Log{
			ID:              uuid.UUID(row.ID.Bytes),
			UserID:          uuid.UUID(row.UserID.Bytes),
			Description:     postgres.TextPointer(row.Description),
			LanguageCode:    row.LanguageCode.String,
			LanguageName:    row.LanguageName.String,
			Activity:        activities.Activity{ID: int32(row.ActivityID.Int16)},
			UnitID:          uuid.UUID(row.UnitID.Bytes),
			UnitKey:         row.UnitKey.String,
			UnitName:        row.UnitName.String,
			Tags:            row.Tags,
			Amount:          row.Amount.Float32,
			Modifier:        row.Modifier.Float32,
			Score:           row.Score.Float32,
			DurationSeconds: postgres.Int4Pointer(row.DurationSeconds),
			CreatedAt:       row.CreatedAt.Time,
			Deleted:         row.DeletedAt.Valid,
			UserDisplayName: &row.UserDisplayName.String,
		})
	}
	if int64(parameters.offset())+int64(parameters.PageSize) < int64(result.TotalSize) {
		result.NextPageToken = fmt.Sprint(parameters.Page + 1)
	}
	return result, nil
}

func (r *LogsRepository) ListActiveUserLogsAfter(
	ctx context.Context,
	userID uuid.UUID,
	after *exportCursor,
	limit int32,
) ([]ExportedLog, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	parameters := queries.ListActiveLogsForUserAfterParams{
		UserID:    postgres.UUID(userID),
		BatchSize: limit,
	}
	if after != nil {
		parameters.AfterCreatedAt = postgres.Timestamp(after.createdAt)
		parameters.AfterID = postgres.UUID(after.id)
	}

	rows, err := queries.New(executor).ListActiveLogsForUserAfter(ctx, parameters)
	if err != nil {
		return nil, fmt.Errorf("list active logs for export: %w", err)
	}

	result := make([]ExportedLog, 0, len(rows))
	for _, row := range rows {
		log := ExportedLog{
			ID:              uuid.UUID(row.ID.Bytes),
			CreatedAt:       row.CreatedAt.Time,
			LanguageCode:    row.LanguageCode,
			LanguageName:    row.LanguageName,
			Activity:        activities.Activity{ID: int32(row.LogActivityID)},
			DurationSeconds: postgres.Int4Pointer(row.DurationSeconds),
			Score:           row.Score.Float32,
			Tags:            row.Tags,
			Description:     postgres.TextPointer(row.Description),
		}
		if row.Amount.Valid {
			log.Amount = &ExportedAmount{Value: row.Amount.Float32, UnitName: row.UnitName.String}
		}
		result = append(result, log)
	}
	return result, nil
}
