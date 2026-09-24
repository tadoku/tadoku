package e2e_test

import (
	"net/http"
	"testing"
	"time"
)

func TestImmersionLogContestRegistrationUpdate(t *testing.T) {
	tests := []struct {
		description          []string
		want                 int
		scoringEngineEnabled bool
		at                   time.Time
	}{
		{description: []string{"disabled", "attach", "stored", "provenance"}, want: http.StatusOK},
		{description: []string{"enabled", "inherit", "override", "replace"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"enabled", "pinned", "fallback"}, want: http.StatusOK, scoringEngineEnabled: true},
		{description: []string{"noop", "bypasses", "locks"}, want: http.StatusOK},
		{description: []string{"detach", "existing", "tracking"}, want: http.StatusOK},
		{description: []string{"admin", "nonowner"}, want: http.StatusOK},
		{description: []string{"grace", "boundary", "excluded"}, want: http.StatusBadRequest, at: time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)},
		{description: []string{"registration", "language"}, want: http.StatusBadRequest},
		{description: []string{"registration", "activity"}, want: http.StatusBadRequest},
		// Registrations are validated against the log owner, so an admin cannot attach their own registration.
		{description: []string{"admin", "own", "registration"}, want: http.StatusBadRequest},
		{description: []string{"frozen", "mutation"}, want: http.StatusConflict},
		{description: []string{"account", "deletion", "locked", "mutation"}, want: http.StatusConflict},
		{description: []string{"missing", "log"}, want: http.StatusNotFound},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionLogContestRegistrationUpdate", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			handler := http.Handler(api.handler)
			if test.scoringEngineEnabled {
				handler = scoringEnabledHandler
			}
			at := test.at
			if at.IsZero() {
				at = fixtureInstant
			}
			runCaseAt(t, api, name, test.want, at,
				implementation{name: "tadoku-api", handler: handler},
			)
		})
	}
}

func TestImmersionLogDeleteByID(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"owner"}, want: http.StatusOK},
		{description: []string{"admin", "nonowner"}, want: http.StatusOK},
		{description: []string{"contest", "end", "day"}, want: http.StatusOK},
		{description: []string{"contest", "ended", "previous", "day"}, want: http.StatusForbidden},
		{description: []string{"mixed", "active", "and", "ended"}, want: http.StatusForbidden},
		{description: []string{"nonowner"}, want: http.StatusForbidden},
		{description: []string{"frozen"}, want: http.StatusConflict},
		{description: []string{"account", "deletion", "locked"}, want: http.StatusConflict},
		{description: []string{"missing"}, want: http.StatusNotFound},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
	}
	for _, test := range tests {
		name := APITestName("ImmersionLogDeleteByID", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}

func TestImmersionContestModerationDetachLog(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"contest", "owner", "blank", "reason"}, want: http.StatusOK},
		{description: []string{"admin", "nonowner", "long", "reason"}, want: http.StatusOK},
		{description: []string{"absent", "link"}, want: http.StatusOK},
		{description: []string{"frozen", "log"}, want: http.StatusOK},
		{description: []string{"log", "owner", "not", "moderator"}, want: http.StatusForbidden},
		{description: []string{"non", "moderator", "missing", "log"}, want: http.StatusForbidden},
		{description: []string{"missing", "contest", "before", "log"}, want: http.StatusNotFound},
		{description: []string{"deleted", "contest"}, want: http.StatusNotFound},
		{description: []string{"missing", "log", "after", "authorization"}, want: http.StatusNotFound},
		{description: []string{"target", "owner", "account", "locked"}, want: http.StatusConflict},
		{description: []string{"guest", "existing"}, want: http.StatusUnauthorized},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionContestModerationDetachLog", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
