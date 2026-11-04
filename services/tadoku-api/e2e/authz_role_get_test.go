package e2e_test

import (
	"net/http"
	"testing"
)

func TestAuthzRoleGet(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		canonical   bool
	}{
		{description: []string{"guest"}, want: http.StatusOK},
		{description: []string{"user"}, want: http.StatusOK},
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"banned"}, want: http.StatusOK},
		{description: []string{"admin", "banned"}, want: http.StatusOK},
		{description: []string{"test", "production", "admin", "not", "inherited"}, want: http.StatusOK, canonical: true},
		{description: []string{"production", "branch", "admin"}, want: http.StatusOK, canonical: true},
		{description: []string{"test", "branch", "admin"}, want: http.StatusOK},
		{description: []string{"test", "production", "ban", "not", "inherited"}, want: http.StatusOK, canonical: true},
		{description: []string{"test", "local", "banned"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("AuthzRoleGet", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			impl := implementation{name: "tadoku-api", handler: api.handler}
			if test.canonical {
				runCase(t, api, name, test.want, impl)
				return
			}

			runScenarioCase(t, name, test.want, impl, alphaAsOther)
		})
	}
}
