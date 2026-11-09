package e2e_test

import (
	"net/http"
	"path/filepath"
	"testing"
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
