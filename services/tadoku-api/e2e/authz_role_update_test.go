package e2e_test

import (
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	legacyrepository "github.com/tadoku/tadoku/services/authz-api/storage/postgres/repository"
	transport "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

func TestAuthzRoleUpdate(t *testing.T) {
	nativeAuditUnavailable, legacyAuditUnavailable := auditUnavailableRoleUpdateHandlers(t)
	tests := []struct {
		description      []string
		want             int
		skipParity       string
		auditUnavailable bool
	}{
		{description: []string{"ban", "user"}, want: http.StatusOK},
		{description: []string{"unban", "user"}, want: http.StatusOK},
		{description: []string{"whitespace", "reason"}, want: http.StatusOK},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "role"}, want: http.StatusBadRequest},
		{description: []string{"missing", "reason"}, want: http.StatusBadRequest},
		{description: []string{"guest", "empty", "body"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin", "invalid", "role"}, want: http.StatusForbidden},
		{
			description: []string{"banned", "malformed", "json"},
			want:        http.StatusForbidden,
			skipParity:  "the native shared ban gate rejects before decoding while legacy reports malformed JSON",
		},
		{description: []string{"missing", "user", "marked", "admin"}, want: http.StatusNotFound},
		{description: []string{"target", "admin"}, want: http.StatusForbidden},
		{description: []string{"audit", "unavailable"}, want: http.StatusInternalServerError, auditUnavailable: true},
	}

	for _, test := range tests {
		name := APITestName("AuthzRoleUpdate", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			nativeHandler := api.handler
			legacyHandler := legacyAuthz.handler
			if test.auditUnavailable {
				nativeHandler = nativeAuditUnavailable
				legacyHandler = legacyAuditUnavailable
			}

			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: nativeHandler},
				implementation{name: "authz-api", handler: legacyHandler, skip: test.skipParity},
			)
		})
	}
}

func auditUnavailableRoleUpdateHandlers(t *testing.T) (*transport.Router, http.Handler) {
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

	config, err := pgx.ParseConfig(api.db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	closedDB := stdlib.OpenDB(*config)
	if err := closedDB.Close(); err != nil {
		t.Fatal(err)
	}
	legacy, err := newLegacyAuthzHandler(
		authenticationJWKS.URL,
		keto.ReadURL(),
		keto.WriteURL(),
		api.kratos.CursorClient(),
		legacyrepository.NewRepository(closedDB),
	)
	if err != nil {
		t.Fatal(err)
	}

	return native, legacy
}
