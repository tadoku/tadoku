package e2e_test

import (
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestIdentitySettingsLeavesPoolAvailableDuringEvaluation(t *testing.T) {
	resetIdentity(t)
	cookie := identityLoginCookie(t)
	config := api.db.AppPool.Config()
	config.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	observer := &waitingIdentityObserver{entered: make(chan struct{}), release: make(chan struct{})}
	handler, _, _, err := newTestRouterWithLogger(
		t.Context(), pool, pool, api.keto,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		func(endpoints *routerEndpoints) { endpoints.identityObserver = observer },
	)
	if err != nil {
		t.Fatal(err)
	}
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		response := identityRequest(t, handler, identityBranchA, "PUT", "/identity/settings/password",
			`{"password":"Replacement-839!"}`, cookie, nil)
		if response.Code != http.StatusNoContent {
			t.Errorf("password status=%d", response.Code)
		}
	}()
	defer func() { close(observer.release); <-finished }()
	select {
	case <-observer.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("settings did not reach evaluation")
	}
	if held := pool.Stat().AcquiredConns(); held != 0 {
		t.Errorf("settings holds %d database connections while evaluation waits, want 0", held)
	}
}

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
	for _, v1 := range []bool{true, false} {
		key, _ := tenant.Parse(identityBranchA)
		if err := flipt.SetIdentityV2(key, !v1); err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct{ method, path, body string }{
			{"PUT", "/identity/settings/password", `{"password":"Replacement-839!"}`},
			{"PATCH", "/identity/settings/profile", `{"email":"changed@example.test"}`},
		} {
			response := identityRequest(t, handler, identityBranchA, tc.method, tc.path, tc.body, "", nil)
			want := http.StatusUnauthorized
			if v1 {
				want = http.StatusNotImplemented
			}
			if response.Code != want {
				t.Errorf("v1=%t %s: status=%d want=%d", v1, tc.path, response.Code, want)
			}
		}
	}
}
