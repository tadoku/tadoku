package e2e_test

import (
	"net/http"
	"testing"
)

func TestContestCreatePermissionCheck(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		serial      bool
	}{
		{description: []string{"old", "account", "under", "limit"}, want: http.StatusOK},
		{description: []string{"exact", "one", "month", "account", "age"}, want: http.StatusOK},
		{description: []string{"young", "account"}, want: http.StatusForbidden, serial: true},
		{description: []string{"yearly", "quota", "reached"}, want: http.StatusForbidden},
		{description: []string{"admin", "bypasses", "age", "and", "quota"}, want: http.StatusOK},
		{description: []string{"missing", "identity"}, want: http.StatusNotFound, serial: true},
		{description: []string{"guest", "subject"}, want: http.StatusUnauthorized},
	}

	for _, test := range tests {
		name := APITestName("ContestCreatePermissionCheck", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			impl := implementation{name: "tadoku-api", handler: api.handler}
			if test.serial {
				runCase(t, api, name, test.want, impl)
				return
			}

			runScenarioCase(t, name, test.want, impl)
		})
	}
}
