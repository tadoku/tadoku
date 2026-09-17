package e2e_test

import (
	"net/http"
	"testing"
)

func TestFindAnnouncementByID(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"null", "href"}, want: http.StatusOK},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"not", "found"}, want: http.StatusNotFound},
		{description: []string{"wrong", "namespace"}, want: http.StatusNotFound},
		{description: []string{"deleted"}, want: http.StatusNotFound},
	}

	for _, test := range tests {
		name := APITestName("FindAnnouncementByID", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler},
			)
		})
	}
}

func TestListAnnouncements(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  bool
	}{
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"without", "credentials"}, want: http.StatusBadRequest},
		{description: []string{"without", "announcements"}, want: http.StatusOK},
		{description: []string{"default", "page", "and", "namespace"}, want: http.StatusOK},
		{description: []string{"second", "page"}, want: http.StatusOK},
		{description: []string{"page", "size", "capped"}, want: http.StatusOK},
		{description: []string{"invalid", "page", "size"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "page"}, want: http.StatusBadRequest},
		{description: []string{"negative", "page", "size"}, want: http.StatusBadRequest, skipParity: true},
		{description: []string{"negative", "page"}, want: http.StatusBadRequest, skipParity: true},
		{description: []string{"offset", "overflow"}, want: http.StatusOK, skipParity: true},
		{description: []string{"page", "beyond", "total", "size"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("ListAnnouncements", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			legacy := implementation{name: "content-api", handler: legacyContent.handler}
			if test.skipParity {
				legacy.skip = "legacy overflows pagination offsets"
				if test.want == http.StatusBadRequest {
					legacy.skip = "legacy returns 500; a malformed query parameter is a client error"
				}
			}
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				legacy,
			)
		})
	}
}

func TestListActiveAnnouncements(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest"}, want: http.StatusOK},
		{description: []string{"escaped", "slash", "namespace"}, want: http.StatusOK},
		{description: []string{"escaped", "space", "namespace"}, want: http.StatusOK},
		{description: []string{"unicode", "namespace"}, want: http.StatusOK},
		{description: []string{"without", "announcements"}, want: http.StatusOK},
		{description: []string{"active", "undeleted", "newest", "ten"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("ListActiveAnnouncements", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler},
			)
		})
	}
}
