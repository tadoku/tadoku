package e2e_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	accountidentity "github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func seedIdentityReads(t *testing.T) {
	t.Helper()
	api.reset(t, "")
	if err := api.db.Reset(t.Context()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, err := api.db.Pool.Exec(context.Background(), `
			delete from tenants where key = 'e2e/read-0123abcd';
			delete from accounts
			where tenant = 'tadoku/prod'
			  and user_id in (
			    '11111111-1111-4111-8111-111111111111',
			    '22222222-2222-4222-8222-222222222222'
			  )`)
		if err != nil {
			t.Error(err)
		}
	})
	_, err := api.db.Pool.Exec(t.Context(), `
 insert into tenants (key, kind) values ('e2e/read-0123abcd', 'test');
 insert into users (tenant, id, display_name, created_at, updated_at)
 values
 ('tadoku/prod', '11111111-1111-4111-8111-111111111111', 'Reader One', '2020-01-01', '2020-01-01'),
 ('tadoku/prod', '22222222-2222-4222-8222-222222222222', 'Admin One', '2020-01-01', '2020-01-01'),
 ('e2e/read-0123abcd', '11111111-1111-4111-8111-111111111111', 'Branch Reader', '2020-01-01', '2020-01-01'),
 ('e2e/read-0123abcd', '99999999-9999-4999-8999-999999999999', 'Foreign Reader', '2020-01-01', '2020-01-01');
 insert into accounts (tenant, user_id, email, email_normalized, password_hash, created_at)
 values
 ('tadoku/prod', '11111111-1111-4111-8111-111111111111',
  'local@example.test', 'local@example.test', 'unused', '2020-02-03'),
 ('tadoku/prod', '22222222-2222-4222-8222-222222222222',
  'admin@example.test', 'admin@example.test', 'unused', '2020-02-03'),
 ('e2e/read-0123abcd', '11111111-1111-4111-8111-111111111111',
  'branch@example.test', 'branch@example.test', 'unused', '2021-02-03'),
 ('e2e/read-0123abcd', '99999999-9999-4999-8999-999999999999',
  'foreign@example.test', 'foreign@example.test', 'unused', '2021-02-03');
 `)
	if err != nil {
		t.Fatal(err)
	}
	if err := api.keto.Reset(t.Context(), "testdata/ProfileUsersList/relationships.json"); err != nil {
		t.Fatal(err)
	}
}

func TestIdentityReaderTenantPaging(t *testing.T) {
	seedIdentityReads(t)
	var reader accountidentity.Reader = accountidentity.NewReader(api.db.AppPool)
	ctx := tenant.WithKey(t.Context(), tenant.Production())
	first, err := reader.ListAccounts(ctx, accountidentity.ListAccounts{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Accounts) != 1 ||
		first.Accounts[0].ID.String() != "11111111-1111-4111-8111-111111111111" ||
		first.NextCursor == "" {
		t.Fatalf("first page=%+v", first)
	}
	second, err := reader.ListAccounts(ctx, accountidentity.ListAccounts{Limit: 1, Cursor: first.NextCursor})
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Accounts) != 1 || second.Accounts[0].ID == first.Accounts[0].ID || second.NextCursor != "" {
		t.Fatalf("second page=%+v", second)
	}
	for _, query := range []string{"LOCAL@", "reader"} {
		page, err := reader.ListAccounts(ctx, accountidentity.ListAccounts{Query: query})
		if err != nil || len(page.Accounts) != 1 || page.Accounts[0].ID != first.Accounts[0].ID {
			t.Fatalf("query=%q page=%+v err=%v", query, page, err)
		}
	}
	page, err := reader.ListAccounts(ctx, accountidentity.ListAccounts{Query: "%"})
	if err != nil || len(page.Accounts) != 0 {
		t.Fatalf("literal wildcard page=%+v err=%v", page, err)
	}
	_, err = reader.ListAccounts(ctx, accountidentity.ListAccounts{Cursor: "broken"})
	if errx.KindOf(err) != errx.InvalidInput {
		t.Fatalf("invalid cursor error=%v", err)
	}
	exists, err := reader.UserExists(ctx, uuid.MustParse("99999999-9999-4999-8999-999999999999"))
	if err != nil || exists {
		t.Fatalf("foreign account exists=%v err=%v", exists, err)
	}
	branch, _ := tenant.Parse("e2e/read-0123abcd")
	for _, invalid := range []struct {
		ctx   context.Context
		query accountidentity.ListAccounts
	}{
		{tenant.WithKey(t.Context(), branch), accountidentity.ListAccounts{Cursor: first.NextCursor}},
		{ctx, accountidentity.ListAccounts{Query: "Reader", Cursor: first.NextCursor}},
	} {
		_, err := reader.ListAccounts(invalid.ctx, invalid.query)
		if errx.KindOf(err) != errx.InvalidInput {
			t.Fatalf("cursor must bind tenant and search: %v", err)
		}
	}
	account, err := reader.FindAccount(tenant.WithKey(t.Context(), branch), first.Accounts[0].ID)
	if err != nil || account.Email.String() != "branch@example.test" || account.DisplayName.String() != "Branch Reader" {
		t.Fatalf("branch account=%+v err=%v", account, err)
	}
	_, err = api.db.Pool.Exec(t.Context(), `
		update users
		set deleted_at = '2026-01-01'
		where tenant = 'tadoku/prod'
		  and id = '22222222-2222-4222-8222-222222222222'`)
	if err != nil {
		t.Fatal(err)
	}
	deletedID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	exists, err = reader.UserExists(ctx, deletedID)
	if err != nil || exists {
		t.Fatalf("deleted user exists=%v err=%v", exists, err)
	}
	_, err = reader.FindAccount(ctx, deletedID)
	if errx.KindOf(err) != errx.NotFound {
		t.Fatalf("deleted account lookup=%v", err)
	}
	remaining, err := reader.ListAccounts(ctx, accountidentity.ListAccounts{})
	if err != nil || len(remaining.Accounts) != 1 || remaining.Accounts[0].ID != first.Accounts[0].ID {
		t.Fatalf("deleted user listed: %+v err=%v", remaining, err)
	}
	_, err = api.db.Pool.Exec(t.Context(), `
		update accounts
		set state = 'inactive'
		where tenant = 'tadoku/prod'
		  and user_id = '11111111-1111-4111-8111-111111111111'`)
	if err != nil {
		t.Fatal(err)
	}
	exists, err = reader.UserExists(ctx, first.Accounts[0].ID)
	if err != nil || exists {
		t.Fatalf("inactive account exists=%v err=%v", exists, err)
	}
}

func TestIdentityCallersUseTenantAccounts(t *testing.T) {
	seedIdentityReads(t)
	source := readHTTPRequest(t, filepath.Join("testdata", "ImmersionProfileFindByUserID", "200_user_populated"))
	request := httptest.NewRequest(http.MethodGet, source.URL.String(), nil)
	request.Header = source.Header.Clone()
	response := httptest.NewRecorder()
	atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })
	var profile struct {
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &profile); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || profile.CreatedAt != "2020-02-03T00:00:00Z" {
		t.Fatalf("profile=%d %s", response.Code, response.Body)
	}
	admin := readHTTPRequest(t, filepath.Join("testdata", "ProfileUsersList", "200_admin"))
	for _, test := range []struct {
		path, method, body string
		status             int
	}{
		{"/profile/users", "GET", "", http.StatusOK},
		{
			"/authz/users/55555555-5555-4555-8555-555555555555/role",
			"PUT",
			`{"role":"banned","reason":"not an account in this tenant"}`,
			http.StatusNotFound,
		},
	} {
		request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		request.Header = admin.Header.Clone()
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })
		if response.Code != test.status {
			t.Errorf("%s=%d %s", test.path, response.Code, response.Body)
		}
		if test.path == "/profile/users" {
			var result struct {
				Total int `json:"total_size"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || result.Total != 2 {
				t.Fatalf("local list=%s err=%v", response.Body, err)
			}
		}
	}
}
