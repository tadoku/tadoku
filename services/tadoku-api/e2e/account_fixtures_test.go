package e2e_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func resetAccountCast(t *testing.T) {
	t.Helper()
	api.reset(t, "")
	if err := api.db.Reset(t.Context(), "testdata/accounts.sql"); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"tadoku/prod", identityBranchA, identityBranchB} {
		key, err := tenant.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		if err := flipt.SetIdentityV2(key, true); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAccountCastMatchesSignedSubjects(t *testing.T) {
	resetAccountCast(t)
	cast, err := loadCast(nil)
	if err != nil {
		t.Fatal(err)
	}
	createdAt := map[string]time.Time{
		"11111111-1111-4111-8111-111111111111": time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		"22222222-2222-4222-8222-222222222222": fixtureInstant,
		"33333333-3333-4333-8333-333333333333": fixtureInstant,
		"44444444-4444-4444-8444-444444444444": time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC),
	}
	for _, raw := range []string{"tadoku/prod", identityBranchA, identityBranchB} {
		for who, encoded := range cast {
			t.Run(raw+"/"+string(who), func(t *testing.T) {
				claims := jwt.MapClaims{}
				if _, _, err := new(jwt.Parser).ParseUnverified(encoded, claims); err != nil {
					t.Fatal(err)
				}
				subject := claims["sub"].(string)
				if subject == "guest" {
					return
				}
				traits := claims["session"].(map[string]any)["identity"].(map[string]any)["traits"].(map[string]any)
				body, err := json.Marshal(map[string]any{
					"email":    traits["email"],
					"password": identityFixturePassword,
				})
				if err != nil {
					t.Fatal(err)
				}

				response := identityRequest(t, api.handler, raw, "POST", "/identity/login", string(body), "", nil)
				requireIdentityStatus(t, response, http.StatusOK)
				cookie := response.Result().Cookies()[0]
				response = identityRequest(t, api.handler, raw, "GET", "/identity/session", "", cookie.String(), nil)
				view := identityView(t, response)
				if view.AccountApp != "v2" || view.Account == nil || view.Account.ID != subject ||
					view.Account.DisplayName != traits["display_name"] || view.Account.Email != traits["email"] ||
					!view.Account.CreatedAt.Equal(createdAt[subject]) || !view.Account.AuthenticatedAt.Equal(fixtureInstant) {
					t.Fatalf("local account does not match signed %s fixture: %+v", who, view.Account)
				}
				other := identityBranchA
				if raw == identityBranchA {
					other = identityBranchB
				}
				foreign := identityRequest(t, api.handler, other, "GET", "/identity/session", "", cookie.String(), nil)
				if identityView(t, foreign).Account != nil {
					t.Fatal("fixture session authenticated another tenant")
				}
			})
		}
	}
}

func TestAccountCastUnavailableStateAndReset(t *testing.T) {
	for _, state := range []string{"inactive", "deleted", "locked"} {
		t.Run(state, func(t *testing.T) {
			resetAccountCast(t)
			body := `{"email":"young@example.test","password":"Fixture-login-482!"}`
			login := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login", body, "", nil)
			requireIdentityStatus(t, login, http.StatusOK)
			cookie := login.Result().Cookies()[0].String()
			response := identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", cookie, nil)
			account := identityView(t, response).Account
			if account == nil || account.ID != "55555555-5555-4555-8555-555555555555" ||
				account.DisplayName != "Young User" || !account.CreatedAt.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
				t.Fatal("young account fixture changed")
			}
			query := "update accounts set state = 'inactive' where tenant = $1 and user_id = $2"
			if state == "deleted" {
				query = "update users set deleted_at = $3 where tenant = $1 and id = $2"
			}
			if state == "locked" {
				query = "update users set deletion_locked_at = $3 where tenant = $1 and id = $2"
			}
			args := []any{identityBranchA, account.ID}
			if state != "inactive" {
				args = append(args, fixtureInstant)
			}
			if _, err := api.db.Pool.Exec(t.Context(), query, args...); err != nil {
				t.Fatal(err)
			}

			response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", cookie, nil)
			if identityView(t, response).Account != nil {
				t.Fatal("unavailable account authenticated")
			}
			login = identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login", body, "", nil)
			requireIdentityStatus(t, login, http.StatusUnauthorized)

			resetAccountCast(t)
			response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", cookie, nil)
			if identityView(t, response).Account != nil {
				t.Fatal("reset leaked the prior session")
			}
			login = identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login", body, "", nil)
			requireIdentityStatus(t, login, http.StatusOK)
		})
	}
}
