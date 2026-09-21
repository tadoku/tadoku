package e2e_test

import (
	"net/http"
	"testing"
)

func TestContestConfigurationOptions(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest"}, want: http.StatusOK},
		{description: []string{"admin"}, want: http.StatusOK},
	}

	for _, test := range tests {
		name := APITestName("ContestConfigurationOptions", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler},
			)
		})
	}
}
