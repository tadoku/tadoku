package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionLogExport(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"csv", "populated"}, want: http.StatusOK},
		{description: []string{"json", "populated"}, want: http.StatusOK},
		{description: []string{"csv", "empty"}, want: http.StatusOK},
		{description: []string{"json", "empty"}, want: http.StatusOK},
		{description: []string{"csv", "beyond", "one", "batch"}, want: http.StatusOK},
		{description: []string{"ignores", "user", "id", "query"}, want: http.StatusOK},
		{description: []string{"admin", "own", "logs"}, want: http.StatusOK},
		{description: []string{"invalid", "format"}, want: http.StatusBadRequest},
		{description: []string{"missing", "format"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
	}
	for _, test := range tests {
		name := APITestName("ImmersionLogExport", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
