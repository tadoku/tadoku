package e2e_test

import (
	"net/http"
	"testing"
)

func TestCreatePage(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"admin"}, want: http.StatusCreated, skipParity: "canonical page create returns 201 Created; legacy returns 200 OK"},
		{description: []string{"missing", "html"}, want: http.StatusBadRequest, skipParity: "legacy dereferences missing HTML; native rejects it as invalid input"},
		{description: []string{"null", "html"}, want: http.StatusBadRequest, skipParity: "legacy dereferences null HTML; native rejects it as invalid input"},
		{description: []string{"empty", "html"}, want: http.StatusBadRequest},
		{description: []string{"short", "slug"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"duplicate", "id"}, want: http.StatusConflict, skipParity: "native maps uniqueness conflicts to 409; legacy returns 400"},
		{description: []string{"duplicate", "slug"}, want: http.StatusConflict, skipParity: "native maps uniqueness conflicts to 409; legacy returns 400"},
	}

	for _, test := range tests {
		name := APITestName("CreatePage", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler, skip: test.skipParity},
			)
		})
	}
}
