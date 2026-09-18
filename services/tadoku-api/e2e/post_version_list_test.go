package e2e_test

import (
	"net/http"
	"testing"
)

func TestListPostVersions(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"missing"}, want: http.StatusOK},
		{description: []string{"wrong", "namespace"}, want: http.StatusOK, skipParity: "approved correction: legacy ignores namespace and exposes revision history"},
		{description: []string{"deleted"}, want: http.StatusOK, skipParity: "approved correction: legacy exposes deleted post revision history"},
		{description: []string{"tied", "timestamps"}, want: http.StatusOK, skipParity: "approved deterministic revision ordering by timestamp and ID"},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("ListPostVersions", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler, skip: test.skipParity},
			)
		})
	}
}
