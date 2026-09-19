package e2e_test

import (
	"net/http"
	"testing"
)

func TestCreateLanguage(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"maximum", "byte", "lengths"}, want: http.StatusOK},
		{description: []string{"whitespace", "retained"}, want: http.StatusOK},
		{description: []string{"empty", "code"}, want: http.StatusBadRequest},
		{description: []string{"empty", "name"}, want: http.StatusBadRequest},
		{description: []string{"code", "too", "long"}, want: http.StatusBadRequest},
		{description: []string{"name", "too", "long"}, want: http.StatusBadRequest},
		{description: []string{"unicode", "code", "bytes"}, want: http.StatusBadRequest},
		{description: []string{"unicode", "name", "bytes"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
		{description: []string{"empty", "guest"}, want: http.StatusBadRequest, skipParity: "strict JSON decoding rejects an empty body before legacy application authorization"},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"duplicate", "code"}, want: http.StatusConflict},
	}

	for _, test := range tests {
		name := APITestName("CreateLanguage", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler, skip: test.skipParity},
			)
		})
	}
}
