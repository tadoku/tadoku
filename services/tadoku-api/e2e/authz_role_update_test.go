package e2e_test

import (
	"io"
	"log/slog"
	"net/http"
	"testing"

	transport "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

func TestAuthzRoleUpdate(t *testing.T) {
	auditUnavailable := auditUnavailableRoleUpdateHandler(t)
	tests := []struct {
		description      []string
		want             int
		auditUnavailable bool
	}{
		{description: []string{"ban", "user"}, want: http.StatusOK},
		{description: []string{"unban", "user"}, want: http.StatusOK},
		{description: []string{"whitespace", "reason"}, want: http.StatusOK},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "role"}, want: http.StatusBadRequest},
		{description: []string{"missing", "reason"}, want: http.StatusBadRequest},
		{
			description: []string{"guest", "empty", "body"},
			want:        http.StatusBadRequest,
		},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin", "invalid", "role"}, want: http.StatusForbidden},
		{description: []string{"missing", "user", "marked", "admin"}, want: http.StatusNotFound},
		{description: []string{"target", "admin"}, want: http.StatusForbidden},
		{description: []string{"audit", "unavailable"}, want: http.StatusInternalServerError, auditUnavailable: true},
	}

	for _, test := range tests {
		name := APITestName("AuthzRoleUpdate", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			nativeHandler := api.handler
			if test.auditUnavailable {
				nativeHandler = auditUnavailable
			}

			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: nativeHandler},
			)
		})
	}
}

func auditUnavailableRoleUpdateHandler(t *testing.T) *transport.Router {
	t.Helper()

	closedPool := openClosedPool(t, api.db.DSN)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	native, _, _, err := newTestRouterWithLogger(
		t.Context(),
		api.db.Pool,
		closedPool,
		keto,
		api.kratos,
		logger,
	)
	if err != nil {
		t.Fatal(err)
	}

	return native
}
