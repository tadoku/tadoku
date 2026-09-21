package e2e_test

import (
	"net/http"
	"testing"
)

func TestAuthzPermissionCheck(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"missing", "relation"}, want: http.StatusBadRequest},
		{
			description: []string{"guest", "empty", "body"},
			want:        http.StatusBadRequest,
		},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"not", "allowlisted"}, want: http.StatusForbidden},
		{
			description: []string{"banned", "malformed", "json"},
			want:        http.StatusForbidden,
		},
	}

	for _, test := range tests {
		name := APITestName("AuthzPermissionCheck", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
