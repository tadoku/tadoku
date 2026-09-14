package e2e_test

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestListActiveAnnouncements(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{
			description: []string{"without", "auth"},
			want:        http.StatusOK,
		},
		{
			description: []string{"escaped", "slash", "namespace"},
			want:        http.StatusOK,
		},
		{
			description: []string{"escaped", "space", "namespace"},
			want:        http.StatusOK,
		},
		{
			description: []string{"unicode", "namespace"},
			want:        http.StatusOK,
		},
		{
			description: []string{"without", "announcements"},
			want:        http.StatusOK,
		},
		{
			description: []string{"active", "undeleted", "newest", "ten"},
			want:        http.StatusOK,
		},
	}

	for _, test := range tests {
		name := APITestName("ListActiveAnnouncements", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("testdata", name)
			for _, implementation := range []struct {
				name    string
				handler http.Handler
			}{
				{name: "tadoku-api", handler: api.handler},
				{name: "content-api", handler: legacyContent.handler},
			} {
				t.Run(implementation.name, func(t *testing.T) {
					resetCase(t, path)
					timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
						checkHTTPGolden(t, implementation.handler, path, test.want)
					})

					if api.proxied.Load() != 0 {
						t.Error("native read contacted an upstream")
					}
				})
			}
		})
	}
}
