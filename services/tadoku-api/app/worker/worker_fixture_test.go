package worker

import (
	"context"
	"errors"
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
	"github.com/tadoku/tadoku/services/tadoku-api/app/permissionmodel"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contestevents"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant/alltenants"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
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
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application, err := NewApplication(
		workerDependencies(t, f.db, leaderboard.NewCache(client, providerTimeout)),
		Config{
			Scope:           baseWorkerScope(t),
			Concurrency:     4,
			Logger:          logger,
			Metrics:         NewMetrics(prometheus.NewRegistry()),
			ShutdownTimeout: shutdown,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return application
}

// workerDependencies gives email no sender, so the worker holds every message.
func workerDependencies(t *testing.T, pool *pgxpool.Pool, cache *leaderboard.Cache) Dependencies {
	t.Helper()

	model, err := permissionmodel.New()
	if err != nil {
		t.Fatal(err)
	}

	return Dependencies{
		DB:               pool,
		JobQueue:         jobqueue.NewService(jobqueue.NewRepository(pool)),
		Leaderboard:      leaderboard.NewService(leaderboard.NewRepository(pool), cache),
		Email:            email.NewService(email.NewRepository(pool), nil),
		IdentityExpiry:   identity.NewExpiry(identity.NewRepository(pool)),
		IdentityDeletion: identity.NewDeletion(identity.NewRepository(pool)),
		Profile:          profile.NewService(profile.NewRepository(pool), identity.NewReader(pool)),
		Logs:             logs.NewService(logs.NewLogsRepository(pool), false),
		Contests:         contests.NewService(contests.NewContestsRepository(pool)),
		Relationships:    permissions.NewRelationships(authz.NewRepository(pool), model),
		Audit:            audit.NewService(audit.NewRepository(pool)),
		ContestEvents:    contestevents.NewService(contestevents.NewRepository(pool)),
	}
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
