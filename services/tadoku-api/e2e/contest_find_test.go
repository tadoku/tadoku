package e2e_test

import (
	"net/http"
	"testing"
)

func TestFindContestByID(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"guest", "can", "see", "private"}, want: http.StatusOK},
		{description: []string{"guest", "cannot", "see", "deleted"}, want: http.StatusNotFound},
		{description: []string{"admin", "sees", "deleted", "and", "deleted", "organizer"}, want: http.StatusOK},
		{description: []string{"missing"}, want: http.StatusNotFound},
		{description: []string{"invalid", "stored", "activity"}, want: http.StatusBadRequest},
		{
			description: []string{"invalid", "path", "id"},
			want:        http.StatusBadRequest,
			skipParity:  "native responses omit parser details and reflected parameter input",
		},
	}

	for _, test := range tests {
		name := APITestName("FindContestByID", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler, skip: test.skipParity},
			)
		})
	}
}
