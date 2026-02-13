package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FetchTagSuggestionsForUser(ctx context.Context, userID uuid.UUID, query string) ([]string, error) {
	rows, err := r.q.ListTagSuggestionsForUser(ctx, postgres.ListTagSuggestionsForUserParams{
		UserID: userID,
		Query:  sql.NullString{String: query, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	tags := make([]string, len(rows))
	for i, row := range rows {
		tags[i] = row.Tag
	}
	return tags, nil
}

func (r *Repository) FetchDefaultTagsMatching(ctx context.Context, query string) ([]string, error) {
	query = strings.ToLower(query)
	return r.q.ListDefaultTagsMatching(ctx, sql.NullString{String: query, Valid: true})
}
