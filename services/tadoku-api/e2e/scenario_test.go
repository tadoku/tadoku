package e2e_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"log/slog"
	mathrand "math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app/tenantlifecycle"
	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	fliptclient "github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/uuidx"
)

// scenario owns one independent E2E scenario: a private test tenant per semantic role, their seeds, signed cast,
// Flipt namespaces, cache keys and workers. Create it with newScenario after any t.Parallel call. Its cleanup stops the
// scenario's workers and tears down every tenant it allocated, including partially provisioned ones, without
// touching another scenario. Shared reference data, the canonical tenant and process-wide fixtures stay read-only.
type scenario struct {
	tenants map[string]tenant.TestKey

	// clock starts at fixtureInstant; set or advance it to move this scenario's business time alone.
	clock *timex.Clock
	ids   *scenarioIDs

	mu      sync.Mutex
	workers []*scenarioWorker
}

type scenarioIDs struct {
	mu     sync.Mutex
	source io.Reader
}

// newScenarioIDs seeds a scenario's UUID stream from its first tenant, so concurrent scenarios draw different IDs.
func newScenarioIDs(key tenant.TestKey) *scenarioIDs {
	seed := fnv.New64a()
	seed.Write([]byte(key.String()))
	return &scenarioIDs{source: mathrand.New(mathrand.NewSource(int64(seed.Sum64())))}
}

func (ids *scenarioIDs) Read(p []byte) (int, error) {
	ids.mu.Lock()
	defer ids.mu.Unlock()
	return ids.source.Read(p)
}

// useIDs replaces the scenario's UUID stream, such as with a golden case's fixed stream. Call it before the scenario's
// first request or worker.
func (s *scenario) useIDs(source io.Reader) {
	s.ids = &scenarioIDs{source: source}
}

type scenarioWorker struct {
	cancel context.CancelFunc
	done   chan error
	once   sync.Once
	err    error
}

// tenantSeed lists one role's fixture files. The cast seed always runs between setup and finalize, and
// relationships rows must omit their tenant: the role supplies it.
type tenantSeed struct {
	setup         []string
	finalize      []string
	relationships []string
}

const scenarioTeardownTimeout = 30 * time.Second

// Test safety: scenario seeds may not change session state or name the tenant bindings, so a seed cannot widen or
// switch its tenant even if it restores the binding before it ends.
var sessionStatePattern = regexp.MustCompile(`(?i)set_config\s*\(|tadoku\.(tenant|all_tenants)|(^|;)\s*(set|reset)\b`)

var sqlComments = regexp.MustCompile(`(?s)--[^\n]*|/\*.*?\*/`)

var allocatedScenarioTenants sync.Map

var scenarioFeatures = sync.OnceValues(func() ([]fliptmanagement.Resource, error) {
	path, err := bazel.Runfile("k8s/dev/base/flipt/features.yaml")
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return fliptmanagement.ParseFeatures(file)
})

// newScenario provisions one fresh test tenant for each distinct role, such as "primary" and "other".
func newScenario(t testing.TB, roles ...string) *scenario {
	t.Helper()
	if len(roles) == 0 {
		t.Fatal("a scenario needs at least one tenant role")
	}

	s := &scenario{tenants: make(map[string]tenant.TestKey, len(roles)), clock: timex.NewClock(fixtureInstant)}
	keys := make([]tenant.TestKey, 0, len(roles))
	for _, role := range roles {
		if _, duplicate := s.tenants[role]; duplicate {
			t.Fatalf("scenario role %q is listed twice", role)
		}
		key, err := allocateScenarioTenant(role)
		if err != nil {
			t.Fatal(err)
		}
		s.tenants[role] = key
		keys = append(keys, key)
	}
	s.ids = newScenarioIDs(keys[0])

	lifecycle := scenarioLifecycle(t)
	t.Cleanup(func() {
		s.stopWorkers(t)

		for _, key := range keys {
			ctx, cancel := context.WithTimeout(context.Background(), scenarioTeardownTimeout)
			err := lifecycle.Teardown(ctx, key)
			cancel()
			if err != nil {
				t.Errorf("tear down scenario tenant %s: %v", key, err)
				continue
			}
			allocatedScenarioTenants.Delete(key)
		}
	})

	features, err := scenarioFeatures()
	if err != nil {
		t.Fatalf("read scenario Flipt features: %v", err)
	}
	for _, key := range keys {
		if err := provisionScenarioTenant(t.Context(), lifecycle, key, features); err != nil {
			t.Fatalf("provision scenario tenant %s: %v", key, err)
		}
	}
	return s
}

// Test safety: registering the tenant with the migration-time platform scoring snapshot makes Provision skip copying
// the canonical tenant's current rule set.
const registerScenarioTenantSQL = `with registered as (
	insert into tenants (key, kind) values ($1, 'test')
	returning key
), active as (
	select rule_set.*
	from tadoku_test_platform_scoring_config_baseline config
	join tadoku_test_scoring_rule_sets_baseline rule_set on rule_set.id = config.active_rule_set_id
	where config.tenant = 'tadoku/prod'
		and rule_set.tenant = 'tadoku/prod'
		and rule_set.scope = 'platform'
), rule_set as (
	insert into scoring_rule_sets (tenant, id, scope, version, status, created_at, published_at)
	select registered.key, gen_random_uuid(), scope, version, status, created_at, published_at
	from active, registered
	returning tenant, id
), rules as (
	insert into scoring_rules (
		tenant, rule_set_id, priority, stackable, activity_id, unit_key, language_code, tag, score_source, rate
	)
	select rule_set.tenant, rule_set.id, rule.priority, rule.stackable, rule.activity_id, rule.unit_key,
		rule.language_code, rule.tag, rule.score_source, rule.rate
	from tadoku_test_scoring_rules_baseline rule, active, rule_set
	where rule.rule_set_id = active.id
)
insert into platform_scoring_config (tenant, singleton, active_rule_set_id)
select tenant, true, id from rule_set`

// provisionScenarioTenant leaves the tenant with the migration-time platform scoring and no pages, whatever the
// canonical tenant holds at the time.
func provisionScenarioTenant(
	ctx context.Context,
	lifecycle *tenantlifecycle.Application,
	key tenant.TestKey,
	features []fliptmanagement.Resource,
) error {
	tag, err := api.db.Pool.Exec(ctx, registerScenarioTenantSQL, key.String())
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return errors.New("register: the migration baseline has no active platform rule set")
	}

	if err := lifecycle.Provision(ctx, key, features); err != nil {
		return err
	}
	return api.db.SeedTenant(ctx, key, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "delete from pages_content; delete from pages")
		return err
	})
}

func allocateScenarioTenant(role string) (tenant.TestKey, error) {
	for {
		suffix := make([]byte, 4)
		if _, err := rand.Read(suffix); err != nil {
			return tenant.TestKey{}, err
		}
		key, err := tenant.ParseTestTenant("e2e/" + role + "-" + hex.EncodeToString(suffix))
		if err != nil {
			return tenant.TestKey{}, fmt.Errorf("scenario role %q: %w", role, err)
		}
		if _, taken := allocatedScenarioTenants.LoadOrStore(key, struct{}{}); !taken {
			return key, nil
		}
	}
}

func scenarioLifecycle(t testing.TB) *tenantlifecycle.Application {
	t.Helper()
	targets, _ := fliptclient.NewTargets("local", "default", "test")
	management, err := fliptmanagement.NewClient(fliptmanagement.Config{
		URL: flipt.URL(), Targets: targets, HTTPClient: &http.Client{Timeout: time.Second},
	})
	if err != nil {
		t.Fatal(err)
	}
	return tenantlifecycle.NewApplication(
		api.db.Pool,
		leaderboard.NewCache(leaderboardValkey.client, time.Second),
		management,
	)
}

func (s *scenario) tenant(t testing.TB, role string) tenant.TestKey {
	t.Helper()
	key, ok := s.tenants[role]
	if !ok {
		t.Fatalf("scenario has no tenant role %q", role)
	}
	return key
}

// seed runs the role's fixture files and the account cast in one transaction bound to the role's tenant under the
// restricted application role, so row-level security rejects writes to any other tenant or to shared tables.
// Missing files are skipped.
func (s *scenario) seed(t testing.TB, role string, files tenantSeed) {
	t.Helper()
	var relationships []fixtureRelationship
	for _, file := range files.relationships {
		rows, err := readRelationships(file)
		if err != nil {
			t.Fatal(err)
		}
		for index, relationship := range rows {
			if relationship.Tenant != "" {
				t.Fatalf("scenario relationship seed %s[%d] names tenant %q", file, index, relationship.Tenant)
			}
		}
		relationships = append(relationships, rows...)
	}
	s.seedTenant(t, role, files.setup, files.finalize, relationships)
}

// seedTenant is seed with relationships already read; their tenant must be empty.
func (s *scenario) seedTenant(
	t testing.TB,
	role string,
	setup, finalize []string,
	relationships []fixtureRelationship,
) {
	t.Helper()
	key := s.tenant(t, role)

	batch := &pgx.Batch{}
	for _, relationship := range relationships {
		relationship.queue(batch, key.String())
	}

	statements := map[string]string{}
	var paths []string
	for _, path := range slices.Concat(setup, []string{"testdata/tenant_accounts.sql"}, finalize) {
		contents, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if sessionStatePattern.MatchString(sqlComments.ReplaceAllString(string(contents), "")) {
			t.Fatalf("scenario seed %s changes session state; scenario seeds run bound to their tenant", path)
		}
		statements[path] = string(contents)
		paths = append(paths, path)
	}

	err := api.db.SeedTenant(t.Context(), key, func(ctx context.Context, tx pgx.Tx) error {
		for _, path := range paths {
			if _, err := tx.Exec(ctx, statements[path]); err != nil {
				return fmt.Errorf("execute seed %s: %w", path, err)
			}
		}
		return tx.SendBatch(ctx, batch).Close()
	})
	if err != nil {
		t.Fatalf("seed scenario tenant %s: %v", key, err)
	}
}

// verifyRows runs a verify query in a read-only transaction bound to the role's tenant under the restricted
// application role, so row-level security confines every table it reads, and returns its rows as one JSON array.
func (s *scenario) verifyRows(t testing.TB, role, query string) string {
	t.Helper()
	key := s.tenant(t, role)

	var rows string
	err := pgx.BeginTxFunc(t.Context(), api.db.AppPool, pgx.TxOptions{AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		if _, err := tx.Exec(t.Context(), "select set_config('tadoku.tenant', $1, true)", key.String()); err != nil {
			return err
		}
		return tx.QueryRow(t.Context(), wrapVerifyQuery(query)).Scan(&rows)
	})
	if err != nil {
		t.Fatalf("run verify query in %s: %v", key, err)
	}
	return rows
}

// serve sends request through the production router as the cast member in the role's tenant, with a freshly signed
// token; none sends no credentials.
func (s *scenario) serve(t testing.TB, role string, as member, request *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	return s.serveWith(t, api.handler, role, as, request)
}

// serveWith is serve through handler, such as a private router built on the shared pools.
func (s *scenario) serveWith(
	t testing.TB,
	handler http.Handler,
	role string,
	as member,
	request *http.Request,
) *httptest.ResponseRecorder {
	t.Helper()
	key := s.tenant(t, role)

	request.Header.Del("Authorization")
	if as != none {
		token, err := signCastToken(key.Key(), as)
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+token)
	}
	return s.serveRequest(handler, request)
}

// serveRequest sends request through handler in the scenario's execution context with its credentials as they are.
func (s *scenario) serveRequest(handler http.Handler, request *http.Request) *httptest.ResponseRecorder {
	request = request.WithContext(s.executionContext(request.Context()))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

// startWorker runs a worker that claims only the role's jobs until stop or the scenario's cleanup returns.
func (s *scenario) startWorker(t testing.TB, role string) (stop func()) {
	t.Helper()
	scope, err := jobqueue.OnlyTenant(s.tenant(t, role).Key())
	if err != nil {
		t.Fatal(err)
	}
	application, err := worker.NewApplication(
		workerDependencies(t, api.db.AppPool),
		worker.Config{
			Scope:           scope,
			Concurrency:     4,
			ShutdownTimeout: 2 * time.Second,
			Logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
			Metrics:         worker.NewMetrics(prometheus.NewRegistry()),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(s.executionContext(context.Background()))
	running := &scenarioWorker{cancel: cancel, done: make(chan error, 1)}
	s.mu.Lock()
	s.workers = append(s.workers, running)
	s.mu.Unlock()
	go func() { running.done <- application.Run(ctx) }()

	return func() {
		t.Helper()
		if err := running.stop(); err != nil {
			t.Error(err)
		}
	}
}

// runWorker completes every job of the role's tenant, including jobs those jobs enqueue, and fails on a failed job.
func (s *scenario) runWorker(t testing.TB, role string) {
	t.Helper()
	stop := s.startWorker(t, role)
	defer stop()
	if failed := waitForWorker(t, api, s.tenant(t, role).String()); failed != 0 {
		t.Fatalf("scenario worker failed %d jobs", failed)
	}
}

func (s *scenario) executionContext(ctx context.Context) context.Context {
	return uuidx.WithSource(timex.WithClock(ctx, s.clock), s.ids)
}

func (s *scenario) stopWorkers(t testing.TB) {
	s.mu.Lock()
	workers := slices.Clone(s.workers)
	s.mu.Unlock()
	for _, running := range workers {
		if err := running.stop(); err != nil {
			t.Error(err)
		}
	}
}

func (w *scenarioWorker) stop() error {
	w.once.Do(func() {
		w.cancel()
		select {
		case err := <-w.done:
			if err != nil {
				w.err = fmt.Errorf("scenario worker shutdown: %w", err)
			}
		case <-time.After(10 * time.Second):
			w.err = errors.New("scenario worker did not stop")
		}
	})
	return w.err
}
