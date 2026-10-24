package e2e_test

import (
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIdentitySettingsRejectBeforeDatabase(t *testing.T) {
	resetIdentity(t)
	pool, err := pgxpool.NewWithConfig(t.Context(), api.db.AppPool.Config())
	if err != nil {
		t.Fatal(err)
	}
	handler, _, _, err := newTestRouterWithLogger(
		t.Context(), pool, pool, api.keto,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		pool.Close()
		t.Fatal(err)
	}
	pool.Close()
	for _, tc := range []struct{ method, path, body string }{
		{"PUT", "/identity/settings/password", `{"password":"Replacement-839!"}`},
		{"PATCH", "/identity/settings/profile", `{"email":"changed@example.test"}`},
	} {
		response := identityRequest(t, handler, identityBranchA, tc.method, tc.path, tc.body, "", nil)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("%s: status=%d want=%d", tc.path, response.Code, http.StatusUnauthorized)
		}
	}
}
