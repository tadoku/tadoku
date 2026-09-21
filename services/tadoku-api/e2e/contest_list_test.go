package e2e_test

import (
	"net/http"
	"testing"
)

func TestListContests(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"guest", "defaults", "hide", "private", "but", "count", "it"}, want: http.StatusOK},
		{description: []string{"guest", "owner", "filter", "shows", "private", "unofficial"}, want: http.StatusOK},
		{description: []string{"guest", "includes", "deleted"}, want: http.StatusOK},
		{description: []string{"admin", "includes", "private", "and", "deleted"}, want: http.StatusOK},
		{description: []string{"first", "page", "token", "counts", "hidden", "private"}, want: http.StatusOK},
		{
			description: []string{"invalid", "user", "id"},
			want:        http.StatusBadRequest,
			skipParity:  "native responses omit parser details and reflected parameter input",
		},
		{description: []string{"malformed", "official"}, want: http.StatusBadRequest, skipParity: "native responses omit parser details and reflected parameter input"},
		{description: []string{"malformed", "page"}, want: http.StatusBadRequest, skipParity: "native responses omit parser details and reflected parameter input"},
		{description: []string{"malformed", "page", "omits", "reflected", "input"}, want: http.StatusBadRequest, skipParity: "native responses omit parser details and reflected parameter input"},
		{description: []string{"empty", "owner", "filter"}, want: http.StatusOK},
		{description: []string{"page", "beyond", "visible", "rows"}, want: http.StatusOK},
		{description: []string{"negative", "page"}, want: http.StatusInternalServerError},
		{description: []string{"negative", "page", "size"}, want: http.StatusInternalServerError},
		{description: []string{"explicit", "zero", "page", "size"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("ListContests", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler, skip: test.skipParity},
			)
		})
	}
}
