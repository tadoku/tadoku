package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

// newPairedIdentityScenario provisions identity-a and identity-b tenants, each holding the account cast.
func newPairedIdentityScenario(t *testing.T) *scenario {
	t.Helper()
	s := newScenario(t, "identity-a", "identity-b")
	s.seed(t, "identity-a", tenantSeed{})
	s.seed(t, "identity-b", tenantSeed{})
	return s
}

// guestIdentityRequest sends an identity request as a guest of the scenario role's tenant.
func guestIdentityRequest(
	t *testing.T,
	s *scenario,
	role, method, path, body, cookie string,
	edit func(*http.Request),
) *httptest.ResponseRecorder {
	t.Helper()
	return s.serve(t, role, guest, newIdentityRequest(t, "", method, path, body, cookie, edit))
}

func recoverFromRole(t *testing.T, s *scenario, role, email, peer, forwarded string) *httptest.ResponseRecorder {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"return_to":%q}`, email, recoverySettingsURL)
	return guestIdentityRequest(t, s, role, "POST", "/identity/recovery", body, "", fromPeer(peer, forwarded))
}

func loginCookieIn(t *testing.T, s *scenario, role string) string {
	t.Helper()
	response := guestIdentityRequest(t, s, role, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
	return response.Result().Cookies()[0].String()
}

// writeTenantRows runs sql in tenantKey: a test tenant through the confined seeding login, the canonical tenant
// through the owner pool for serial tests only.
func writeTenantRows(t *testing.T, tenantKey, sql string, arguments ...any) {
	t.Helper()
	if tenantKey == tenant.Production().String() {
		if _, err := api.db.Pool.Exec(t.Context(), sql, arguments...); err != nil {
			t.Fatal(err)
		}
		return
	}

	key, err := tenant.ParseTestTenant(tenantKey)
	if err != nil {
		t.Fatal(err)
	}
	err = api.db.SeedTenant(t.Context(), key, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sql, arguments...)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

// seedRecoveryTokenIn stores a recovery token for the cast reader of the scenario role's tenant.
func seedRecoveryTokenIn(
	t *testing.T,
	s *scenario,
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

// recoveryLocation is the Settings URL a consumed recovery link of the scenario role's tenant opens.
func recoveryLocation(t *testing.T, s *scenario, role string) string {
	t.Helper()
	_, route, _ := strings.Cut(s.tenant(t, role).String(), "/")
	return "https://" + route + ".account.tadoku.dev.lab/?edit=password"
}

func serveAsyncIn(t *testing.T, s *scenario, role string, request *http.Request) <-chan *httptest.ResponseRecorder {
	t.Helper()
	served := make(chan *httptest.ResponseRecorder, 1)
	go func() { served <- s.serve(t, role, guest, request) }()
	return served
}

// publishedJWKS returns the local JWKS, including the scenario signing key, for a private router's JWKS server.
// Call it outside any synctest bubble.
func publishedJWKS(t *testing.T) []byte {
	t.Helper()
	response, err := http.Get(authenticationJWKS.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	jwks, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return jwks
}
