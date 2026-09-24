package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionProfileYearlyActivitySplitByUserID(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest", "populated"}, want: http.StatusOK},
		{description: []string{"user", "populated"}, want: http.StatusOK},
		{description: []string{"user2", "populated"}, want: http.StatusOK},
		{description: []string{"admin", "populated"}, want: http.StatusOK},
		{description: []string{"unknown", "user"}, want: http.StatusOK},
		{description: []string{"nil", "user"}, want: http.StatusOK},
		{description: []string{"invalid", "user"}, want: http.StatusBadRequest},
		{description: []string{"empty"}, want: http.StatusOK},
		{description: []string{"empty", "year"}, want: http.StatusOK},
		{description: []string{"previous", "year"}, want: http.StatusOK},
		{description: []string{"next", "year"}, want: http.StatusOK},
		{description: []string{"zero", "year"}, want: http.StatusOK},
		{description: []string{"negative", "year"}, want: http.StatusOK},
		{description: []string{"wrapped", "year"}, want: http.StatusOK},
		{description: []string{"invalid", "year"}, want: http.StatusBadRequest},
		{description: []string{"overflow", "year"}, want: http.StatusBadRequest},
		{description: []string{"all", "deleted"}, want: http.StatusOK},
		{description: []string{"invalid", "activity"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionProfileYearlyActivitySplitByUserID", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
