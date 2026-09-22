package e2e_test

import (
	"net/http"
	"testing"
)

func TestCreatePost(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin"}, want: http.StatusCreated},
		{description: []string{"draft"}, want: http.StatusCreated},
		{description: []string{"null", "publication"}, want: http.StatusCreated},
		{description: []string{"scheduled"}, want: http.StatusCreated},
		{description: []string{"unicode", "slug"}, want: http.StatusCreated},
		{description: []string{"whitespace", "fields"}, want: http.StatusCreated},
		{description: []string{"namespace", "isolation"}, want: http.StatusCreated},
		{description: []string{"ignored", "server", "fields"}, want: http.StatusCreated},
		{description: []string{"trailing", "json"}, want: http.StatusCreated},
		{description: []string{"offset", "date"}, want: http.StatusCreated},
		{description: []string{"empty", "slug"}, want: http.StatusBadRequest},
		{description: []string{"short", "slug"}, want: http.StatusBadRequest},
		{description: []string{"uppercase", "slug"}, want: http.StatusBadRequest},
		{description: []string{"one", "rune", "slug"}, want: http.StatusBadRequest},
		{description: []string{"empty", "title"}, want: http.StatusBadRequest},
		{description: []string{"empty", "content"}, want: http.StatusBadRequest},
		{description: []string{"zero", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "time"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
		{description: []string{"null", "body"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"duplicate", "id"}, want: http.StatusConflict},
		{description: []string{"duplicate", "slug"}, want: http.StatusConflict},
	}

	for _, test := range tests {
		name := APITestName("CreatePost", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want, implementation{name: "tadoku-api", handler: api.handler})
		})
	}
}
