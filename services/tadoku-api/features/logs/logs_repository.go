package logs

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type LogsRepository struct{ db *pgxpool.Pool }

func NewLogsRepository(db *pgxpool.Pool) *LogsRepository { return &LogsRepository{db: db} }

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
		var languageCode *string
		if row.LanguageCode.Valid {
			languageCode = &row.LanguageCode.String
		}
		result = append(result, Unit{
			ID:            uuid.UUID(row.ID.Bytes),
			Key:           row.UnitKey,
			LogActivityID: int(row.LogActivityID),
			Name:          row.Name,
			Modifier:      row.Modifier,
			LanguageCode:  languageCode,
		})
	}
	return result, nil
}

func (r *LogsRepository) ListUserLanguageCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	codes, err := queries.New(executor).ListDistinctLanguageCodesForUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
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
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
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

func (r *LogsRepository) YearlyActivity(ctx context.Context, userID uuid.UUID, year int16) ([]ActivityScore, error) {
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

func (r *LogsRepository) YearlyScores(ctx context.Context, userID uuid.UUID, year int16) ([]Score, error) {
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

func (r *LogsRepository) YearlyActivitySplit(ctx context.Context, userID uuid.UUID, year int16) ([]ActivitySplitScore, error) {
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

func (r *LogsRepository) ContestScores(ctx context.Context, userID, contestID uuid.UUID) ([]Score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).FetchScoresForContestProfile(ctx, queries.FetchScoresForContestProfileParams{
		UserID:    pgtype.UUID{Bytes: userID, Valid: true},
		ContestID: pgtype.UUID{Bytes: contestID, Valid: true},
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
		UserID:    pgtype.UUID{Bytes: userID, Valid: true},
		ContestID: pgtype.UUID{Bytes: contestID, Valid: true},
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

func (r *LogsRepository) FindLog(ctx context.Context, id uuid.UUID, includeDeleted bool) (*Log, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindLogByID(ctx, queries.FindLogByIDParams{
		ID:             pgtype.UUID{Bytes: id, Valid: true},
		IncludeDeleted: includeDeleted,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrLogNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find log: %w", err)
	}

	return &Log{
		ID:              uuid.UUID(row.ID.Bytes),
		UserID:          uuid.UUID(row.UserID.Bytes),
		Description:     textPointer(row.Description),
		LanguageCode:    row.LanguageCode,
		LanguageName:    row.LanguageName,
		Activity:        activities.Activity{ID: int32(row.ActivityID)},
		UnitID:          uuid.UUID(row.UnitID.Bytes),
		UnitKey:         row.UnitKey,
		UnitName:        row.UnitName,
		Tags:            legacyLogTags(row.Tags),
		Amount:          row.Amount.Float32,
		Modifier:        row.Modifier.Float32,
		Score:           row.Score.Float32,
		DurationSeconds: intPointer(row.DurationSeconds),
		CreatedAt:       row.CreatedAt.Time,
		Deleted:         row.DeletedAt.Valid,
		UserDisplayName: &row.UserDisplayName,
	}, nil
}

func (r *LogsRepository) AttachedRegistrations(ctx context.Context, id uuid.UUID) ([]RegistrationReference, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).FindAttachedContestRegistrationsForLog(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("find attached registrations: %w", err)
	}

	result := make([]RegistrationReference, 0, len(rows))
	for _, row := range rows {
		result = append(result, RegistrationReference{
			RegistrationID:       uuid.UUID(row.ID.Bytes),
			ContestID:            uuid.UUID(row.ContestID.Bytes),
			ContestEnd:           row.ContestEnd.Time,
			Title:                row.Title,
			OwnerUserDisplayName: row.OwnerUserDisplayName,
			Official:             row.Official,
			Score:                row.Score.Float32,
		})
	}
	return result, nil
}

func (r *LogsRepository) ListUserLogs(ctx context.Context, parameters ListParameters) (*LogList, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var userID pgtype.UUID
	if parameters.UserID != nil {
		userID = pgtype.UUID{Bytes: *parameters.UserID, Valid: true}
	}
	rows, err := queries.New(executor).ListLogsForUser(ctx, queries.ListLogsForUserParams{
		UserID:         userID,
		StartFrom:      int32(parameters.Page * parameters.PageSize),
		PageSize:       int32(parameters.PageSize),
		IncludeDeleted: parameters.IncludeDeleted,
	})
	if err != nil {
		return nil, fmt.Errorf("list logs: %w", err)
	}

	result := &LogList{Logs: make([]Log, 0, len(rows))}
	for _, row := range rows {
		result.Logs = append(result.Logs, Log{
			ID:              uuid.UUID(row.ID.Bytes),
			UserID:          uuid.UUID(row.UserID.Bytes),
			Description:     textPointer(row.Description),
			LanguageCode:    row.LanguageCode,
			LanguageName:    row.LanguageName,
			Activity:        activities.Activity{ID: int32(row.ActivityID)},
			UnitID:          uuid.UUID(row.UnitID.Bytes),
			UnitKey:         row.UnitKey,
			UnitName:        row.UnitName,
			Tags:            legacyLogTags(row.Tags),
			Amount:          row.Amount.Float32,
			Modifier:        row.Modifier.Float32,
			Score:           row.Score.Float32,
			DurationSeconds: intPointer(row.DurationSeconds),
			CreatedAt:       row.CreatedAt.Time,
			Deleted:         row.DeletedAt.Valid,
		})
		result.TotalSize = int(row.TotalSize)
	}
	if parameters.Page*parameters.PageSize+parameters.PageSize < result.TotalSize {
		result.NextPageToken = fmt.Sprint(parameters.Page + 1)
	}
	return result, nil
}

func (r *LogsRepository) ListContestLogs(ctx context.Context, parameters ListParameters) (*LogList, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var userID pgtype.UUID
	if parameters.UserID != nil {
		userID = pgtype.UUID{Bytes: *parameters.UserID, Valid: true}
	}
	rows, err := queries.New(executor).ListLogsForContest(ctx, queries.ListLogsForContestParams{
		UserID:         userID,
		StartFrom:      int32(parameters.Page * parameters.PageSize),
		PageSize:       int32(parameters.PageSize),
		IncludeDeleted: parameters.IncludeDeleted,
		ContestID:      pgtype.UUID{Bytes: parameters.ContestID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list logs: %w", err)
	}

	result := &LogList{Logs: make([]Log, 0, len(rows))}
	for _, row := range rows {
		result.Logs = append(result.Logs, Log{
			ID:              uuid.UUID(row.ID.Bytes),
			UserID:          uuid.UUID(row.UserID.Bytes),
			Description:     textPointer(row.Description),
			LanguageCode:    row.LanguageCode,
			LanguageName:    row.LanguageName,
			Activity:        activities.Activity{ID: int32(row.ActivityID)},
			UnitID:          uuid.UUID(row.UnitID.Bytes),
			UnitKey:         row.UnitKey,
			UnitName:        row.UnitName,
			Tags:            legacyLogTags(row.Tags),
			Amount:          row.Amount.Float32,
			Modifier:        row.Modifier.Float32,
			Score:           row.Score.Float32,
			DurationSeconds: intPointer(row.DurationSeconds),
			CreatedAt:       row.CreatedAt.Time,
			Deleted:         row.DeletedAt.Valid,
			UserDisplayName: &row.UserDisplayName,
		})
		result.TotalSize = int(row.TotalSize)
	}
	if parameters.Page*parameters.PageSize+parameters.PageSize < result.TotalSize {
		result.NextPageToken = fmt.Sprint(parameters.Page + 1)
	}
	return result, nil
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func intPointer(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

// Preserve the legacy array-text decoder, including its handling of escaped
// quotes and backslashes. Changing this requires a separate response-contract fix.
func legacyLogTags(encoded string) []string {
	result := []string{}
	if len(encoded) <= 2 {
		return result
	}
	var current []byte
	quoted := false
	for _, ch := range []byte(encoded[1 : len(encoded)-1]) {
		switch {
		case ch == '"':
			quoted = !quoted
		case ch == ',' && !quoted:
			result = append(result, string(current))
			current = nil
		default:
			current = append(current, ch)
		}
	}
	return append(result, string(current))
}
