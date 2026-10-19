package e2e_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func emailTokens(t *testing.T) map[string]string {
	t.Helper()
	raw, err := os.ReadFile("testdata/journeys/cast.json")
	if err != nil {
		t.Fatal(err)
	}
	var tokens map[string]string
	if err := json.Unmarshal(raw, &tokens); err != nil {
		t.Fatal(err)
	}
	previous := jwt.TimeFunc
	jwt.TimeFunc = func() time.Time { return fixtureInstant }
	t.Cleanup(func() { jwt.TimeFunc = previous })
	return tokens
}

func TestEmailReadAuthorizationAuditAndPagination(t *testing.T) {
	api.reset(t, "")
	if err := api.keto.Reset(t.Context(),
		"testdata/journeys/relationships.json",
		"testdata/journeys/TenantIsolation/relationships.json"); err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"EmailList", "EmailGet"} {
		for _, entry := range []struct {
			status int
			who    string
		}{{401, "guest"}, {403, "user"}} {
			name := APITestName(operation, entry.status, entry.who)
			runCase(t, api, name, entry.status, implementation{name: "tadoku-api", handler: api.handler})
		}
	}
	if err := api.keto.Reset(
		t.Context(),
		"testdata/journeys/relationships.json",
		"testdata/journeys/TenantIsolation/relationships.json",
	); err != nil {
		t.Fatal(err)
	}
	timex.TheWorld(fixtureInstant, func() {
		emailReadAuthorizationAuditAndPagination(t)
	})
}

func emailReadAuthorizationAuditAndPagination(t *testing.T) {
	ctx := tenant.WithKey(t.Context(), tenant.Production())
	tokens := emailTokens(t)

	first := writeEmail(t, ctx, emailService(), recoveryMail())
	second := writeEmail(t, ctx, emailService(), recoveryMail())
	for _, path := range []string{"/emails", "/emails/" + first.String()} {
		for _, tc := range []struct {
			who    string
			status int
		}{{"guest", 401}, {"user", 403}, {"banned", 403}} {
			emailRequest(t, tokens[tc.who], http.MethodGet, path, "", tc.status)
		}
	}
	raw := emailRequest(t,
		tokens["admin"],
		http.MethodGet,
		"/emails?limit=1&recipient=reader%40example.test&template=account.recovery&status=pending&since="+url.QueryEscape(fixtureInstant.Format(time.RFC3339)),
		"",
		200)
	var page struct {
		Emails []struct {
			ID string `json:"id"`
		}
		NextCursor string `json:"next_cursor"`
	}
	if err := json.Unmarshal(raw, &page); err != nil || len(page.Emails) != 1 || page.NextCursor == "" {
		t.Fatalf("page=%s err=%v", raw, err)
	}
	raw = emailRequest(t,
		tokens["admin"],
		http.MethodGet,
		"/emails?limit=1&cursor="+url.QueryEscape(page.NextCursor),
		"",
		200)
	var next struct {
		Emails []struct {
			ID string `json:"id"`
		}
	}
	if err := json.Unmarshal(raw, &next); err != nil || len(next.Emails) != 1 || next.Emails[0].ID == page.Emails[0].ID {
		t.Fatalf("next=%s err=%v", raw, err)
	}
	emailRequest(t, tokens["admin"], http.MethodGet, "/emails/"+second.String(), "", 200)
	for _, filter := range []string{"limit=101", "limit=-1", "cursor=invalid", "status=bad"} {
		emailRequest(t, tokens["admin"], http.MethodGet, "/emails?"+filter, "", 400)
	}
	var count int
	if err := api.db.Pool.QueryRow(ctx,
		"select count(*) from moderation_audit_log where action='read_email'").Scan(&count); err != nil || count != 3 {
		t.Fatalf("audit count=%d err=%v", count, err)
	}
	var unauthorized int
	if err := api.db.Pool.QueryRow(ctx,
		"select count(*) from moderation_audit_log where user_id <> '22222222-2222-4222-8222-222222222222'").Scan(&unauthorized); err != nil || unauthorized != 0 {
		t.Fatalf("unauthorized audits=%d err=%v", unauthorized, err)
	}
	key, _ := tenant.Parse("e2e/alpha-0000000a")
	if _, err := api.db.Pool.Exec(ctx,
		"insert into tenants(key,kind) values ($1,'test') on conflict do nothing",
		key.String()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := api.db.Pool.Exec(cleanup, "delete from tenants where key=$1", key.String()); err != nil {
			t.Error(err)
		}
	})
	emailRequest(t, tokens["alpha_admin"], http.MethodGet, "/emails/"+first.String(), "", 404)
	emailRequest(t, tokens["alpha_admin"], http.MethodGet, "/emails?cursor="+url.QueryEscape(page.NextCursor), "", 400)
	testCtx := tenant.WithKey(t.Context(), key)
	testID := writeEmail(t, testCtx, emailService(), recoveryMail())
	emailRequest(t, tokens["alpha_admin"], http.MethodGet, "/emails/"+testID.String(), "", 200)
	if err := api.db.Pool.QueryRow(ctx,
		"select count(*) from moderation_audit_log where tenant=$1",
		key.String()).Scan(&count); err != nil || count != 0 {
		t.Fatalf("test audit count=%d err=%v", count, err)
	}
	for _, filter := range []string{"recipient=nobody%40example.test",
		"template=unknown",
		"status=sent",
		"since=2099-01-01T00%3A00%3A00Z"} {
		raw = emailRequest(t, tokens["admin"], http.MethodGet, "/emails?"+filter, "", 200)
		var empty struct{ Emails []email.Email }
		if err := json.Unmarshal(raw, &empty); err != nil || len(empty.Emails) != 0 {
			t.Fatalf("filter %s response=%s err=%v", filter, raw, err)
		}
	}
}

func TestEmailRetention(t *testing.T) {
	api.reset(t, "")
	ctx := tenant.WithKey(t.Context(), tenant.Production())
	now := fixtureInstant
	timex.TheWorld(now, func() {
		emailRetention(t, ctx, now)
	})
}

func emailRetention(t *testing.T, ctx context.Context, now time.Time) {
	key, _ := tenant.Parse("e2e/email-retention")
	if _, err := api.db.Pool.Exec(ctx, "insert into tenants(key,kind) values ($1,'test')", key.String()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := api.db.Pool.Exec(cleanup, "delete from tenants where key=$1", key.String()); err != nil {
			t.Error(err)
		}
	})
	for _, tc := range []struct {
		key  tenant.Key
		days int
	}{{tenant.Production(), 3}, {key, 14}} {
		scoped := tenant.WithKey(t.Context(), tc.key)
		for _, age := range []time.Duration{time.Duration(tc.days)*24*time.Hour + time.Microsecond,
			time.Duration(tc.days) * 24 * time.Hour,
			time.Hour} {
			id := writeEmail(t, scoped, emailService(), recoveryMail())
			if _, err := api.db.Pool.Exec(ctx,
				"update emails set created_at=$1 where tenant=$2 and id=$3",
				now.Add(-age),
				tc.key.String(),
				id); err != nil {
				t.Fatal(err)
			}
		}
		if err := emailService().Retain(scoped); err != nil {
			t.Fatal(err)
		}
		var count int
		if err := api.db.Pool.QueryRow(ctx,
			"select count(*) from emails where tenant=$1",
			tc.key.String()).Scan(&count); err != nil || count != 2 {
			t.Fatalf("tenant=%s count=%d err=%v", tc.key, count, err)
		}
	}
	if _, err := api.db.Pool.Exec(ctx, "delete from tenants where key=$1", key.String()); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := api.db.Pool.QueryRow(ctx,
		"select count(*) from emails where tenant=$1",
		key.String()).Scan(&count); err != nil || count != 0 {
		t.Fatalf("cascade count=%d err=%v", count, err)
	}
}

func TestEmailAuditFailureHidesContent(t *testing.T) {
	api.reset(t, "")
	if err := api.keto.Reset(t.Context(),
		"testdata/journeys/relationships.json",
		"testdata/journeys/TenantIsolation/relationships.json"); err != nil {
		t.Fatal(err)
	}
	tokens := emailTokens(t)
	ctx := tenant.WithKey(t.Context(), tenant.Production())
	id := writeEmail(t, ctx, emailService(), recoveryMail())
	closed := openClosedPool(t, api.db.AppPool.Config().ConnString())
	handler, _, _, err := newTestRouterWithLogger(ctx,
		api.db.AppPool,
		closed,
		api.keto,
		api.kratos,
		slog.New(slog.NewTextHandler(io.Discard,
			nil)))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/emails", "/emails/" + id.String()} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("Authorization", "Bearer "+tokens["admin"])
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != 500 || strings.Contains(response.Body.String(), "recovery_url") {
			t.Errorf("audit failure response=%d %s", response.Code, response.Body.String())
		}
	}
}

func emailRequest(t *testing.T, token, method, path, body string, want int) []byte {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	api.handler.ServeHTTP(response, request)
	if response.Code != want {
		t.Fatalf("%s status=%d want=%d body=%s", path, response.Code, want, response.Body.String())
	}
	return response.Body.Bytes()
}
