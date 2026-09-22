package e2e_test

import (
	"net/http"
	"testing"
)

func TestListPosts(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest"}, want: http.StatusOK},
		{description: []string{"non", "admin"}, want: http.StatusOK},
		{description: []string{"admin", "includes", "drafts", "and", "scheduled"}, want: http.StatusOK},
		{description: []string{"guest", "including", "drafts"}, want: http.StatusForbidden},
		{description: []string{"non", "admin", "including", "drafts"}, want: http.StatusForbidden},
		{description: []string{"without", "posts"}, want: http.StatusOK},
		{description: []string{"other", "namespace"}, want: http.StatusOK},
		{description: []string{"default", "page"}, want: http.StatusOK},
		{description: []string{"zero", "page", "size"}, want: http.StatusOK},
		{description: []string{"second", "page"}, want: http.StatusOK},
		{description: []string{"page", "size", "capped"}, want: http.StatusOK},
		{description: []string{"invalid", "page", "size"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "page"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "include", "drafts"}, want: http.StatusBadRequest},
		{description: []string{"negative", "page", "size"}, want: http.StatusBadRequest},
		{description: []string{"negative", "page"}, want: http.StatusBadRequest},
		{description: []string{"offset", "overflow"}, want: http.StatusOK},
		{description: []string{"page", "beyond", "total", "size"}, want: http.StatusOK},
		{description: []string{"tied", "timestamps", "first", "page"}, want: http.StatusOK},
		{description: []string{"tied", "timestamps", "final", "page"}, want: http.StatusOK},
		{description: []string{"scheduled", "posts", "hidden"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("ListPosts", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want, implementation{name: "tadoku-api", handler: api.handler})
		})
	}
}
