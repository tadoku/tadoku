package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionProfileYearlyActivityByUserID(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"guest", "populated"}, want: http.StatusOK},
		{description: []string{"user", "populated"}, want: http.StatusOK},
		{description: []string{"user2", "populated"}, want: http.StatusOK},
		{description: []string{"admin", "populated"}, want: http.StatusOK},
		{description: []string{"unknown", "user"}, want: http.StatusOK},
		{description: []string{"nil", "user"}, want: http.StatusOK},
		{description: []string{"invalid", "user"}, want: http.StatusBadRequest, skipParity: "native responses omit parser details and reflected parameter input"},
		{description: []string{"empty"}, want: http.StatusOK},
		{description: []string{"empty", "year"}, want: http.StatusOK},
		{description: []string{"previous", "year"}, want: http.StatusOK},
		{description: []string{"next", "year"}, want: http.StatusOK},
		{description: []string{"zero", "year"}, want: http.StatusOK},
		{description: []string{"negative", "year"}, want: http.StatusOK},
		{description: []string{"wrapped", "year"}, want: http.StatusOK},
		{description: []string{"invalid", "year"}, want: http.StatusBadRequest, skipParity: "native responses omit parser details and reflected parameter input"},
		{description: []string{"overflow", "year"}, want: http.StatusBadRequest, skipParity: "native responses omit parser details and reflected parameter input"},
		{description: []string{"all", "deleted"}, want: http.StatusOK},
	}
	for _, test := range tests {
		name := APITestName("ImmersionProfileYearlyActivityByUserID", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler, skip: test.skipParity})
		})
	}
}
