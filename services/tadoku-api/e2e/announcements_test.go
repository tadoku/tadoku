package e2e_test

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
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

func TestListAnnouncementsAuthorizationWiring(t *testing.T) {
	for _, test := range []struct {
		fixture string
		want    int
	}{
		{fixture: "204_admin", want: http.StatusOK},
		{fixture: "403_non_admin", want: http.StatusForbidden},
	} {
		t.Run(test.fixture, func(t *testing.T) {
			path := filepath.Join("testdata", "RequireAdmin", test.fixture)
			resetCase(t, path)
			request := listAuthorizationRequest(t, path)

			previous := jwt.TimeFunc
			jwt.TimeFunc = timex.Now
			defer func() { jwt.TimeFunc = previous }()
			timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
				response := httptest.NewRecorder()
				api.authenticatedHandler.ServeHTTP(response, request)
				if response.Code != test.want {
					t.Errorf("status=%d, want %d", response.Code, test.want)
				}
			})
		})
	}
}

func listAuthorizationRequest(t *testing.T, directory string) *http.Request {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(directory, "request.http"))
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(data)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = request.Body.Close() })
	request.URL.Path = "/content/announcements/main"
	request.URL.RawPath = ""
	request.RequestURI = request.URL.RequestURI()
	return request
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
