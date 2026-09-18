package e2e_test

import (
	"net/http"
	"testing"
)

func TestCreatePost(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"admin"}, want: http.StatusCreated, skipParity: "approved 201 Created instead of legacy 200 OK"},
		{description: []string{"draft"}, want: http.StatusCreated, skipParity: "approved 201 Created instead of legacy 200 OK"},
		{description: []string{"null", "publication"}, want: http.StatusCreated, skipParity: "approved 201 Created instead of legacy 200 OK"},
		{description: []string{"scheduled"}, want: http.StatusCreated, skipParity: "approved 201 Created instead of legacy 200 OK"},
		{description: []string{"unicode", "slug"}, want: http.StatusCreated, skipParity: "approved 201 Created instead of legacy 200 OK"},
		{description: []string{"whitespace", "fields"}, want: http.StatusCreated, skipParity: "approved 201 Created instead of legacy 200 OK"},
		{description: []string{"namespace", "isolation"}, want: http.StatusCreated, skipParity: "approved 201 Created instead of legacy 200 OK"},
		{description: []string{"ignored", "server", "fields"}, want: http.StatusCreated, skipParity: "approved 201 Created instead of legacy 200 OK"},
		{description: []string{"trailing", "json"}, want: http.StatusCreated, skipParity: "approved 201 Created instead of legacy 200 OK"},
		{description: []string{"offset", "date"}, want: http.StatusCreated, skipParity: "approved 201 instead of legacy 200; publication timestamps also normalize to UTC"},
		{description: []string{"empty", "slug"}, want: http.StatusBadRequest},
		{description: []string{"short", "slug"}, want: http.StatusBadRequest},
		{description: []string{"uppercase", "slug"}, want: http.StatusBadRequest},
		{description: []string{"one", "rune", "slug"}, want: http.StatusBadRequest},
		{description: []string{"empty", "title"}, want: http.StatusBadRequest},
		{description: []string{"empty", "content"}, want: http.StatusBadRequest},
		{description: []string{"zero", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "time"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest, skipParity: "legacy dereferences the missing ID after binding an empty body"},
		{description: []string{"null", "body"}, want: http.StatusBadRequest, skipParity: "legacy dereferences the missing ID after binding a null body"},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"duplicate", "id"}, want: http.StatusConflict, skipParity: "native returns 409 for a unique violation; legacy returns 400"},
		{description: []string{"duplicate", "slug"}, want: http.StatusConflict, skipParity: "native returns 409 for a unique violation; legacy returns 400"},
	}

	for _, test := range tests {
		name := APITestName("CreatePost", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler, skip: test.skipParity},
			)
		})
	}
}
