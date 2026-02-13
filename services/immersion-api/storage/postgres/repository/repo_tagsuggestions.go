package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FetchTagSuggestionsForUser(ctx context.Context, userID uuid.UUID, query string) ([]string, error) {
	return r.q.ListTagSuggestionsForUser(ctx, postgres.ListTagSuggestionsForUserParams{
		UserID: userID,
		Query:  postgres.NewNullString(&query),
	})
}

func (r *Repository) FetchDefaultTagsMatching(ctx context.Context, query string) ([]string, error) {
	// Convert query to lowercase for case-insensitive matching
	query = strings.ToLower(query)
	return r.q.ListDefaultTagsMatching(ctx, postgres.NewNullString(&query))
}
