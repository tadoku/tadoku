package e2e_test

import (
	"errors"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestListAnnouncements(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"without", "announcements"}, want: http.StatusOK},
		{description: []string{"default", "page", "and", "namespace"}, want: http.StatusOK},
		{description: []string{"second", "page"}, want: http.StatusOK},
		{description: []string{"page", "size", "capped"}, want: http.StatusOK},
		{description: []string{"invalid", "page", "size"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "page"}, want: http.StatusBadRequest},
		{description: []string{"negative", "page", "size"}, want: http.StatusInternalServerError},
		{description: []string{"negative", "page"}, want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		name := APITestName("ListAnnouncements", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("testdata", name)
			for _, implementation := range []struct {
				name    string
				handler http.Handler
			}{
				{name: "tadoku-api", handler: withContractAdminIdentity(api.handler)},
				{name: "content-api", handler: withContractAdminIdentity(legacyContent.handler)},
			} {
				t.Run(implementation.name, func(t *testing.T) {
					resetCase(t, path)
					checkHTTPGolden(t, implementation.handler, path, test.want)

					if api.proxied.Load() != 0 {
						t.Error("native read contacted an upstream")
					}
				})
			}
		})
	}
}

func TestListAnnouncementsAuthorization(t *testing.T) {
	for _, test := range []struct {
		description []string
		want        int
	}{
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
	} {
		name := APITestName("ListAnnouncements", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("testdata", name)
			resetCase(t, path)

			previous := jwt.TimeFunc
			jwt.TimeFunc = timex.Now
			defer func() { jwt.TimeFunc = previous }()
			timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
				checkHTTPGolden(t, api.authenticatedHandler, path, test.want)
			})
			if api.proxied.Load() != 0 {
				t.Error("authorization check fell back to the proxy")
			}
		})
	}
}

func TestListAnnouncementsRejectsFailedBanLookup(t *testing.T) {
	path := filepath.Join("testdata", APITestName("ListAnnouncements", http.StatusServiceUnavailable, "failed", "ban", "lookup"))
	resetCase(t, path)
	// Exercise the application permission check with the shared ban gate's
	// failure context. Provider behavior is covered by the middleware tests.
	handler := withContractAdminIdentity(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := permissions.WithBanLookupError(r.Context(), errors.New("ban lookup failed"))
		api.handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	checkHTTPGolden(t, handler, path, http.StatusServiceUnavailable)
	if api.proxied.Load() != 0 {
		t.Error("failed permission check fell back to the proxy")
	}
}

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
