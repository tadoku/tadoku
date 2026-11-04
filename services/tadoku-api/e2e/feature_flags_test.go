package e2e_test

import (
	"net/http"
	"testing"
)

type featureFlagCase struct {
	description []string
	want        int
	unavailable bool
	canonical   bool
}

func TestImmersionFeatureFlagDecisions(t *testing.T) {
	tests := []featureFlagCase{
		{description: []string{"named", "user", "enabled"}, want: http.StatusOK, canonical: true},
		{description: []string{"test", "tenant", "provider", "disabled"}, want: http.StatusOK, canonical: true},
		{description: []string{"guest", "safe", "default"}, want: http.StatusOK},
		{description: []string{"provider", "unavailable", "safe", "default"}, want: http.StatusOK, unavailable: true},
	}

	for _, test := range tests {
		name := APITestName("ImmersionFeatureFlagDecisions", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			handler := http.Handler(api.handler)
			if test.unavailable {
				handler = withFliptUnavailable(handler)
			}
			impl := implementation{name: "tadoku-api", handler: handler}
			if test.unavailable || test.canonical {
				runCase(t, api, name, test.want, impl)
				return
			}

			runScenarioCase(t, name, test.want, impl)
		})
	}
}

func withFliptUnavailable(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		flipt.SetUnavailable(true)
		defer flipt.SetUnavailable(false)
		next.ServeHTTP(response, request)
	})
}
