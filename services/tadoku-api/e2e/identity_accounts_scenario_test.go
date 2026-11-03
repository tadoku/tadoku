package e2e_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

// Test safety: identityReaderID must match reader@example.test in testdata/tenant_accounts.sql.
const identityReaderID = "11111111-1111-4111-8111-111111111111"

var identityProductionCostSeed = tenantSeed{
	finalize: []string{"testdata/IdentityV2/scenario-production-password-cost.sql"},
}

// identityRequest is the serial identityRequest for a scenario role: it signs the request as the role's guest and
// serves it through handler on the scenario's clock.
func (s *scenario) identityRequest(
	t testing.TB,
	handler http.Handler,
	role, method, path, body, cookie string,
	edit func(*http.Request),
) *httptest.ResponseRecorder {
	t.Helper()
	request := identityHTTPRequest(method, path, body, cookie)
	if edit != nil {
		edit(request)
	}
	return s.serveWith(t, handler, role, guest, request)
}

func (s *scenario) identityLoginCookie(t testing.TB, role string) string {
	t.Helper()
	response := s.identityRequest(t, api.handler, role, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
	return response.Result().Cookies()[0].String()
}

func (s *scenario) registerFrom(
	t testing.TB,
	handler http.Handler,
	role, email, peer, forwarded string,
) *httptest.ResponseRecorder {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"display_name":"Budget Reader","password":"Independent-653!"}`, email)
	return s.identityRequest(t, handler, role, "POST", "/identity/registration", body, "", fromPeer(peer, forwarded))
}

// seedRecoveryToken stores a recovery token for the role's cast reader and returns its raw value.
func (s *scenario) seedRecoveryToken(
	t testing.TB,
	role string,
	marker byte,
	expires time.Time,
	used *time.Time,
) string {
	t.Helper()
	raw := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat(string(marker), 32)))
	digest := sha256.Sum256([]byte(raw))
	_, err := api.db.Pool.Exec(t.Context(), `insert into recovery_tokens
		(tenant, token_hash, user_id, expires_at, used_at)
		values ($1, $2, $3, $4, $5)`,
		s.tenant(t, role).String(), hex.EncodeToString(digest[:]), identityReaderID, expires, used)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func identitySessionCookieName(key tenant.TestKey) string {
	return "tadoku_session_" + strings.ReplaceAll(key.String(), "/", "_")
}
