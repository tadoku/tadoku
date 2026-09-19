package e2e_test

import (
	"net/http"
	"testing"
)

func TestListPages(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"admin", "default", "includes", "drafts"}, want: http.StatusOK},
		{description: []string{"without", "drafts", "includes", "scheduled"}, want: http.StatusOK},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"non", "admin", "without", "drafts"}, want: http.StatusForbidden},
		{description: []string{"empty", "namespace"}, want: http.StatusOK},
		{description: []string{"negative", "page"}, want: http.StatusBadRequest, skipParity: "legacy returns 500; a malformed query parameter is a client error"},
		{description: []string{"offset", "overflow"}, want: http.StatusOK, skipParity: "legacy overflows pagination offsets"},
		{description: []string{"page", "beyond", "total", "size"}, want: http.StatusOK},
		{description: []string{"tied", "timestamps", "first", "page"}, want: http.StatusOK, skipParity: "intentional stable-ordering difference"},
		{description: []string{"tied", "timestamps", "final", "page"}, want: http.StatusOK, skipParity: "intentional stable-ordering difference"},
	}

	for _, test := range tests {
		name := APITestName("ListPages", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler, skip: test.skipParity},
			)
		})
	}
}
