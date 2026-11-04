package e2e_test

import (
	"net/http"
	"testing"
	"time"
)

func TestImmersionLogCreate(t *testing.T) {
	cleanupLogUnit(t, mismatchedActivityUnitID)

	tests := []struct {
		description          []string
		want                 int
		scoringEngineEnabled bool
		canonical            bool
	}{
		{description: []string{"amount", "unit", "tags", "disabled"}, want: http.StatusOK},
		{description: []string{"language", "specific", "unit", "disabled"}, want: http.StatusOK},
		{description: []string{"generic", "unit", "for", "other", "language"}, want: http.StatusOK},
		{description: []string{"language", "specific", "unit", "by", "id"}, want: http.StatusOK},
		{description: []string{"unit", "key", "skips", "stored", "activity", "mismatch"}, want: http.StatusOK, canonical: true},
		{description: []string{"duration", "disabled", "without", "platform"}, want: http.StatusOK},
		{description: []string{"both", "authoritative", "priority", "modifiers"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"authoritative", "contest", "modes"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"authoritative", "pinned", "fallback"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"authoritative", "unmatched"}, want: http.StatusOK, scoringEngineEnabled: true},
		{
			description:          []string{"authoritative", "listening", "minute", "amount", "key"},
			want:                 http.StatusOK,
			scoringEngineEnabled: true,
		},
		{
			description:          []string{"authoritative", "dense", "minute", "amount", "unit", "id"},
			want:                 http.StatusOK,
			scoringEngineEnabled: true,
		},
		{description: []string{"authoritative", "comic", "page", "key"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"authoritative", "two", "column", "unit", "id"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"authoritative", "reading", "pages", "key"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"authoritative", "writing", "pages", "key"}, want: http.StatusOK, scoringEngineEnabled: true},
		{
			description:          []string{"authoritative", "characters", "key", "language", "rate"},
			want:                 http.StatusOK,
			scoringEngineEnabled: true,
		},
		{description: []string{"registration", "within", "grace"}, want: http.StatusOK},
		{description: []string{"authoritative", "missing", "platform"}, want: http.StatusNotFound, scoringEngineEnabled: true},
		{description: []string{"invalid", "tags", "before", "registration"}, want: http.StatusBadRequest},
		{
			description:          []string{"retired", "unit", "tag", "exceeds", "limit"},
			want:                 http.StatusBadRequest,
			scoringEngineEnabled: true,
		},
		{
			description:          []string{"minute", "amount", "rounds", "to", "zero"},
			want:                 http.StatusBadRequest,
			scoringEngineEnabled: true,
		},
		{description: []string{"registration", "not", "owned"}, want: http.StatusBadRequest},
		{description: []string{"registration", "language"}, want: http.StatusBadRequest},
		{description: []string{"registration", "activity"}, want: http.StatusBadRequest},
		{description: []string{"registration", "outside", "grace"}, want: http.StatusBadRequest},
		{description: []string{"description", "too", "long"}, want: http.StatusBadRequest},
		{description: []string{"unit", "id", "other", "language"}, want: http.StatusBadRequest},
		{description: []string{"unit", "id", "other", "activity"}, want: http.StatusBadRequest},
		{description: []string{"unknown", "unit", "id"}, want: http.StatusBadRequest},
		{description: []string{"unit", "key", "other", "language"}, want: http.StatusBadRequest},
		{description: []string{"unit", "id", "stored", "activity", "mismatch"}, want: http.StatusBadRequest, canonical: true},
		{description: []string{"unit", "id", "key", "activity", "mismatch"}, want: http.StatusBadRequest, canonical: true},
		{description: []string{"account", "deletion", "locked"}, want: http.StatusConflict},
		{description: []string{"duration", "unknown", "language", "readback"}, want: http.StatusNotFound},
		{description: []string{"guest", "empty", "body"}, want: http.StatusUnauthorized},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionLogCreate", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			handler := http.Handler(api.handler)
			if test.scoringEngineEnabled {
				handler = scoringEnabledHandler
			}
			impl := implementation{name: "tadoku-api", handler: handler}
			if test.canonical {
				runCase(t, api, name, test.want, impl)
				return
			}

			runScenarioCase(t, name, test.want, impl)
		})
	}
}

func TestImmersionLogUpdate(t *testing.T) {
	t.Parallel()

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
		{description: []string{"authoritative", "comic", "page", "key"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"disabled", "without", "platform"}, want: http.StatusOK},
		{description: []string{"disabled", "contest", "end", "day"}, want: http.StatusOK},
		{description: []string{"authoritative", "missing", "platform"}, want: http.StatusNotFound, scoringEngineEnabled: true},
		{description: []string{"nonowner", "before", "invalid", "tags"}, want: http.StatusForbidden},
		{description: []string{"frozen", "after", "invalid", "tags"}, want: http.StatusBadRequest},
		{description: []string{"description", "too", "long"}, want: http.StatusBadRequest},
		{description: []string{"frozen"}, want: http.StatusConflict},
		{description: []string{"account", "deletion", "locked"}, want: http.StatusConflict},
		{description: []string{"missing"}, want: http.StatusNotFound},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionLogUpdate", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			handler := http.Handler(api.handler)
			if test.scoringEngineEnabled {
				handler = scoringEnabledHandler
			}
			at := test.at
			if at.IsZero() {
				at = fixtureInstant
			}
			runScenarioCaseAt(t, name, test.want, at,
				implementation{name: "tadoku-api", handler: handler},
			)
		})
	}
}
