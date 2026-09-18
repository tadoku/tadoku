package e2e_test

import (
	"net/http"
	"testing"
)

func TestFindPageBySlug(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"guest"}, want: http.StatusOK},
		{description: []string{"publication", "boundary"}, want: http.StatusOK},
		{description: []string{"draft", "slug"}, want: http.StatusNotFound},
		{description: []string{"scheduled", "slug"}, want: http.StatusNotFound},
		{description: []string{"wrong", "namespace"}, want: http.StatusNotFound},
		{description: []string{"deleted"}, want: http.StatusNotFound},
		{description: []string{"missing", "content"}, want: http.StatusNotFound},
		{description: []string{"admin", "draft", "id"}, want: http.StatusOK},
		{description: []string{"guest", "id"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin", "id"}, want: http.StatusForbidden},
		{description: []string{"uuid", "slug", "priority"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("FindPageBySlug", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler, skip: test.skipParity},
			)
		})
	}
}
