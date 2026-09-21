package e2e_test

import (
	"net/http"
	"testing"
)

func TestProfileUsersList(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"pagination"}, want: http.StatusOK},
		{description: []string{"search"}, want: http.StatusOK},
		{description: []string{"empty", "search"}, want: http.StatusOK},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("ProfileUsersList", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "profile-api", handler: legacyProfile.handler},
			)
		})
	}
}
