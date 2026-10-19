package e2e_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestIdentitySettingsStaleSessions(t *testing.T) {
	for _, tc := range []struct{ method, path, body string }{
		{"PATCH", "/identity/settings/profile", `{"email":"changed@example.test"}`},
		{"PUT", "/identity/settings/password", `{"password":"Replacement-839!"}`},
	} {
		t.Run(tc.path, func(t *testing.T) {
			resetIdentity(t)
			cookie := identityLoginCookie(t)
			_, err := api.db.Pool.Exec(t.Context(), `update sessions
				set authenticated_at = $2::timestamptz - interval '1 hour 1 second'
				where tenant = $1`, identityBranchA, fixtureInstant)
			if err != nil {
				t.Fatal(err)
			}
			response := identityRequest(t, api.handler, identityBranchA, tc.method, tc.path, tc.body, cookie, nil)
			requireIdentityStatus(t, response, http.StatusUnauthorized)
			if !strings.Contains(response.Body.String(), `"code":"stale_session"`) {
				t.Fatalf("stale session error=%s", response.Body)
			}
		})
	}
}

func TestIdentitySettingsEmailNoticeAndPersistence(t *testing.T) {
	resetIdentity(t)
	cookie := identityLoginCookie(t)
	response := identityRequest(t, api.handler, identityBranchA, "PATCH", "/identity/settings/profile",
		`{"email":"CHANGED@EXAMPLE.TEST","display_name":"Changed Reader"}`, cookie, nil)
	requireIdentityStatus(t, response, http.StatusOK)
	var account struct {
		Email, DisplayName string
		AuthenticatedAt    time.Time `json:"authenticated_at"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &account); err != nil {
		t.Fatal(err)
	}
	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", cookie, nil)
	view := identityView(t, response)
	if account.Email != "changed@example.test" || !account.AuthenticatedAt.Equal(fixtureInstant) ||
		view.Account == nil || view.Account.Email != "changed@example.test" || view.Account.DisplayName != "Changed Reader" {
		t.Fatal("settings response/session did not persist normalized email and display name")
	}
	var recipient, template, newEmail, changedAt string
	err := api.db.Pool.QueryRow(t.Context(), `select recipient, template, params->>'new_email', params->>'changed_at'
		from emails where tenant = $1`, identityBranchA).Scan(&recipient, &template, &newEmail, &changedAt)
	if err != nil || recipient != "reader@example.test" || template != "account.email_changed" ||
		newEmail != "changed@example.test" || changedAt != fixtureInstant.Format(time.RFC3339) {
		t.Fatalf("email notice recipient=%q template=%q new=%q changed=%q error=%v",
			recipient, template, newEmail, changedAt, err)
	}
	response = identityRequest(t, api.handler, identityBranchA, "POST", "/identity/login",
		`{"email":"changed@example.test","password":"Fixture-login-482!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
}

func TestIdentitySettingsPasswordRevokesOthers(t *testing.T) {
	resetIdentity(t)
	first, second := identityLoginCookie(t), identityLoginCookie(t)
	response := identityRequest(t, api.handler, identityBranchA, "PUT", "/identity/settings/password",
		`{"password":"Replacement-839!"}`, first, nil)
	requireIdentityStatus(t, response, http.StatusNoContent)
	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", second, nil)
	if identityView(t, response).Account != nil {
		t.Fatal("password change left another session active")
	}
	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", first, nil)
	if identityView(t, response).Account == nil {
		t.Fatal("password change revoked the current session")
	}
}

func TestIdentitySettingsValidationAndTenantIsolation(t *testing.T) {
	resetIdentity(t)
	cookie := identityLoginCookie(t)
	for _, tc := range []struct {
		method, path, body, tenant, cookie string
		want                               int
	}{
		{"PATCH", "/identity/settings/profile", `{}`, identityBranchA, cookie, 400},
		{"PATCH", "/identity/settings/profile", `{"display_name":"x"}`, identityBranchA, cookie, 400},
		{"PATCH", "/identity/settings/profile", `{"email":"invalid"}`, identityBranchA, cookie, 400},
		{"PUT", "/identity/settings/password", `{"password":"short"}`, identityBranchA, cookie, 400},
		{"PUT", "/identity/settings/password", `{"password":"reader@example.test!"}`, identityBranchA, cookie, 400},
		{"PUT", "/identity/settings/password", `{"password":"` + strings.Repeat("界", 25) + `"}`, identityBranchA, cookie, 400},
		{"PUT", "/identity/settings/password", `{"password":"Replacement-839!"}`, identityBranchB, cookie, 401},
		{"PATCH", "/identity/settings/profile", `{"display_name":"New Reader"}`, identityBranchA, "", 401},
	} {
		t.Run(tc.method+tc.body+tc.tenant, func(t *testing.T) {
			response := identityRequest(t, api.handler, tc.tenant, tc.method, tc.path, tc.body, tc.cookie, nil)
			requireIdentityStatus(t, response, tc.want)
		})
	}
}

func TestIdentitySettingsDisplayNameDoesNotRequireFreshLogin(t *testing.T) {
	resetIdentity(t)
	cookie := identityLoginCookie(t)
	_, err := api.db.Pool.Exec(t.Context(), `update sessions
		set authenticated_at = $2::timestamptz - interval '2 hours'
		where tenant = $1`, identityBranchA, fixtureInstant)
	if err != nil {
		t.Fatal(err)
	}
	response := identityRequest(t, api.handler, identityBranchA, "PATCH", "/identity/settings/profile",
		`{"display_name":"Fresh Name"}`, cookie, nil)
	requireIdentityStatus(t, response, http.StatusOK)
}

func TestIdentityRecoveryAndEmailSettingsRollbackOnQueueFailure(t *testing.T) {
	for _, operation := range []string{"recovery", "email settings"} {
		t.Run(operation, func(t *testing.T) {
			resetIdentity(t)
			cookie := identityLoginCookie(t)
			_, err := api.db.Pool.Exec(t.Context(), `alter table jobs add constraint identity_d_reject_email
				check (task_type <> 'email.send.v1') not valid`)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if _, err := api.db.Pool.Exec(ctx, "alter table jobs drop constraint identity_d_reject_email"); err != nil {
					t.Error(err)
				}
			})
			method, path := "POST", "/identity/recovery"
			body := `{"email":"reader@example.test","return_to":"https://authv2.tadoku.dev.lab/"}`
			if operation == "email settings" {
				method, path, body = "PATCH", "/identity/settings/profile", `{"email":"changed@example.test"}`
			}
			response := identityRequest(t, api.handler, identityBranchA, method, path, body, cookie, nil)
			requireIdentityStatus(t, response, http.StatusInternalServerError)
			var email string
			var writes int
			err = api.db.Pool.QueryRow(t.Context(), `select email,
				(select count(*) from emails where tenant = $1) +
				(select count(*) from recovery_tokens where tenant = $1)
				from accounts where tenant = $1`, identityBranchA).Scan(&email, &writes)
			if err != nil || email != "reader@example.test" || writes != 0 {
				t.Fatalf("failed enqueue committed identity/email writes: email=%q writes=%d err=%v", email, writes, err)
			}
		})
	}
}

func TestIdentityPasswordChangeFencesInFlightLogin(t *testing.T) {
	resetIdentity(t)
	tx, err := api.db.Pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	_, err = tx.Exec(t.Context(), `update accounts set password_hash = 'changed-credential'
		where tenant = $1`, identityBranchA)
	if err != nil {
		t.Fatal(err)
	}

	var response *httptest.ResponseRecorder
	atFixtureInstant(func() {
		request := readHTTPRequest(t, "testdata/BannedUsers/200_guest")
		contents, err := os.ReadFile("testdata/IdentityV2/tokens.json")
		if err != nil {
			t.Fatal(err)
		}
		var tokens map[string]string
		if err := json.Unmarshal(contents, &tokens); err != nil {
			t.Fatal(err)
		}
		request.Method, request.URL.Path = "POST", "/identity/login"
		request.Body = io.NopCloser(strings.NewReader(`{"email":"reader@example.test","password":"Fixture-login-482!"}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", "https://tadoku.dev.lab")
		request.Header.Set("Authorization", "Bearer "+tokens[identityBranchA])
		response = httptest.NewRecorder()
		finished := make(chan struct{})
		go func() {
			defer close(finished)
			api.handler.ServeHTTP(response, request)
		}()

		deadline := time.Now().Add(5 * time.Second)
		waiting := false
		for time.Now().Before(deadline) {
			select {
			case <-finished:
				waiting = true
			default:
				err = api.db.Pool.QueryRow(t.Context(), `select exists (
					select 1 from pg_stat_activity
					where datname = current_database() and wait_event_type = 'Lock'
					and query like '%insert into sessions%')`).Scan(&waiting)
				if err != nil {
					t.Error(err)
				}
			}
			if waiting {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		if !waiting {
			t.Error("login never reached the session creation boundary")
		}
		if err := tx.Commit(t.Context()); err != nil {
			t.Error(err)
		}
		<-finished
	})
	requireIdentityStatus(t, response, http.StatusUnauthorized)
	if response.Header().Get("Set-Cookie") != "" {
		t.Fatal("in-flight old-password login minted a session after password replacement")
	}
}
