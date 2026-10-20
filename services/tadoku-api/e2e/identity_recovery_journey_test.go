package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
)

func TestIdentityGeneratedRecoveryJourney(t *testing.T) {
	resetIdentity(t)
	if err := api.keto.Reset(t.Context(), "testdata/journeys/relationships.json"); err != nil {
		t.Fatal(err)
	}
	old := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, old, http.StatusOK)
	oldCookie := old.Result().Cookies()[0].String()
	destination := "https://admin.tadoku.dev.lab/users?view=mine&sort=new#account"
	for n := 0; n < 2; n++ {
		body, _ := json.Marshal(map[string]string{"email": "READER@EXAMPLE.TEST", "return_to": destination})
		response := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/recovery", string(body), "", nil)
		requireIdentityStatus(t, response, http.StatusAccepted)
	}

	scope, err := jobqueue.AllTenantsExcept(jobqueue.WorkerComponent)
	if err != nil {
		t.Fatal(err)
	}
	application := emailWorker(t, emailService(), jobqueue.NewRepository(api.db.AppPool), scope)
	atFixtureInstant(func() {
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		go func() { done <- application.Run(ctx) }()
		defer func() {
			cancel()
			if err := <-done; err != nil {
				t.Error(err)
			}
		}()
		deadline := time.Now().Add(10 * time.Second)
		for {
			var held int
			err := api.db.Pool.QueryRow(t.Context(), `select count(*) from emails e join jobs j
   on j.tenant = e.tenant and j.payload->>'email_id' = e.id::text
   where e.tenant = 'tadoku/prod' and e.status = 'held' and j.state = 'completed'`).Scan(&held)
			if err != nil {
				t.Fatal(err)
			}
			if held == 2 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("recovery jobs did not hold both messages")
			}
			time.Sleep(10 * time.Millisecond)
		}

	})

	tokens := emailTokens(t)
	raw := emailRequest(t, tokens["admin"], "GET", "/emails?template=account.recovery&status=held", "", 200)
	var page struct {
		Emails []struct {
			ID string `json:"id"`
		} `json:"emails"`
	}
	if err := json.Unmarshal(raw, &page); err != nil || len(page.Emails) != 2 {
		t.Fatal("authorized held recovery email listing failed")
	}
	var link *url.URL
	for i, item := range page.Emails {
		raw := emailRequest(t, tokens["admin"], "GET", "/emails/"+item.ID, "", 200)
		var message struct {
			Params map[string]string `json:"params"`
			Status string            `json:"status"`
		}
		if err := json.Unmarshal(raw, &message); err != nil || message.Status != "held" {
			t.Fatal("held email detail invalid")
		}
		parsed, err := url.Parse(message.Params["recovery_url"])
		if err != nil || parsed.Host != "tadoku.dev.lab" || parsed.Query().Get("return_to") != destination {
			t.Fatal("generated recovery link lost its tenant or continuation")
		}
		token := strings.TrimPrefix(parsed.Path, "/api/internal/identity/recovery/")
		digest := sha256.Sum256([]byte(token))
		var matches int
		err = api.db.Pool.QueryRow(t.Context(), `select count(*) from recovery_tokens
   where tenant = 'tadoku/prod' and token_hash = $1`, hex.EncodeToString(digest[:])).Scan(&matches)
		if err != nil || matches != 1 {
			t.Fatal("generated token does not match its stored SHA-256")
		}
		if i == 0 {
			link = parsed
		}
	}

	path := strings.TrimPrefix(link.EscapedPath(), "/api/internal") + "?" + link.RawQuery
	response := identityRequest(t, api.handler, "tadoku/prod", "GET", path, "", "", nil)
	requireIdentityStatus(t, response, http.StatusSeeOther)
	cookie := response.Result().Cookies()[0]
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatal("insecure recovery session")
	}
	location, err := url.Parse(response.Header().Get("Location"))
	if err != nil || location.Scheme != "https" || location.Host != "authv2.tadoku.dev.lab" ||
		location.Path != "/" || location.Query().Get("return_to") != destination || len(location.String()) > 2048 {
		t.Error("consumption must land on tenant Settings with an intact continuation")
	}
	response = identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", oldCookie, nil)
	if identityView(t, response).Account != nil {
		t.Fatal("old session survived recovery")
	}
	var unused int
	err = api.db.Pool.QueryRow(t.Context(), `select count(*) from recovery_tokens
  where tenant = 'tadoku/prod' and used_at is null`).Scan(&unused)
	if err != nil || unused != 0 {
		t.Fatal("recovery left a usable sibling token")
	}
	response = identityRequest(t, api.handler, "tadoku/prod", "GET", path, "", "", nil)
	requireIdentityStatus(t, response, http.StatusBadRequest)
	response = identityRequest(t, api.handler, "tadoku/prod", "PUT", "/identity/settings/password",
		`{"password":"Replacement-839!"}`, cookie.String(), nil)
	requireIdentityStatus(t, response, http.StatusNoContent)
	response = identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Replacement-839!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
	response = identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusUnauthorized)
}

func TestIdentityRecoverySettingsLocationBounds(t *testing.T) {
	root := "https://identity-a-0123abcd.authv2.tadoku.dev.lab/"
	prefix := "https://tadoku.dev.lab/?q="
	boundary := prefix + strings.Repeat("a", 2048-len(root+"?return_to=")-len(url.QueryEscape(prefix)))
	for _, destination := range []string{"", "https://admin.tadoku.app/users?a=1&b=2#here", boundary} {
		resetIdentity(t)
		token := seedRecoveryToken(t, 'a', fixtureInstant.Add(time.Hour), nil)
		path := "/identity/recovery/" + token
		if destination != "" {
			path += "?return_to=" + url.QueryEscape(destination)
		}
		response := identityRequest(t, api.handler, identityBranchA, "GET", path, "", "", nil)
		requireIdentityStatus(t, response, http.StatusSeeOther)
		want := root
		if destination != "" {
			want += "?return_to=" + url.QueryEscape(destination)
		}
		if response.Header().Get("Location") != want || len(want) > 2048 {
			t.Error("Settings URL or continuation changed")
		}
	}
	for _, destination := range []string{boundary + "a", prefix + strings.Repeat("a", 2048-len(prefix)),
		prefix + strings.Repeat("&", 2048-len(prefix)), "https://evil.test/", "https://foreign.example/"} {
		resetIdentity(t)
		token := seedRecoveryToken(t, 'a', fixtureInstant.Add(time.Hour), nil)
		for _, email := range []string{"reader@example.test", "unknown@example.test"} {
			body, _ := json.Marshal(map[string]string{"email": email, "return_to": destination})
			response := identityRequest(
				t, api.handler, identityBranchA, "POST", "/identity/recovery", string(body), "", nil,
			)
			requireIdentityStatus(t, response, http.StatusBadRequest)
		}
		response := identityRequest(t, api.handler, identityBranchA, "GET",
			"/identity/recovery/"+token+"?return_to="+url.QueryEscape(destination), "", "", nil)
		requireIdentityStatus(t, response, http.StatusBadRequest)
		response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/recovery/"+token, "", "", nil)
		requireIdentityStatus(t, response, http.StatusSeeOther)
	}
	response := identityRequest(t, api.handler, identityBranchB, "GET",
		"/identity/recovery/"+seedRecoveryToken(t, 'b', fixtureInstant.Add(time.Hour), nil), "", "", nil)
	requireIdentityStatus(t, response, http.StatusBadRequest)
}
