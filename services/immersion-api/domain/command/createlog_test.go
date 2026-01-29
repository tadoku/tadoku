package command_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
)

func TestValidateAndNormalizeTags(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		expected    []string
		expectError bool
	}{
		{
			name:        "normalizes to lowercase",
			input:       []string{"Test", "UPPER"},
			expected:    []string{"test", "upper"},
			expectError: false,
		},
		{
			name:        "trims whitespace",
			input:       []string{"  spaced  ", "normal"},
			expected:    []string{"spaced", "normal"},
			expectError: false,
		},
		{
			name:        "removes empty strings",
			input:       []string{"", "valid", "  "},
			expected:    []string{"valid"},
			expectError: false,
		},
		{
			name:        "deduplicates case-insensitively",
			input:       []string{"Test", "test", "TEST"},
			expected:    []string{"test"},
			expectError: false,
		},
		{
			name:        "skips tags over 50 chars",
			input:       []string{strings.Repeat("a", 51)},
			expected:    []string{},
			expectError: false,
		},
		{
			name:        "allows exactly 50 chars",
			input:       []string{strings.Repeat("a", 50)},
			expected:    []string{strings.Repeat("a", 50)},
			expectError: false,
		},
		{
			name:        "allows exactly 10 tags",
			input:       []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			expected:    []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			expectError: false,
		},
		{
			name:        "errors on 11+ tags",
			input:       []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"},
			expected:    nil,
			expectError: true,
		},
		{
			name:        "handles nil input",
			input:       nil,
			expected:    []string{},
			expectError: false,
		},
		{
			name:        "handles empty slice input",
			input:       []string{},
			expected:    []string{},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := command.ValidateAndNormalizeTags(test.input)

			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}
