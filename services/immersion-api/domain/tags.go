package domain

import (
	"fmt"
	"strings"
)

const (
	MaxTagsPerLog = 10
	MaxTagLength  = 50
)

// ValidateAndNormalizeTags validates and normalizes a list of tags.
// It trims whitespace, converts to lowercase, deduplicates, and enforces limits.
func ValidateAndNormalizeTags(tags []string) ([]string, error) {
	if len(tags) == 0 {
		return []string{}, nil
	}

	seen := make(map[string]struct{})
	normalized := make([]string, 0, len(tags))

	for _, tag := range tags {
		// Trim whitespace and convert to lowercase
		tag = strings.ToLower(strings.TrimSpace(tag))

		// Skip empty tags
		if tag == "" {
			continue
		}

		// Check length
		if len(tag) > MaxTagLength {
			return nil, fmt.Errorf("tag %q exceeds maximum length of %d characters: %w", tag, MaxTagLength, ErrInvalidTags)
		}

		// Deduplicate (case-insensitive, already lowercased)
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}

		normalized = append(normalized, tag)
	}

	// Check count after deduplication
	if len(normalized) > MaxTagsPerLog {
		return nil, fmt.Errorf("too many tags: maximum is %d, got %d: %w", MaxTagsPerLog, len(normalized), ErrInvalidTags)
	}

	return normalized, nil
}
