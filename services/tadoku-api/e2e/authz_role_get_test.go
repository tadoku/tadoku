package e2e_test

import (
	"net/http"
	"testing"
)

func TestAuthzRoleGet(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest"}, want: http.StatusOK},
		{description: []string{"user"}, want: http.StatusOK},
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"banned"}, want: http.StatusOK},
		{description: []string{"admin", "banned"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("AuthzRoleGet", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
