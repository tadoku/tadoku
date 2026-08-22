package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func TestAllRequestedLanguagesExist(t *testing.T) {
	tests := []struct {
		name      string
		codes     []string
		languages []postgres.Language
		want      bool
	}{
		{
			name: "empty request",
			want: true,
		},
		{
			name:  "all requested languages are present",
			codes: []string{"jpn", "kor"},
			languages: []postgres.Language{
				{Code: "kor"},
				{Code: "jpn"},
			},
			want: true,
		},
		{
			name:  "requested language is missing",
			codes: []string{"jpn", "kor"},
			languages: []postgres.Language{
				{Code: "jpn"},
			},
			want: false,
		},
		{
			name:  "duplicate requested languages are present",
			codes: []string{"jpn", "jpn", "kor"},
			languages: []postgres.Language{
				{Code: "jpn"},
				{Code: "kor"},
			},
			want: true,
		},
		{
			name:  "duplicate request still detects another missing language",
			codes: []string{"jpn", "jpn", "kor"},
			languages: []postgres.Language{
				{Code: "jpn"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, allRequestedLanguagesExist(tt.codes, tt.languages))
		})
	}
}
