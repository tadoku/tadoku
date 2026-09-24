package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionScoringRuleSetPublish(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin", "platform", "draft"}, want: http.StatusOK},
		{description: []string{"owner", "contest", "draft"}, want: http.StatusOK},
		{description: []string{"admin", "contest", "draft"}, want: http.StatusOK},
		{description: []string{"published"}, want: http.StatusConflict},
		{description: []string{"member", "platform"}, want: http.StatusForbidden},
		{description: []string{"nonowner", "contest"}, want: http.StatusForbidden},
		{description: []string{"contest", "already", "started"}, want: http.StatusConflict},
		{description: []string{"missing"}, want: http.StatusNotFound},
		{description: []string{"guest", "existing"}, want: http.StatusUnauthorized},
		{description: []string{"guest", "missing"}, want: http.StatusUnauthorized},
	}
	for _, test := range tests {
		name := APITestName("ImmersionScoringRuleSetPublish", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}

func TestImmersionScoringRuleSetActivate(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin", "platform", "published"}, want: http.StatusNoContent},
		{description: []string{"repeat", "platform", "activation"}, want: http.StatusNoContent},
		{description: []string{"owner", "contest", "published"}, want: http.StatusNoContent},
		{description: []string{"admin", "contest", "published"}, want: http.StatusNoContent},
		{description: []string{"draft"}, want: http.StatusConflict},
		{description: []string{"member", "platform"}, want: http.StatusForbidden},
		{description: []string{"nonowner", "contest"}, want: http.StatusForbidden},
		{description: []string{"contest", "already", "started"}, want: http.StatusConflict},
		{description: []string{"missing"}, want: http.StatusNotFound},
		{description: []string{"guest", "existing"}, want: http.StatusUnauthorized},
		{description: []string{"guest", "missing"}, want: http.StatusUnauthorized},
	}
	for _, test := range tests {
		name := APITestName("ImmersionScoringRuleSetActivate", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
