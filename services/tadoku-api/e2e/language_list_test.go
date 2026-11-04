package e2e_test

import (
	"net/http"
	"testing"
)

func TestListLanguages(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		canonical   bool
	}{
		{description: []string{"admin", "ordered"}, want: http.StatusOK},
		{description: []string{"test", "tenant", "admin"}, want: http.StatusOK, canonical: true},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("ListLanguages", test.want, test.description...)
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
