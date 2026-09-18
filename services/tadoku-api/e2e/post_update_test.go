package e2e_test

import (
	"net/http"
	"testing"
)

func TestUpdatePost(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"replace", "fields"}, want: http.StatusOK},
		{description: []string{"title", "only"}, want: http.StatusOK},
		{description: []string{"content", "only"}, want: http.StatusOK},
		{description: []string{"metadata", "only"}, want: http.StatusOK},
		{description: []string{"null", "published", "at"}, want: http.StatusOK},
		{description: []string{"omitted", "published", "at"}, want: http.StatusOK},
		{description: []string{"offset", "date"}, want: http.StatusOK, skipParity: "legacy preserves the submitted timestamp offset instead of normalizing to UTC"},
		{description: []string{"schedule", "publication"}, want: http.StatusOK},
		{description: []string{"update", "draft"}, want: http.StatusOK},
		{description: []string{"whitespace", "fields"}, want: http.StatusOK},
		{description: []string{"trailing", "json"}, want: http.StatusOK},
		{description: []string{"missing", "slug"}, want: http.StatusBadRequest},
		{description: []string{"missing", "title"}, want: http.StatusBadRequest},
		{description: []string{"missing", "content"}, want: http.StatusBadRequest},
		{description: []string{"single", "rune", "slug"}, want: http.StatusBadRequest},
		{description: []string{"uppercase", "slug"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "date"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "body", "id"}, want: http.StatusBadRequest},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
		{description: []string{"null", "body"}, want: http.StatusBadRequest},
		{description: []string{"guest", "empty", "body"}, want: http.StatusBadRequest, skipParity: "intentional JSON-only decoding difference"},
		{description: []string{"non", "admin", "empty", "body"}, want: http.StatusBadRequest, skipParity: "intentional JSON-only decoding difference"},
		{description: []string{"malformed", "json", "guest"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "id", "guest"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"non", "admin", "invalid", "fields"}, want: http.StatusForbidden},
		{description: []string{"not", "found"}, want: http.StatusNotFound},
		{description: []string{"zero", "id"}, want: http.StatusNotFound},
		{description: []string{"wrong", "namespace"}, want: http.StatusNotFound},
		{description: []string{"deleted"}, want: http.StatusNotFound},
		{description: []string{"duplicate", "slug", "content"}, want: http.StatusConflict, skipParity: "intentional conflict response instead of legacy bad request"},
		{description: []string{"duplicate", "slug", "metadata"}, want: http.StatusConflict, skipParity: "intentional conflict response instead of legacy bad request"},
		{description: []string{"write", "failure"}, want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		name := APITestName("UpdatePost", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler, skip: test.skipParity},
			)
		})
	}
}
