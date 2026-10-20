package e2e_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"golang.org/x/crypto/bcrypt"
)

const importedAccountID = "dddddddd-dddd-4ddd-8ddd-dddddddddddd"
const importedPassword = "Imported-fixture-317!"

var importedCreatedAt = time.Date(2019, 1, 2, 9, 34, 5, 123456000, time.UTC)

type importedEmailCase struct {
	name string
	// raw is the source Kratos trait. stored, when set, replaces the imported address afterwards to model a
	// stored address the importer itself refuses, such as a v1 trait.
	raw, stored string
	login       string
	wire        string
}

func TestIdentityImportedAccountEmailResponses(t *testing.T) {
	kelvin := strings.Repeat("a", 316) + "K@x"
	kelvinLogin := strings.Repeat("a", 316) + "k@x"
	dottedI := strings.Repeat("a", 317) + "İ@x"
	dottedILogin := strings.Repeat("a", 317) + "i@x"
	longPrefix := strings.Repeat("　", 350) + "imported@example.test"
	cases := []importedEmailCase{
		{name: "ordinary", raw: "imported@example.test", login: "imported@example.test", wire: "imported@example.test"},
		{name: "mixed case", raw: "Imported@Example.Test", login: "imported@example.test", wire: "Imported@Example.Test"},
		{
			name:   "quoted local part",
			raw:    "imported@example.test",
			stored: `"imported"@example.test`,
			login:  "imported@example.test",
			wire:   "imported@example.test",
		},
		{
			name:   "leading space",
			raw:    "imported@example.test",
			stored: " imported@example.test",
			login:  "imported@example.test",
			wire:   "imported@example.test",
		},
		{
			name:   "trailing tab",
			raw:    "imported@example.test",
			stored: "imported@example.test\t",
			login:  "imported@example.test",
			wire:   "imported@example.test",
		},
		{
			name:  "ideographic space prefix",
			raw:   "　imported@example.test",
			login: "imported@example.test",
			wire:  "　imported@example.test",
		},
		{name: "kelvin sign", raw: kelvinLogin, stored: kelvin, login: kelvinLogin, wire: kelvin},
		{name: "dotted capital I", raw: dottedILogin, stored: dottedI, login: dottedILogin, wire: dottedI},
		{
			name:   "long ideographic prefix",
			raw:    "imported@example.test",
			stored: longPrefix,
			login:  "imported@example.test",
			wire:   longPrefix,
		},
		{
			name:   "leading vertical tab",
			raw:    "imported@example.test",
			stored: "\vimported@example.test",
			login:  "imported@example.test",
			wire:   "\vimported@example.test",
		},
	}
	source := newKratosSource(t)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			importAccount(t, source, tc)

			login := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login",
				`{"email":`+strconv.Quote(tc.login)+`,"password":"`+importedPassword+`"}`, "", nil)
			checkImportedAccount(t, "login", login, tc.wire, "Imported Reader")
			cookies := login.Result().Cookies()
			if len(cookies) != 1 || cookies[0].Name != "tadoku_session" {
				t.Errorf("login set cookies %d, want one production session cookie", len(cookies))
			}
			if sessions := importedSessions(t); sessions != 1 {
				t.Errorf("login stored %d sessions, want 1", sessions)
			}

			cookie := insertImportedSession(t)
			session := identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
			checkImportedAccount(t, "session", session, tc.wire, "Imported Reader")
			checkImportedCallback(t, cookie, tc.wire)

			foreign := identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", cookie, nil)
			if view := identityView(t, foreign); view.Account != nil {
				t.Error("imported production cookie authenticated a branch tenant")
			}

			profile := identityRequest(t, api.handler, "tadoku/prod", "PATCH", "/identity/settings/profile",
				`{"display_name":"Renamed Import"}`, cookie, nil)
			checkImportedAccount(t, "profile", profile, tc.wire, "Renamed Import")
			checkImportedRow(t, tc, "Renamed Import")
		})
	}
}

func newKratosSource(t *testing.T) *testpostgres.Database {
	t.Helper()

	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if db.Close() != nil {
			t.Error("dispose Kratos source database failed")
		}
	})

	path, err := runfiles.Rlocation("tadoku/services/tadoku-api/cmd/tadoku-identity-import/testdata/kratos-source.sql")
	if err != nil {
		t.Fatal(err)
	}
	schema, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	role := db.AppPool.Config().ConnConfig.User
	for _, statement := range []string{
		"drop schema public cascade; create schema public",
		`grant usage on schema public to "` + role + `"`,
		string(schema),
		`grant select on all tables in schema public to "` + role + `"`,
		"insert into networks (id, created_at, updated_at) values ('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee', now(), now())",
	} {
		if _, err := db.Pool.Exec(t.Context(), statement); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func importAccount(t *testing.T, source *testpostgres.Database, tc importedEmailCase) {
	t.Helper()

	resetIdentity(t)
	_, err := api.db.Pool.Exec(t.Context(), "delete from accounts where tenant = 'tadoku/prod'")
	if err != nil {
		t.Fatal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(importedPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	_, err = source.Pool.Exec(t.Context(), "delete from identity_credentials; delete from identities")
	if err != nil {
		t.Fatal(err)
	}
	_, err = source.Pool.Exec(t.Context(), `insert into identities (id, nid, schema_id, traits, state, created_at, updated_at)
		select $1, id, 'user', jsonb_build_object('email', $2::text, 'display_name', 'Imported Reader'),
		'active', $3, $3 from networks`,
		importedAccountID, tc.raw, importedCreatedAt)
	if err != nil {
		t.Fatal(err)
	}
	_, err = source.Pool.Exec(t.Context(), `insert into identity_credentials
		(id, nid, identity_id, identity_credential_type_id, config, created_at, updated_at)
		select 'ffffffff-ffff-4fff-8fff-ffffffffffff', i.nid, i.id, ct.id,
		jsonb_build_object('hashed_password', $2::text), i.created_at, i.updated_at
		from identities i
		cross join identity_credential_types ct
		where i.id = $1
		  and ct.name = 'password'`,
		importedAccountID, string(hash))
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	err = identity.ImportKratosAccounts(t.Context(), source.AppPool, api.db.AppPool, 1, true, &output)
	if err != nil {
		t.Fatalf("import rejected the source account: %v\n%s", err, output.String())
	}
	checkImportedRow(t, importedEmailCase{raw: tc.raw, login: tc.login}, "Imported Reader")
	if tc.stored != "" {
		_, err = api.db.Pool.Exec(t.Context(), `update accounts set email = $2
			where tenant = 'tadoku/prod' and user_id = $1`, importedAccountID, tc.stored)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func checkImportedRow(t *testing.T, tc importedEmailCase, displayName string) {
	t.Helper()

	want := tc.raw
	if tc.stored != "" {
		want = tc.stored
	}
	var email, normalized, hash, name string
	var created time.Time
	err := api.db.Pool.QueryRow(t.Context(), `select a.email, a.email_normalized, a.password_hash, a.created_at,
		u.display_name
		from accounts a
		join users u on u.tenant = a.tenant and u.id = a.user_id
		where a.tenant = 'tadoku/prod' and a.user_id = $1`, importedAccountID,
	).Scan(&email, &normalized, &hash, &created, &name)
	if err != nil {
		t.Fatal(err)
	}
	if email != want || normalized != tc.login || !created.Equal(importedCreatedAt) || name != displayName {
		t.Errorf(
			"stored account email=%q normalized=%q created=%s display_name=%q",
			email,
			normalized,
			created,
			name,
		)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(importedPassword)) != nil {
		t.Error("stored account lost its imported password hash")
	}
}

func checkImportedAccount(
	t *testing.T,
	operation string,
	response *httptest.ResponseRecorder,
	wire, displayName string,
) {
	t.Helper()

	if response.Code != http.StatusOK {
		t.Errorf("%s status=%d body=%q, want 200", operation, response.Code, response.Body)
		return
	}
	if response.Header().Get("Cache-Control") != "private, no-store" {
		t.Errorf("%s account response may be cached", operation)
	}

	body := response.Body.Bytes()
	if operation != "profile" {
		var view struct {
			AccountApp string          `json:"account_app"`
			Account    json.RawMessage `json:"account"`
		}
		if err := json.Unmarshal(body, &view); err != nil || view.AccountApp != "v2" {
			t.Errorf("%s body=%q is not a v2 session", operation, response.Body)
			return
		}
		body = view.Account
	}
	var account struct {
		ID          string    `json:"id"`
		Email       *string   `json:"email"`
		DisplayName string    `json:"display_name"`
		CreatedAt   time.Time `json:"created_at"`
	}
	if err := json.Unmarshal(body, &account); err != nil || account.Email == nil {
		t.Errorf("%s body=%q has no account email", operation, response.Body)
		return
	}
	if *account.Email != wire || account.ID != importedAccountID || account.DisplayName != displayName ||
		!account.CreatedAt.Equal(importedCreatedAt) {
		t.Errorf("%s account=%+v email=%q, want email %q", operation, account, *account.Email, wire)
	}
}

func checkImportedCallback(t *testing.T, cookie, wire string) {
	t.Helper()

	request := httptest.NewRequest("GET", "/identity/internal/v1/session", nil)
	request.Header.Set("Authorization", "Bearer "+callbackToken)
	request.Header.Set("Cookie", cookie)
	response := httptest.NewRecorder()
	atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })

	if response.Code != http.StatusOK {
		t.Errorf("callback status=%d body=%q, want 200", response.Code, response.Body)
		return
	}
	var result struct {
		Identity struct {
			ID     string `json:"id"`
			Traits struct {
				Email       string `json:"email"`
				DisplayName string `json:"display_name"`
			} `json:"traits"`
		} `json:"identity"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Errorf("callback body=%q is not JSON", response.Body)
		return
	}
	if result.Identity.ID != importedAccountID || result.Identity.Traits.Email != wire ||
		result.Identity.Traits.DisplayName != "Imported Reader" {
		t.Errorf("callback identity=%+v, want email %q", result.Identity, wire)
	}
}

func insertImportedSession(t *testing.T) string {
	t.Helper()

	value := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x5a}, 32))
	digest := sha256.Sum256([]byte(value))
	_, err := api.db.Pool.Exec(t.Context(), `insert into sessions
		(tenant, token_hash, user_id, authenticated_at, expires_at, created_at)
		values ('tadoku/prod', $1, $2, $3, $3::timestamptz + interval '1 day', $3)`,
		hex.EncodeToString(digest[:]), importedAccountID, fixtureInstant)
	if err != nil {
		t.Fatal(err)
	}
	return "tadoku_session=" + value
}

func importedSessions(t *testing.T) int {
	t.Helper()

	var count int
	err := api.db.Pool.QueryRow(t.Context(), `select count(*) from sessions
		where tenant = 'tadoku/prod' and user_id = $1`, importedAccountID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}
