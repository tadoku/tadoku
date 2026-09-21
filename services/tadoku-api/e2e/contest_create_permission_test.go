package e2e_test

import (
	"net/http"
	"testing"
)

func TestContestCreatePermissionCheck(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"old", "account", "under", "limit"}, want: http.StatusOK},
		{description: []string{"exact", "one", "month", "account", "age"}, want: http.StatusOK},
		{description: []string{"young", "account"}, want: http.StatusInternalServerError},
		{description: []string{"yearly", "quota", "reached"}, want: http.StatusForbidden},
		{description: []string{"admin", "bypasses", "age", "and", "quota"}, want: http.StatusOK},
		{description: []string{"missing", "identity"}, want: http.StatusNotFound},
		{description: []string{"guest", "subject"}, want: http.StatusInternalServerError},
		{description: []string{"banned"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("ContestCreatePermissionCheck", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler},
			)
		})
	}
}
