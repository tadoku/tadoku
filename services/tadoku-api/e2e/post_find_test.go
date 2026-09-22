package e2e_test

import (
	"net/http"
	"testing"
)

func TestFindPostBySlug(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest"}, want: http.StatusOK},
		{description: []string{"admin", "slug"}, want: http.StatusOK},
		{description: []string{"publication", "boundary"}, want: http.StatusOK},
		{description: []string{"draft", "slug"}, want: http.StatusNotFound},
		{description: []string{"scheduled", "slug"}, want: http.StatusNotFound},
		{description: []string{"missing"}, want: http.StatusNotFound},
		{description: []string{"wrong", "namespace"}, want: http.StatusNotFound},
		{description: []string{"deleted"}, want: http.StatusNotFound},
		{description: []string{"missing", "content"}, want: http.StatusNotFound},
		{description: []string{"admin", "id"}, want: http.StatusOK},
		{description: []string{"admin", "draft", "id"}, want: http.StatusOK},
		{description: []string{"admin", "scheduled", "id"}, want: http.StatusOK},
		{description: []string{"guest", "id"}, want: http.StatusNotFound},
		{description: []string{"non", "admin", "id"}, want: http.StatusForbidden},
		{description: []string{"missing", "id"}, want: http.StatusNotFound},
		{description: []string{"wrong", "namespace", "id"}, want: http.StatusNotFound},
		{description: []string{"deleted", "id"}, want: http.StatusNotFound},
		{description: []string{"uuid", "slug", "priority"}, want: http.StatusOK},
		{description: []string{"unpublished", "uuid", "slug", "fallback"}, want: http.StatusOK},
		{description: []string{"escaped", "slug"}, want: http.StatusOK},
		{description: []string{"escaped", "namespace"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("FindPostBySlug", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want, implementation{name: "tadoku-api", handler: api.handler})
		})
	}
}
