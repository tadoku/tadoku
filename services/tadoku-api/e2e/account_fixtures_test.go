package e2e_test

import (
	"encoding/json"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestAccountCastMatchesSignedSubjects(t *testing.T) {
	api.reset(t, "")
	cast, err := loadCast(nil)
	if err != nil {
		t.Fatal(err)
	}
	var members []member
	for who := range cast {
		members = append(members, who)
	}
	slices.Sort(members)
	seen := make(map[[2]string]bool)
	for _, who := range members {
		encoded := cast[who]
		claims := jwt.MapClaims{}
		if _, _, err := new(jwt.Parser).ParseUnverified(encoded, claims); err != nil {
			t.Fatal(err)
		}
		raw := claims["tenant"].(string)
		subject := claims["sub"].(string)
		binding := [2]string{raw, subject}
		if subject == "guest" || seen[binding] {
			continue
		}
		seen[binding] = true
		signed := func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+encoded) }
		t.Run(raw+"/"+string(who), func(t *testing.T) {
			traits := claims["session"].(map[string]any)["identity"].(map[string]any)["traits"].(map[string]any)
			body, err := json.Marshal(map[string]any{
				"email":    traits["email"],
				"password": identityFixturePassword,
			})
			if err != nil {
				t.Fatal(err)
			}

			response := identityRequest(t, api.handler, raw, "POST", "/identity/login", string(body), "", signed)
			requireIdentityStatus(t, response, http.StatusOK)
			cookie := response.Result().Cookies()[0]
			response = identityRequest(t, api.handler, raw, "GET", "/identity/session", "", cookie.String(), signed)
			view := identityView(t, response)
			if view.Account == nil || view.Account.ID != subject ||
				view.Account.DisplayName != traits["display_name"] || view.Account.Email != traits["email"] ||
				!view.Account.AuthenticatedAt.Equal(fixtureInstant) {
				t.Fatalf("local account does not match signed %s fixture: %+v", who, view.Account)
			}
			other, guestToken := "tadoku/prod", cast["guest"]
			if raw == "tadoku/prod" {
				other, guestToken = "e2e/alpha-0000000a", cast["alpha_guest"]
			}
			cookie.Name = "tadoku_session_" + strings.ReplaceAll(other, "/", "_")
			foreign := identityRequest(t, api.handler, other, "GET", "/identity/session", "", cookie.String(),
				func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+guestToken) })
			if identityView(t, foreign).Account != nil {
				t.Fatal("fixture session authenticated another tenant")
			}
		})
	}
}

func TestAccountCastSeedComposition(t *testing.T) {
	const caseDir = "testdata/ImmersionContestListLogs/200_default_limit"
	const readerID = "11111111-1111-4111-8111-111111111111"
	for _, state := range []string{"active", "deleted", "locked", "inactive"} {
		t.Run(state, func(t *testing.T) {
			resetIdentity(t, "testdata/accounts.sql")
			seeds := append(fixtureSeeds(caseDir, "setup.sql"), "testdata/accounts.sql", "testdata/accounts.sql")
			if err := api.db.Reset(t.Context(), seeds...); err != nil {
				t.Fatal(err)
			}

			query := "update users set deleted_at = $1 where tenant = 'tadoku/prod' and id = $2"
			if state == "locked" {
				query = "update users set deletion_locked_at = $1 where tenant = 'tadoku/prod' and id = $2"
			}
			if state == "deleted" || state == "locked" {
				if _, err := api.db.Pool.Exec(t.Context(), query, fixtureInstant, readerID); err != nil {
					t.Fatal(err)
				}
			}
			if state == "inactive" {
				_, err := api.db.Pool.Exec(t.Context(), `update accounts set state = 'inactive'
					where tenant = 'tadoku/prod' and user_id = $1`, readerID)
				if err != nil {
					t.Fatal(err)
				}
			}

			const snapshotSQL = `select to_jsonb(u)::text from users u
				where tenant = 'tadoku/prod' and id = $1`
			var before, after string
			if err := api.db.Pool.QueryRow(t.Context(), snapshotSQL, readerID).Scan(&before); err != nil {
				t.Fatal(err)
			}
			seed, err := os.ReadFile("testdata/accounts.sql")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := api.db.Pool.Exec(t.Context(), string(seed)); err != nil {
				t.Fatal(err)
			}
			if err := api.db.Pool.QueryRow(t.Context(), snapshotSQL, readerID).Scan(&after); err != nil {
				t.Fatal(err)
			}
			if before != after {
				t.Fatalf("account seed changed case-specific user: before=%s after=%s", before, after)
			}

			response := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login",
				`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
			if state != "active" {
				requireIdentityStatus(t, response, http.StatusUnauthorized)
				return
			}
			view := identityView(t, response)
			if view.Account == nil || view.Account.ID != readerID || view.Account.DisplayName != "Stored Reader" ||
				!view.Account.CreatedAt.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
				t.Fatalf("account seed lost case-specific name or timestamp: %+v", view.Account)
			}
		})
	}
}

func TestAccountCastUnavailableStateAndReset(t *testing.T) {
	for _, state := range []string{"inactive", "deleted", "locked"} {
		t.Run(state, func(t *testing.T) {
			resetIdentity(t, "testdata/accounts.sql")
			body := `{"email":"young@example.test","password":"Fixture-login-482!"}`
			login := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login", body, "", nil)
			requireIdentityStatus(t, login, http.StatusOK)
			cookie := login.Result().Cookies()[0].String()
			response := identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
			account := identityView(t, response).Account
			if account == nil {
				t.Fatal("fixture login did not authenticate before the state change")
			}
			query := "update accounts set state = 'inactive' where tenant = $1 and user_id = $2"
			if state == "deleted" {
				query = "update users set deleted_at = $3 where tenant = $1 and id = $2"
			}
			if state == "locked" {
				query = "update users set deletion_locked_at = $3 where tenant = $1 and id = $2"
			}
			args := []any{"tadoku/prod", account.ID}
			if state != "inactive" {
				args = append(args, fixtureInstant)
			}
			if _, err := api.db.Pool.Exec(t.Context(), query, args...); err != nil {
				t.Fatal(err)
			}

			response = identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
			if identityView(t, response).Account != nil {
				t.Fatal("unavailable account authenticated")
			}
			login = identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login", body, "", nil)
			requireIdentityStatus(t, login, http.StatusUnauthorized)

			resetIdentity(t, "testdata/accounts.sql")
			response = identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
			if identityView(t, response).Account != nil {
				t.Fatal("reset leaked the prior session")
			}
			login = identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login", body, "", nil)
			requireIdentityStatus(t, login, http.StatusOK)
		})
	}
}
