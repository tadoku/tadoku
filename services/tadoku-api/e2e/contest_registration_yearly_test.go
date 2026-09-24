package e2e_test

import (
	"net/http"
	"testing"
)

func TestListYearlyContestRegistrations(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"owner", "mixed", "history"}, want: http.StatusOK},
		{description: []string{"admin", "mixed", "history"}, want: http.StatusOK},
		{description: []string{"other", "user", "mixed", "history"}, want: http.StatusOK},
		{description: []string{"guest", "mixed", "history"}, want: http.StatusOK},
		{description: []string{"invalid", "signed", "subject", "public", "history"}, want: http.StatusOK},
		{description: []string{"missing", "contest", "owner"}, want: http.StatusOK},
		{description: []string{"deleted", "user"}, want: http.StatusOK},
		{description: []string{"empty", "year"}, want: http.StatusOK},
		{description: []string{"unknown", "user"}, want: http.StatusOK},
		{description: []string{"nil", "user"}, want: http.StatusOK},
		{description: []string{"negative", "year"}, want: http.StatusOK},
		{description: []string{"zero", "year"}, want: http.StatusOK},
		{description: []string{"int32", "wrapped", "year"}, want: http.StatusOK},
		{description: []string{"start", "year", "boundaries"}, want: http.StatusOK},
		{description: []string{"deleted", "contests", "and", "registrations"}, want: http.StatusOK},
		{description: []string{"stored", "ordering", "and", "hydration"}, want: http.StatusOK},
		{description: []string{"empty", "stored", "arrays"}, want: http.StatusOK},
		{description: []string{"invalid", "stored", "activity"}, want: http.StatusBadRequest},
		{description: []string{"hidden", "invalid", "activity"}, want: http.StatusOK},
		{description: []string{"invalid", "user", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "year"}, want: http.StatusBadRequest},
		{description: []string{"overflowing", "year"}, want: http.StatusBadRequest},
	}

	for _, test := range tests {
		name := APITestName("ListYearlyContestRegistrations", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
