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
	userTags []domain.TagSuggestion
	userErr  error
}

func (m *mockTagSuggestionsRepository) FetchTagSuggestionsForUser(ctx context.Context, userID uuid.UUID, query string) ([]domain.TagSuggestion, error) {
	return m.userTags, m.userErr
}

func TestTagSuggestions_Execute(t *testing.T) {
	userID := uuid.New()

	t.Run("returns user tags for authenticated user", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{
			userTags: []domain.TagSuggestion{{Tag: "book", Count: 5}, {Tag: "fiction", Count: 3}},
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
			userTags: []domain.TagSuggestion{{Tag: "book", Count: 5}},
		}
		svc := domain.NewTagSuggestions(repo)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: ""})

		require.NoError(t, err)
		assert.Equal(t, []domain.TagSuggestion{
			{Tag: "book", Count: 5},
			{Tag: "anime", Count: 0},
			{Tag: "audiobook", Count: 0},
			{Tag: "chat", Count: 0},
		}, result.Suggestions[:4])
	})

	t.Run("deduplicates suggestions case-insensitively", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{
			userTags: []domain.TagSuggestion{{Tag: "Book", Count: 3}, {Tag: "fiction", Count: 1}},
		}
		svc := domain.NewTagSuggestions(repo)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: ""})

		require.NoError(t, err)
		assert.Equal(t, []domain.TagSuggestion{
			{Tag: "Book", Count: 3},
			{Tag: "fiction", Count: 1},
			{Tag: "anime", Count: 0},
			{Tag: "audiobook", Count: 0},
			{Tag: "chat", Count: 0},
		}, result.Suggestions[:5])
		for _, suggestion := range result.Suggestions {
			assert.NotEqual(t, "book", suggestion.Tag)
		}
	})

	t.Run("filters default tags case-insensitively", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{}
		svc := domain.NewTagSuggestions(repo)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: "VID"})

		require.NoError(t, err)
		assert.Equal(t, []domain.TagSuggestion{
			{Tag: "online video", Count: 0},
		}, result.Suggestions)
	})

	t.Run("caps total suggestions at 30", func(t *testing.T) {
		userTags := make([]domain.TagSuggestion, 29)
		for i := range userTags {
			userTags[i] = domain.TagSuggestion{Tag: string(rune('a' + i)), Count: 1}
		}
		repo := &mockTagSuggestionsRepository{userTags: userTags}
		svc := domain.NewTagSuggestions(repo)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: ""})

		require.NoError(t, err)
		require.Len(t, result.Suggestions, 30)
		assert.Equal(t, domain.TagSuggestion{Tag: "anime", Count: 0}, result.Suggestions[29])
	})

	t.Run("returns unauthorized for unauthenticated user", func(t *testing.T) {
		repo := &mockTagSuggestionsRepository{}
		svc := domain.NewTagSuggestions(repo)

		ctx := context.Background()

		_, err := svc.Execute(ctx, &domain.TagSuggestionsRequest{Query: "bo"})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})
}
