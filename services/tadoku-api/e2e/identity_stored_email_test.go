package e2e_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const storedAccountID = "dddddddd-dddd-4ddd-8ddd-dddddddddddd"
const storedPassword = "Stored-fixture-317!"

var storedCreatedAt = time.Date(2019, 1, 2, 9, 34, 5, 123456000, time.UTC)

type storedEmailCase struct {
	name, stored, login, wire string
}

func TestIdentityStoredAccountEmailResponses(t *testing.T) {
	kelvin := strings.Repeat("a", 316) + "K@x"
	kelvinLogin := strings.Repeat("a", 316) + "k@x"
	dottedI := strings.Repeat("a", 317) + "İ@x"
	dottedILogin := strings.Repeat("a", 317) + "i@x"
	longPrefix := strings.Repeat("　", 350) + "stored@example.test"
	cases := []storedEmailCase{
		{name: "ordinary", stored: "stored@example.test", login: "stored@example.test", wire: "stored@example.test"},
		{name: "mixed case", stored: "Stored@Example.Test", login: "stored@example.test", wire: "Stored@Example.Test"},
		{
			name:   "quoted local part",
			stored: `"stored"@example.test`,
			login:  "stored@example.test",
			wire:   "stored@example.test",
		},
		{
			name:   "leading space",
			stored: " stored@example.test",
			login:  "stored@example.test",
			wire:   "stored@example.test",
		},
		{
			name:   "trailing tab",
			stored: "stored@example.test\t",
			login:  "stored@example.test",
			wire:   "stored@example.test",
		},
		{
			name:   "ideographic space prefix",
			stored: "　stored@example.test",
			login:  "stored@example.test",
			wire:   "　stored@example.test",
		},
		{name: "kelvin sign", stored: kelvin, login: kelvinLogin, wire: kelvin},
		{name: "dotted capital I", stored: dottedI, login: dottedILogin, wire: dottedI},
		{
			name:   "long ideographic prefix",
			stored: longPrefix,
			login:  "stored@example.test",
			wire:   longPrefix,
		},
		{
			name:   "leading vertical tab",
			stored: "\vstored@example.test",
			login:  "stored@example.test",
			wire:   "\vstored@example.test",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seedStoredAccount(t, tc)

			login := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login",
				`{"email":`+strconv.Quote(tc.login)+`,"password":"`+storedPassword+`"}`, "", nil)
			checkStoredAccount(t, "login", login, tc.wire, "Stored Reader")
			cookies := login.Result().Cookies()
			if len(cookies) != 1 || cookies[0].Name != "tadoku_session" {
				t.Errorf("login set cookies %d, want one production session cookie", len(cookies))
			}
			if sessions := storedSessions(t); sessions != 1 {
				t.Errorf("login stored %d sessions, want 1", sessions)
			}

			cookie := insertStoredSession(t)
			session := identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
			checkStoredAccount(t, "session", session, tc.wire, "Stored Reader")
			checkStoredCallback(t, cookie, tc.wire)

			foreign := identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", cookie, nil)
			if view := identityView(t, foreign); view.Account != nil {
				t.Error("stored production cookie authenticated a branch tenant")
			}

			profile := identityRequest(t, api.handler, "tadoku/prod", "PATCH", "/identity/settings/profile",
				`{"display_name":"Renamed Reader"}`, cookie, nil)
			checkStoredAccount(t, "profile", profile, tc.wire, "Renamed Reader")
			checkStoredRow(t, tc, "Renamed Reader")
		})
	}
}

func seedStoredAccount(t *testing.T, tc storedEmailCase) {
	t.Helper()

	resetIdentity(t)
	hash, err := bcrypt.GenerateFromPassword([]byte(storedPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	_, err = api.db.Pool.Exec(t.Context(), `with stored_user as (
		insert into users (tenant, id, display_name)
		values ('tadoku/prod', $1, 'Stored Reader')
		returning tenant, id
	)
	insert into accounts (tenant, user_id, email, email_normalized, password_hash, created_at)
	select tenant, id, $2, $3, $4, $5
	from stored_user`, storedAccountID, tc.stored, tc.login, string(hash), storedCreatedAt)
	if err != nil {
		t.Fatal(err)
	}
}

func checkStoredRow(t *testing.T, tc storedEmailCase, displayName string) {
	t.Helper()

	var email, normalized, hash, name string
	var created time.Time
	err := api.db.Pool.QueryRow(t.Context(), `select a.email, a.email_normalized, a.password_hash, a.created_at,
		u.display_name
		from accounts a
		join users u on u.tenant = a.tenant and u.id = a.user_id
		where a.tenant = 'tadoku/prod' and a.user_id = $1`, storedAccountID,
	).Scan(&email, &normalized, &hash, &created, &name)
	if err != nil {
		t.Fatal(err)
	}
	if email != tc.stored || normalized != tc.login || !created.Equal(storedCreatedAt) || name != displayName {
		t.Errorf(
			"stored account email=%q normalized=%q created=%s display_name=%q",
			email,
			normalized,
			created,
			name,
		)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(storedPassword)) != nil {
		t.Error("stored account lost its stored password hash")
	}
}

func checkStoredAccount(
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
			Account json.RawMessage `json:"account"`
		}
		if err := json.Unmarshal(body, &view); err != nil || view.Account == nil {
			t.Errorf("%s body=%q is not a session", operation, response.Body)
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
	if *account.Email != wire || account.ID != storedAccountID || account.DisplayName != displayName ||
		!account.CreatedAt.Equal(storedCreatedAt) {
		t.Errorf("%s account=%+v email=%q, want email %q", operation, account, *account.Email, wire)
	}
}

func checkStoredCallback(t *testing.T, cookie, wire string) {
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
	if result.Identity.ID != storedAccountID || result.Identity.Traits.Email != wire ||
		result.Identity.Traits.DisplayName != "Stored Reader" {
		t.Errorf("callback identity=%+v, want email %q", result.Identity, wire)
	}
}

func insertStoredSession(t *testing.T) string {
	t.Helper()

	value := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x5a}, 32))
	digest := sha256.Sum256([]byte(value))
	_, err := api.db.Pool.Exec(t.Context(), `insert into sessions
		(tenant, token_hash, user_id, authenticated_at, expires_at, created_at)
		values ('tadoku/prod', $1, $2, $3, $3::timestamptz + interval '1 day', $3)`,
		hex.EncodeToString(digest[:]), storedAccountID, fixtureInstant)
	if err != nil {
		t.Fatal(err)
	}
	return "tadoku_session=" + value
}

func storedSessions(t *testing.T) int {
	t.Helper()

	var count int
	err := api.db.Pool.QueryRow(t.Context(), `select count(*) from sessions
		where tenant = 'tadoku/prod' and user_id = $1`, storedAccountID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}
