package e2e_test

import (
	"net/http"
	"testing"
)

func TestBannedUsers(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest"}, want: http.StatusOK},
		{description: []string{"without", "tuple"}, want: http.StatusOK},
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"banned"}, want: http.StatusForbidden},
		{description: []string{"admin", "and", "banned"}, want: http.StatusForbidden},
		{description: []string{"other", "user", "banned"}, want: http.StatusOK},
		{description: []string{"test", "guest"}, want: http.StatusForbidden},
		{description: []string{"test", "without", "access"}, want: http.StatusForbidden},
		{description: []string{"test", "tester"}, want: http.StatusOK},
		{description: []string{"test", "production", "admin"}, want: http.StatusOK},
		{description: []string{"test", "production", "banned"}, want: http.StatusForbidden},
		{description: []string{"test", "local", "banned"}, want: http.StatusForbidden},
		{description: []string{"production", "locally", "banned"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("BannedUsers", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
