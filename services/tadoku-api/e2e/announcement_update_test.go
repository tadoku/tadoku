package e2e_test

import (
	"net/http"
	"testing"
)

func TestUpdateAnnouncement(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"replace", "fields"}, want: http.StatusOK},
		{description: []string{"null", "href"}, want: http.StatusOK},
		{description: []string{"omitted", "href"}, want: http.StatusOK},
		{description: []string{"empty", "href"}, want: http.StatusOK},
		{description: []string{"whitespace", "fields"}, want: http.StatusOK},
		{description: []string{"trailing", "json"}, want: http.StatusOK},
		// Legacy stores timestamp offsets as wall-clock fields instead of UTC instants.
		{description: []string{"offset", "dates"}, want: http.StatusOK, skipParity: "legacy stores timestamp offsets as wall-clock fields"},
		{description: []string{"whitespace", "body", "guest"}, want: http.StatusBadRequest},
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
		// The generated JSON decoder no longer receives an injected empty object.
		{description: []string{"guest", "empty", "body"}, want: http.StatusBadRequest, skipParity: "intentional JSON-only decoding difference"},
		{description: []string{"non", "admin", "empty", "body"}, want: http.StatusBadRequest, skipParity: "intentional JSON-only decoding difference"},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"non", "admin", "invalid", "fields"}, want: http.StatusForbidden},
		{description: []string{"not", "found"}, want: http.StatusNotFound},
		{description: []string{"wrong", "namespace"}, want: http.StatusNotFound},
		{description: []string{"deleted"}, want: http.StatusNotFound},
		{description: []string{"write", "failure"}, want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		name := APITestName("UpdateAnnouncement", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			legacy := implementation{name: "content-api", handler: legacyContent.handler, skip: test.skipParity}
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				legacy,
			)
		})
	}
}
