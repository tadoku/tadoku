package e2e_test

import (
	"net/http"
	"testing"
)

func TestFindOngoingContestRegistrations(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"empty"}, want: http.StatusOK},
		{description: []string{"end", "day", "included", "with", "stored", "ordering"}, want: http.StatusOK},
		{description: []string{"start", "day", "included"}, want: http.StatusOK},
		{description: []string{"ended", "and", "future", "excluded"}, want: http.StatusOK},
		{description: []string{"deleted", "private", "contest", "remains", "visible"}, want: http.StatusOK},
		{description: []string{"invalid", "stored", "activity"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"invalid", "signed", "subject"}, want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		name := APITestName("FindOngoingContestRegistrations", test.want, test.description...)

		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
