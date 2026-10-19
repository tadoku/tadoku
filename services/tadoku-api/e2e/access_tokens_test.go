package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestAccessTokenEmailRoutes(t *testing.T) {
	api.reset(t, "")
	key, _ := tenant.Parse("e2e/tokens-0123abcd")
	other, _ := tenant.Parse("e2e/other-0123abcd")
	if _, err := api.db.Pool.Exec(t.Context(), "insert into tenants(key,kind) values ($1,'test'),($2,'test')", key.String(), other.String()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := api.db.Pool.Exec(ctx, "delete from tenants where key=any($1)", []string{key.String(), other.String()}); err != nil {
			t.Error(err)
		}
	})
	id := writeEmail(t, tenant.WithKey(t.Context(), key), emailService(), recoveryMail())
	otherID := writeEmail(t, tenant.WithKey(t.Context(), other), emailService(), recoveryMail())
	raw := "tdk_at_e2e_tokens-0123abcd_" + strings.Repeat("a", 43)
	digest := sha256.Sum256([]byte(raw))
	tokenID := uuid.New()
	if _, err := api.db.Pool.Exec(t.Context(), `insert into access_tokens(tenant,id,name,token_hash,scopes,expires_at)
  values ($1,$2,'test',$3,array['email:read'],now()+interval '1 hour')`, key.String(), tokenID, hex.EncodeToString(digest[:])); err != nil {
		t.Fatal(err)
	}
	deployment, _ := tenant.ParseDeployment(key.String())
	handler, _, _, err := newTestRouterWithLeaderboardService(t.Context(), api.db.AppPool, api.db.AppPool, api.keto, api.kratos, slog.New(slog.NewTextHandler(io.Discard, nil)), false, nil, deployment)
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
		h, _, _, err := newTestRouterWithLeaderboardService(t.Context(), api.db.AppPool, api.db.AppPool, api.keto, api.kratos, slog.New(slog.NewTextHandler(io.Discard, nil)), false, nil, dep)
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
	for _, tc := range []struct{ name, sql string }{
		{"expired", "update access_tokens set expires_at=now()-interval '1 second' where id=$1"},
		{"revoked", "update access_tokens set expires_at=now()+interval '1 hour',revoked_at=now() where id=$1"},
		{"wrong scope", "update access_tokens set revoked_at=null,scopes=array['other'] where id=$1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := api.db.Pool.Exec(t.Context(), tc.sql, tokenID); err != nil {
				t.Fatal(err)
			}
			request(t, handler, raw, "/emails", 401)
		})
	}
	t.Run("unauthenticated", func(t *testing.T) { request(t, handler, "", "/emails", 400) })
	t.Run("forged prefix", func(t *testing.T) {
		request(t, api.handler, strings.Replace(raw, "e2e_tokens", "e2e_other", 1), "/emails", 401)
	})
}
