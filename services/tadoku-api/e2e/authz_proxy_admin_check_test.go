package e2e_test

import (
	"net/http"
	"testing"
)

func TestAuthzProxyAdminCheck(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"admin", "banned"}, want: http.StatusOK},
		{description: []string{"missing", "subject"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "subject"}, want: http.StatusBadRequest},
		{description: []string{"missing", "credential", "malformed", "json"}, want: http.StatusUnauthorized},
		{description: []string{"wrong", "credential"}, want: http.StatusUnauthorized},
		{description: []string{"user", "jwt"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("AuthzProxyProxyAdminCheck", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
