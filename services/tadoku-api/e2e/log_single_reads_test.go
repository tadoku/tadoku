package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionLogFindByID(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"quoted", "tags"}, want: http.StatusOK},
		{description: []string{"minimum", "duration"}, want: http.StatusOK},
		{description: []string{"score", "provenance"}, want: http.StatusOK},
		{description: []string{"guest", "populated"}, want: http.StatusOK},
		{description: []string{"user", "populated"}, want: http.StatusOK},
		{description: []string{"user2", "populated"}, want: http.StatusOK},
		{description: []string{"admin", "populated"}, want: http.StatusOK},
		{description: []string{"unknown"}, want: http.StatusNotFound},
		{description: []string{"nil", "id"}, want: http.StatusNotFound},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "signed", "subject"}, want: http.StatusUnauthorized},
		{description: []string{"invalid", "activity"}, want: http.StatusBadRequest},
		{description: []string{"missing", "unit"}, want: http.StatusOK},
		{description: []string{"missing", "language"}, want: http.StatusNotFound},
		{description: []string{"missing", "local", "user"}, want: http.StatusNotFound},
		{description: []string{"deleted", "local", "user"}, want: http.StatusOK},
		{description: []string{"duration", "only"}, want: http.StatusOK},
		{description: []string{"amount", "only"}, want: http.StatusOK},
		{description: []string{"unattached"}, want: http.StatusOK},
		{description: []string{"deleted", "organizer"}, want: http.StatusOK},
		{description: []string{"deleted", "registration"}, want: http.StatusOK},
		{description: []string{"deleted", "contest"}, want: http.StatusOK},
		{description: []string{"deleted", "guest"}, want: http.StatusNotFound},
		{description: []string{"deleted", "user"}, want: http.StatusNotFound},
		{description: []string{"deleted", "user2"}, want: http.StatusNotFound},
		{description: []string{"deleted", "admin"}, want: http.StatusOK},
	}
	for _, test := range tests {
		name := APITestName("ImmersionLogFindByID", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
