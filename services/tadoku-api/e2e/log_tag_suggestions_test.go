package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionLogTagSuggestions(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"empty_history"}, want: http.StatusOK},
		{description: []string{"empty_query"}, want: http.StatusOK},
		{description: []string{"omitted_query"}, want: http.StatusOK},
		{description: []string{"matching_case"}, want: http.StatusOK},
		{description: []string{"duplicate_youtube_default"}, want: http.StatusOK},
		{description: []string{"default_only_youtube"}, want: http.StatusOK},
		{description: []string{"nonmatching"}, want: http.StatusOK},
		{description: []string{"frequency_and_ties"}, want: http.StatusOK},
		{description: []string{"underscore_wildcard"}, want: http.StatusOK},
		{description: []string{"wildcard_with_text"}, want: http.StatusOK},
		{description: []string{"escaped_percent"}, want: http.StatusOK},
		{description: []string{"space_not_trimmed"}, want: http.StatusOK},
		{description: []string{"other_user"}, want: http.StatusOK},
		{description: []string{"limit_with_default"}, want: http.StatusOK},
		{description: []string{"limit_without_default"}, want: http.StatusOK},
		{description: []string{"default_fill_to_limit"}, want: http.StatusOK},
		{description: []string{"invalid_signed_subject"}, want: http.StatusUnauthorized},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
	}
	for _, test := range tests {
		name := APITestName("ImmersionLogTagSuggestions", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler},
			)
		})
	}
}
