package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"

	valkeygo "github.com/valkey-io/valkey-go"
)

const (
	recoveryRecipientBudget = 3
	recoveryClientBudget    = 10
)

func recoverFrom(
	t *testing.T,
	handler http.Handler,
	tenantKey, email, peer, forwarded string,
) *httptest.ResponseRecorder {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"return_to":%q}`, email, recoverySettingsURL)
	return identityRequest(t, handler, tenantKey, "POST", "/identity/recovery", body, "", fromPeer(peer, forwarded))
}

// recoverIn requests recovery as a guest of the scenario's identity-a tenant.
func recoverIn(
	t *testing.T,
	s *scenario,
	handler http.Handler,
	email, peer string,
	edit func(*http.Request),
) *httptest.ResponseRecorder {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"return_to":%q}`, email, recoverySettingsURL)
	request := newIdentityRequest(t, s.tenant(t, "identity-a").String(), "POST", "/identity/recovery", body, "",
		fromPeer(peer, ""))
	if edit != nil {
		edit(request)
	}
	return s.serveWith(t, handler, "identity-a", guest, request)
}

func fromPeer(peer, forwarded string) func(*http.Request) {
	return func(r *http.Request) {
		r.RemoteAddr = peer + ":12345"
		if forwarded != "" {
			r.Header.Set("X-Real-IP", forwarded)
		}
	}
}

// newIdentityScenario provisions one identity-a tenant holding the account cast.
func newIdentityScenario(t *testing.T) *scenario {
	t.Helper()
	s := newScenario(t, "identity-a")
	s.seed(t, "identity-a", tenantSeed{})
	return s
}

func requireRecoveryAccepted(t *testing.T, response *httptest.ResponseRecorder) http.Header {
	t.Helper()
	requireIdentityStatus(t, response, http.StatusAccepted)
	if response.Body.Len() != 0 ||
		response.Header().Get("Set-Cookie") != "" ||
		response.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatalf("recovery answered with body %q or headers %v", response.Body, response.Header())
	}
	return response.Header()
}

func requireRecoveryRejected(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	requireIdentityStatus(t, response, status)
	var failure struct{ Code string }
	if err := json.Unmarshal(response.Body.Bytes(), &failure); err != nil {
		t.Fatal(err)
	}
	if failure.Code != code || response.Header().Get("Set-Cookie") != "" {
		t.Fatalf("rejected recovery returned code %q, want %q, or issued a session", failure.Code, code)
	}
}

type recoveryWriteCounts struct{ tokens, emails, jobs int }

func recoveryWrites(t *testing.T, tenantKey string) recoveryWriteCounts {
	t.Helper()
	var counts recoveryWriteCounts
	err := api.db.Pool.QueryRow(t.Context(), `select
		(select count(*) from recovery_tokens where tenant = $1),
		(select count(*) from emails where tenant = $1 and template = 'account.recovery'),
		(select count(*) from jobs where tenant = $1 and task_type = 'email.send.v1')`,
		tenantKey,
	).Scan(&counts.tokens, &counts.emails, &counts.jobs)
	if err != nil {
		t.Fatal(err)
	}
	return counts
}

func requireRecoveryWrites(t *testing.T, tenantKey string, want int) {
	t.Helper()
	got := recoveryWrites(t, tenantKey)
	if got != (recoveryWriteCounts{want, want, want}) {
		t.Fatalf("%s recovery tokens/emails/jobs = %+v, want %d each", tenantKey, got, want)
	}
}

func seedSecondRecoveryAccount(t *testing.T, tenantKey string) {
	t.Helper()
	writeTenantRows(t, tenantKey, `with reader as (
			select password_hash from accounts where tenant = $1 and email_normalized = 'reader@example.test'
		), created as (
			insert into users (tenant, id, display_name, created_at, updated_at)
			values ($1, 'dddddddd-dddd-4ddd-8ddd-dddddddddddd', 'Second Reader', '2026-01-04', '2026-01-04')
			returning tenant, id, created_at
		)
		insert into accounts (tenant, user_id, email, email_normalized, password_hash, created_at, updated_at)
		select created.tenant, created.id, 'second@example.test', 'second@example.test',
			reader.password_hash, created.created_at, created.created_at
		from created cross join reader`,
		tenantKey,
	)
}

func TestIdentityRecoveryRecipientSuppression(t *testing.T) {
	t.Parallel()
	s := newPairedIdentityScenario(t)
	tenantA, tenantB := s.tenant(t, "identity-a").String(), s.tenant(t, "identity-b").String()
	seedSecondRecoveryAccount(t, tenantA)
	variants := []string{"reader@example.test", "READER@example.test", "Reader@Example.Test", "reader@EXAMPLE.TEST"}
	peers := []struct{ peer, forwarded string }{
		{"198.51.100.30", ""},
		{"198.51.100.31", ""},
		{"10.42.0.9", "198.51.100.32"},
		{"203.0.113.33", ""},
	}

	var known http.Header
	for i := 0; i <= recoveryRecipientBudget; i++ {
		response := recoverFromRole(t, s, "identity-a", variants[i], peers[i].peer, peers[i].forwarded)
		known = requireRecoveryAccepted(t, response)
	}
	requireRecoveryWrites(t, tenantA, recoveryRecipientBudget)

	for i := 0; i <= recoveryRecipientBudget; i++ {
		email := "nobody" + strings.TrimPrefix(strings.ToLower(variants[i]), "reader")
		unknown := requireRecoveryAccepted(t,
			recoverFromRole(t, s, "identity-a", email, peers[i].peer, peers[i].forwarded))
		if !reflect.DeepEqual(unknown, known) {
			t.Fatalf("unknown recovery headers %v differ from suppressed known headers %v", unknown, known)
		}
	}
	requireRecoveryWrites(t, tenantA, recoveryRecipientBudget)

	requireRecoveryAccepted(t, recoverFromRole(t, s, "identity-a", "second@example.test", "198.51.100.30", ""))
	requireRecoveryWrites(t, tenantA, recoveryRecipientBudget+1)

	requireRecoveryAccepted(t, recoverFromRole(t, s, "identity-b", "READER@example.test", "198.51.100.30", ""))
	requireRecoveryWrites(t, tenantB, 1)
}

func TestIdentityRecoveryClientBudget(t *testing.T) {
	t.Parallel()
	s := newPairedIdentityScenario(t)
	tenantA, tenantB := s.tenant(t, "identity-a").String(), s.tenant(t, "identity-b").String()
	const peer = "198.51.100.40"
	for _, body := range []string{
		`{"email":"not-an-address","return_to":"https://account.tadoku.dev.lab/"}`,
		`{"email":"reader@example.test","return_to":"https://evil.example/"}`,
		`{"email":"reader@example.test","return_to":"https://account.tadoku.dev.lab/` + strings.Repeat("a", 2040) + `"}`,
	} {
		response := guestIdentityRequest(t, s, "identity-a", "POST", "/identity/recovery", body, "",
			fromPeer(peer, ""))
		requireIdentityStatus(t, response, http.StatusBadRequest)
	}

	for i := 0; i < recoveryClientBudget-1; i++ {
		requireRecoveryAccepted(t,
			recoverFromRole(t, s, "identity-a", fmt.Sprintf("budget%d@example.test", i), peer, ""))
	}
	requireRecoveryAccepted(t, recoverFromRole(t, s, "identity-a", "reader@example.test", peer, ""))
	requireRecoveryWrites(t, tenantA, 1)

	for _, attempt := range []struct{ email, peer, forwarded string }{
		{"reader@example.test", peer, ""},
		{"unknown@example.test", peer, ""},
		{"reader@example.test", peer, "198.51.100.77"},
		{"reader@example.test", "10.42.0.9", peer},
	} {
		response := recoverFromRole(t, s, "identity-a", attempt.email, attempt.peer, attempt.forwarded)
		requireRecoveryRejected(t, response, http.StatusTooManyRequests, "throttled")
	}
	requireRecoveryWrites(t, tenantA, 1)

	requireRecoveryAccepted(t, recoverFromRole(t, s, "identity-a", "reader@example.test", "203.0.113.41", ""))
	requireRecoveryAccepted(t,
		recoverFromRole(t, s, "identity-a", "reader@example.test", "10.42.0.9", "198.51.100.42"))
	requireRecoveryWrites(t, tenantA, 3)
	requireRecoveryAccepted(t, recoverFromRole(t, s, "identity-b", "reader@example.test", peer, ""))
	requireRecoveryWrites(t, tenantB, 1)

	response := guestIdentityRequest(t, s, "identity-a", "POST", "/identity/login",
		`{"email":"reader@example.test","password":"Fixture-login-482!"}`, "", fromPeer(peer, ""))
	requireIdentityStatus(t, response, http.StatusOK)
	response = guestIdentityRequest(t, s, "identity-a", "POST", "/identity/registration",
		`{"email":"registered@example.test","display_name":"Budget Reader","password":"Independent-653!"}`, "",
		fromPeer(peer, ""))
	requireIdentityStatus(t, response, http.StatusCreated)
}

func TestIdentityRecoveryAdmissionConcurrency(t *testing.T) {
	t.Parallel()
	s := newIdentityScenario(t)
	const attempts = 8

	responses := make([]*httptest.ResponseRecorder, attempts)
	var group sync.WaitGroup
	start := make(chan struct{})
	for i := range responses {
		body := fmt.Sprintf(`{"email":"reader@example.test","return_to":%q}`, recoverySettingsURL)
		request := httptest.NewRequest("POST", "https://tadoku.dev.lab/identity/recovery", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", "https://tadoku.dev.lab")
		request.RemoteAddr = fmt.Sprintf("198.51.100.%d:12345", 50+i)
		group.Go(func() {
			<-start
			responses[i] = s.serve(t, "identity-a", guest, request)
		})
	}
	close(start)
	group.Wait()

	for _, response := range responses {
		requireRecoveryAccepted(t, response)
	}
	requireRecoveryWrites(t, s.tenant(t, "identity-a").String(), recoveryRecipientBudget)
}

func identityThrottleKeys(t *testing.T, tenantKey, operation string) map[string]int64 {
	t.Helper()
	client := leaderboardValkey.client
	keys, err := client.Do(t.Context(),
		client.B().Keys().Pattern("tenant:"+tenantKey+":identity:"+operation+":*").Build()).AsStrSlice()
	if err != nil {
		t.Fatal(err)
	}
	ttls := make(map[string]int64, len(keys))
	for _, key := range keys {
		ttl, err := client.Do(t.Context(), client.B().Pttl().Key(key).Build()).AsInt64()
		if err != nil {
			t.Fatal(err)
		}
		ttls[key] = ttl
	}
	return ttls
}

func TestIdentityRecoveryAdmissionWindow(t *testing.T) {
	t.Parallel()
	s := newIdentityScenario(t)
	tenantKey := s.tenant(t, "identity-a").String()

	const peer = "198.51.100.60"
	for i := 0; i < recoveryRecipientBudget; i++ {
		requireRecoveryAccepted(t, recoverIn(t, s, api.handler, "reader@example.test", peer, nil))
	}

	first := identityThrottleKeys(t, tenantKey, "recovery")
	if len(first) != 2 {
		t.Fatalf("recovery admission keys = %v, want one client and one recipient counter", first)
	}
	for key, ttl := range first {
		if strings.Contains(key, "reader") || strings.Contains(key, "198.51") || ttl <= 0 || ttl > 900_000 {
			t.Fatalf("recovery admission key %q exposes input or has window %dms, want (0, 900000]", key, ttl)
		}
	}
	requireRecoveryAccepted(t, recoverIn(t, s, api.handler, "reader@example.test", peer, nil))
	for key, ttl := range identityThrottleKeys(t, tenantKey, "recovery") {
		if ttl > first[key] {
			t.Fatalf("suppressed request extended %q from %dms to %dms", key, first[key], ttl)
		}
	}
	requireRecoveryWrites(t, tenantKey, recoveryRecipientBudget)

	client := leaderboardValkey.client
	for key := range first {
		if err := client.Do(t.Context(), client.B().Del().Key(key).Build()).Error(); err != nil {
			t.Fatal(err)
		}
	}
	requireRecoveryAccepted(t, recoverIn(t, s, api.handler, "reader@example.test", peer, nil))
	requireRecoveryWrites(t, tenantKey, recoveryRecipientBudget+1)
}

type stalledRecoveryValkey struct {
	valkeygo.Client
	cancel context.CancelFunc
}

func (client *stalledRecoveryValkey) Do(ctx context.Context, _ valkeygo.Completed) valkeygo.ValkeyResult {
	if client.cancel != nil {
		client.cancel()
	}
	<-ctx.Done()
	return valkeygo.NewErrorResult(ctx.Err())
}

func TestIdentityRecoveryAdmissionFailsClosed(t *testing.T) {
	t.Parallel()
	s := newIdentityScenario(t)
	tenantKey := s.tenant(t, "identity-a").String()

	closed, err := newClosedLeaderboardValkeyClient()
	if err != nil {
		t.Fatal(err)
	}
	stalled := &stalledRecoveryValkey{Client: leaderboardValkey.client}
	cases := []struct {
		name   string
		client valkeygo.Client
		cancel bool
		status int
	}{
		{"unavailable", closed, false, http.StatusServiceUnavailable},
		{"timeout", stalled, false, http.StatusGatewayTimeout},
		{"cancelled", stalled, true, 499},
	}

	for _, tc := range cases {
		handler, _, err := newTestRouterWithLogger(
			t.Context(), api.db.AppPool, api.db.AppPool,
			slog.New(slog.NewTextHandler(io.Discard, nil)), withIdentityValkey(tc.client),
		)
		if err != nil {
			t.Fatal(err)
		}
		var answers []string
		for _, email := range []string{"reader@example.test", "unknown@example.test"} {
			response := recoverIn(t, s, handler, email, "198.51.100.20", func(r *http.Request) {
				if tc.cancel {
					ctx, cancel := context.WithCancel(r.Context())
					stalled.cancel = cancel
					*r = *r.WithContext(ctx)
				}
			})
			stalled.cancel = nil
			requireIdentityStatus(t, response, tc.status)
			answers = append(answers, fmt.Sprintf("%v %s", response.Header(), response.Body))
		}
		if answers[0] != answers[1] {
			t.Errorf("%s: known and unknown failures differ:\n%s\n%s", tc.name, answers[0], answers[1])
		}
		requireRecoveryWrites(t, tenantKey, 0)
	}

	requireRecoveryAccepted(t, recoverIn(t, s, api.handler, "reader@example.test", "198.51.100.20", nil))
	requireRecoveryWrites(t, tenantKey, 1)
}

func TestIdentityRecoveryRetryAfterQueueFailure(t *testing.T) {
	resetIdentity(t)
	_, err := api.db.Pool.Exec(t.Context(), `alter table jobs add constraint identity_recovery_reject_email
		check (task_type <> 'email.send.v1') not valid`)
	if err != nil {
		t.Fatal(err)
	}
	dropped := false
	drop := func(ctx context.Context) {
		if dropped {
			return
		}
		dropped = true
		_, err := api.db.Pool.Exec(ctx, "alter table jobs drop constraint identity_recovery_reject_email")
		if err != nil {
			t.Error(err)
		}
	}
	t.Cleanup(func() { drop(context.Background()) })

	response := recoverFrom(t, api.handler, identityBranchA, "reader@example.test", "198.51.100.70", "")
	requireIdentityStatus(t, response, http.StatusInternalServerError)
	requireRecoveryWrites(t, identityBranchA, 0)
	drop(t.Context())

	for i := 0; i < recoveryRecipientBudget; i++ {
		requireRecoveryAccepted(t, recoverFrom(t, api.handler, identityBranchA, "reader@example.test",
			fmt.Sprintf("198.51.100.%d", 71+i), ""))
	}
	requireRecoveryWrites(t, identityBranchA, recoveryRecipientBudget-1)

	var link string
	err = api.db.Pool.QueryRow(t.Context(), `select params->>'recovery_url' from emails
		where tenant = $1 order by created_at desc limit 1`, identityBranchA).Scan(&link)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	response = identityRequest(t, api.handler, identityBranchA, "GET",
		strings.TrimPrefix(parsed.Path, "/api/internal")+"?"+parsed.RawQuery, "", "", nil)
	requireIdentityStatus(t, response, http.StatusSeeOther)
}
