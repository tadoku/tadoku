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
	userTags    []string
	userErr     error
	defaultTags []string
	defaultErr  error
}

func (m *mockTagSuggestionsRepository) FetchTagSuggestionsForUser(ctx context.Context, userID uuid.UUID, query string) ([]string, error) {
	return m.userTags, m.userErr
}

func (m *mockTagSuggestionsRepository) FetchDefaultTagsMatching(ctx context.Context, query string) ([]string, error) {
	return m.defaultTags, m.defaultErr
}

func TestTagSuggestions_Execute(t *testing.T) {
	userID := uuid.New()

	t.Run("returns user tags for authenticated user", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{
			userTags:    []string{"book", "fiction"},
			defaultTags: []string{"comic", "ebook"},
		}
		svc := domain.NewTagSuggestions(repo)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: "bo"})

		require.NoError(t, err)
		assert.Contains(t, result.Suggestions, "book")
		assert.Contains(t, result.Suggestions, "fiction")
	})

	t.Run("appends default tags after user tags", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{
			userTags:    []string{"book"},
			defaultTags: []string{"fiction", "non-fiction"},
		}
		svc := domain.NewTagSuggestions(repo)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: ""})

		require.NoError(t, err)
		assert.Equal(t, []string{"book", "fiction", "non-fiction"}, result.Suggestions)
	})

	t.Run("deduplicates suggestions case-insensitively", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{
			userTags:    []string{"Book", "fiction"},
			defaultTags: []string{"book", "ebook"}, // "book" is case-insensitive duplicate of "Book"
		}
		svc := domain.NewTagSuggestions(repo)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: ""})

		require.NoError(t, err)
		assert.Equal(t, []string{"Book", "fiction", "ebook"}, result.Suggestions)
	})

	t.Run("returns only default tags for unauthenticated user", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{
			defaultTags: []string{"book", "fiction"},
		}
		svc := domain.NewTagSuggestions(repo)

		ctx := context.Background()

		result, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: "bo"})

		require.NoError(t, err)
		assert.Contains(t, result.Suggestions, "book")
		assert.Contains(t, result.Suggestions, "fiction")
	})
}
