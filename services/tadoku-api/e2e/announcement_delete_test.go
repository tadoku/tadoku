package e2e_test

import (
	"net/http"
	"testing"
)

func TestDeleteAnnouncement(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin"}, want: http.StatusNoContent},
		{description: []string{"missing"}, want: http.StatusNoContent},
		{description: []string{"already", "deleted"}, want: http.StatusNoContent},
		{description: []string{"wrong", "namespace"}, want: http.StatusNoContent},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"active", "as", "id"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("DeleteAnnouncement", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler},
			)
		})
	}
}
