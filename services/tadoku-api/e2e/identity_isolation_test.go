package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIdentityV2CookieTenantHostMatrix(t *testing.T) {
	resetIdentity(t)
	tenants := []string{"tadoku/prod", identityBranchA, identityBranchB}
	cookies := make([]*http.Cookie, len(tenants))
	ids := []string{
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		"cccccccc-cccc-4ccc-8ccc-cccccccccccc",
	}
	for i, key := range tenants {
		response := identityRequest(t, api.handler, key, "POST", "/identity/login",
			`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
		requireIdentityStatus(t, response, http.StatusOK)
		cookies[i] = response.Result().Cookies()[0]
		if key == "tadoku/prod" && (cookies[i].Name != "tadoku_session" || cookies[i].MaxAge != 720*60*60) {
			t.Error("production cookie name or lifetime differs")
		}
	}

	for actor, key := range tenants {
		for credential, cookie := range cookies {
			for _, host := range []string{
				"tadoku.dev.lab",
				"identity-a-0123abcd.tadoku.dev.lab",
				"identity-b-0123abcd.tadoku.dev.lab",
			} {
				t.Run(fmt.Sprintf("tenant%d/cookie%d/%s", actor, credential, host), func(t *testing.T) {
					response := identityRequest(t, api.handler, key, "GET", "/identity/session", "", cookie.String(),
						func(r *http.Request) { r.Host = host })
					view := identityView(t, response)
					if actor != credential {
						if view.Account != nil {
							t.Error("foreign cookie authenticated")
						}
					} else if view.Account == nil || view.Account.ID != ids[actor] {
						t.Error("host changed the signed tenant's selected account")
					}
				})
			}
			if actor != credential {
				renamed := cookies[actor].Name + "=" + cookie.Value
				response := identityRequest(t, api.handler, key, "GET", "/identity/session", "", renamed, nil)
				if identityView(t, response).Account != nil {
					t.Error("renaming a foreign cookie bypassed tenant storage")
				}
			}
		}

		both := cookies[0].String() + "; " + cookies[1].String() + "; " + cookies[2].String()
		response := identityRequest(t, api.handler, key, "GET", "/identity/session", "", both, nil)
		view := identityView(t, response)
		if view.Account == nil || view.Account.ID != ids[actor] {
			t.Error("coexisting production/branch cookies selected the wrong account")
		}
		request := httptest.NewRequest("GET", "/identity/internal/v1/session", nil)
		request.Header.Set("Authorization", "Bearer "+callbackToken)
		request.Header.Set("Cookie", both)
		if actor != 0 {
			request.Header.Set("X-Dev-Branch", strings.TrimPrefix(key, "tadoku/"))
		}
		callback := httptest.NewRecorder()
		atFixtureInstant(func() { api.handler.ServeHTTP(callback, request) })
		requireIdentityStatus(t, callback, http.StatusOK)
		var result struct {
			Identity struct{ ID string } `json:"identity"`
		}
		if err := json.Unmarshal(callback.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Identity.ID != ids[actor] {
			t.Error("callback selected another tenant's session")
		}
	}
}

func TestIdentitySessionCallbackGuestAndSelection(t *testing.T) {
	resetIdentity(t)
	for _, branch := range []string{"", "identity-a-0123abcd", "identity-b-0123abcd"} {
		for _, cookie := range []string{"", "unrelated=present", "ory_kratos_session=foreign", "tadoku_session=invalid"} {
			request := httptest.NewRequest("GET", "/identity/internal/v1/session", nil)
			request.Header.Set("Authorization", "Bearer "+callbackToken)
			request.Header.Set("X-Dev-Branch", branch)
			request.Header.Set("Cookie", cookie)
			response := httptest.NewRecorder()
			atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })
			requireIdentityStatus(t, response, http.StatusOK)
			if response.Body.String() != "{\"identity\":{\"id\":\"guest\"}}\n" {
				t.Errorf("unrelated/invalid cookie did not return guest: %s", response.Body)
			}
		}
	}
}

func TestIdentityV2ExpiredSessionIsGuest(t *testing.T) {
	resetIdentity(t)
	response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
	cookie := response.Result().Cookies()[0].String()
	_, err := api.db.Pool.Exec(t.Context(), `update sessions
		set authenticated_at = $2::timestamptz - interval '1 day', expires_at = $2
		where tenant = $1`, identityBranchA, fixtureInstant)
	if err != nil {
		t.Fatal(err)
	}

	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", cookie, nil)
	if identityView(t, response).Account != nil {
		t.Error("expired session still authenticates")
	}
	request := httptest.NewRequest("GET", "/identity/internal/v1/session", nil)
	request.Header.Set("Authorization", "Bearer "+callbackToken)
	request.Header.Set("X-Dev-Branch", strings.TrimPrefix(identityBranchA, "tadoku/"))
	request.Header.Set("Cookie", cookie)
	callback := httptest.NewRecorder()
	atFixtureInstant(func() { api.handler.ServeHTTP(callback, request) })
	requireIdentityStatus(t, callback, http.StatusOK)
	if callback.Body.String() != "{\"identity\":{\"id\":\"guest\"}}\n" {
		t.Error("expired session still authenticates through the callback")
	}
}

func TestIdentityGuestGatewayAdmissionPreservesBanGate(t *testing.T) {
	resetIdentity(t)
	response := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
	cookie := response.Result().Cookies()[0].String()
	if err := resetRelationships(
		t.Context(),
		api.db.Pool,
		"testdata/IdentityV2/banned.json",
	); err != nil {
		t.Fatal(err)
	}

	protected := readHTTPRequest(t, "testdata/BannedUsers/403_banned")
	denied := httptest.NewRecorder()
	atFixtureInstant(func() { api.handler.ServeHTTP(denied, protected) })
	requireIdentityStatus(t, denied, http.StatusForbidden)

	response = identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
	view := identityView(t, response)
	if view.Account == nil || view.Account.ID != "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" {
		t.Fatal("anonymous gateway admission failed to return the banned cookie account")
	}
	response = identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/logout", "", cookie, nil)
	requireIdentityStatus(t, response, http.StatusNoContent)
	response = identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
	if identityView(t, response).Account != nil {
		t.Error("banned cookie account could not revoke its session")
	}
}
