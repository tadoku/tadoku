package e2e_test

import (
	"net/http"
	"path/filepath"
	"testing"
)

func TestUpdateAnnouncement(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"replace", "fields"}, want: http.StatusOK},
		{description: []string{"null", "href"}, want: http.StatusOK},
		{description: []string{"omitted", "href"}, want: http.StatusOK},
		{description: []string{"empty", "href"}, want: http.StatusOK},
		{description: []string{"whitespace", "fields"}, want: http.StatusOK},
		{description: []string{"trailing", "json"}, want: http.StatusOK},
		{description: []string{"xml"}, want: http.StatusOK},
		{description: []string{"form"}, want: http.StatusBadRequest},
		{description: []string{"guest", "form"}, want: http.StatusUnauthorized},
		{description: []string{"whitespace", "body", "guest"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "xml", "guest"}, want: http.StatusBadRequest},
		{description: []string{"missing", "title"}, want: http.StatusBadRequest},
		{description: []string{"missing", "content"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "style"}, want: http.StatusBadRequest},
		{description: []string{"missing", "starts", "at"}, want: http.StatusBadRequest},
		{description: []string{"missing", "ends", "at"}, want: http.StatusBadRequest},
		{description: []string{"equal", "dates"}, want: http.StatusBadRequest},
		{description: []string{"reversed", "dates"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "date"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "href"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "body", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "id", "guest"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "json", "guest"}, want: http.StatusBadRequest},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
		{description: []string{"null", "body"}, want: http.StatusBadRequest},
		{description: []string{"unsupported", "content", "type"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"guest", "empty", "body"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"non", "admin", "invalid", "fields"}, want: http.StatusForbidden},
		{description: []string{"non", "admin", "empty", "body"}, want: http.StatusForbidden},
		{description: []string{"not", "found"}, want: http.StatusNotFound},
		{description: []string{"wrong", "namespace"}, want: http.StatusNotFound},
		{description: []string{"deleted"}, want: http.StatusNotFound},
		{description: []string{"write", "failure"}, want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		name := APITestName("UpdateAnnouncement", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler},
			)
		})
	}
}

func TestUpdateAnnouncementReadback(t *testing.T) {
	for _, test := range []struct {
		description []string
		want        int
	}{
		{description: []string{"replace", "fields"}, want: http.StatusOK},
		{description: []string{"null", "href"}, want: http.StatusOK},
		{description: []string{"omitted", "href"}, want: http.StatusOK},
		{description: []string{"empty", "href"}, want: http.StatusOK},
		{description: []string{"write", "failure"}, want: http.StatusInternalServerError},
	} {
		name := APITestName("UpdateAnnouncement", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			for _, impl := range []implementation{
				{name: "tadoku-api", handler: api.handler},
				{name: "content-api", handler: legacyContent.handler},
			} {
				t.Run(impl.name, func(t *testing.T) {
					dir := filepath.Join("testdata", name)
					api.reset(t, dir)
					atFixtureInstant(func() {
						checkHTTPGolden(t, impl.handler, dir, test.want)
						checkHTTPGolden(t, impl.handler, filepath.Join(dir, "readback"), http.StatusOK)
					})
					if api.proxied.Load() != 0 {
						t.Error("handler contacted an upstream")
					}
				})
			}
		})
	}
}
