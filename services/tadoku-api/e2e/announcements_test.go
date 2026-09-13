package e2e_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestListActiveAnnouncements(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
	}{
		{
			name:    "lists active announcements without authentication",
			fixture: "200_plain",
		},
		{
			name:    "decodes a slash in the namespace",
			fixture: "200_escaped_slash",
		},
		{
			name:    "decodes a space in the namespace",
			fixture: "200_escaped_space",
		},
		{
			name:    "accepts a Unicode namespace",
			fixture: "200_unicode",
		},
		{
			name:    "returns an empty array when no announcements exist",
			fixture: "200_empty",
		},
		{
			name:    "filters publication windows and deleted rows, returning the newest ten",
			fixture: "200_publication_window_and_limit",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join("testdata", "list_active_announcements", test.fixture)
			reset(t, filepath.Join(path, "setup.sql"))
			timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
				checkHTTPGolden(t, api.handler, path)
			})

			if api.proxied.Load() != 0 {
				t.Error("native read contacted an upstream")
			}
		})
	}
}
