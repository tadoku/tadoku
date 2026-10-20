package e2e_test

import (
	"io"
	"log/slog"
	"net/http"
	"testing"
)

func TestIdentityV2InfrastructureErrorBody(t *testing.T) {
	resetIdentity(t)
	pool := openUnreachablePool(t, api.db.AppPool.Config().ConnString())
	handler, _, _, err := newTestRouterWithLogger(
		t.Context(),
		pool,
		pool,
		api.keto,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatal(err)
	}
	response := identityRequest(t, handler, identityBranchA, "GET", "/identity/session", "",
		"tadoku_session_tadoku_identity-a-0123abcd=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", nil)
	requireIdentityStatus(t, response, http.StatusServiceUnavailable)
	if response.Body.String() != "{\"message\":\"identity service unavailable\"}\n" {
		t.Errorf("infrastructure error body=%q, want the sanitized identity error", response.Body.String())
	}
}

func TestIdentityV2NonJSONErrorBody(t *testing.T) {
	resetIdentity(t)
	response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"wrong-password"}`, "", func(request *http.Request) {
			request.Header.Set("Content-Type", "text/plain")
		})
	requireIdentityStatus(t, response, http.StatusBadRequest)
	if response.Body.String() != "{\"message\":\"identity requests must use application/json\"}\n" {
		t.Errorf("content-type error body=%q, want the identity error", response.Body.String())
	}
}
