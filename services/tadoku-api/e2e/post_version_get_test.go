package e2e_test

import (
	"net/http"
	"testing"
)

func TestGetPostVersion(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"first", "version"}, want: http.StatusOK, skipParity: "legacy returns version 0; revisions now use their 1-based history number"},
		{description: []string{"tied", "timestamps"}, want: http.StatusOK, skipParity: "legacy returns version 0; revisions now use their stable 1-based history number"},
		{description: []string{"latest", "version"}, want: http.StatusOK, skipParity: "legacy returns version 0; revisions now use their 1-based history number"},
		{description: []string{"draft", "empty", "content"}, want: http.StatusOK, skipParity: "legacy returns version 0; revisions now use their 1-based history number"},
		{description: []string{"scheduled"}, want: http.StatusOK, skipParity: "legacy returns version 0; revisions now use their 1-based history number"},
		{description: []string{"invalid", "post", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "content", "id"}, want: http.StatusBadRequest, skipParity: "native responses omit parser details and reflected parameter input"},
		{description: []string{"missing", "post"}, want: http.StatusNotFound},
		{description: []string{"missing", "version"}, want: http.StatusNotFound},
		{description: []string{"wrong", "post"}, want: http.StatusNotFound},
		{description: []string{"wrong", "namespace"}, want: http.StatusNotFound, skipParity: "legacy ignores the namespace; revisions now require a matching namespace"},
		{description: []string{"deleted"}, want: http.StatusNotFound, skipParity: "legacy reads deleted posts; revisions now require an undeleted post"},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("GetPostVersion", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler, skip: test.skipParity},
			)
		})
	}
}
