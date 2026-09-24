package e2e_test

import (
	"net/http"
	"testing"
	"time"
)

func TestImmersionLogCreate(t *testing.T) {
	tests := []struct {
		description          []string
		want                 int
		scoringEngineEnabled bool
	}{
		{description: []string{"amount", "unit", "tags", "disabled"}, want: http.StatusOK},
		{description: []string{"language", "specific", "unit", "disabled"}, want: http.StatusOK},
		{description: []string{"duration", "disabled", "without", "platform"}, want: http.StatusOK},
		{description: []string{"both", "authoritative", "priority", "modifiers"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"authoritative", "contest", "modes"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"authoritative", "pinned", "fallback"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"authoritative", "unmatched"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"registration", "within", "grace"}, want: http.StatusOK},
		{description: []string{"authoritative", "missing", "platform"}, want: http.StatusNotFound, scoringEngineEnabled: true},
		{description: []string{"invalid", "tags", "before", "registration"}, want: http.StatusBadRequest},
		{description: []string{"registration", "not", "owned"}, want: http.StatusBadRequest},
		{description: []string{"registration", "language"}, want: http.StatusBadRequest},
		{description: []string{"registration", "activity"}, want: http.StatusBadRequest},
		{description: []string{"registration", "outside", "grace"}, want: http.StatusBadRequest},
		{description: []string{"account", "deletion", "locked"}, want: http.StatusConflict},
		{description: []string{"duration", "unknown", "language", "readback"}, want: http.StatusNotFound},
		{description: []string{"guest", "empty", "body"}, want: http.StatusUnauthorized},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionLogCreate", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			native := http.Handler(api.handler)
			if test.scoringEngineEnabled {
				native = scoringEnabledHandler
			}
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: native},
			)
		})
	}
}

func TestImmersionLogUpdate(t *testing.T) {
	tests := []struct {
		description          []string
		want                 int
		scoringEngineEnabled bool
		at                   time.Time
	}{
		{description: []string{"owner", "duration", "disabled"}, want: http.StatusOK},
		{description: []string{"admin", "amount", "disabled"}, want: http.StatusOK},
		{description: []string{"authoritative", "ongoing", "and", "ended"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"authoritative", "exact", "contest", "end"}, want: http.StatusOK, scoringEngineEnabled: true, at: time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)},
		{description: []string{"authoritative", "ended", "today", "only"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"disabled", "without", "platform"}, want: http.StatusOK},
		{description: []string{"disabled", "contest", "end", "day"}, want: http.StatusOK},
		{description: []string{"authoritative", "missing", "platform"}, want: http.StatusNotFound, scoringEngineEnabled: true},
		{description: []string{"nonowner", "before", "invalid", "tags"}, want: http.StatusForbidden},
		{description: []string{"frozen", "after", "invalid", "tags"}, want: http.StatusBadRequest},
		{description: []string{"frozen"}, want: http.StatusConflict},
		{description: []string{"account", "deletion", "locked"}, want: http.StatusConflict},
		{description: []string{"missing"}, want: http.StatusNotFound},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionLogUpdate", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			native := http.Handler(api.handler)
			if test.scoringEngineEnabled {
				native = scoringEnabledHandler
			}
			at := test.at
			if at.IsZero() {
				at = fixtureInstant
			}
			runCaseAt(t, api, name, test.want, at,
				implementation{name: "tadoku-api", handler: native},
			)
		})
	}
}
