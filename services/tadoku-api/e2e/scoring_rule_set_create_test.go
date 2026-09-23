package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionScoringRuleSetCreatePlatform(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"admin", "multi", "rule", "normalizes"}, want: http.StatusOK},
		{description: []string{"admin", "empty", "rules"}, want: http.StatusOK},
		{description: []string{"admin", "ignores", "contest", "fields"}, want: http.StatusOK},
		{description: []string{"member"}, want: http.StatusForbidden},
		{description: []string{"unknown", "signing", "key"}, want: http.StatusUnauthorized},
		{description: []string{"guest"}, want: http.StatusUnauthorized, skipParity: "Tadoku API consistently requires authentication before administrator authorization; legacy returns forbidden for a signed guest"},
	}
	seed := int64(1)
	for _, test := range tests {
		name := APITestName("ImmersionScoringRuleSetCreatePlatform", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler, uuidSeed: &seed},
				implementation{name: "immersion-api", handler: legacyImmersion.handler, uuidSeed: &seed, skip: test.skipParity},
			)
		})
	}
}

func TestImmersionScoringRuleSetCreateContest(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"owner", "replace", "zero", "rate"}, want: http.StatusOK},
		{description: []string{"owner", "override", "published", "fallback"}, want: http.StatusOK},
		{description: []string{"replace", "forbids", "fallback"}, want: http.StatusBadRequest},
		{description: []string{"override", "requires", "fallback"}, want: http.StatusBadRequest},
		{description: []string{"missing", "fallback"}, want: http.StatusInternalServerError},
		{description: []string{"draft", "fallback"}, want: http.StatusBadRequest},
		{description: []string{"contest", "fallback"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "priority"}, want: http.StatusBadRequest},
		{description: []string{"duplicate", "priorities"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "rate"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "activity"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "source"}, want: http.StatusBadRequest},
		{description: []string{"unknown", "language"}, want: http.StatusBadRequest},
		{description: []string{"unit", "activity", "mismatch"}, want: http.StatusBadRequest},
		{description: []string{"nonowner"}, want: http.StatusForbidden},
		{description: []string{"admin", "nonowner"}, want: http.StatusOK},
		{description: []string{"already", "started"}, want: http.StatusConflict},
		{description: []string{"unknown", "signing", "key"}, want: http.StatusUnauthorized},
		{description: []string{"guest", "existing"}, want: http.StatusUnauthorized, skipParity: "Tadoku API consistently requires authentication before administrator authorization; legacy returns forbidden for a signed guest"},
		{description: []string{"guest", "missing"}, want: http.StatusNotFound},
	}
	seed := int64(1)
	for _, test := range tests {
		name := APITestName("ImmersionScoringRuleSetCreateContest", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler, uuidSeed: &seed},
				implementation{name: "immersion-api", handler: legacyImmersion.handler, uuidSeed: &seed, skip: test.skipParity},
			)
		})
	}
}
