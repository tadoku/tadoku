package e2e_test

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

func TestIdentityV2IPThrottleAndFailOpen(t *testing.T) {
	resetIdentity(t)
	for i := 0; i < 51; i++ {
		body := fmt.Sprintf(`{"email":"unknown%d@example.test","password":"Independent-653!"}`, i)
		response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login", body, "",
			func(r *http.Request) {
				r.RemoteAddr = fmt.Sprintf("10.42.0.%d:12345", i+1)
				r.Header.Set("X-Real-IP", "198.51.100.30")
			})
		want := http.StatusUnauthorized
		if i == 50 {
			want = http.StatusTooManyRequests
		}
		requireIdentityStatus(t, response, want)
	}
	for i := 0; i < 51; i++ {
		body := fmt.Sprintf(`{"email":"unknown%d@example.test","password":"Independent-653!"}`, i)
		response := identityRequest(t, api.handler, identityBranchB, "POST", "/identity/login", body, "",
			func(r *http.Request) { r.Header.Set("X-Real-IP", fmt.Sprintf("198.51.100.%d", i+1)) })
		want := http.StatusUnauthorized
		if i == 50 {
			want = http.StatusTooManyRequests
		}
		requireIdentityStatus(t, response, want)
	}

	closed, err := newClosedLeaderboardValkeyClient()
	if err != nil {
		t.Fatal(err)
	}
	handler, _, _, err := newTestRouterWithLogger(
		t.Context(), api.db.AppPool, api.db.AppPool, api.keto,
		slog.New(slog.NewTextHandler(io.Discard, nil)), withIdentityValkey(closed),
	)
	if err != nil {
		t.Fatal(err)
	}
	response := identityRequest(t, handler, identityBranchA, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
}
