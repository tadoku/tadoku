package e2e_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

const storedReturnTo = "https://tadoku.dev.lab/settings"
const storedRecoverySeed = "existing stored recovery link"

type storedDeliveryCase struct {
	name, stored, login string
	// mailbox is the recovery and old-address notice recipient. Empty means the stored address names no mailbox
	// the email writer accepts, so mail requests fail without writing anything.
	mailbox string
	// keeps reports that submitting the normalized address leaves the stored address and its recipient unchanged.
	keeps bool
}

func (tc storedDeliveryCase) account() storedEmailCase {
	return storedEmailCase{stored: tc.stored, login: tc.login}
}

func storedDeliveryCases() []storedDeliveryCase {
	boundary := strings.Repeat("a", 307) + "@example.test"
	kelvin := strings.Repeat("a", 316) + "\u212a@x"
	dottedI := strings.Repeat("a", 317) + "İ@x"
	longPrefix := strings.Repeat("　", 350) + "stored@example.test"
	return []storedDeliveryCase{
		{
			name:    "ordinary",
			stored:  "stored@example.test",
			login:   "stored@example.test",
			mailbox: "stored@example.test",
			keeps:   true,
		},
		{
			name:    "mixed case",
			stored:  "Stored@Example.Test",
			login:   "stored@example.test",
			mailbox: "Stored@Example.Test",
		},
		{
			name:    "leading space",
			stored:  " stored@example.test",
			login:   "stored@example.test",
			mailbox: "stored@example.test",
			keeps:   true,
		},
		{
			name:    "trailing tab",
			stored:  "stored@example.test\t",
			login:   "stored@example.test",
			mailbox: "stored@example.test",
			keeps:   true,
		},
		{
			name:    "leading tab and mixed case",
			stored:  "\tStored@Example.Test",
			login:   "stored@example.test",
			mailbox: "Stored@Example.Test",
		},
		{
			name:    "leading space at the 320-byte recipient limit",
			stored:  " " + boundary,
			login:   boundary,
			mailbox: boundary,
			keeps:   true,
		},
		{
			name:    "ideographic space prefix",
			stored:  "　stored@example.test",
			login:   "stored@example.test",
			mailbox: "　stored@example.test",
		},
		{
			name:   "leading vertical tab",
			stored: "\vstored@example.test",
			login:  "stored@example.test",
		},
		{
			name:   "leading form feed",
			stored: "\fstored@example.test",
			login:  "stored@example.test",
		},
		{
			name:   "leading CRLF",
			stored: "\r\nstored@example.test",
			login:  "stored@example.test",
		},
		{
			name:   "trailing LF",
			stored: "stored@example.test\n",
			login:  "stored@example.test",
		},
		{
			name:   "kelvin sign",
			stored: kelvin,
			login:  strings.Repeat("a", 316) + "k@x",
		},
		{
			name:   "dotted capital I",
			stored: dottedI,
			login:  strings.Repeat("a", 317) + "i@x",
		},
		{
			name:   "long ideographic prefix",
			stored: longPrefix,
			login:  "stored@example.test",
		},
	}
}

func TestIdentityStoredAccountRecoveryDelivery(t *testing.T) {
	for _, tc := range storedDeliveryCases() {
		t.Run(tc.name, func(t *testing.T) {
			original := tc.account()
			seedStoredAccount(t, original)
			seedStoredRecoveryToken(t)

			body := fmt.Sprintf(`{"email":%s,"return_to":%q}`, strconv.Quote(tc.login), storedReturnTo)
			response := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/recovery", body, "", nil)

			checkStoredRow(t, original, "Stored Reader")
			if seededStoredTokenUsed(t) {
				t.Error("recovery request consumed an existing recovery link")
			}
			if tc.mailbox == "" {
				requireStoredMailRejected(t, response)
				return
			}
			requireRecoveryAccepted(t, response)
			messages := storedMessages(t, "account.recovery")
			if len(messages) != 1 || messages[0].recipient != tc.mailbox {
				t.Fatalf("recovery mail=%+v, want one message to %q", messages, tc.mailbox)
			}
			if tokens := storedRecoveryTokens(t); tokens != 2 {
				t.Fatalf("recovery stored %d tokens, want the existing one and one new token", tokens)
			}
			followStoredRecoveryLink(t, messages[0])
		})
	}

	t.Run("unknown address", func(t *testing.T) {
		seedStoredAccount(t, storedEmailCase{stored: "stored@example.test", login: "stored@example.test"})

		body := fmt.Sprintf(`{"email":"unknown@example.test","return_to":%q}`, storedReturnTo)
		response := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/recovery", body, "", nil)

		requireRecoveryAccepted(t, response)
		if messages := storedMessages(t, "account.recovery"); len(messages) != 0 {
			t.Fatalf("unknown address received mail %+v", messages)
		}
	})
}

func TestIdentityStoredAccountEmailChange(t *testing.T) {
	for _, tc := range storedDeliveryCases() {
		t.Run(tc.name+"/different address", func(t *testing.T) {
			original := tc.account()
			seedStoredAccount(t, original)
			seedStoredRecoveryToken(t)
			cookie := insertStoredSession(t)

			response := identityRequest(t, api.handler, "tadoku/prod", "PATCH", "/identity/settings/profile",
				`{"email":"changed@example.test","display_name":"Changed Reader"}`, cookie, nil)

			if tc.mailbox == "" {
				requireStoredMailRejected(t, response)
				checkStoredRow(t, original, "Stored Reader")
				if seededStoredTokenUsed(t) {
					t.Error("rejected email change invalidated recovery links")
				}
				return
			}
			checkStoredAccount(t, "profile", response, "changed@example.test", "Changed Reader")
			checkStoredRow(t, storedEmailCase{stored: "changed@example.test", login: "changed@example.test"},
				"Changed Reader")
			if !seededStoredTokenUsed(t) {
				t.Error("email change kept a recovery link for the old address")
			}
			checkEmailChangedNotice(t, tc.mailbox, "changed@example.test")
		})

		t.Run(tc.name+"/normalized address", func(t *testing.T) {
			original := tc.account()
			seedStoredAccount(t, original)
			seedStoredRecoveryToken(t)
			cookie := insertStoredSession(t)

			response := identityRequest(t, api.handler, "tadoku/prod", "PATCH", "/identity/settings/profile",
				`{"email":`+strconv.Quote(tc.login)+`,"display_name":"Renamed Reader"}`, cookie, nil)

			switch {
			case tc.mailbox == "":
				requireStoredMailRejected(t, response)
				checkStoredRow(t, original, "Stored Reader")
				if seededStoredTokenUsed(t) {
					t.Error("rejected email change invalidated recovery links")
				}
			case tc.keeps:
				checkStoredAccount(t, "profile", response, tc.mailbox, "Renamed Reader")
				checkStoredRow(t, original, "Renamed Reader")
				if seededStoredTokenUsed(t) {
					t.Error("unchanged email invalidated recovery links")
				}
				if messages := storedMessages(t, "account.email_changed"); len(messages) != 0 {
					t.Errorf("unchanged email sent notices %+v", messages)
				}
			default:
				checkStoredAccount(t, "profile", response, tc.login, "Renamed Reader")
				checkStoredRow(t, storedEmailCase{stored: tc.login, login: tc.login}, "Renamed Reader")
				if !seededStoredTokenUsed(t) {
					t.Error("email change kept a recovery link for the old address")
				}
				checkEmailChangedNotice(t, tc.mailbox, tc.login)
			}
		})
	}

	t.Run("stale session", func(t *testing.T) {
		original := storedEmailCase{stored: " stored@example.test", login: "stored@example.test"}
		seedStoredAccount(t, original)
		cookie := insertStoredSession(t)
		_, err := api.db.Pool.Exec(t.Context(), `update sessions set authenticated_at = $2
			where tenant = 'tadoku/prod' and user_id = $1`, storedAccountID, fixtureInstant.Add(-2*time.Hour))
		if err != nil {
			t.Fatal(err)
		}

		response := identityRequest(t, api.handler, "tadoku/prod", "PATCH", "/identity/settings/profile",
			`{"email":"changed@example.test"}`, cookie, nil)

		requireIdentityStatus(t, response, http.StatusUnauthorized)
		if !strings.Contains(response.Body.String(), `"code":"stale_session"`) {
			t.Errorf("stale session error=%s", response.Body)
		}
		checkStoredRow(t, original, "Stored Reader")
	})

	t.Run("foreign tenant", func(t *testing.T) {
		original := storedEmailCase{stored: " stored@example.test", login: "stored@example.test"}
		seedStoredAccount(t, original)
		cookie := insertStoredSession(t)

		response := identityRequest(t, api.handler, identityBranchA, "PATCH", "/identity/settings/profile",
			`{"email":"changed@example.test"}`, cookie, nil)

		requireIdentityStatus(t, response, http.StatusUnauthorized)
		checkStoredRow(t, original, "Stored Reader")
		if messages := storedMessages(t, "account.email_changed"); len(messages) != 0 {
			t.Errorf("foreign tenant request sent notices %+v", messages)
		}
	})
}

type storedMessage struct {
	id, recipient, body string
	params              map[string]string
}

func storedMessages(t *testing.T, template string) []storedMessage {
	t.Helper()

	rows, err := api.db.Pool.Query(t.Context(), `select e.id, e.recipient, e.body, e.params
		from emails e
		where e.tenant = 'tadoku/prod' and e.template = $1
		order by e.created_at, e.id`, template)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var messages []storedMessage
	for rows.Next() {
		var message storedMessage
		if err := rows.Scan(&message.id, &message.recipient, &message.body, &message.params); err != nil {
			t.Fatal(err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	var jobs, idOnly int
	err = api.db.Pool.QueryRow(t.Context(), `select count(*),
		count(*) filter (where payload::jsonb = jsonb_build_object('email_id', e.id::text))
		from jobs j
		join emails e on e.tenant = j.tenant and e.id::text = j.payload::jsonb->>'email_id'
		where j.tenant = 'tadoku/prod' and j.task_type = 'email.send.v1' and e.template = $1`, template,
	).Scan(&jobs, &idOnly)
	if err != nil {
		t.Fatal(err)
	}
	var orphans int
	err = api.db.Pool.QueryRow(t.Context(), `select count(*) from jobs j
		where j.tenant = 'tadoku/prod' and j.task_type = 'email.send.v1'
		  and not exists (
		    select 1 from emails e
		    where e.tenant = j.tenant and e.id::text = j.payload::jsonb->>'email_id'
		  )`,
	).Scan(&orphans)
	if err != nil {
		t.Fatal(err)
	}
	if jobs != len(messages) || idOnly != len(messages) || orphans != 0 {
		t.Fatalf("%s emails=%d jobs=%d ID-only=%d orphan jobs=%d", template, len(messages), jobs, idOnly, orphans)
	}
	return messages
}

func checkEmailChangedNotice(t *testing.T, recipient, newEmail string) {
	t.Helper()

	messages := storedMessages(t, "account.email_changed")
	if len(messages) != 1 || messages[0].recipient != recipient || messages[0].params["new_email"] != newEmail ||
		!strings.Contains(messages[0].body, newEmail) {
		t.Fatalf("email change notices=%+v, want one to %q announcing %q", messages, recipient, newEmail)
	}
}

func requireStoredMailRejected(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	var failure struct{ Message string }
	if err := json.Unmarshal(response.Body.Bytes(), &failure); err != nil {
		t.Fatalf("status=%d body=%q is not an error response", response.Code, response.Body)
	}
	if response.Code != http.StatusBadRequest || failure.Message != "invalid email recipient" ||
		response.Header().Get("Set-Cookie") != "" {
		t.Fatalf("status=%d message=%q, want 400 invalid email recipient", response.Code, failure.Message)
	}
	for _, template := range []string{"account.recovery", "account.email_changed"} {
		if messages := storedMessages(t, template); len(messages) != 0 {
			t.Fatalf("rejected request wrote %s mail %+v", template, messages)
		}
	}
	if tokens := storedRecoveryTokens(t); tokens != 1 {
		t.Fatalf("rejected request left %d recovery tokens, want only the existing one", tokens)
	}
}

func seedStoredRecoveryToken(t *testing.T) {
	t.Helper()

	digest := sha256.Sum256([]byte(storedRecoverySeed))
	_, err := api.db.Pool.Exec(t.Context(), `insert into recovery_tokens (tenant, token_hash, user_id, expires_at)
		values ('tadoku/prod', $1, $2, $3)`,
		hex.EncodeToString(digest[:]), storedAccountID, fixtureInstant.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
}

func seededStoredTokenUsed(t *testing.T) bool {
	t.Helper()

	digest := sha256.Sum256([]byte(storedRecoverySeed))
	var used bool
	err := api.db.Pool.QueryRow(t.Context(), `select used_at is not null from recovery_tokens
		where tenant = 'tadoku/prod' and token_hash = $1`, hex.EncodeToString(digest[:])).Scan(&used)
	if err != nil {
		t.Fatal(err)
	}
	return used
}

func storedRecoveryTokens(t *testing.T) int {
	t.Helper()

	var count int
	err := api.db.Pool.QueryRow(t.Context(), `select count(*) from recovery_tokens
		where tenant = 'tadoku/prod' and user_id = $1`, storedAccountID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func followStoredRecoveryLink(t *testing.T, message storedMessage) {
	t.Helper()

	link, err := url.Parse(message.params["recovery_url"])
	if err != nil || !strings.Contains(message.body, message.params["recovery_url"]) || link.Host != "tadoku.dev.lab" {
		t.Fatalf("recovery mail does not carry a usable link (parse error %v)", err)
	}
	response := identityRequest(t, api.handler, "tadoku/prod", "GET",
		strings.TrimPrefix(link.Path, "/api/internal")+"?"+link.RawQuery, "", "", nil)
	requireIdentityStatus(t, response, http.StatusSeeOther)

	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("recovery set %d cookies, want one session", len(cookies))
	}
	session := identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", cookies[0].String(), nil)
	if view := identityView(t, session); view.Account == nil || view.Account.ID != storedAccountID {
		t.Fatal("recovery link did not sign in the stored account")
	}
}
