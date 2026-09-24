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
		{description: []string{"missing", "active", "config"}, want: http.StatusNotFound},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
	}
	for _, test := range tests {
		name := APITestName("ImmersionScoringRuleSetListPlatform", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
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
		{description: []string{"missing"}, want: http.StatusNotFound},
		{description: []string{"guest", "existing"}, want: http.StatusUnauthorized},
		{description: []string{"guest", "missing"}, want: http.StatusUnauthorized},
	}
	for _, test := range tests {
		name := APITestName("ImmersionScoringRuleSetListContest", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
