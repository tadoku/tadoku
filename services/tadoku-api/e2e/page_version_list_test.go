package e2e_test

import (
	"net/http"
	"testing"
)

func TestListPageVersions(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin"}, want: http.StatusOK},
		{description: []string{"missing"}, want: http.StatusOK},
		{description: []string{"wrong", "namespace"}, want: http.StatusOK},
		{description: []string{"deleted"}, want: http.StatusOK},
		{description: []string{"tied", "timestamps"}, want: http.StatusOK},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("ListPageVersions", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want, implementation{name: "tadoku-api", handler: api.handler})
		})
	}
}
