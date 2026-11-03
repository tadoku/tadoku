package e2e_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
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
)

// scenario owns one independent E2E scenario: a private test tenant per semantic role, their seeds, signed cast,
// Flipt namespaces, cache keys and workers. Create it with newScenario after any t.Parallel call. Its cleanup stops the
// scenario's workers and tears down every tenant it allocated, including partially provisioned ones, without
// touching another scenario. Shared reference data, the canonical tenant and process-wide fixtures stay read-only.
type scenario struct {
	tenants map[string]tenant.TestKey

	// executionContext, when set, derives every request and worker context; runners use it to attach the
	// scenario's business clock and ID source.
	executionContext func(context.Context) context.Context

	mu      sync.Mutex
	workers []*scenarioWorker
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

	s := &scenario{tenants: make(map[string]tenant.TestKey, len(roles))}
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

	lifecycle := scenarioLifecycle()
	t.Cleanup(func() {
		s.stopWorkers(t)

		ctx, cancel := context.WithTimeout(context.Background(), scenarioTeardownTimeout)
		defer cancel()
		for _, key := range keys {
			if err := lifecycle.Teardown(ctx, key); err != nil {
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
	return inScenarioTenant(ctx, key, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "delete from pages_content; delete from pages")
		return err
	})
}

// inScenarioTenant runs work in an owner-pool transaction bound to key that has switched to the restricted application
// role, so row-level security confines every statement to key's rows and refuses shared-table writes.
func inScenarioTenant(ctx context.Context, key tenant.TestKey, work func(pgx.Tx) error) error {
	applicationRole := pgx.Identifier{api.db.AppPool.Config().ConnConfig.User}.Sanitize()
	return pgx.BeginFunc(ctx, api.db.Pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "select set_config('tadoku.tenant', $1, true)", key.String()); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "set local role "+applicationRole); err != nil {
			return err
		}
		return work(tx)
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

func scenarioLifecycle() *tenantlifecycle.Application {
	targets, _ := fliptclient.NewTargets("local", "default", "test")
	return tenantlifecycle.NewApplication(
		api.db.Pool,
		leaderboard.NewCache(leaderboardValkey.client, time.Second),
		fliptmanagement.NewClient(fliptmanagement.Config{URL: flipt.URL(), Targets: targets}),
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
	key := s.tenant(t, role)

	sqlFiles := slices.Concat(files.setup, []string{"testdata/tenant_accounts.sql"}, files.finalize)
	batch := &pgx.Batch{}
	for _, file := range files.relationships {
		relationships, err := readRelationships(file)
		if err != nil {
			t.Fatal(err)
		}
		for index, relationship := range relationships {
			if relationship.Tenant != "" {
				t.Fatalf("scenario relationship seed %s[%d] names tenant %q", file, index, relationship.Tenant)
			}
			relationship.queue(batch, key.String())
		}
	}

	err := inScenarioTenant(t.Context(), key, func(tx pgx.Tx) error {
		for _, path := range sqlFiles {
			contents, err := os.ReadFile(path)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return err
			}
			if _, err := tx.Exec(t.Context(), string(contents)); err != nil {
				return fmt.Errorf("execute seed %s: %w", path, err)
			}
		}
		return tx.SendBatch(t.Context(), batch).Close()
	})
	if err != nil {
		t.Fatalf("seed scenario tenant %s: %v", key, err)
	}
}

// serve sends request through the production router as the cast member in the role's tenant, with a freshly signed
// token; none sends no credentials.
func (s *scenario) serve(t testing.TB, role string, as member, request *http.Request) *httptest.ResponseRecorder {
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
	if s.executionContext != nil {
		request = request.WithContext(s.executionContext(request.Context()))
	}

	recorder := httptest.NewRecorder()
	api.handler.ServeHTTP(recorder, request)
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

	ctx, cancel := context.WithCancel(context.Background())
	if s.executionContext != nil {
		ctx = s.executionContext(ctx)
	}
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
