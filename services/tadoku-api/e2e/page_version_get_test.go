package e2e_test

import (
	"net/http"
	"testing"
)

func TestGetPageVersion(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"first", "version"}, want: http.StatusOK, skipParity: "legacy returns version 0; revisions now use their 1-based history number"},
		{description: []string{"tied", "timestamps"}, want: http.StatusOK, skipParity: "legacy returns version 0; revisions now use their stable 1-based history number"},
		{description: []string{"latest", "version"}, want: http.StatusOK, skipParity: "legacy returns version 0; revisions now use their 1-based history number"},
		{description: []string{"draft", "empty", "html"}, want: http.StatusOK, skipParity: "legacy returns version 0; revisions now use their 1-based history number"},
		{description: []string{"scheduled"}, want: http.StatusOK, skipParity: "legacy returns version 0; revisions now use their 1-based history number"},
		{description: []string{"invalid", "page", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "content", "id"}, want: http.StatusBadRequest, skipParity: "native responses omit parser details and reflected parameter input"},
		{description: []string{"missing", "page"}, want: http.StatusNotFound},
		{description: []string{"missing", "version"}, want: http.StatusNotFound},
		{description: []string{"wrong", "page"}, want: http.StatusNotFound},
		{description: []string{"wrong", "namespace"}, want: http.StatusNotFound, skipParity: "legacy ignores namespace; revisions require a matching namespace"},
		{description: []string{"deleted"}, want: http.StatusNotFound, skipParity: "legacy reads deleted pages; revisions require an undeleted page"},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("GetPageVersion", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler, skip: test.skipParity},
			)
		})
	}
}
