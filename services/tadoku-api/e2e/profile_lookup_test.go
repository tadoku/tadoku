package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionProfileFindByUserID(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"guest", "populated"}, want: http.StatusOK},
		{description: []string{"user", "populated"}, want: http.StatusOK},
		{description: []string{"user2", "populated"}, want: http.StatusOK},
		{description: []string{"admin", "populated"}, want: http.StatusOK},
		{description: []string{"unknown", "user"}, want: http.StatusInternalServerError},
		{description: []string{"nil", "user"}, want: http.StatusInternalServerError},
		{description: []string{"invalid", "user"}, want: http.StatusBadRequest, skipParity: "native responses omit parser details and reflected parameter input"},
	}
	for _, test := range tests {
		name := APITestName("ImmersionProfileFindByUserID", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler, skip: test.skipParity})
		})
	}
}
