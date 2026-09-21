package e2e_test

import (
	"net/http"
	"testing"
)

func TestAuthzPermissionCheck(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		configured  bool
	}{
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"missing", "relation"}, want: http.StatusBadRequest},
		{
			description: []string{"guest", "empty", "body"},
			want:        http.StatusBadRequest,
		},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"not", "allowlisted"}, want: http.StatusForbidden},
		{description: []string{"configured", "allowed"}, want: http.StatusOK, configured: true},
		{description: []string{"configured", "denied"}, want: http.StatusOK, configured: true},
		{description: []string{"configured", "namespace", "not", "allowlisted"}, want: http.StatusForbidden, configured: true},
		{description: []string{"configured", "relation", "not", "allowlisted"}, want: http.StatusForbidden, configured: true},
		{
			description: []string{"banned", "malformed", "json"},
			want:        http.StatusForbidden,
		},
	}

	for _, test := range tests {
		name := APITestName("AuthzPermissionCheck", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			nativeHandler := http.Handler(api.handler)
			if test.configured {
				nativeHandler = configuredAuthz
			}
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: nativeHandler},
			)
		})
	}
}
