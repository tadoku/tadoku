package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionContestProfileFetchScores(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
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
		{description: []string{"missing", "registration"}, want: http.StatusNotFound},
		{description: []string{"deleted", "registration"}, want: http.StatusNotFound},
		{description: []string{"missing", "contest"}, want: http.StatusNotFound},
		{description: []string{"missing", "local", "user"}, want: http.StatusNotFound},
		{description: []string{"unknown", "user"}, want: http.StatusNotFound},
		{description: []string{"unknown", "contest"}, want: http.StatusNotFound},
		{description: []string{"empty", "logs"}, want: http.StatusOK},
		{description: []string{"deleted", "logs"}, want: http.StatusOK},
		{description: []string{"deleted", "local", "user"}, want: http.StatusOK},
		{description: []string{"missing", "owner"}, want: http.StatusOK},
		{description: []string{"language", "hydration"}, want: http.StatusOK},
		{description: []string{"empty", "registration", "languages"}, want: http.StatusOK},
		{description: []string{"description"}, want: http.StatusOK},
		{description: []string{"invalid", "user"}, want: http.StatusBadRequest, skipParity: "native responses omit parser details and reflected parameter input"},
		{description: []string{"invalid", "contest"}, want: http.StatusBadRequest, skipParity: "native responses omit parser details and reflected parameter input"},
	}
	for _, test := range tests {
		name := APITestName("ImmersionContestProfileFetchScores", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler, skip: test.skipParity})
		})
	}
}
