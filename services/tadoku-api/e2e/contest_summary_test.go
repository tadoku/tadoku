package e2e_test

import (
	"net/http"
	"testing"
)

func TestFetchContestSummary(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"populated", "counts", "distinct", "and", "scores"}, want: http.StatusOK},
		{description: []string{"empty"}, want: http.StatusOK},
		{description: []string{"missing"}, want: http.StatusOK},
		{description: []string{"guest", "reads", "private"}, want: http.StatusOK},
		{description: []string{"guest", "reads", "deleted"}, want: http.StatusOK},
		{
			description: []string{"invalid", "id"},
			want:        http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		name := APITestName("FetchContestSummary", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
