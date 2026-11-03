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

// seedRecoveryToken stores a recovery token for the cast reader of the role's tenant and returns its raw value.
func (s *scenario) seedRecoveryToken(
	t *testing.T,
	role string,
	marker byte,
	expires time.Time,
	used *time.Time,
) string {
	t.Helper()
	raw := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat(string(marker), 32)))
	digest := sha256.Sum256([]byte(raw))
	writeTenantRows(t, s.tenant(t, role).String(), `insert into recovery_tokens
		(token_hash, user_id, expires_at, used_at)
		values ($1, $2, $3, $4)`,
		hex.EncodeToString(digest[:]), scenarioCast[user].subject, expires, used)
	return raw
}

func identitySessionCookieName(key tenant.TestKey) string {
	return "tadoku_session_" + strings.ReplaceAll(key.String(), "/", "_")
}
