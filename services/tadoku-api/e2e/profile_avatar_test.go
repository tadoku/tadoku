package e2e_test

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

// TestImmersionProfileAvatarByUserID sends each case's request twice, so calls counts the Gravatar requests that the
// cache did not absorb.
func TestImmersionProfileAvatarByUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description []string
		want        int
		email       string
		calls       int
	}{
		{description: []string{"guest", "gravatar"}, want: http.StatusOK, email: "reader@example.test", calls: 1},
		{description: []string{"normalized", "email"}, want: http.StatusOK, email: "mixed@example.test", calls: 1},
		{description: []string{"no", "gravatar"}, want: http.StatusNotFound, email: "user2@example.test", calls: 1},
		{description: []string{"unknown", "user"}, want: http.StatusNotFound},
		{description: []string{"deleted", "user"}, want: http.StatusNotFound, email: "deleted@example.test", calls: 0},
		{description: []string{"upstream", "error"}, want: http.StatusServiceUnavailable, email: "error@example.test", calls: 2},
		{
			description: []string{"disallowed", "content", "type"},
			want:        http.StatusServiceUnavailable,
			email:       "svg@example.test",
			calls:       2,
		},
		{
			description: []string{"oversized", "body"},
			want:        http.StatusServiceUnavailable,
			email:       "oversized@example.test",
			calls:       2,
		},
		{description: []string{"upstream", "timeout"}, want: http.StatusGatewayTimeout, email: "slow@example.test", calls: 2},
		{description: []string{"invalid", "user"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionProfileAvatarByUserID", test.want, test.description...)
		t.Run(name+"/tadoku-api", func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join("testdata", name)
			impl := implementation{name: "tadoku-api", handler: api.handler}
			s := newFixtureScenario(t)

			s.runCase(t, dir, test.want, fixtureInstant, impl)
			s.checkCase(t, dir, test.want, impl)

			if test.email == "" {
				return
			}
			if got := gravatar.callsFor(test.email); got != test.calls {
				t.Errorf("Gravatar requests for %s = %d; want %d", test.email, got, test.calls)
			}
		})
	}
}

func TestImmersionProfileAvatarRefresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description []string
		want        int
		email       string
	}{
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"refreshed"}, want: http.StatusNoContent, email: "refresh@example.test"},
		{description: []string{"no", "gravatar"}, want: http.StatusNoContent, email: "refresh-none@example.test"},
		{description: []string{"upstream", "error"}, want: http.StatusServiceUnavailable, email: "refresh-error@example.test"},
		{description: []string{"upstream", "timeout"}, want: http.StatusGatewayTimeout, email: "refresh-slow@example.test"},
	}
	for _, test := range tests {
		name := APITestName("ImmersionProfileAvatarRefresh", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runScenarioCase(t, name, test.want, implementation{name: "tadoku-api", handler: api.handler})

			if test.email == "" {
				return
			}
			if got := gravatar.callsFor(test.email); got != 1 {
				t.Errorf("Gravatar requests for %s = %d; want 1", test.email, got)
			}
		})
	}
}

// TestProfileAvatarRefreshJourney starts every cast member's avatar cache with a stale image, so a read shows whether
// a refresh replaced it.
func TestProfileAvatarRefreshJourney(t *testing.T) {
	runScenarioJourneyWithSetup(t, api.handler, "ProfileAvatarRefresh", func(t *testing.T, s *fixtureScenario) {
		for _, userID := range []string{
			"11111111-1111-4111-8111-111111111111",
			"22222222-2222-4222-8222-222222222222",
			"44444444-4444-4444-8444-444444444444",
		} {
			seedAvatarCache(t, s.tenant(t, primaryRole).String(), userID, "image/png\nstale avatar")
		}
	}, nil, []step{
		{request: "read_stale", as: guest, want: http.StatusOK},
		{
			request: "refresh",
			as:      user,
			want:    http.StatusNoContent,
			others:  cast{guest: http.StatusUnauthorized, banned: http.StatusForbidden},
		},
		{request: "read_refreshed", as: guest, want: http.StatusOK},
		{request: "refresh_without_gravatar", as: user2, want: http.StatusNoContent},
		{request: "read_without_gravatar", as: guest, want: http.StatusNotFound},
		{request: "refresh_failure", as: admin, want: http.StatusServiceUnavailable},
		{request: "read_after_failure", as: guest, want: http.StatusOK},
	})
	if t.Failed() {
		return
	}

	for _, email := range []string{
		"journey-avatar@example.test",
		"journey-none@example.test",
		"journey-error@example.test",
	} {
		if got := gravatar.callsFor(email); got != 1 {
			t.Errorf("Gravatar requests for %s = %d; want 1", email, got)
		}
	}
}

func seedAvatarCache(t *testing.T, tenantKey, userID, value string) {
	t.Helper()
	client := leaderboardValkey.client
	key := "tenant:" + tenantKey + ":profile:avatar:" + userID
	if err := client.Do(t.Context(), client.B().Set().Key(key).Value(value).Ex(time.Hour).Build()).Error(); err != nil {
		t.Fatal(err)
	}
}
