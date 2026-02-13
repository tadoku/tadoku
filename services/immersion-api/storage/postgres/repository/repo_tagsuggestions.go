package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FetchTagSuggestionsForUser(ctx context.Context, userID uuid.UUID, query string) ([]domain.TagSuggestion, error) {
	rows, err := r.q.ListTagSuggestionsForUser(ctx, postgres.ListTagSuggestionsForUserParams{
		UserID: userID,
		Query:  sql.NullString{String: query, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	tags := make([]domain.TagSuggestion, len(rows))
	for i, row := range rows {
		tags[i] = domain.TagSuggestion{Tag: row.Tag, Count: int(row.UsageCount)}
	}
	return tags, nil
}

func (r *Repository) FetchDefaultTagsMatching(ctx context.Context, query string) ([]string, error) {
	query = strings.ToLower(query)
	return r.q.ListDefaultTagsMatching(ctx, sql.NullString{String: query, Valid: true})
}
