package e2e_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestIdentityV2IPThrottleAndFailOpen(t *testing.T) {
	resetIdentity(t)
	for i := 0; i < 51; i++ {
		body := fmt.Sprintf(`{"email":"unknown%d@example.test","password":"Independent-653!"}`, i)
		response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login", body, "",
			func(r *http.Request) {
				r.RemoteAddr = fmt.Sprintf("10.42.0.%d:12345", i+1)
				r.Header.Set("X-Real-IP", "198.51.100.30")
			})
		want := http.StatusUnauthorized
		if i == 50 {
			want = http.StatusTooManyRequests
		}
		requireIdentityStatus(t, response, want)
	}
	for i := 0; i < 51; i++ {
		body := fmt.Sprintf(`{"email":"unknown%d@example.test","password":"Independent-653!"}`, i)
		response := identityRequest(t, api.handler, identityBranchB, "POST", "/identity/login", body, "",
			func(r *http.Request) { r.Header.Set("X-Real-IP", fmt.Sprintf("198.51.100.%d", i+1)) })
		want := http.StatusUnauthorized
		if i == 50 {
			want = http.StatusTooManyRequests
		}
		requireIdentityStatus(t, response, want)
	}

	closed, err := newClosedLeaderboardValkeyClient()
	if err != nil {
		t.Fatal(err)
	}
	handler, _, _, err := newTestRouterWithLogger(
		t.Context(), api.db.AppPool, api.db.AppPool, api.keto, api.kratos,
		slog.New(slog.NewTextHandler(io.Discard, nil)), withIdentityValkey(closed),
	)
	if err != nil {
		t.Fatal(err)
	}
	response := identityRequest(t, handler, identityBranchA, "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
}

func TestIdentityV1ProviderSessionAndLogout(t *testing.T) {
	resetIdentity(t)
	resetKratosAfter(t, api)
	if err := flipt.SetIdentityV2(tenant.Production(), false); err != nil {
		t.Fatal(err)
	}
	provider := *api.kratos.HTTPClient()
	provider.Jar, _ = cookiejar.New(nil)
	request, err := http.NewRequestWithContext(t.Context(), "POST", "http://kratos.test/admin/identities",
		strings.NewReader(`{"schema_id":"user","traits":{"email":"legacy@example.test","display_name":"Legacy Reader"},"credentials":{"password":{"config":{"password":"Independent-653!"}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	created, err := provider.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer created.Body.Close()
	if created.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(created.Body)
		t.Fatalf("owned Kratos identity setup status=%d body=%s", created.StatusCode, body)
	}
	var legacy struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(created.Body).Decode(&legacy); err != nil {
		t.Fatal(err)
	}

	flowRequest, _ := http.NewRequestWithContext(t.Context(), "GET", "http://kratos-public.test/self-service/login/browser", nil)
	flowRequest.Header.Set("Accept", "application/json")
	flow, err := provider.Do(flowRequest)
	if err != nil {
		t.Fatal(err)
	}
	defer flow.Body.Close()
	var loginFlow struct {
		ID string `json:"id"`
		UI struct {
			Nodes []struct {
				Attributes struct {
					Name  string
					Value string
				}
			}
		}
	}
	if err := json.NewDecoder(flow.Body).Decode(&loginFlow); err != nil {
		t.Fatal(err)
	}
	csrf := ""
	for _, node := range loginFlow.UI.Nodes {
		if node.Attributes.Name == "csrf_token" {
			csrf = node.Attributes.Value
		}
	}
	loginBody, _ := json.Marshal(map[string]string{
		"method":     "password",
		"identifier": "legacy@example.test",
		"password":   "Independent-653!",
		"csrf_token": csrf,
	})
	loginRequest, _ := http.NewRequestWithContext(t.Context(), "POST",
		"http://kratos-public.test/self-service/login?flow="+loginFlow.ID,
		strings.NewReader(string(loginBody)))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginRequest.Header.Set("Accept", "application/json")
	loginResponse, err := provider.Do(loginRequest)
	if err != nil {
		t.Fatal(err)
	}
	defer loginResponse.Body.Close()
	if loginResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(loginResponse.Body)
		t.Fatalf("owned Kratos login setup status=%d body=%s", loginResponse.StatusCode, body)
	}
	cookie := ""
	for _, issued := range loginResponse.Cookies() {
		if issued.Name == "ory_kratos_session" {
			cookie = issued.Name + "=" + issued.Value
		}
	}
	if cookie == "" {
		t.Fatal("owned Kratos browser login did not issue a session cookie")
	}
	response := identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
	view := identityView(t, response)
	if view.AccountApp != "v1" || view.Account == nil || view.Account.ID != legacy.ID ||
		view.Account.Email != "legacy@example.test" || view.Account.DisplayName != "Legacy Reader" {
		t.Fatal("v1 session did not come from real Kratos")
	}
	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", cookie, nil)
	if identityView(t, response).Account != nil {
		t.Error("v2 tenant accepted a Kratos session")
	}
	response = identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/logout", "", cookie, nil)
	requireIdentityStatus(t, response, http.StatusNoContent)
	response = identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
	if identityView(t, response).Account != nil {
		t.Error("v1 logout did not revoke the provider session")
	}
}
