package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionProfileListLogs(t *testing.T) {
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
		{description: []string{"unknown"}, want: http.StatusOK},
		{description: []string{"nil", "id"}, want: http.StatusOK},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "activity"}, want: http.StatusBadRequest},
		{description: []string{"missing", "unit"}, want: http.StatusOK},
		{description: []string{"missing", "language"}, want: http.StatusOK},
		{description: []string{"missing", "local", "user"}, want: http.StatusOK},
		{description: []string{"deleted", "local", "user"}, want: http.StatusOK},
		{description: []string{"include", "deleted", "guest"}, want: http.StatusForbidden},
		{description: []string{"include", "deleted", "user"}, want: http.StatusForbidden},
		{description: []string{"include", "deleted", "user2"}, want: http.StatusForbidden},
		{description: []string{"include", "deleted", "admin"}, want: http.StatusOK},
		{description: []string{"explicit", "false"}, want: http.StatusOK},
		{description: []string{"empty"}, want: http.StatusOK},
		{description: []string{"page", "one"}, want: http.StatusOK},
		{description: []string{"out", "of", "range"}, want: http.StatusOK},
		{description: []string{"last", "page"}, want: http.StatusOK},
		{description: []string{"negative", "page"}, want: http.StatusInternalServerError},
		{description: []string{"wrapped", "offset"}, want: http.StatusOK},
		{description: []string{"default", "limit"}, want: http.StatusOK},
		{description: []string{"zero", "limit"}, want: http.StatusOK},
		{description: []string{"negative", "limit"}, want: http.StatusOK},
		{description: []string{"capped", "limit"}, want: http.StatusOK},
		{description: []string{"second", "large", "page"}, want: http.StatusOK},
		{description: []string{"malformed", "page"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "page", "size"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "include", "deleted"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "page", "overflow"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionProfileListLogs", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
