package e2e_test

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestAccessTokenEmailRoutes(t *testing.T) {
	t.Parallel()
	s := newScenario(t, "tokens", "other")
	key := s.tenant(t, "tokens").Key()
	other := s.tenant(t, "other").Key()
	now := s.clock.Now()

	id := writeEmail(t, s.executionContext(tenant.WithKey(t.Context(), key)), emailService(), recoveryMail())
	otherID := writeEmail(t, s.executionContext(tenant.WithKey(t.Context(), other)), emailService(), recoveryMail())

	prefix := func(key tenant.Key) string { return "tdk_at_" + strings.ReplaceAll(key.String(), "/", "_") + "_" }
	raw := prefix(key) + strings.Repeat("a", 43)
	digest := sha256.Sum256([]byte(raw))
	tokenID := uuid.New()
	_, err := api.db.Pool.Exec(t.Context(), `insert into access_tokens(tenant,id,name,token_hash,scopes,expires_at)
 values ($1,$2,'test',$3,array['email:read'],$4::timestamptz+interval '1 hour')`,
		key.String(), tokenID, hex.EncodeToString(digest[:]), now)
	if err != nil {
		t.Fatal(err)
	}

	deployment, _ := tenant.ParseDeployment(key.String())
	handler, _, err := newTestRouterWithLeaderboardService(
		t.Context(), api.db.AppPool, api.db.AppPool,
		slog.New(slog.NewTextHandler(io.Discard, nil)), false, nil, deployment,
	)
	if err != nil {
		t.Fatal(err)
	}

	request := func(t *testing.T, h http.Handler, token, path string, status int) []byte {
		t.Helper()
		req := httptest.NewRequest("GET", path, nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Set("X-Dev-Branch", other.String())
		req = req.WithContext(s.executionContext(req.Context()))

		response := httptest.NewRecorder()

		h.ServeHTTP(response, req)

		if response.Code != status {
			t.Errorf("%s status=%d want=%d body=%s", path, response.Code, status, response.Body)
		}

		return response.Body.Bytes()
	}

	t.Run("list and detail", func(t *testing.T) {
		body := request(t, handler, raw, "/emails", 200)
		if !strings.Contains(string(body), id.String()) || strings.Contains(string(body), otherID.String()) {
			t.Error("token list did not isolate tenant email")
		}

		request(t, handler, raw, "/emails/"+id.String(), 200)
		request(t, handler, raw, "/emails/"+otherID.String(), 404)
	})

	t.Run("other deployment", func(t *testing.T) {
		dep, _ := tenant.ParseDeployment(other.String())
		h, _, err := newTestRouterWithLeaderboardService(
			t.Context(), api.db.AppPool, api.db.AppPool,
			slog.New(slog.NewTextHandler(io.Discard, nil)), false, nil, dep,
		)
		if err != nil {
			t.Fatal(err)
		}
		request(t, h, raw, "/emails", 421)
	})

	t.Run("outside email", func(t *testing.T) {
		for _, path := range []string{"/identity/session", "/authz/current-user/role", "/profile/users"} {
			request(t, handler, raw, path, 401)
		}
	})

	for _, tc := range []struct {
		name, sql string
		status    int
	}{
		{"expired", "update access_tokens set expires_at=@now::timestamptz-interval '1 second' where id=@id", 401},
		{
			"revoked",
			"update access_tokens set expires_at=@now::timestamptz+interval '1 hour',revoked_at=@now where id=@id",
			401,
		},
		{"wrong scope", "update access_tokens set revoked_at=null,scopes=array['other'] where id=@id", 401},
		{"extra scope", "update access_tokens set scopes=array['email:read','other'] where id=@id", 401},
		{"no scope", "update access_tokens set scopes=array[]::text[] where id=@id", 401},
		{"nonstandard bounds", "update access_tokens set scopes='[0:0]={email:read}'::text[] where id=@id", 401},
		{"two dimensions", "update access_tokens set scopes=array[['email:read']] where id=@id", 401},
		{"null scope", "update access_tokens set scopes=array[null]::text[] where id=@id", 401},
		{"restored scope", "update access_tokens set scopes=array['email:read'] where id=@id", 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			seeded := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			if _, err := api.db.Pool.Exec(t.Context(), tc.sql, pgx.NamedArgs{"id": tokenID, "now": now}); err != nil {
				t.Fatal(err)
			}
			_, err := api.db.Pool.Exec(t.Context(), "update access_tokens set last_used_at=$2 where id=$1", tokenID, seeded)
			if err != nil {
				t.Fatal(err)
			}

			body := request(t, handler, raw, "/emails", tc.status)

			if tc.status == 401 && len(body) != 0 {
				t.Errorf("rejected token body=%s, want empty", body)
			}

			var lastUsed time.Time
			err = api.db.Pool.QueryRow(t.Context(), "select last_used_at from access_tokens where id=$1", tokenID).
				Scan(&lastUsed)
			if err != nil {
				t.Fatal(err)
			}
			if used := !lastUsed.Equal(seeded); used != (tc.status == 200) {
				t.Errorf("last_used_at=%s after status %d, seeded %s", lastUsed, tc.status, seeded)
			}
		})
	}

	t.Run("unauthenticated", func(t *testing.T) { request(t, handler, "", "/emails", 400) })

	t.Run("forged prefix", func(t *testing.T) {
		request(t, api.handler, strings.Replace(raw, prefix(key), prefix(other), 1), "/emails", 401)
	})
}

func TestAccessTokenCanonicalDeploymentAndAudit(t *testing.T) {
	api.reset(t, "")
	key := tenant.Production()

	id := writeEmail(t, tenant.WithKey(t.Context(), key), emailService(), recoveryMail())
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}

	raw := "tdk_at_tadoku_prod_" + base64.RawURLEncoding.EncodeToString(secret)
	hash := sha256.Sum256([]byte(raw))
	tokenID := uuid.New()
	_, err := api.db.Pool.Exec(t.Context(), `insert into access_tokens(tenant,id,name,token_hash,scopes,expires_at)
  values ('tadoku/prod',$1,'canonical development',$2,array['email:read'],now()+interval '1 hour')`,
		tokenID, hex.EncodeToString(hash[:]))
	if err != nil {
		t.Fatal(err)
	}

	check := func(handler http.Handler, status int) []byte {
		t.Helper()
		request := httptest.NewRequest("GET", "/emails/"+id.String(), nil)
		request.Header.Set("Authorization", "Bearer "+raw)

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != status {
			t.Errorf("canonical email status=%d want=%d", response.Code, status)
		}

		return response.Body.Bytes()
	}

	check(api.handler, 401)

	development := func(endpoints *routerEndpoints) {
		value := false
		endpoints.production = &value
	}
	handler, _, err := newTestRouterWithLogger(t.Context(), api.db.AppPool, api.db.AppPool,
		slog.New(slog.NewTextHandler(io.Discard, nil)), development)
	if err != nil {
		t.Fatal(err)
	}

	check(handler, 200)

	var audits int
	err = api.db.Pool.QueryRow(t.Context(), `select count(*) from moderation_audit_log
  where tenant='tadoku/prod' and action='read_email' and user_id=$1
    and metadata->>'access_token_id'=$2`, tokenID, tokenID.String()).Scan(&audits)
	if err != nil || audits != 1 {
		t.Errorf("canonical token audit count=%d error=%v, want 1", audits, err)
	}

	closed := openClosedPool(t, api.db.AppPool.Config().ConnString())
	failed, _, err := newTestRouterWithLogger(t.Context(), api.db.AppPool, closed,
		slog.New(slog.NewTextHandler(io.Discard, nil)), development)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Contains(check(failed, 500), []byte(id.String())) {
		t.Error("failed canonical token audit leaked email content")
	}
}
