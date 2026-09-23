package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionLogGetConfigurations(t *testing.T) {
	tests := []struct {
		description          []string
		want                 int
		scoringEngineEnabled bool
	}{
		{description: []string{"empty_history_disabled"}, want: http.StatusOK, scoringEngineEnabled: false},
		{description: []string{"empty_history_enabled"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"populated_disabled"}, want: http.StatusOK, scoringEngineEnabled: false},
		{description: []string{"populated_enabled"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"other_user_history"}, want: http.StatusOK, scoringEngineEnabled: false},
		{description: []string{"invalid_signed_subject"}, want: http.StatusUnauthorized},
		{description: []string{"guest"}, want: http.StatusUnauthorized, scoringEngineEnabled: false},
	}
	for _, test := range tests {
		name := APITestName("ImmersionLogGetConfigurations", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			var native http.Handler = api.handler
			if test.scoringEngineEnabled {
				native = scoringEnabledHandler
			}
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: native},
			)
		})
	}
}
