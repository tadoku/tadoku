package e2e_test

import (
	"net/http"
	"testing"
)

func TestProfileAccountDeletionCreate(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"normalized", "email"}, want: http.StatusAccepted},
		{description: []string{"confirmation", "mismatch"}, want: http.StatusBadRequest},
		{description: []string{"missing", "email"}, want: http.StatusBadRequest},
		{description: []string{"missing", "reason"}, want: http.StatusBadRequest},
		{description: []string{"blank", "reason"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"unknown", "user"}, want: http.StatusNotFound},
		{description: []string{"deleted", "user"}, want: http.StatusNotFound},
		{description: []string{"target", "admin"}, want: http.StatusConflict},
		{description: []string{"deletion", "in", "progress"}, want: http.StatusConflict},
	}

	for _, test := range tests {
		name := APITestName("ProfileAccountDeletionCreate", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
