package e2e_test

import (
	"net/http"
	"testing"
)

func TestGetPageVersion(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"first", "version"}, want: http.StatusOK},
		{description: []string{"tied", "timestamps"}, want: http.StatusOK},
		{description: []string{"latest", "version"}, want: http.StatusOK},
		{description: []string{"draft", "empty", "html"}, want: http.StatusOK},
		{description: []string{"scheduled"}, want: http.StatusOK},
		{description: []string{"invalid", "page", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "content", "id"}, want: http.StatusBadRequest},
		{description: []string{"missing", "page"}, want: http.StatusNotFound},
		{description: []string{"missing", "version"}, want: http.StatusNotFound},
		{description: []string{"wrong", "page"}, want: http.StatusNotFound},
		{description: []string{"wrong", "namespace"}, want: http.StatusNotFound},
		{description: []string{"deleted"}, want: http.StatusNotFound},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("GetPageVersion", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want, implementation{name: "tadoku-api", handler: api.handler})
		})
	}
}
