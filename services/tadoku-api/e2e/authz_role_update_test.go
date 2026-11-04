package e2e_test

import (
	"net/http"
	"testing"
)

func TestAuthzRoleUpdate(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		canonical   bool
	}{
		{description: []string{"ban", "user"}, want: http.StatusOK},
		{description: []string{"unban", "user"}, want: http.StatusOK},
		{description: []string{"whitespace", "reason"}, want: http.StatusOK},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "role"}, want: http.StatusBadRequest},
		{description: []string{"missing", "reason"}, want: http.StatusBadRequest},
		{
			description: []string{"guest", "empty", "body"},
			want:        http.StatusBadRequest,
		},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin", "invalid", "role"}, want: http.StatusForbidden},
		{description: []string{"missing", "user", "marked", "admin"}, want: http.StatusNotFound},
		{description: []string{"target", "admin"}, want: http.StatusForbidden},
		{
			description: []string{"test", "production", "admin", "not", "inherited"},
			want:        http.StatusForbidden,
			canonical:   true,
		},
	}

	for _, test := range tests {
		name := APITestName("AuthzRoleUpdate", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			impl := implementation{name: "tadoku-api", handler: api.handler}
			if test.canonical {
				runCase(t, api, name, test.want, impl)
				return
			}

			runScenarioCase(t, name, test.want, impl)
		})
	}
}
