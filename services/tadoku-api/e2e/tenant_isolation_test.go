package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	valkeygo "github.com/valkey-io/valkey-go"
)

const isolationTenant = "e2e/isolation-0123abcd"

func TestTenantIsolation(t *testing.T) {
	api.reset(t, "")
	if err := api.db.Reset(t.Context(), "testdata/TenantIsolation/setup.sql"); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := api.db.Pool.Exec(
			ctx,
			"delete from tenants where key = $1 and kind = 'test'",
			isolationTenant,
		); err != nil {
			t.Error(err)
		}
	})

	if err := api.keto.Reset(t.Context(), "testdata/CreatePage/relationships.json"); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile("testdata/TenantIsolation/tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	var tokens map[string]string
	if err := json.Unmarshal(contents, &tokens); err != nil {
		t.Fatal(err)
	}

	previous := jwt.TimeFunc
	jwt.TimeFunc = func() time.Time { return fixtureInstant }
	defer func() { jwt.TimeFunc = previous }()

	key, err := tenant.Parse(isolationTenant)
	if err != nil {
		t.Fatal(err)
	}
	ctx := tenant.WithKey(t.Context(), key)

	var logIDs [2]string
	for index, who := range []string{"production", "test"} {
		body := isolationRequest(
			t,
			tokens[who],
			http.MethodPost,
			"/immersion/logs",
			`{"language_code":"eng","activity_id":1,"amount":10,"unit_key":"reading_page","description":"`+who+`"}`,
			http.StatusOK,
		)
		var log struct{ ID string }
		if err := json.Unmarshal(body, &log); err != nil || log.ID == "" {
			t.Fatalf("created log: id=%q error=%v", log.ID, err)
		}
		logIDs[index] = log.ID
	}

	for index, who := range []string{"production", "test"} {
		isolationRequest(t, tokens[who], http.MethodGet, "/immersion/logs/"+logIDs[index], "", http.StatusOK)
		isolationRequest(t, tokens[who], http.MethodGet, "/immersion/logs/"+logIDs[1-index], "", http.StatusNotFound)

		pageID := fmt.Sprintf("c0000000-0000-4000-8000-%012d", index+1)
		isolationRequest(
			t,
			tokens[who],
			http.MethodPost,
			"/content/pages/main",
			`{"id":"`+pageID+`","slug":"isolation-page","title":"`+who+`","html":"<p>`+who+`</p>","published_at":"2026-09-11T12:00:00Z"}`,
			http.StatusCreated,
		)
		body := isolationRequest(t, tokens[who], http.MethodGet, "/content/pages/main/isolation-page", "", http.StatusOK)
		if !strings.Contains(string(body), "<p>"+who+"</p>") &&
			!strings.Contains(string(body), `\u003cp\u003e`+who+`\u003c/p\u003e`) {
			t.Fatalf("page leaked another tenant: %s", body)
		}
	}

	var synchronized int
	if err := api.db.Pool.QueryRow(t.Context(), `
		select count(*) from users
		where id = '11111111-1111-4111-8111-111111111111' and tenant in ('tadoku/prod', $1)`,
		isolationTenant,
	).Scan(&synchronized); err != nil || synchronized != 2 {
		t.Fatalf("same synchronized identity across tenants: rows=%d error=%v", synchronized, err)
	}

	executor, err := postgres.Executor(ctx, api.db.AppPool)
	if err != nil {
		t.Fatal(err)
	}

	_, err = executor.Exec(ctx, "insert into languages (code, name) values ('iso', 'forbidden')")
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != "42501" {
		t.Fatalf("test-tenant shared insert=%v, want PostgreSQL42501", err)
	}

	tag, err := executor.Exec(ctx, "update languages set name = 'forbidden' where code = 'eng'")
	if err != nil || tag.RowsAffected() != 0 {
		t.Fatalf("test-tenant shared update=%s error=%v", tag, err)
	}

	queue := jobqueue.NewService(jobqueue.NewRepository(api.db.AppPool))
	for _, queueContext := range []context.Context{tenant.WithKey(t.Context(), tenant.Production()), ctx} {
		if err := queue.Enqueue(
			queueContext,
			jobs.InvalidateContestLeaderboardV1{ContestID: uuid.New()},
			jobs.InvalidateOfficialLeaderboardV1{Year: 2026},
		); err != nil {
			t.Fatal(err)
		}
	}

	var productionJobs, testJobs int
	if err := api.db.Pool.QueryRow(t.Context(), `
		select
			count(*) filter (where tenant = 'tadoku/prod'),
			count(*) filter (where tenant = $1)
		from jobs`,
		isolationTenant,
	).Scan(&productionJobs, &testJobs); err != nil || productionJobs == 0 || testJobs == 0 {
		t.Fatalf("both tenant queues: production=%d test=%d error=%v", productionJobs, testJobs, err)
	}

	branch, err := tenant.ParseDeployment(isolationTenant)
	if err != nil {
		t.Fatal(err)
	}
	runIsolationWorker(t, branch, 0, testJobs)

	if err := queue.Enqueue(ctx, jobs.InvalidateOfficialLeaderboardV1{Year: 2026}); err != nil {
		t.Fatal(err)
	}
	testJobs++
	runIsolationWorker(t, tenant.Deployment{}, productionJobs, testJobs)

	if err := queue.Enqueue(ctx, jobs.InvalidateOfficialLeaderboardV1{Year: 2026}); err != nil {
		t.Fatal(err)
	}
	claimed, err := queue.Claim(ctx, jobs.LeaderboardInvalidateOfficialV1, 1, time.Second, 5)
	if err != nil || len(claimed) != 1 || claimed[0].Tenant != key {
		t.Fatalf("own-tenant replay fixture claim=%v error=%v", claimed, err)
	}
	if changed, err := queue.Fail(ctx, claimed[0], "isolation_fixture"); err != nil || !changed {
		t.Fatalf("fail replay fixture: changed=%t error=%v", changed, err)
	}

	replayed, err := worker.Replay(ctx, queue, claimed[0].ID, "isolation-test", "cascade replay chain")
	if err != nil || replayed <= claimed[0].ID {
		t.Fatalf("replay chain: old=%d new=%d error=%v", claimed[0].ID, replayed, err)
	}

	before := isolationOwnedRows(t, "tadoku/prod")
	tag, err = executor.Exec(ctx, "delete from tenants where key = $1 and kind = 'test'", isolationTenant)
	if err != nil || tag.RowsAffected() != 1 {
		t.Fatalf("delete own test tenant: %s error=%v", tag, err)
	}
	for table, rows := range isolationOwnedRows(t, isolationTenant) {
		if rows != "[]" {
			t.Errorf("cascade left %s rows=%s", table, rows)
		}
	}
	if after := isolationOwnedRows(t, "tadoku/prod"); !reflect.DeepEqual(before, after) {
		t.Fatalf("test deletion changed canonical rows: before=%v after=%v", before, after)
	}

	tag, err = executor.Exec(ctx, "delete from tenants where key = 'tadoku/prod'")
	if err != nil || tag.RowsAffected() != 0 {
		t.Fatalf("production tenant delete: %s error=%v", tag, err)
	}

	var chain int
	if err := api.db.Pool.QueryRow(
		t.Context(),
		"select count(*) from jobs where id in ($1, $2)",
		claimed[0].ID,
		replayed,
	).Scan(&chain); err != nil || chain != 0 {
		t.Fatalf("replayed job chain remains: count=%d error=%v", chain, err)
	}

	var registry int
	if err := api.db.Pool.QueryRow(
		t.Context(),
		"select count(*) from tenants where key = 'tadoku/prod' and kind = 'production'",
	).Scan(&registry); err != nil || registry != 1 {
		t.Fatalf("canonical registry: rows=%d error=%v", registry, err)
	}
}

func isolationRequest(t *testing.T, token, method, path, body string, want int) []byte {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	timex.TheWorld(fixtureInstant, func() { api.handler.ServeHTTP(response, request) })
	if response.Code != want {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, response.Code, want, response.Body.String())
	}
	return response.Body.Bytes()
}

func isolationOwnedRows(t *testing.T, key string) map[string]string {
	t.Helper()
	rows, err := api.db.Pool.Query(t.Context(), `
		select c.relname
		from pg_class c
		join pg_namespace n on n.oid = c.relnamespace
		join pg_attribute a on a.attrelid = c.oid and a.attname = 'tenant' and not a.attisdropped
		where n.nspname = current_schema()
			and c.relkind in ('r', 'p')
			and c.relname <> 'tenant_overrides'
			and not starts_with(c.relname, 'tadoku_test_')
		order by c.relname`)
	if err != nil {
		t.Fatal(err)
	}

	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) == 0 {
		t.Fatal("empty tenant-owned table inventory")
	}

	values := make(map[string]string, len(tables))
	for _, table := range tables {
		var rows string
		query := "select coalesce(jsonb_agg(to_jsonb(row) order by to_jsonb(row)::text), '[]'::jsonb)::text" +
			" from " + pgx.Identifier{table}.Sanitize() + " row where tenant = $1"
		if err := api.db.Pool.QueryRow(t.Context(), query, key).Scan(&rows); err != nil {
			t.Fatal(err)
		}
		values[table] = rows
	}

	t.Logf("complete tenant row snapshot: tenant=%s tables=%d", key, len(tables))
	return values
}

func runIsolationWorker(t *testing.T, scope tenant.Deployment, production, test int) {
	t.Helper()
	provider := &isolationValkey{Client: leaderboardValkey.client, tenants: make(map[string]int)}
	service := leaderboard.NewService(
		leaderboard.NewRepository(api.db.AppPool),
		leaderboard.NewCache(provider, time.Second, ""),
	)
	application, err := worker.NewApplication(jobqueue.NewService(jobqueue.NewRepository(api.db.AppPool)), service, worker.Config{
		Scope:           scope,
		Concurrency:     4,
		ShutdownTimeout: 2 * time.Second,
		Logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		Metrics:         worker.NewMetrics(prometheus.NewRegistry()),
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- application.Run(ctx) }()
	defer func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(10 * time.Second):
			t.Error("isolation worker did not stop")
		}
	}()

	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()

	for {
		var canonicalCompleted, testCompleted, failed, canonicalClaimed int
		err := api.db.Pool.QueryRow(t.Context(), `
			select
				count(*) filter (where tenant = 'tadoku/prod' and state = 'completed'),
				count(*) filter (where tenant = $1 and state = 'completed'),
				count(*) filter (where state = 'failed'),
				count(*) filter (where tenant = 'tadoku/prod' and attempts > 0)
			from jobs`,
			isolationTenant,
		).Scan(&canonicalCompleted, &testCompleted, &failed, &canonicalClaimed)
		if err != nil || failed != 0 {
			t.Fatalf("worker outcomes: failed=%d error=%v", failed, err)
		}
		if _, branch := scope.Key(); branch && canonicalClaimed != 0 {
			t.Fatalf("branch worker claimed canonical jobs=%d", canonicalClaimed)
		}

		if canonicalCompleted == production && testCompleted == test {
			provider.mu.Lock()
			defer provider.mu.Unlock()
			if provider.tenants[isolationTenant] == 0 || provider.tenants[""] != 0 {
				t.Fatalf("real Valkey handler scope=%v", provider.tenants)
			}
			if _, branch := scope.Key(); branch {
				if len(provider.tenants) != 1 {
					t.Fatalf("branch handler leaked scope=%v", provider.tenants)
				}
			} else if provider.tenants["tadoku/prod"] == 0 || len(provider.tenants) != 2 {
				t.Fatalf("base worker did not dispatch both tenant contexts=%v", provider.tenants)
			}
			t.Logf("real Valkey delegated provider tenant contexts=%v", provider.tenants)
			return
		}

		select {
		case <-deadline.C:
			t.Fatalf(
				"worker completed production=%d/%d test=%d/%d",
				canonicalCompleted, production, testCompleted, test,
			)
		case <-tick.C:
		}
	}
}

type isolationValkey struct {
	valkeygo.Client
	mu      sync.Mutex
	tenants map[string]int
}

func (client *isolationValkey) Do(ctx context.Context, command valkeygo.Completed) valkeygo.ValkeyResult {
	key, _ := tenant.FromContext(ctx)
	client.mu.Lock()
	client.tenants[key.String()]++
	client.mu.Unlock()
	return client.Client.Do(ctx, command)
}
