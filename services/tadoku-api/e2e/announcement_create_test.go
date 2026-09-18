package e2e_test

import (
	"net/http"
	"testing"
)

func TestCreateAnnouncement(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"admin"}, want: http.StatusCreated},
		{description: []string{"null", "href"}, want: http.StatusCreated},
		{description: []string{"omitted", "href"}, want: http.StatusCreated},
		{description: []string{"empty", "href"}, want: http.StatusCreated},
		{description: []string{"relative", "href"}, want: http.StatusCreated},
		{description: []string{"ignored", "server", "fields"}, want: http.StatusCreated},
		{description: []string{"trailing", "json"}, want: http.StatusCreated},
		// Legacy stores timestamp offsets as wall-clock fields instead of UTC instants.
		{description: []string{"offset", "dates"}, want: http.StatusCreated, skipParity: "legacy stores timestamp offsets as wall-clock fields"},
		{description: []string{"empty", "title"}, want: http.StatusBadRequest},
		{description: []string{"empty", "content"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "style"}, want: http.StatusBadRequest},
		{description: []string{"missing", "start"}, want: http.StatusBadRequest},
		{description: []string{"missing", "end"}, want: http.StatusBadRequest},
		{description: []string{"equal", "dates"}, want: http.StatusBadRequest},
		{description: []string{"reversed", "dates"}, want: http.StatusBadRequest},
		{description: []string{"zero", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "time"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "href"}, want: http.StatusBadRequest},
		{description: []string{"javascript", "href"}, want: http.StatusBadRequest, skipParity: "intentional native href validation difference"},
		{description: []string{"data", "href"}, want: http.StatusBadRequest, skipParity: "intentional native href validation difference"},
		{description: []string{"href", "too", "long"}, want: http.StatusBadRequest, skipParity: "intentional native href validation difference"},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "guest"}, want: http.StatusBadRequest},
		{description: []string{"whitespace", "guest"}, want: http.StatusBadRequest},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
		{description: []string{"null", "body"}, want: http.StatusBadRequest},
		// Empty JSON input is no longer rewritten to an empty object for the
		// legacy binder; reject it before application-level authorization.
		{description: []string{"empty", "guest"}, want: http.StatusBadRequest, skipParity: "intentional JSON-only decoding difference"},
		{description: []string{"empty", "non", "admin"}, want: http.StatusBadRequest, skipParity: "intentional JSON-only decoding difference"},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"null", "guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"invalid", "non", "admin"}, want: http.StatusForbidden},
		// legacy returns 500 for a unique violation.
		{description: []string{"duplicate", "id"}, want: http.StatusConflict, skipParity: "native returns 409 for a unique violation; legacy returns 500"},
	}

	for _, test := range tests {
		name := APITestName("CreateAnnouncement", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler, skip: test.skipParity},
			)
		})
	}
}
