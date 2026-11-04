package e2e_test

import (
	"net/http"
	"testing"
)

func TestUpdateLanguage(t *testing.T) {
	resetAfter(t)

	tests := []struct {
		description []string
		want        int
		canonical   bool
	}{
		{description: []string{"admin"}, want: http.StatusOK, canonical: true},
		{description: []string{"maximum", "byte", "length"}, want: http.StatusOK, canonical: true},
		{description: []string{"whitespace", "retained"}, want: http.StatusOK, canonical: true},
		{description: []string{"empty", "name"}, want: http.StatusBadRequest, canonical: true},
		{description: []string{"name", "too", "long"}, want: http.StatusBadRequest, canonical: true},
		{description: []string{"unicode", "name", "bytes"}, want: http.StatusBadRequest, canonical: true},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
		{description: []string{"empty", "guest"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"test", "tenant", "admin"}, want: http.StatusForbidden, canonical: true},
		{description: []string{"test", "tenant", "production", "admin", "not", "inherited"}, want: http.StatusForbidden, canonical: true},
		{description: []string{"missing", "code"}, want: http.StatusNotFound, canonical: true},
		{description: []string{"case", "sensitive", "code"}, want: http.StatusNotFound, canonical: true},
	}

	for _, test := range tests {
		name := APITestName("UpdateLanguage", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			impl := implementation{name: "tadoku-api", handler: api.handler}
			if test.canonical {
				runCase(t, api, name, test.want, impl)
				return
			}

			runScenarioCase(t, name, test.want, impl)
		})
	}
}
