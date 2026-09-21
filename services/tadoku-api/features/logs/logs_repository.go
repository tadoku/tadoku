package logs

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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
