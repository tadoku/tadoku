package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func TestValidateAndNormalizeTags(t *testing.T) {
	t.Run("returns empty slice for nil input", func(t *testing.T) {
		result, err := domain.ValidateAndNormalizeTags(nil)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("returns empty slice for empty input", func(t *testing.T) {
		result, err := domain.ValidateAndNormalizeTags([]string{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		result, err := domain.ValidateAndNormalizeTags([]string{"  book  ", "\tfiction\n"})
		require.NoError(t, err)
		assert.Equal(t, []string{"book", "fiction"}, result)
	})

	t.Run("converts to lowercase", func(t *testing.T) {
		result, err := domain.ValidateAndNormalizeTags([]string{"Book", "FICTION", "Non-Fiction"})
		require.NoError(t, err)
		assert.Equal(t, []string{"book", "fiction", "non-fiction"}, result)
	})

	t.Run("deduplicates tags case-insensitively", func(t *testing.T) {
		result, err := domain.ValidateAndNormalizeTags([]string{"book", "Book", "BOOK", "fiction"})
		require.NoError(t, err)
		assert.Equal(t, []string{"book", "fiction"}, result)
	})

	t.Run("skips empty tags after trimming", func(t *testing.T) {
		result, err := domain.ValidateAndNormalizeTags([]string{"book", "", "  ", "\t", "fiction"})
		require.NoError(t, err)
		assert.Equal(t, []string{"book", "fiction"}, result)
	})

	t.Run("returns error for tag exceeding max length", func(t *testing.T) {
		longTag := strings.Repeat("a", domain.MaxTagLength+1)
		_, err := domain.ValidateAndNormalizeTags([]string{longTag})
		assert.ErrorIs(t, err, domain.ErrInvalidTags)
	})

	t.Run("allows tag at max length", func(t *testing.T) {
		maxTag := strings.Repeat("a", domain.MaxTagLength)
		result, err := domain.ValidateAndNormalizeTags([]string{maxTag})
		require.NoError(t, err)
		assert.Equal(t, []string{maxTag}, result)
	})

	t.Run("returns error when exceeding max tags", func(t *testing.T) {
		tags := make([]string, domain.MaxTagsPerLog+1)
		for i := range tags {
			tags[i] = string(rune('a' + i))
		}
		_, err := domain.ValidateAndNormalizeTags(tags)
		assert.ErrorIs(t, err, domain.ErrInvalidTags)
	})

	t.Run("allows exactly max tags", func(t *testing.T) {
		tags := make([]string, domain.MaxTagsPerLog)
		for i := range tags {
			tags[i] = string(rune('a' + i))
		}
		result, err := domain.ValidateAndNormalizeTags(tags)
		require.NoError(t, err)
		assert.Len(t, result, domain.MaxTagsPerLog)
	})

	t.Run("counts unique tags against limit after deduplication", func(t *testing.T) {
		// 12 tags but only 6 unique after deduplication
		tags := []string{"a", "A", "b", "B", "c", "C", "d", "D", "e", "E", "f", "F"}
		result, err := domain.ValidateAndNormalizeTags(tags)
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c", "d", "e", "f"}, result)
	})

	t.Run("preserves order of first occurrence", func(t *testing.T) {
		result, err := domain.ValidateAndNormalizeTags([]string{"fiction", "book", "ebook"})
		require.NoError(t, err)
		assert.Equal(t, []string{"fiction", "book", "ebook"}, result)
	})
}
