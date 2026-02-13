package domain_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockTagSuggestionsRepository struct {
	userTags    []domain.TagSuggestion
	userErr     error
	defaultTags []string
	defaultErr  error
}

func (m *mockTagSuggestionsRepository) FetchTagSuggestionsForUser(ctx context.Context, userID uuid.UUID, query string) ([]domain.TagSuggestion, error) {
	return m.userTags, m.userErr
}

func (m *mockTagSuggestionsRepository) FetchDefaultTagsMatching(ctx context.Context, query string) ([]string, error) {
	return m.defaultTags, m.defaultErr
}

func TestTagSuggestions_Execute(t *testing.T) {
	userID := uuid.New()

	t.Run("returns user tags for authenticated user", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{
			userTags:    []domain.TagSuggestion{{Tag: "book", Count: 5}, {Tag: "fiction", Count: 3}},
			defaultTags: []string{"comic", "ebook"},
		}
		svc := domain.NewTagSuggestions(repo)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: "bo"})

		require.NoError(t, err)
		assert.Equal(t, domain.TagSuggestion{Tag: "book", Count: 5}, result.Suggestions[0])
		assert.Equal(t, domain.TagSuggestion{Tag: "fiction", Count: 3}, result.Suggestions[1])
	})

	t.Run("appends default tags after user tags with count 0", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{
			userTags:    []domain.TagSuggestion{{Tag: "book", Count: 5}},
			defaultTags: []string{"fiction", "non-fiction"},
		}
		svc := domain.NewTagSuggestions(repo)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: ""})

		require.NoError(t, err)
		assert.Equal(t, []domain.TagSuggestion{
			{Tag: "book", Count: 5},
			{Tag: "fiction", Count: 0},
			{Tag: "non-fiction", Count: 0},
		}, result.Suggestions)
	})

	t.Run("deduplicates suggestions case-insensitively", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{
			userTags:    []domain.TagSuggestion{{Tag: "Book", Count: 3}, {Tag: "fiction", Count: 1}},
			defaultTags: []string{"book", "ebook"}, // "book" is case-insensitive duplicate of "Book"
		}
		svc := domain.NewTagSuggestions(repo)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: ""})

		require.NoError(t, err)
		assert.Equal(t, []domain.TagSuggestion{
			{Tag: "Book", Count: 3},
			{Tag: "fiction", Count: 1},
			{Tag: "ebook", Count: 0},
		}, result.Suggestions)
	})

	t.Run("returns unauthorized for unauthenticated user", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{}
		svc := domain.NewTagSuggestions(repo)

		ctx := context.Background()

		_, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: "bo"})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})
}
