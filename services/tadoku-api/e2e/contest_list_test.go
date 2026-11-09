package e2e_test

import (
	"net/http"
	"testing"
)

func TestListContests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest", "defaults", "hide", "private", "from", "total"}, want: http.StatusOK},
		{description: []string{"guest", "owner", "filter", "hides", "private", "unofficial"}, want: http.StatusOK},
		{description: []string{"user", "owner", "filter", "hides", "other", "users", "private"}, want: http.StatusOK},
		{description: []string{"owner", "filter", "shows", "own", "private"}, want: http.StatusOK},
		{description: []string{"owner", "without", "filter", "shows", "own", "private"}, want: http.StatusOK},
		{description: []string{"guest", "include", "deleted", "is", "ignored"}, want: http.StatusOK},
		{description: []string{"user", "include", "deleted", "is", "ignored"}, want: http.StatusOK},
		{description: []string{"admin", "includes", "private", "and", "deleted"}, want: http.StatusOK},
		{description: []string{"admin", "owner", "filter", "shows", "other", "users", "private"}, want: http.StatusOK},
		{description: []string{"first", "page", "token", "excludes", "hidden", "private"}, want: http.StatusOK},
		{
			description: []string{"invalid", "user", "id"},
			want:        http.StatusBadRequest,
		},
		{description: []string{"malformed", "official"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "page"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "page", "omits", "reflected", "input"}, want: http.StatusBadRequest},
		{description: []string{"empty", "owner", "filter"}, want: http.StatusOK},
		{description: []string{"page", "beyond", "visible", "rows"}, want: http.StatusOK},
		{description: []string{"negative", "page"}, want: http.StatusBadRequest},
		{description: []string{"negative", "page", "size"}, want: http.StatusBadRequest},
		{description: []string{"offset", "overflow"}, want: http.StatusOK},
		{description: []string{"maximum", "page"}, want: http.StatusOK},
		{description: []string{"explicit", "zero", "page", "size"}, want: http.StatusOK},
		{description: []string{"upcoming", "soonest", "first"}, want: http.StatusOK},
		{description: []string{"upcoming", "second", "page"}, want: http.StatusOK},
		{description: []string{"upcoming", "none"}, want: http.StatusOK},
		{description: []string{"upcoming", "false", "keeps", "created", "order"}, want: http.StatusOK},
		{description: []string{"malformed", "upcoming"}, want: http.StatusBadRequest},
	}

	for _, test := range tests {
		name := APITestName("ListContests", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runScenarioCase(t, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
