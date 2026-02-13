package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringArrayFromInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []string
	}{
		{
			name:     "nil returns empty slice",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "empty byte slice returns empty slice",
			input:    []byte{},
			expected: []string{},
		},
		{
			name:     "empty array literal returns empty slice",
			input:    []byte("{}"),
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []byte("{foo}"),
			expected: []string{"foo"},
		},
		{
			name:     "multiple elements",
			input:    []byte("{foo,bar,baz}"),
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "quoted element with comma",
			input:    []byte(`{"foo,bar",baz}`),
			expected: []string{"foo,bar", "baz"},
		},
		{
			name:     "string input works too",
			input:    "{foo,bar}",
			expected: []string{"foo", "bar"},
		},
		{
			name:     "[]string passthrough",
			input:    []string{"foo", "bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "nil []string returns empty slice",
			input:    ([]string)(nil),
			expected: []string{},
		},
		{
			name:     "[]any with strings",
			input:    []any{"foo", "bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "unknown type returns empty slice",
			input:    123,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringArrayFromInterface(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
