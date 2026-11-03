package e2e_test

import (
	"io"
	"log/slog"
	"net/http"
	"testing"
)

func TestIdentityV2InfrastructureErrorBody(t *testing.T) {
	t.Parallel()
	s := newIdentityScenario(t)

	pool := openUnreachablePool(t, api.db.AppPool.Config().ConnString())
	handler, _, err := newTestRouterWithLogger(
		t.Context(),
		pool,
		pool,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatal(err)
	}
	response := s.identityRequest(t, handler, "identity-a", "GET", "/identity/session", "",
		identitySessionCookieName(s.tenant(t, "identity-a"))+"=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", nil)
	requireIdentityStatus(t, response, http.StatusServiceUnavailable)
	if response.Body.String() != "{\"message\":\"identity service unavailable\"}\n" {
		t.Errorf("infrastructure error body=%q, want the sanitized identity error", response.Body.String())
	}
}

func TestIdentityV2NonJSONErrorBody(t *testing.T) {
	t.Parallel()
	s := newIdentityScenario(t)

	response := s.identityRequest(t, api.handler, "identity-a", "POST", "/identity/login",
		`{"email":"reader@example.test","password":"wrong-password"}`, "", func(request *http.Request) {
			request.Header.Set("Content-Type", "text/plain")
		})
	requireIdentityStatus(t, response, http.StatusBadRequest)
	if response.Body.String() != "{\"message\":\"identity requests must use application/json\"}\n" {
		t.Errorf("content-type error body=%q, want the identity error", response.Body.String())
	}
}
