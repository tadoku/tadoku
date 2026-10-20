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

const importedReturnTo = "https://tadoku.dev.lab/settings"
const importedRecoverySeed = "existing imported recovery link"

type importedDeliveryCase struct {
	name, raw, login string
	// mailbox is the recovery and old-address notice recipient. Empty means the stored address names no mailbox
	// the email writer accepts, so mail requests fail without writing anything.
	mailbox string
	// keeps reports that submitting the normalized address leaves the stored address and its recipient unchanged.
	keeps bool
	// refused marks a raw address the importer refuses; the case imports the login address and then stores raw.
	refused bool
}

func (tc importedDeliveryCase) account() importedEmailCase {
	if tc.refused {
		return importedEmailCase{raw: tc.login, stored: tc.raw, login: tc.login}
	}
	return importedEmailCase{raw: tc.raw, login: tc.login}
}

func importedDeliveryCases() []importedDeliveryCase {
	boundary := strings.Repeat("a", 307) + "@example.test"
	kelvin := strings.Repeat("a", 316) + "\u212a@x"
	dottedI := strings.Repeat("a", 317) + "İ@x"
	longPrefix := strings.Repeat("　", 350) + "imported@example.test"
	return []importedDeliveryCase{
		{
			name:    "ordinary",
			raw:     "imported@example.test",
			login:   "imported@example.test",
			mailbox: "imported@example.test",
			keeps:   true,
		},
		{
			name:    "mixed case",
			raw:     "Imported@Example.Test",
			login:   "imported@example.test",
			mailbox: "Imported@Example.Test",
		},
		{
			name:    "leading space",
			raw:     " imported@example.test",
			login:   "imported@example.test",
			mailbox: "imported@example.test",
			keeps:   true,
			refused: true,
		},
		{
			name:    "trailing tab",
			raw:     "imported@example.test\t",
			login:   "imported@example.test",
			mailbox: "imported@example.test",
			keeps:   true,
			refused: true,
		},
		{
			name:    "leading tab and mixed case",
			raw:     "\tImported@Example.Test",
			login:   "imported@example.test",
			mailbox: "Imported@Example.Test",
			refused: true,
		},
		{
			name:    "leading space at the 320-byte recipient limit",
			raw:     " " + boundary,
			login:   boundary,
			mailbox: boundary,
			keeps:   true,
			refused: true,
		},
		{
			name:    "ideographic space prefix",
			raw:     "　imported@example.test",
			login:   "imported@example.test",
			mailbox: "　imported@example.test",
		},
		{
			name:    "leading vertical tab",
			raw:     "\vimported@example.test",
			login:   "imported@example.test",
			refused: true,
		},
		{
			name:    "leading form feed",
			raw:     "\fimported@example.test",
			login:   "imported@example.test",
			refused: true,
		},
		{
			name:    "leading CRLF",
			raw:     "\r\nimported@example.test",
			login:   "imported@example.test",
			refused: true,
		},
		{
			name:    "trailing LF",
			raw:     "imported@example.test\n",
			login:   "imported@example.test",
			refused: true,
		},
		{
			name:    "kelvin sign",
			raw:     kelvin,
			login:   strings.Repeat("a", 316) + "k@x",
			refused: true,
		},
		{
			name:    "dotted capital I",
			raw:     dottedI,
			login:   strings.Repeat("a", 317) + "i@x",
			refused: true,
		},
		{
			name:    "long ideographic prefix",
			raw:     longPrefix,
			login:   "imported@example.test",
			refused: true,
		},
	}
}

func TestIdentityImportedAccountRecoveryDelivery(t *testing.T) {
	source := newKratosSource(t)

	for _, tc := range importedDeliveryCases() {
		t.Run(tc.name, func(t *testing.T) {
			original := tc.account()
			importAccount(t, source, original)
			seedImportedRecoveryToken(t)

			body := fmt.Sprintf(`{"email":%s,"return_to":%q}`, strconv.Quote(tc.login), importedReturnTo)
			response := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/recovery", body, "", nil)

			checkImportedRow(t, original, "Imported Reader")
			if seededImportedTokenUsed(t) {
				t.Error("recovery request consumed an existing recovery link")
			}
			if tc.mailbox == "" {
				requireImportedMailRejected(t, response)
				return
			}
			requireRecoveryAccepted(t, response)
			messages := importedMessages(t, "account.recovery")
			if len(messages) != 1 || messages[0].recipient != tc.mailbox {
				t.Fatalf("recovery mail=%+v, want one message to %q", messages, tc.mailbox)
			}
			if tokens := importedRecoveryTokens(t); tokens != 2 {
				t.Fatalf("recovery stored %d tokens, want the existing one and one new token", tokens)
			}
			followImportedRecoveryLink(t, messages[0])
		})
	}

	t.Run("unknown address", func(t *testing.T) {
		importAccount(t, source, importedEmailCase{raw: "imported@example.test", login: "imported@example.test"})

		body := fmt.Sprintf(`{"email":"unknown@example.test","return_to":%q}`, importedReturnTo)
		response := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/recovery", body, "", nil)

		requireRecoveryAccepted(t, response)
		if messages := importedMessages(t, "account.recovery"); len(messages) != 0 {
			t.Fatalf("unknown address received mail %+v", messages)
		}
	})
}

func TestIdentityImportedAccountEmailChange(t *testing.T) {
	source := newKratosSource(t)

	for _, tc := range importedDeliveryCases() {
		t.Run(tc.name+"/different address", func(t *testing.T) {
			original := tc.account()
			importAccount(t, source, original)
			seedImportedRecoveryToken(t)
			cookie := insertImportedSession(t)

			response := identityRequest(t, api.handler, "tadoku/prod", "PATCH", "/identity/settings/profile",
				`{"email":"changed@example.test","display_name":"Changed Import"}`, cookie, nil)

			if tc.mailbox == "" {
				requireImportedMailRejected(t, response)
				checkImportedRow(t, original, "Imported Reader")
				if seededImportedTokenUsed(t) {
					t.Error("rejected email change invalidated recovery links")
				}
				return
			}
			checkImportedAccount(t, "profile", response, "changed@example.test", "Changed Import")
			checkImportedRow(t, importedEmailCase{raw: "changed@example.test", login: "changed@example.test"},
				"Changed Import")
			if !seededImportedTokenUsed(t) {
				t.Error("email change kept a recovery link for the old address")
			}
			checkEmailChangedNotice(t, tc.mailbox, "changed@example.test")
		})

		t.Run(tc.name+"/normalized address", func(t *testing.T) {
			original := tc.account()
			importAccount(t, source, original)
			seedImportedRecoveryToken(t)
			cookie := insertImportedSession(t)

			response := identityRequest(t, api.handler, "tadoku/prod", "PATCH", "/identity/settings/profile",
				`{"email":`+strconv.Quote(tc.login)+`,"display_name":"Renamed Import"}`, cookie, nil)

			switch {
			case tc.mailbox == "":
				requireImportedMailRejected(t, response)
				checkImportedRow(t, original, "Imported Reader")
				if seededImportedTokenUsed(t) {
					t.Error("rejected email change invalidated recovery links")
				}
			case tc.keeps:
				checkImportedAccount(t, "profile", response, tc.mailbox, "Renamed Import")
				checkImportedRow(t, original, "Renamed Import")
				if seededImportedTokenUsed(t) {
					t.Error("unchanged email invalidated recovery links")
				}
				if messages := importedMessages(t, "account.email_changed"); len(messages) != 0 {
					t.Errorf("unchanged email sent notices %+v", messages)
				}
			default:
				checkImportedAccount(t, "profile", response, tc.login, "Renamed Import")
				checkImportedRow(t, importedEmailCase{raw: tc.login, login: tc.login}, "Renamed Import")
				if !seededImportedTokenUsed(t) {
					t.Error("email change kept a recovery link for the old address")
				}
				checkEmailChangedNotice(t, tc.mailbox, tc.login)
			}
		})
	}

	t.Run("stale session", func(t *testing.T) {
		original := importedEmailCase{
			raw:    "imported@example.test",
			stored: " imported@example.test",
			login:  "imported@example.test",
		}
		importAccount(t, source, original)
		cookie := insertImportedSession(t)
		_, err := api.db.Pool.Exec(t.Context(), `update sessions set authenticated_at = $2
			where tenant = 'tadoku/prod' and user_id = $1`, importedAccountID, fixtureInstant.Add(-2*time.Hour))
		if err != nil {
			t.Fatal(err)
		}

		response := identityRequest(t, api.handler, "tadoku/prod", "PATCH", "/identity/settings/profile",
			`{"email":"changed@example.test"}`, cookie, nil)

		requireIdentityStatus(t, response, http.StatusUnauthorized)
		if !strings.Contains(response.Body.String(), `"code":"stale_session"`) {
			t.Errorf("stale session error=%s", response.Body)
		}
		checkImportedRow(t, original, "Imported Reader")
	})

	t.Run("foreign tenant", func(t *testing.T) {
		original := importedEmailCase{
			raw:    "imported@example.test",
			stored: " imported@example.test",
			login:  "imported@example.test",
		}
		importAccount(t, source, original)
		cookie := insertImportedSession(t)

		response := identityRequest(t, api.handler, identityBranchA, "PATCH", "/identity/settings/profile",
			`{"email":"changed@example.test"}`, cookie, nil)

		requireIdentityStatus(t, response, http.StatusUnauthorized)
		checkImportedRow(t, original, "Imported Reader")
		if messages := importedMessages(t, "account.email_changed"); len(messages) != 0 {
			t.Errorf("foreign tenant request sent notices %+v", messages)
		}
	})
}

type importedMessage struct {
	id, recipient, body string
	params              map[string]string
}

func importedMessages(t *testing.T, template string) []importedMessage {
	t.Helper()

	rows, err := api.db.Pool.Query(t.Context(), `select e.id, e.recipient, e.body, e.params
		from emails e
		where e.tenant = 'tadoku/prod' and e.template = $1
		order by e.created_at, e.id`, template)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var messages []importedMessage
	for rows.Next() {
		var message importedMessage
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

	messages := importedMessages(t, "account.email_changed")
	if len(messages) != 1 || messages[0].recipient != recipient || messages[0].params["new_email"] != newEmail ||
		!strings.Contains(messages[0].body, newEmail) {
		t.Fatalf("email change notices=%+v, want one to %q announcing %q", messages, recipient, newEmail)
	}
}

func requireImportedMailRejected(t *testing.T, response *httptest.ResponseRecorder) {
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
		if messages := importedMessages(t, template); len(messages) != 0 {
			t.Fatalf("rejected request wrote %s mail %+v", template, messages)
		}
	}
	if tokens := importedRecoveryTokens(t); tokens != 1 {
		t.Fatalf("rejected request left %d recovery tokens, want only the existing one", tokens)
	}
}

func seedImportedRecoveryToken(t *testing.T) {
	t.Helper()

	digest := sha256.Sum256([]byte(importedRecoverySeed))
	_, err := api.db.Pool.Exec(t.Context(), `insert into recovery_tokens (tenant, token_hash, user_id, expires_at)
		values ('tadoku/prod', $1, $2, $3)`,
		hex.EncodeToString(digest[:]), importedAccountID, fixtureInstant.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
}

func seededImportedTokenUsed(t *testing.T) bool {
	t.Helper()

	digest := sha256.Sum256([]byte(importedRecoverySeed))
	var used bool
	err := api.db.Pool.QueryRow(t.Context(), `select used_at is not null from recovery_tokens
		where tenant = 'tadoku/prod' and token_hash = $1`, hex.EncodeToString(digest[:])).Scan(&used)
	if err != nil {
		t.Fatal(err)
	}
	return used
}

func importedRecoveryTokens(t *testing.T) int {
	t.Helper()

	var count int
	err := api.db.Pool.QueryRow(t.Context(), `select count(*) from recovery_tokens
		where tenant = 'tadoku/prod' and user_id = $1`, importedAccountID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func followImportedRecoveryLink(t *testing.T, message importedMessage) {
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
	if view := identityView(t, session); view.Account == nil || view.Account.ID != importedAccountID {
		t.Fatal("recovery link did not sign in the imported account")
	}
}
