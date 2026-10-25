package e2e_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIdentityV2UniqueAccountTenantIsolation(t *testing.T) {
	resetIdentity(t)
	registration := `{"email":"unique@example.test","display_name":"Unique Reader","password":"Independent-653!"}`
	response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/registration", registration, "", nil)
	requireIdentityStatus(t, response, http.StatusCreated)
	for _, key := range []string{"tadoku/prod", identityBranchB} {
		response = identityRequest(t, api.handler, key, "POST", "/identity/login",
			`{"email":"unique@example.test","password":"Independent-653!"}`, "", nil)
		requireIdentityStatus(t, response, http.StatusUnauthorized)
		var failure struct{ Code string }
		if err := json.Unmarshal(response.Body.Bytes(), &failure); err != nil {
			t.Fatal(err)
		}
		if failure.Code != "wrong_password" || response.Header().Get("Set-Cookie") != "" {
			t.Errorf("foreign tenant %s accepted the account or disclosed its existence", key)
		}
	}

	response = identityRequest(t, api.handler, identityBranchB, "POST", "/identity/registration", registration, "", nil)
	requireIdentityStatus(t, response, http.StatusCreated)
}

func TestIdentityV2DuplicateRegistrationIsAtomic(t *testing.T) {
	resetIdentity(t)
	var before, after int
	err := api.db.Pool.QueryRow(t.Context(), "select count(*) from users where tenant = $1", identityBranchA).Scan(&before)
	if err != nil {
		t.Fatal(err)
	}

	response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/registration",
		`{"email":"READER@EXAMPLE.TEST","display_name":"Duplicate Reader","password":"Independent-653!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusConflict)
	err = api.db.Pool.QueryRow(t.Context(), "select count(*) from users where tenant = $1", identityBranchA).Scan(&after)
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Errorf("duplicate registration left a user behind: before=%d after=%d", before, after)
	}
}

func TestIdentityRegistrationPreservesUTCInNonUTCSession(t *testing.T) {
	resetIdentity(t)
	config := api.db.AppPool.Config()
	config.ConnConfig.RuntimeParams["timezone"] = "Asia/Tokyo"
	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	var zone string
	if err := pool.QueryRow(t.Context(), "show timezone").Scan(&zone); err != nil {
		t.Fatal(err)
	}
	if zone != "Asia/Tokyo" {
		t.Fatalf("database session timezone=%q, want Asia/Tokyo", zone)
	}
	handler, _, err := newTestRouterWithLogger(
		t.Context(), pool, pool,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatal(err)
	}

	instant := time.Date(2026, time.September, 12, 12, 0, 0, 0, time.UTC)
	response := identityRequest(t, handler, identityBranchA, "POST", "/identity/registration",
		`{"email":"utc@example.test","display_name":"UTC Reader","password":"Independent-653!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusCreated)
	cookie := response.Result().Cookies()[0]

	response = identityRequest(t, handler, identityBranchA, "GET", "/identity/session", "", cookie.String(), nil)
	view := identityView(t, response)
	if view.Account == nil {
		t.Fatal("registered session has no account")
	}
	if !view.Account.CreatedAt.Equal(instant) {
		t.Errorf("session account created_at=%s, want %s", view.Account.CreatedAt, instant)
	}

	response = identityRequest(t, handler, identityBranchA, "POST", "/identity/login",
		`{"email":"utc@example.test","password":"Independent-653!"}`, "", nil)
	view = identityView(t, response)
	if view.Account == nil || !view.Account.CreatedAt.Equal(instant) {
		t.Errorf("login account=%+v, want created_at=%s", view.Account, instant)
	}
}
