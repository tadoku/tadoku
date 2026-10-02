package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant/alltenants"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	valkeygo "github.com/valkey-io/valkey-go"
)

type workerFixture struct {
	database *testpostgres.Database
	db       *pgxpool.Pool
	dsn      string
	client   valkeygo.Client
	key      tenant.Key
	prefix   string
}

type workerQueryScope struct {
	key tenant.Key
	all bool
}

type workerScopeTracer struct {
	mu      sync.Mutex
	queries map[string][]workerQueryScope
}

func (tracer *workerScopeTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	tracer.record(ctx, data.SQL)
	return ctx
}

func (*workerScopeTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (*workerScopeTracer) TraceBatchStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceBatchStartData) context.Context {
	return ctx
}

func (tracer *workerScopeTracer) TraceBatchQuery(ctx context.Context, _ *pgx.Conn, data pgx.TraceBatchQueryData) {
	tracer.record(ctx, data.SQL)
}

func (*workerScopeTracer) TraceBatchEnd(context.Context, *pgx.Conn, pgx.TraceBatchEndData) {}

func (tracer *workerScopeTracer) record(ctx context.Context, sql string) {
	for _, name := range []string{
		"Claim",
		"Complete",
		"Retry",
		"Fail",
		"Renew",
		"CleanupCompleted",
		"Stats",
		"UnsupportedStats",
	} {
		if !strings.HasPrefix(sql, "-- name: "+name+" :") {
			continue
		}

		key, _ := tenant.FromContext(ctx)
		tracer.mu.Lock()
		tracer.queries[name] = append(tracer.queries[name], workerQueryScope{key: key, all: alltenants.Enabled(ctx)})
		tracer.mu.Unlock()
		break
	}
}

func (tracer *workerScopeTracer) scopes(name string) []workerQueryScope {
	tracer.mu.Lock()
	defer tracer.mu.Unlock()
	return append([]workerQueryScope(nil), tracer.queries[name]...)
}

func tracedWorkerPool(t *testing.T, dsn string) (*pgxpool.Pool, *workerScopeTracer) {
	t.Helper()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	tracer := &workerScopeTracer{queries: make(map[string][]workerQueryScope)}
	config.ConnConfig.Tracer = tracer
	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool, tracer
}

func TestWorkerProcessUsesPersistedTenant(t *testing.T) {
	f, err := newWorkerFixture(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	key, err := tenant.Parse("e2e/worker-0123abcd")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(t.Context(), `insert into tenants (key, kind) values ($1, 'test')`, key.String()); err != nil {
		t.Fatal(err)
	}

	pool, tracer := tracedWorkerPool(t, f.dsn)
	repository := jobqueue.NewRepository(pool)

	for _, test := range []struct {
		name       string
		handlerErr error
		state      string
		transition string
	}{
		{"complete", nil, "completed", "Complete"},
		{"retry", errors.New("temporary fixture error"), "pending", "Retry"},
		{"fail", &permanentError{err: errors.New("permanent fixture error")}, "failed", "Fail"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var id int64
			if err := f.db.QueryRow(t.Context(), `
				insert into jobs (tenant, task_type, payload)
				values ($1, $2, '{"year":2026}')
				returning id`,
				key.String(), string(jobs.LeaderboardInvalidateOfficialV1),
			).Scan(&id); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				if _, err := f.db.Exec(ctx, `delete from jobs where id=$1`, id); err != nil {
					t.Error(err)
				}
			})

			claimed, err := repository.Claim(
				alltenants.With(t.Context()),
				jobs.LeaderboardInvalidateOfficialV1,
				1,
				30*time.Second,
				3,
			)
			if err != nil || len(claimed) != 1 || claimed[0].ID != id {
				t.Fatalf("claim fixture job: count=%d error=%v", len(claimed), err)
			}

			var observedKey tenant.Key
			var observedAll bool
			var storedTenant, storedAll string
			handlers, err := newRegistry(handle(func(ctx context.Context, _ jobs.InvalidateOfficialLeaderboardV1) error {
				observedKey, _ = tenant.FromContext(ctx)
				observedAll = alltenants.Enabled(ctx)
				executor, err := postgres.Executor(ctx, pool)
				if err != nil {
					return err
				}
				if err := executor.QueryRow(ctx, `
					select
						current_setting('tadoku.tenant'),
						coalesce(current_setting('tadoku.all_tenants', true), '')`,
				).Scan(&storedTenant, &storedAll); err != nil {
					return err
				}
				return test.handlerErr
			}, Policy{Concurrency: 1, Timeout: time.Second, MaxAttempts: 3}))
			if err != nil {
				t.Fatal(err)
			}

			runtime := &runner{
				queue:    jobqueue.NewService(repository),
				handlers: handlers,
				logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
				metrics:  NewMetrics(prometheus.NewRegistry()),
			}

			spec := handlers.ordered[0].spec
			spec.lease = time.Minute
			renewals := len(tracer.scopes("Renew"))
			err = runtime.process(tenant.WithKey(t.Context(), tenant.Production()), claimed[0], spec)
			if !errors.Is(err, test.handlerErr) {
				t.Errorf("handler result = %v; want %v", err, test.handlerErr)
			}

			if observedKey != key || observedAll || storedTenant != key.String() || storedAll == "on" {
				t.Errorf(
					"handler context tenant=%s all=%t; database tenant=%s all=%s; want tenant=%s without all-tenants",
					observedKey, observedAll, storedTenant, storedAll, key,
				)
			}

			var state string
			if err := f.db.QueryRow(t.Context(), `select state from jobs where id=$1`, id).Scan(&state); err != nil {
				t.Fatal(err)
			}
			if state != test.state {
				t.Errorf("job state = %s; want %s", state, test.state)
			}

			observed := tracer.scopes(test.transition)
			if len(observed) != 1 || observed[0].key != key || observed[0].all {
				t.Errorf("%s query contexts = %+v; want job tenant without all-tenants", test.transition, observed)
			}

			observed = tracer.scopes("Renew")[renewals:]
			if len(observed) != 1 || observed[0].key != key || observed[0].all {
				t.Errorf("pre-dispatch renewal contexts = %+v; want job tenant without all-tenants", observed)
			}
		})
	}
}

func TestBaseWorkerHousekeepingUsesAllTenants(t *testing.T) {
	f, err := newWorkerFixture(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	if _, err := f.db.Exec(t.Context(), `
		insert into tenants (key, kind) values ('e2e/worker-0123abcd', 'test');

		insert into jobs (tenant, task_type, payload, state, created_at, completed_at)
		select
			key, 'leaderboard.invalidate_official.v1', '{"year":2026}'::jsonb, 'completed',
			now()-interval '1 year', now()-interval '1 year'
		from tenants
		where key in ('tadoku/prod', 'e2e/worker-0123abcd');

		insert into jobs (tenant, task_type, payload, state, failed_at)
		select key, 'leaderboard.invalidate_official.v1', '{"year":2026}'::jsonb, 'failed', now()
		from tenants
		where key in ('tadoku/prod', 'e2e/worker-0123abcd');

		insert into jobs (tenant, task_type, payload)
		select key, 'future.worker.v1', '{}'::jsonb
		from tenants
		where key in ('tadoku/prod', 'e2e/worker-0123abcd');`); err != nil {
		t.Fatal(err)
	}

	pool, tracer := tracedWorkerPool(t, f.dsn)
	registry := prometheus.NewRegistry()
	handlers, err := newRegistry(handle(
		func(context.Context, jobs.InvalidateOfficialLeaderboardV1) error { return nil },
		Policy{Concurrency: 1, Timeout: time.Second, MaxAttempts: 3},
	))
	if err != nil {
		t.Fatal(err)
	}

	application := &Application{runner: &runner{
		queue:           jobqueue.NewService(jobqueue.NewRepository(pool)),
		handlers:        handlers,
		logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		metrics:         NewMetrics(registry),
		concurrency:     1,
		shutdownTimeout: time.Second,
	}}

	startWorker(t, application)
	if err := waitFor(t.Context(), func() (bool, error) { return application.Ready(), nil }); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"Claim", "CleanupCompleted", "Stats", "UnsupportedStats"} {
		observed := tracer.scopes(name)
		if len(observed) == 0 {
			t.Errorf("%s did not reach PostgreSQL", name)
		}
		for _, scope := range observed {
			if scope.key != (tenant.Key{}) || !scope.all {
				t.Errorf("%s scope = %+v; want only all-tenants", name, scope)
			}
		}
	}

	var completed int
	if err := f.db.QueryRow(t.Context(), `select count(*) from jobs where state='completed'`).Scan(&completed); err != nil {
		t.Fatal(err)
	}
	if completed != 0 {
		t.Errorf("expired completed jobs remaining = %d; want 0 across both tenants", completed)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"tadoku_worker_failed_jobs", "tadoku_worker_unsupported_pending_jobs"} {
		found := false
		for _, family := range families {
			if family.GetName() == name {
				found = true
				if len(family.Metric) != 1 || family.Metric[0].GetGauge().GetValue() != 2 {
					t.Errorf("%s = %v; want 2 across both tenants", name, family.Metric)
				}
			}
		}
		if !found {
			t.Errorf("missing metric %s", name)
		}
	}
}

func newWorkerFixture(ctx context.Context) (_ workerFixture, err error) {
	database, err := testpostgres.New(ctx)
	if err != nil {
		return workerFixture{}, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, database.Close())
		}
	}()
	key, err := tenant.Parse("e2e/worker-" + uuid.NewString())
	if err != nil {
		return workerFixture{}, err
	}
	_, err = database.Pool.Exec(ctx, `insert into tenants (key, kind) values ($1, 'test')`, key.String())
	if err != nil {
		return workerFixture{}, err
	}

	rawURL, err := testvalkey.URL()
	if err != nil {
		return workerFixture{}, err
	}
	option, err := valkeygo.ParseURL(rawURL)
	if err != nil {
		return workerFixture{}, err
	}
	option.SelectDB = 13
	option.ForceSingleClient = true
	option.DisableRetry = true
	client, err := valkeygo.NewClient(option)
	if err != nil {
		return workerFixture{}, err
	}
	return workerFixture{
		database: database,
		db:       database.Pool,
		dsn:      database.DSN,
		client:   client,
		key:      key,
		prefix:   "tenant:" + key.String() + ":",
	}, nil
}

func (f workerFixture) Close() error {
	ctx, stop := context.WithTimeout(tenant.WithKey(context.Background(), tenant.Production()), time.Second)
	defer stop()

	var cleanupErr error
	var cursor uint64
	for {
		page, err := f.client.Do(ctx, f.client.B().Scan().Cursor(cursor).Match(f.prefix+"*").Count(100).Build()).AsScanEntry()
		if err != nil {
			cleanupErr = err
			break
		}
		if len(page.Elements) > 0 {
			if err := f.client.Do(ctx, f.client.B().Del().Key(page.Elements...).Build()).Error(); err != nil {
				cleanupErr = err
				break
			}
		}
		if page.Cursor == 0 {
			break
		}
		cursor = page.Cursor
	}

	f.client.Close()
	return errors.Join(cleanupErr, f.database.Close())
}

func (f workerFixture) runner(t *testing.T, client valkeygo.Client, providerTimeout, shutdown time.Duration) *Application {
	service := leaderboard.NewService(
		leaderboard.NewRepository(f.db),
		leaderboard.NewCache(client, providerTimeout, ""),
	)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application, err := NewApplication(jobqueue.NewService(jobqueue.NewRepository(f.db)), service, Config{
		Concurrency:     4,
		Logger:          logger,
		Metrics:         NewMetrics(prometheus.NewRegistry()),
		ShutdownTimeout: shutdown,
	})
	if err != nil {
		t.Fatal(err)
	}
	return application
}

func startWorker(t *testing.T, runner *Application) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); runner.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Error("worker did not stop")
		}
	})
}

type blockingValkey struct {
	valkeygo.Client
	prefix   string
	started  chan struct{}
	release  <-chan struct{}
	canceled chan struct{}
}

type acquireStartTracer struct{ starts chan struct{} }

func (t *acquireStartTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	return ctx
}
func (t *acquireStartTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}
func (t *acquireStartTracer) TraceAcquireStart(ctx context.Context, _ *pgxpool.Pool, _ pgxpool.TraceAcquireStartData) context.Context {
	select {
	case t.starts <- struct{}{}:
	default:
	}
	return ctx
}
func (t *acquireStartTracer) TraceAcquireEnd(context.Context, *pgxpool.Pool, pgxpool.TraceAcquireEndData) {
}

func (c *blockingValkey) Do(ctx context.Context, command valkeygo.Completed) valkeygo.ValkeyResult {
	args := command.Commands()
	if len(args) > 3 && (args[0] == "EVALSHA" || args[0] == "EVAL") {
		for _, key := range args[3:] {
			if strings.HasPrefix(key, c.prefix) {
				select {
				case c.started <- struct{}{}:
				default:
				}
				select {
				case <-c.release:
				case <-ctx.Done():
					if c.canceled != nil {
						select {
						case c.canceled <- struct{}{}:
						default:
						}
					}
					return valkeygo.NewErrorResult(ctx.Err())
				}
				break
			}
		}
	}
	return c.Client.Do(ctx, command)
}

func TestWorkerStartsWithoutValkey(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	tenantCtx = tenant.WithKey(t.Context(), f.key)
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	rawURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}
	option, err := valkeygo.ParseURL(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	option.SelectDB = 13
	option.ForceSingleClient = true
	option.DisableRetry = true
	unavailable, err := valkeygo.NewClient(option)
	if err != nil {
		t.Fatal(err)
	}
	unavailable.Close()

	application := f.runner(t, unavailable, time.Second, time.Second)
	invalidID, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":0}`, false)
	if err != nil {
		t.Fatal(err)
	}
	validID, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":2025}`, false)
	if err != nil {
		t.Fatal(err)
	}

	startWorker(t, application)
	defer func() {
		if !t.Failed() {
			return
		}
		var pending, attempts int
		err := f.db.QueryRow(tenantCtx, `select count(*) filter (where state = 'pending'), sum(attempts)
			from jobs where id = any($1)`, []int64{invalidID, validID}).Scan(&pending, &attempts)
		t.Logf("worker with closed Valkey: ready=%t pending=%d attempts=%d query error=%v", application.Ready(), pending, attempts, err)
	}()

	if err := waitFor(tenantCtx, func() (bool, error) {
		var invalidFailed, validRetried bool
		err := f.db.QueryRow(tenantCtx, `select
			exists(select 1 from jobs where id = $1 and state = 'failed' and last_error = 'invalid_payload'),
			exists(select 1 from jobs where id = $2 and attempts > 0 and last_error = 'handler_error')`, invalidID, validID).Scan(&invalidFailed, &validRetried)
		return application.Ready() && invalidFailed && validRetried, err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerCancelsHandlerAfterLostLeaseAndReclaims(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	tenantCtx = tenant.WithKey(t.Context(), f.key)
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	release := make(chan struct{})
	var released sync.Once
	releaseAll := func() { released.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)
	blocked := &blockingValkey{
		Client:   f.client,
		prefix:   f.prefix + "leaderboard:contest:",
		started:  make(chan struct{}, 2),
		release:  release,
		canceled: make(chan struct{}, 1),
	}
	runner := f.runner(t, blocked, 8*time.Second, 2*time.Second)

	startWorker(t, runner)
	if err := waitFor(tenantCtx, func() (bool, error) {
		return runner.Ready(), nil
	}); err != nil {
		t.Fatal(err)
	}

	id, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateContestV1), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-blocked.started:
	case <-time.After(5 * time.Second):
		t.Fatal("contest handler did not start")
	}

	newToken := uuid.New()
	if _, err := f.db.Exec(tenantCtx, `update jobs set claim_token = $1, lease_expires_at = now() + interval '1 minute'
		where id = $2 and state = 'running'`, newToken, id); err != nil {
		t.Fatal(err)
	}
	select {
	case <-blocked.canceled:
	case <-time.After(6 * time.Second):
		t.Fatal("handler was not canceled after losing lease")
	}

	var state string
	var token uuid.UUID
	if err := f.db.QueryRow(tenantCtx, `select state, claim_token from jobs where id = $1`, id).Scan(&state, &token); err != nil {
		t.Fatal(err)
	}
	if state != "running" || token != newToken {
		t.Fatalf("lost claim was transitioned: state=%s token=%s", state, token)
	}

	releaseAll()
	if _, err := f.db.Exec(tenantCtx, `update jobs set lease_expires_at = now() - interval '1 second' where id = $1`, id); err != nil {
		t.Fatal(err)
	}
	if err := waitFor(tenantCtx, func() (bool, error) {
		var state string
		var attempts int
		err := f.db.QueryRow(tenantCtx, `select state, attempts from jobs where id = $1`, id).Scan(&state, &attempts)
		return state == "completed" && attempts == 2, err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerDeadlineExhaustionAndReplay(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	tenantCtx = tenant.WithKey(t.Context(), f.key)
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	release := make(chan struct{})
	var released sync.Once
	releaseAll := func() { released.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)
	blocked := &blockingValkey{
		Client:  f.client,
		prefix:  f.prefix + "leaderboard:contest:",
		started: make(chan struct{}, 1),
		release: release,
	}
	runner := f.runner(t, blocked, 8*time.Second, 2*time.Second)

	var id int64
	err := f.db.QueryRow(tenantCtx, `
		insert into jobs (tenant, task_type, payload, attempts)
		values ($3, $1, $2::jsonb, 4)
		returning id`,
		string(jobs.LeaderboardInvalidateContestV1),
		fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()),
		f.key.String(),
	).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	repository := jobqueue.NewRepository(f.db)
	claimed, err := repository.Claim(tenantCtx, jobs.LeaderboardInvalidateContestV1, 1, time.Second, 5)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("claim exhausted job: jobs=%d error=%v", len(claimed), err)
	}

	spec := handlerSpec{
		typeName:    jobs.LeaderboardInvalidateContestV1,
		limit:       2,
		timeout:     50 * time.Millisecond,
		lease:       time.Second,
		maxAttempts: 5,
	}
	if err := runner.runner.process(tenantCtx, claimed[0], spec); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("deadline handler error = %v", err)
	}

	var state, code string
	var attempts int
	if err := f.db.QueryRow(tenantCtx, `select state, attempts, last_error from jobs where id = $1`, id).Scan(&state, &attempts, &code); err != nil {
		t.Fatal(err)
	}
	if state != "failed" || attempts != 5 || code != "deadline_exceeded" {
		t.Fatalf("exhausted job: state=%q attempts=%d code=%q", state, attempts, code)
	}

	replayedID, err := Replay(tenantCtx, jobqueue.NewService(repository), id, "worker-e2e", "deadline repaired")
	if err != nil {
		t.Fatal(err)
	}

	releaseAll()
	startWorker(t, runner)
	if err := waitFor(tenantCtx, func() (bool, error) {
		var replayState string
		err := f.db.QueryRow(tenantCtx, `select state from jobs where id = $1`, replayedID).Scan(&replayState)
		return replayState == "completed", err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerShutdownCancelsActiveJob(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	tenantCtx = tenant.WithKey(t.Context(), f.key)
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	release := make(chan struct{})
	var released sync.Once
	releaseAll := func() { released.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)
	blocked := &blockingValkey{
		Client:  f.client,
		prefix:  f.prefix + "leaderboard:contest:",
		started: make(chan struct{}, 1),
		release: release,
	}
	runner := f.runner(t, blocked, 8*time.Second, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); runner.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Error("worker did not stop during cleanup")
		}
	})

	if err := waitFor(tenantCtx, func() (bool, error) {
		return runner.Ready(), nil
	}); err != nil {
		t.Fatal(err)
	}

	id, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateContestV1), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-blocked.started:
	case <-time.After(5 * time.Second):
		t.Fatal("contest handler did not start")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not stop after active handler cancellation")
	}

	var state string
	if err := f.db.QueryRow(tenantCtx, `select state from jobs where id = $1`, id).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state == "completed" {
		t.Error("canceled job was completed")
	}
	if runner.Ready() {
		t.Error("worker remains ready after shutdown")
	}
}

func TestWorkerDoesNotDispatchAfterClaimLeaseExpires(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	tenantCtx = tenant.WithKey(t.Context(), f.key)
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	release := make(chan struct{})
	close(release)
	observed := &blockingValkey{
		Client:  f.client,
		prefix:  f.prefix + "leaderboard:contest:",
		started: make(chan struct{}, 1),
		release: release,
	}
	runner := f.runner(t, observed, time.Second, time.Second)

	id, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateContestV1), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := jobqueue.NewRepository(f.db).Claim(tenantCtx, jobs.LeaderboardInvalidateContestV1, 1, 50*time.Millisecond, 5)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("claim: jobs=%d error=%v", len(claimed), err)
	}
	if err := waitFor(tenantCtx, func() (bool, error) { return time.Until(claimed[0].LeaseExpiresAt) <= 0, nil }); err != nil {
		t.Fatal(err)
	}

	spec := handlerSpec{
		typeName:    jobs.LeaderboardInvalidateContestV1,
		limit:       2,
		timeout:     time.Second,
		lease:       50 * time.Millisecond,
		maxAttempts: 5,
	}
	if err := runner.runner.process(tenantCtx, claimed[0], spec); err == nil {
		t.Error("expired claim was dispatched")
	}

	select {
	case <-observed.started:
		t.Error("Valkey invalidation began after lease expiry")
	default:
	}

	var state string
	if err := f.db.QueryRow(tenantCtx, `select state from jobs where id = $1`, id).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "running" {
		t.Errorf("expired claim state = %s; want running for recovery", state)
	}
}

func TestWorkerCompletesWhenRenewalIsCanceledByFinishedHandler(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	tenantCtx = tenant.WithKey(t.Context(), f.key)
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	poolConfig, err := pgxpool.ParseConfig(f.dsn)
	if err != nil {
		t.Fatal(err)
	}
	poolConfig.MaxConns = 1
	tracer := &acquireStartTracer{starts: make(chan struct{}, 1)}
	poolConfig.ConnConfig.Tracer = tracer
	limitedPool, err := pgxpool.NewWithConfig(tenantCtx, poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(limitedPool.Close)

	release := make(chan struct{})
	var released sync.Once
	releaseAll := func() { released.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)
	blocked := &blockingValkey{
		Client:  f.client,
		prefix:  f.prefix + "leaderboard:contest:",
		started: make(chan struct{}, 1),
		release: release,
	}

	service := leaderboard.NewService(
		leaderboard.NewRepository(limitedPool),
		leaderboard.NewCache(blocked, 8*time.Second, ""),
	)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repository := jobqueue.NewRepository(limitedPool)
	runner, err := NewApplication(jobqueue.NewService(repository), service, Config{
		Concurrency:     4,
		Logger:          logger,
		Metrics:         NewMetrics(prometheus.NewRegistry()),
		ShutdownTimeout: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	id, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateContestV1), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := repository.Claim(tenantCtx, jobs.LeaderboardInvalidateContestV1, 1, 3*time.Second, 5)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("claim: jobs=%d error=%v", len(claimed), err)
	}

	conn, err := limitedPool.Acquire(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	var returned sync.Once
	returnConn := func() { returned.Do(conn.Release) }
	defer returnConn()
	baseline := limitedPool.Stat().CanceledAcquireCount()
	select {
	case <-tracer.starts:
	default:
	}

	spec := handlerSpec{
		typeName:    jobs.LeaderboardInvalidateContestV1,
		limit:       2,
		timeout:     8 * time.Second,
		lease:       3 * time.Second,
		maxAttempts: 5,
	}
	done := make(chan error, 1)
	go func() { done <- runner.runner.process(tenantCtx, claimed[0], spec) }()

	select {
	case <-blocked.started:
	case <-time.After(3 * time.Second):
		t.Fatal("handler did not begin")
	}
	select {
	case <-tracer.starts:
	case <-time.After(3 * time.Second):
		t.Fatal("lease renewal did not request the held pool connection")
	}

	releaseAll()
	if err := waitFor(tenantCtx, func() (bool, error) {
		return limitedPool.Stat().CanceledAcquireCount() > baseline, nil
	}); err != nil {
		t.Fatal(err)
	}

	returnConn()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("completed handler lost claim: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("handler did not complete")
	}

	if got := limitedPool.Stat().CanceledAcquireCount(); got <= baseline {
		t.Errorf("renewal did not contend on the held connection: canceled acquires %d -> %d", baseline, got)
	}

	var state string
	if err := f.db.QueryRow(tenantCtx, `select state from jobs where id = $1`, id).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "completed" {
		t.Errorf("job state = %s; want completed", state)
	}
}

func TestWorkerFairClaimsWithoutPrefetch(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	tenantCtx = tenant.WithKey(t.Context(), f.key)
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	release := make(chan struct{})
	var released sync.Once
	releaseAll := func() { released.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)
	blocked := &blockingValkey{
		Client:  f.client,
		prefix:  f.prefix + "leaderboard:contest:",
		started: make(chan struct{}, 4),
		release: release,
	}
	runner := f.runner(t, blocked, 8*time.Second, 2*time.Second)

	startWorker(t, runner)
	if err := waitFor(tenantCtx, func() (bool, error) {
		return runner.Ready(), nil
	}); err != nil {
		t.Fatal(err)
	}

	contestIDs := make([]int64, 3)
	for i := range contestIDs {
		id, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateContestV1), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
		if err != nil {
			t.Fatal(err)
		}
		contestIDs[i] = id
	}
	officialID, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":2025}`, false)
	if err != nil {
		t.Fatal(err)
	}

	for range 2 {
		select {
		case <-blocked.started:
		case <-time.After(5 * time.Second):
			t.Fatal("two contest handlers did not start")
		}
	}

	if err := waitFor(tenantCtx, func() (bool, error) {
		var state string
		err := f.db.QueryRow(tenantCtx, `select state from jobs where id = $1`, officialID).Scan(&state)
		return state == "completed", err
	}); err != nil {
		t.Fatal(err)
	}

	var running, pending int
	if err := f.db.QueryRow(tenantCtx, `select count(*) filter (where state = 'running'), count(*) filter (where state = 'pending')
		from jobs where id = any($1)`, contestIDs).Scan(&running, &pending); err != nil {
		t.Fatal(err)
	}
	if running != 2 || pending != 1 {
		t.Fatalf("contest claims: running=%d pending=%d; want 2 running, 1 pending", running, pending)
	}

	releaseAll()
	if err := waitFor(tenantCtx, func() (bool, error) {
		var completed int
		err := f.db.QueryRow(tenantCtx, `select count(*) from jobs where id = any($1) and state = 'completed'`, contestIDs).Scan(&completed)
		return completed == 3, err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerGlobalLimitLeavesDueRowsUnclaimed(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	tenantCtx = tenant.WithKey(t.Context(), f.key)
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	release := make(chan struct{})
	var released sync.Once
	releaseAll := func() { released.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)
	blocked := &blockingValkey{
		Client:  f.client,
		prefix:  f.prefix + "leaderboard:",
		started: make(chan struct{}, 6),
		release: release,
	}

	service := leaderboard.NewService(
		leaderboard.NewRepository(f.db),
		leaderboard.NewCache(blocked, 8*time.Second, ""),
	)
	application, err := NewApplication(jobqueue.NewService(jobqueue.NewRepository(f.db)), service, Config{
		Concurrency:     3,
		ShutdownTimeout: 2 * time.Second,
		Logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}

	startWorker(t, application)
	if err := waitFor(tenantCtx, func() (bool, error) {
		return application.Ready(), nil
	}); err != nil {
		t.Fatal(err)
	}

	ids := make([]int64, 0, 6)
	for range 3 {
		contestID, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateContestV1), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
		if err != nil {
			t.Fatal(err)
		}
		officialID, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":2025}`, false)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, contestID, officialID)
	}

	for range 3 {
		select {
		case <-blocked.started:
		case <-time.After(5 * time.Second):
			t.Fatal("three handlers did not start")
		}
	}

	var running, pending int
	if err := f.db.QueryRow(tenantCtx, `select count(*) filter (where state = 'running'), count(*) filter (where state = 'pending')
		from jobs where id = any($1)`, ids).Scan(&running, &pending); err != nil {
		t.Fatal(err)
	}
	if running != 3 || pending != 3 {
		t.Fatalf("global claims: running=%d pending=%d; want 3 running, 3 pending", running, pending)
	}

	releaseAll()
	if err := waitFor(tenantCtx, func() (bool, error) {
		var completed int
		err := f.db.QueryRow(tenantCtx, `select count(*) from jobs where id = any($1) and state = 'completed'`, ids).Scan(&completed)
		return completed == len(ids), err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerRetainsSlotUntilCanceledHandlerReturns(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	started := make(chan struct{}, 2)
	canceled := make(chan struct{}, 2)
	release := make(chan struct{})
	var released sync.Once
	releaseAll := func() { released.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)

	handlers, err := newRegistry(handle(func(ctx context.Context, _ jobs.InvalidateOfficialLeaderboardV1) error {
		started <- struct{}{}
		<-ctx.Done()
		canceled <- struct{}{}
		<-release
		return nil
	}, Policy{Concurrency: 1, Timeout: 50 * time.Millisecond, MaxAttempts: 1}))
	if err != nil {
		t.Fatal(err)
	}

	runtime := &runner{
		queue:           jobqueue.NewService(jobqueue.NewRepository(f.db)),
		handlers:        handlers,
		concurrency:     1,
		shutdownTimeout: time.Second,
		logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		metrics:         NewMetrics(prometheus.NewRegistry()),
	}

	first, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":2025}`, false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":2026}`, false)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() { defer close(done); runtime.run(ctx) }()
	t.Cleanup(func() {
		releaseAll()
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("worker did not stop")
		}
	})

	select {
	case <-canceled:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not receive deadline cancellation")
	}
	select {
	case <-started:
	default:
		t.Fatal("first handler did not start")
	}

	runtime.ready.Store(false)
	if err := waitFor(tenantCtx, func() (bool, error) { return runtime.ready.Load(), nil }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
		t.Fatal("worker reused the canceled handler's occupied slot")
	default:
	}

	var running, pending int
	if err := f.db.QueryRow(tenantCtx, `select count(*) filter (where state = 'running'), count(*) filter (where state = 'pending') from jobs where id = any($1)`, []int64{first, second}).Scan(&running, &pending); err != nil {
		t.Fatal(err)
	}
	if running != 1 || pending != 1 {
		t.Fatalf("canceled handler claims: running=%d pending=%d", running, pending)
	}

	releaseAll()
	if err := waitFor(tenantCtx, func() (bool, error) {
		var failed int
		err := f.db.QueryRow(tenantCtx, `select count(*) from jobs where id = any($1) and state = 'failed' and last_error = 'deadline_exceeded'`, []int64{first, second}).Scan(&failed)
		return failed == 2, err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerRenewedDeadlineSchedulesRetry(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	handlers, err := newRegistry(handle(func(ctx context.Context, _ jobs.InvalidateOfficialLeaderboardV1) error {
		<-ctx.Done()
		return ctx.Err()
	}, Policy{Concurrency: 1, Timeout: 4 * time.Second, MaxAttempts: 2}))
	if err != nil {
		t.Fatal(err)
	}

	repository := jobqueue.NewRepository(f.db)
	runtime := &runner{
		queue:    jobqueue.NewService(repository),
		handlers: handlers,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		metrics:  NewMetrics(prometheus.NewRegistry()),
	}

	id, err := insertJob(tenantCtx, f.db, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":2025}`, false)
	if err != nil {
		t.Fatal(err)
	}

	spec := handlerSpec{
		typeName:    jobs.LeaderboardInvalidateOfficialV1,
		limit:       1,
		timeout:     4 * time.Second,
		lease:       3 * time.Second,
		maxAttempts: 2,
	}
	claims, err := repository.Claim(tenantCtx, spec.typeName, 1, spec.lease, spec.maxAttempts)
	if err != nil || len(claims) != 1 {
		t.Fatalf("claim deadline job: count=%d error=%v", len(claims), err)
	}

	if err := runtime.process(tenantCtx, claims[0], spec); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("deadline handler error = %v", err)
	}

	var state, code string
	if err := f.db.QueryRow(tenantCtx, `select state, coalesce(last_error, '') from jobs where id = $1`, id).Scan(&state, &code); err != nil {
		t.Fatal(err)
	}
	if state != "pending" || code != "deadline_exceeded" {
		t.Fatalf("renewed deadline did not schedule retry: state=%q code=%q", state, code)
	}
}

func TestWorkerCleansUpExpiredCompletedJobsAtStartup(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	var id int64
	if err := f.db.QueryRow(tenantCtx, `
		insert into jobs (tenant, task_type, payload, state, created_at, completed_at)
		values ('tadoku/prod', $1, '{"year":2025}', 'completed', now() - interval '1 year', now() - interval '1 year')
		returning id`,
		string(jobs.LeaderboardInvalidateOfficialV1),
	).Scan(&id); err != nil {
		t.Fatal(err)
	}

	startWorker(t, f.runner(t, f.client, time.Second, time.Second))
	if err := waitFor(tenantCtx, func() (bool, error) {
		var exists bool
		err := f.db.QueryRow(tenantCtx, `select exists(select 1 from jobs where id = $1)`, id).Scan(&exists)
		return !exists, err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerCleanupRetainsThreeMonthsAndFailures(t *testing.T) {
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	f, fixtureErr := newWorkerFixture(tenantCtx)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	runtime := &runner{
		queue:  jobqueue.NewService(jobqueue.NewRepository(f.db)),
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	now := time.Date(2026, time.May, 31, 12, 0, 0, 0, time.UTC)
	recent := time.Date(2026, time.April, 30, 12, 0, 0, 0, time.UTC)
	old := now.AddDate(-1, 0, 0)
	var recentID, oldID, failedID, replayID, pendingID, runningID int64

	for _, item := range []struct {
		id        *int64
		completed time.Time
	}{{&recentID, recent}, {&oldID, old}} {
		if err := f.db.QueryRow(tenantCtx, `
			insert into jobs (tenant, task_type, payload, state, created_at, completed_at)
			values ('tadoku/prod', $1, '{"year":2025}', 'completed', $2, $2)
			returning id`,
			string(jobs.LeaderboardInvalidateOfficialV1), item.completed,
		).Scan(item.id); err != nil {
			t.Fatal(err)
		}
	}

	if err := f.db.QueryRow(tenantCtx, `
		insert into jobs (tenant, task_type, payload, state, created_at, failed_at)
		values ('tadoku/prod', $1, '{"year":2025}', 'failed', $2, $2)
		returning id`,
		string(jobs.LeaderboardInvalidateOfficialV1), old,
	).Scan(&failedID); err != nil {
		t.Fatal(err)
	}

	if err := f.db.QueryRow(tenantCtx, `
		insert into jobs (
			tenant, task_type, payload, state, created_at, completed_at,
			replay_of_id, replay_actor, replay_reason
		)
		values ('tadoku/prod', $1, '{"year":2025}', 'completed', $2, $2, $3, 'retention-test', 'repaired')
		returning id`,
		string(jobs.LeaderboardInvalidateOfficialV1), old, failedID,
	).Scan(&replayID); err != nil {
		t.Fatal(err)
	}

	if _, err := f.db.Exec(tenantCtx, `
		insert into jobs (tenant, task_type, payload, state, created_at, completed_at)
		select 'tadoku/prod', $1, '{"year":2025}'::jsonb, 'completed', $2, $2
		from generate_series(1, 101)`,
		string(jobs.LeaderboardInvalidateOfficialV1), old,
	); err != nil {
		t.Fatal(err)
	}

	if err := f.db.QueryRow(tenantCtx, `
		insert into jobs (tenant, task_type, payload, created_at)
		values ('tadoku/prod', $1, '{"year":2025}', $2)
		returning id`,
		string(jobs.LeaderboardInvalidateOfficialV1), old,
	).Scan(&pendingID); err != nil {
		t.Fatal(err)
	}

	if err := f.db.QueryRow(tenantCtx, `
		insert into jobs (tenant, task_type, payload, state, created_at, claim_token, lease_expires_at)
		values ('tadoku/prod', $1, '{"year":2025}', 'running', $2, $3, now() + interval '1 hour')
		returning id`,
		string(jobs.LeaderboardInvalidateOfficialV1), old, uuid.New(),
	).Scan(&runningID); err != nil {
		t.Fatal(err)
	}

	timex.TheWorld(now, func() { runtime.cleanupCompleted(tenantCtx) })

	var expired int
	if err := f.db.QueryRow(tenantCtx, `select count(*) from jobs where state = 'completed' and completed_at = $1`, old).Scan(&expired); err != nil {
		t.Fatal(err)
	}
	if expired != 0 {
		t.Errorf("completed jobs remaining after batched cleanup = %d; want 0", expired)
	}

	for _, item := range []struct {
		name   string
		id     int64
		retain bool
	}{
		{"one-month-old success", recentID, true},
		{"expired success", oldID, false},
		{"failed original", failedID, true},
		{"expired successful replay", replayID, false},
		{"pending job", pendingID, true},
		{"running job", runningID, true},
	} {
		var exists bool
		if err := f.db.QueryRow(tenantCtx, `select exists(select 1 from jobs where id = $1)`, item.id).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists != item.retain {
			t.Errorf("%s retained=%t; want %t", item.name, exists, item.retain)
		}
	}
}
