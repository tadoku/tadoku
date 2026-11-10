package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionContestProfileFetchSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"public", "guest"}, want: http.StatusOK},
		{description: []string{"public", "user"}, want: http.StatusOK},
		{description: []string{"public", "user2"}, want: http.StatusOK},
		{description: []string{"public", "admin"}, want: http.StatusOK},
		{description: []string{"private", "guest"}, want: http.StatusOK},
		{description: []string{"private", "user"}, want: http.StatusOK},
		{description: []string{"private", "user2"}, want: http.StatusOK},
		{description: []string{"private", "admin"}, want: http.StatusOK},
		{description: []string{"deleted", "guest"}, want: http.StatusOK},
		{description: []string{"deleted", "user"}, want: http.StatusOK},
		{description: []string{"deleted", "user2"}, want: http.StatusOK},
		{description: []string{"deleted", "admin"}, want: http.StatusOK},
		{description: []string{"missing", "registration"}, want: http.StatusOK},
		{description: []string{"deleted", "registration"}, want: http.StatusOK},
		{description: []string{"missing", "contest"}, want: http.StatusOK},
		{description: []string{"missing", "local", "user"}, want: http.StatusOK},
		{description: []string{"unknown", "user"}, want: http.StatusOK},
		{description: []string{"unknown", "contest"}, want: http.StatusOK},
		{description: []string{"empty", "logs"}, want: http.StatusOK},
		{description: []string{"deleted", "logs"}, want: http.StatusOK},
		{description: []string{"deleted", "local", "user"}, want: http.StatusOK},
		{description: []string{"missing", "owner"}, want: http.StatusOK},
		{description: []string{"filter", "language"}, want: http.StatusOK},
		{description: []string{"filter", "activity"}, want: http.StatusOK},
		{description: []string{"filter", "tag"}, want: http.StatusOK},
		{description: []string{"filter", "comma", "tag"}, want: http.StatusOK},
		{description: []string{"combined", "filters"}, want: http.StatusOK},
		{description: []string{"filters", "with", "several", "values"}, want: http.StatusOK},
		{description: []string{"filter", "no", "match"}, want: http.StatusOK},
		{description: []string{"malformed", "activity", "filter"}, want: http.StatusBadRequest},
		{description: []string{"long", "tag"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "user"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "contest"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionContestProfileFetchSummary", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runScenarioCase(t, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
