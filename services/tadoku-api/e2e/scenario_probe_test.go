package e2e_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	fliptclient "github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

const scenarioPassword = "Fixture-login-482!"

// populateScenario gives the role's tenant the cast, relationships, sessions, a pending account deletion job, a
// Flipt grant and leaderboard and identity throttle keys, through the fixture and real HTTP requests only.
func populateScenario(t testing.TB, s *scenario, role string) {
	t.Helper()
	key := s.tenant(t, role)
	s.seed(t, role, scenarioIsolationSeed)

	targets, err := fliptclient.NewTargets("local", "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	management := fliptmanagement.NewClient(fliptmanagement.Config{URL: flipt.URL(), Targets: targets})
	_, err = management.SetNamedUserAccess(
		tenant.WithKey(t.Context(), key.Key()),
		fliptmanagement.Segment{
			Key:         "release-webv3-access",
			Name:        "Webv3 preview access",
			Description: "Native account IDs explicitly granted Webv3 preview access.",
		},
		uuid.MustParse(scenarioCast[user].subject),
		true,
	)
	if err != nil {
		t.Fatal(err)
	}

	login(t, s, role, "reader@example.test", "wrong", http.StatusUnauthorized)
	login(t, s, role, "reader@example.test", scenarioPassword, http.StatusOK)
	session := login(t, s, role, "user2@example.test", scenarioPassword, http.StatusOK)
	deletion := scenarioRequest(http.MethodPost, "/identity/settings/account-deletion", `{"email":"user2@example.test"}`)
	deletion.AddCookie(session)
	requireScenarioStatus(t, s.serve(t, role, guest, deletion), http.StatusAccepted)
	leaderboard := scenarioRequest(http.MethodGet, "/immersion/leaderboard/global", "")
	requireScenarioStatus(t, s.serve(t, role, guest, leaderboard), http.StatusOK)

	state := scenarioState(t, key)
	for _, prefix := range []string{"leaderboard:", "identity:login:"} {
		found := false
		for name := range state {
			found = found || strings.HasPrefix(name, "valkey:tenant:"+key.String()+":"+prefix)
		}
		if !found {
			t.Fatalf("populated tenant has no %s keys: %v", prefix, slices.Sorted(maps.Keys(state)))
		}
	}
	for _, table := range []string{"sessions", "jobs", "permission_relationships"} {
		if state["table:"+table] == "[]" {
			t.Fatalf("populated tenant has no %s rows", table)
		}
	}
}

func login(t testing.TB, s *scenario, role, email, password string, want int) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	response := s.serve(t, role, guest, scenarioRequest(http.MethodPost, "/identity/login", string(body)))
	requireScenarioStatus(t, response, want)
	if cookies := response.Result().Cookies(); len(cookies) != 0 {
		return cookies[0]
	}
	return nil
}

func requireScenarioReadable(t testing.TB, s *scenario, role string) {
	t.Helper()
	checks := []struct {
		as   member
		path string
		want string
	}{
		{as: admin, path: "/authz/current-user/role", want: `{"role":"admin"}`},
		{as: banned, path: "/authz/current-user/role", want: `{"role":"banned"}`},
		{as: user, path: "/immersion/feature-flags", want: `{"decisions":{"release-webv3":true}}`},
		{as: guest, path: "/immersion/leaderboard/global", want: `"total_size":0`},
	}
	for _, check := range checks {
		response := s.serve(t, role, check.as, scenarioRequest(http.MethodGet, check.path, ""))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), check.want) {
			t.Errorf("%s as %s: status=%d body=%s, want 200 containing %s",
				check.path, check.as, response.Code, response.Body, check.want)
		}
	}
	login(t, s, role, "reader@example.test", scenarioPassword, http.StatusOK)
}

func requireScenarioStatus(t testing.TB, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status=%d body=%s, want %d", response.Code, response.Body, want)
	}
}

func scenarioRequest(method, path, body string) *http.Request {
	request := httptest.NewRequest(method, "https://tadoku.dev.lab"+path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Origin", "https://account.tadoku.dev.lab")
	return request
}

// scenarioState snapshots every tenant-owned table, the tenant's registry row, its Valkey keys with their values,
// and its Flipt namespace.
func scenarioState(t testing.TB, key tenant.TestKey) map[string]string {
	t.Helper()
	state := map[string]string{}
	for _, table := range tenantTables(t) {
		state["table:"+table] = tenantRows(t, key.Key(), table)
	}

	var registry string
	err := api.db.Pool.QueryRow(t.Context(),
		"select coalesce(jsonb_agg(to_jsonb(row))::text, '[]') from tenants row where key = $1",
		key.String(),
	).Scan(&registry)
	if err != nil {
		t.Fatal(err)
	}
	state["registry"] = registry

	client := leaderboardValkey.client
	keys, err := client.Do(t.Context(), client.B().Keys().Pattern("*"+key.String()+"*").Build()).AsStrSlice()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range keys {
		value, err := client.Do(t.Context(), client.B().Dump().Key(name).Build()).ToString()
		if err != nil {
			t.Fatal(err)
		}
		state["valkey:"+name] = hex.EncodeToString([]byte(value))
	}

	encoded, _ := json.Marshal(map[string]any{
		"exists":    flipt.Exists(key.Key()),
		"resources": flipt.ResourceCount(key.Key()),
	})
	state["flipt"] = string(encoded)
	return state
}

func requireNoScenarioResidue(t testing.TB, key tenant.TestKey) {
	t.Helper()
	for name, value := range scenarioState(t, key) {
		empty := value == "[]" || (name == "flipt" && value == `{"exists":false,"resources":0}`)
		if !empty {
			t.Errorf("torn-down tenant %s left %s=%s", key, name, value)
		}
	}
}

func tenantTables(t testing.TB) []string {
	t.Helper()
	rows, err := api.db.Pool.Query(t.Context(), `select c.relname
		from pg_class c
		join pg_namespace n on n.oid = c.relnamespace
		join pg_attribute a on a.attrelid = c.oid and a.attname = 'tenant' and not a.attisdropped
		where n.nspname = current_schema()
			and c.relkind in ('r', 'p')
			and not starts_with(c.relname, 'tadoku_test_')
		order by c.relname`)
	if err != nil {
		t.Fatal(err)
	}
	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil || len(tables) == 0 {
		t.Fatalf("tenant-owned inventory=%v error=%v", tables, err)
	}
	return tables
}

func tenantRows(t testing.TB, key tenant.Key, table string) string {
	t.Helper()
	var rows string
	query := "select coalesce(jsonb_agg(to_jsonb(row) order by to_jsonb(row)::text), '[]'::jsonb)::text" +
		" from " + pgx.Identifier{table}.Sanitize() + " row where tenant = $1"
	if err := api.db.Pool.QueryRow(t.Context(), query, key.String()).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

type jobCounts struct {
	total     int
	completed int
}

func scenarioJobs(t testing.TB, key tenant.TestKey) jobCounts {
	t.Helper()
	var counts jobCounts
	err := api.db.Pool.QueryRow(t.Context(),
		"select count(*), count(*) filter (where state = 'completed') from jobs where tenant = $1",
		key.String(),
	).Scan(&counts.total, &counts.completed)
	if err != nil {
		t.Fatal(err)
	}
	return counts
}

func allocatedTenants() map[tenant.TestKey]bool {
	allocated := map[tenant.TestKey]bool{}
	allocatedScenarioTenants.Range(func(key, _ any) bool {
		allocated[key.(tenant.TestKey)] = true
		return true
	})
	return allocated
}

func newlyAllocated(before map[tenant.TestKey]bool) []tenant.TestKey {
	var keys []tenant.TestKey
	for key := range allocatedTenants() {
		if !before[key] {
			keys = append(keys, key)
		}
	}
	return keys
}

// abortedTest stands in for a test that stops early. Fatal and FailNow end the body's goroutine as in the testing
// package; cleanups then run in reverse order after Context is cancelled, and every failure is recorded instead of
// failing the enclosing test.
type abortedTest struct {
	testing.TB
	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.Mutex
	cleanups []func()
	failures []string
}

// runAborted runs body as an abortedTest, then beforeCleanup, then the body's cleanups, and returns the failures.
func runAborted(t *testing.T, body func(*abortedTest), beforeCleanup func()) []string {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	test := &abortedTest{TB: t, ctx: ctx, cancel: cancel}

	done := make(chan struct{})
	go func() {
		defer close(done)
		body(test)
	}()
	<-done

	beforeCleanup()
	cancel()
	for _, cleanup := range slices.Backward(test.cleanups) {
		cleanup()
	}
	return test.failures
}

func (a *abortedTest) Context() context.Context { return a.ctx }

func (a *abortedTest) Cleanup(cleanup func()) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cleanups = append(a.cleanups, cleanup)
}

func (a *abortedTest) Errorf(format string, args ...any) { a.record(fmt.Sprintf(format, args...)) }

func (a *abortedTest) Error(args ...any) { a.record(fmt.Sprint(args...)) }

func (a *abortedTest) Fatalf(format string, args ...any) {
	a.Errorf(format, args...)
	a.FailNow()
}

func (a *abortedTest) Fatal(args ...any) {
	a.Error(args...)
	a.FailNow()
}

func (a *abortedTest) Fail() { a.record("Fail called") }

func (a *abortedTest) FailNow() { runtime.Goexit() }

func (a *abortedTest) Failed() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.failures) != 0
}

func (a *abortedTest) record(failure string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.failures = append(a.failures, failure)
}
