package e2e_test

import (
	"net/http"
	"testing"
)

func TestUpdateLanguage(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"maximum", "byte", "length"}, want: http.StatusOK},
		{description: []string{"whitespace", "retained"}, want: http.StatusOK},
		{description: []string{"empty", "name"}, want: http.StatusBadRequest},
		{description: []string{"name", "too", "long"}, want: http.StatusBadRequest},
		{description: []string{"unicode", "name", "bytes"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
		{description: []string{"empty", "guest"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"missing", "code"}, want: http.StatusNotFound},
		{description: []string{"case", "sensitive", "code"}, want: http.StatusNotFound},
	}

	for _, test := range tests {
		name := APITestName("UpdateLanguage", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
