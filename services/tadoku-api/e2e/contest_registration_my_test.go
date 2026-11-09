package e2e_test

import (
	"net/http"
	"testing"
)

func TestListMyContestRegistrations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"empty"}, want: http.StatusOK},
		{description: []string{"mixed", "history", "across", "years"}, want: http.StatusOK},
		{description: []string{"deleted", "contests", "and", "registrations"}, want: http.StatusOK},
		{description: []string{"live", "deleted", "and", "missing", "organizers"}, want: http.StatusOK},
		{description: []string{"stored", "hydration"}, want: http.StatusOK},
		{description: []string{"empty", "stored", "arrays"}, want: http.StatusOK},
		{description: []string{"invalid", "stored", "activity"}, want: http.StatusBadRequest},
		{description: []string{"first", "page", "has", "next", "page", "token"}, want: http.StatusOK},
		{description: []string{"last", "page", "omits", "next", "page", "token"}, want: http.StatusOK},
		{description: []string{"page", "beyond", "registrations"}, want: http.StatusOK},
		{description: []string{"default", "page", "size", "caps", "at", "50"}, want: http.StatusOK},
		{description: []string{"page", "size", "above", "maximum", "capped"}, want: http.StatusOK},
		{description: []string{"negative", "page", "size"}, want: http.StatusBadRequest},
		{description: []string{"negative", "page"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
	}

	for _, test := range tests {
		name := APITestName("ListMyContestRegistrations", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runScenarioCase(t, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
