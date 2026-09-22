package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionScoringRuleSetListPlatform(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"member", "published", "only"}, want: http.StatusOK},
		{description: []string{"admin", "includes", "draft", "ordered"}, want: http.StatusOK},
		{description: []string{"missing", "active", "config"}, want: http.StatusInternalServerError},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
	}
	for _, test := range tests {
		name := APITestName("ImmersionScoringRuleSetListPlatform", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler},
			)
		})
	}
}

func TestImmersionScoringRuleSetListContest(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"owner", "versions", "ordered"}, want: http.StatusOK},
		{description: []string{"admin", "nonowner"}, want: http.StatusOK},
		{description: []string{"member", "nonowner"}, want: http.StatusForbidden},
		{description: []string{"unknown", "signing", "key"}, want: http.StatusUnauthorized},
		{description: []string{"missing"}, want: http.StatusNotFound},
		{description: []string{"guest", "existing"}, want: http.StatusForbidden},
		{description: []string{"guest", "missing"}, want: http.StatusNotFound},
	}
	for _, test := range tests {
		name := APITestName("ImmersionScoringRuleSetListContest", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "immersion-api", handler: legacyImmersion.handler},
			)
		})
	}
}
