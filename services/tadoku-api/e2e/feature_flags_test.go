package e2e_test

import (
	"net/http"
	"testing"
)

type featureAccessCase struct {
	description []string
	want        int
	unavailable bool
}

func TestImmersionFeatureFlagDecisions(t *testing.T) {
	tests := []featureAccessCase{
		{description: []string{"named", "user", "enabled"}, want: http.StatusOK},
		{description: []string{"guest", "safe", "default"}, want: http.StatusOK},
		{description: []string{"provider", "unavailable", "safe", "default"}, want: http.StatusOK, unavailable: true},
	}

	for _, test := range tests {
		name := APITestName("ImmersionFeatureFlagDecisions", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			native := http.Handler(api.handler)
			if test.unavailable {
				native = withFliptUnavailable(native)
			}
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: native},
			)
		})
	}
}

func TestImmersionFeatureAccessGet(t *testing.T) {
	tests := []featureAccessCase{
		{description: []string{"enabled"}, want: http.StatusOK},
		{description: []string{"invalid", "flag"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "user", "id"}, want: http.StatusBadRequest},
		{description: []string{"zero", "user", "id"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"provider", "unavailable"}, want: http.StatusBadGateway, unavailable: true},
	}
	runFeatureAccessCases(t, "ImmersionFeatureAccessGet", tests)
}

func TestImmersionFeatureAccessGrant(t *testing.T) {
	tests := []featureAccessCase{
		{description: []string{"disabled", "user"}, want: http.StatusOK},
		{description: []string{"already", "enabled"}, want: http.StatusOK},
		{description: []string{"invalid", "flag"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"provider", "unavailable"}, want: http.StatusBadGateway, unavailable: true},
	}
	runFeatureAccessCases(t, "ImmersionFeatureAccessGrant", tests)
}

func TestImmersionFeatureAccessRevoke(t *testing.T) {
	tests := []featureAccessCase{
		{description: []string{"enabled", "user"}, want: http.StatusOK},
		{description: []string{"already", "disabled"}, want: http.StatusOK},
		{description: []string{"invalid", "flag"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"provider", "unavailable"}, want: http.StatusBadGateway, unavailable: true},
	}
	runFeatureAccessCases(t, "ImmersionFeatureAccessRevoke", tests)
}

func runFeatureAccessCases(t *testing.T, operation string, tests []featureAccessCase) {
	t.Helper()
	for _, test := range tests {
		name := APITestName(operation, test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			native := http.Handler(api.handler)
			if test.unavailable {
				native = withFliptUnavailable(native)
			}
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: native},
			)
		})
	}
}

func withFliptUnavailable(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		flipt.SetUnavailable(true)
		next.ServeHTTP(response, request)
	})
}
