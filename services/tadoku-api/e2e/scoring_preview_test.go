package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionScorePreview(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"platform", "amount", "tag", "priority"}, want: http.StatusOK},
		{description: []string{"amount", "precedes", "duration"}, want: http.StatusOK},
		{description: []string{"duration", "source"}, want: http.StatusOK},
		{description: []string{"uncovered", "input"}, want: http.StatusOK},
		{description: []string{"contest", "override", "fallback"}, want: http.StatusOK},
		{description: []string{"contest", "inherits", "platform"}, want: http.StatusOK},
		{description: []string{"contest", "replace", "no", "match"}, want: http.StatusOK},
		{description: []string{"duplicate", "registrations", "preserve", "order"}, want: http.StatusOK},
		{description: []string{"registration", "not", "owned"}, want: http.StatusBadRequest},
		{description: []string{"registration", "language", "not", "allowed"}, want: http.StatusBadRequest},
		{description: []string{"registration", "activity", "not", "allowed"}, want: http.StatusBadRequest},
		{description: []string{"registration", "not", "ongoing"}, want: http.StatusBadRequest},
		{description: []string{"deleted", "contest"}, want: http.StatusNotFound},
		{description: []string{"unit", "id", "key", "mismatch"}, want: http.StatusBadRequest},
		{description: []string{"duration", "with", "unit", "without", "amount"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "duration", "with", "amount"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "tags"}, want: http.StatusInternalServerError},
		{description: []string{"invalid", "tags", "before", "unknown", "registration"}, want: http.StatusInternalServerError},
		{description: []string{"missing", "language", "before", "invalid", "tags"}, want: http.StatusBadRequest},
		{description: []string{"missing", "language", "before", "missing", "active", "platform"}, want: http.StatusBadRequest},
		{description: []string{"missing", "active", "platform"}, want: http.StatusNotFound},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
	}
	for _, test := range tests {
		name := APITestName("ImmersionScorePreview", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
