package e2e_test

import (
	"net/http"
	"testing"
)

func TestFindLatestOfficialContest(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"greatest", "start", "includes", "future", "deleted"}, want: http.StatusOK},
		{description: []string{"empty"}, want: http.StatusNotFound},
	}

	for _, test := range tests {
		name := APITestName("FindLatestOfficialContest", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler},
			)
		})
	}
}
