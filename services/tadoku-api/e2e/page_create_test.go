package e2e_test

import (
	"net/http"
	"testing"
)

func TestCreatePage(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin"}, want: http.StatusCreated},
		{description: []string{"missing", "html"}, want: http.StatusBadRequest},
		{description: []string{"null", "html"}, want: http.StatusBadRequest},
		{description: []string{"empty", "html"}, want: http.StatusBadRequest},
		{description: []string{"short", "slug"}, want: http.StatusBadRequest},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"duplicate", "id"}, want: http.StatusConflict},
		{description: []string{"duplicate", "slug"}, want: http.StatusConflict},
	}

	for _, test := range tests {
		name := APITestName("CreatePage", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want, implementation{name: "tadoku-api", handler: api.handler})
		})
	}
}
