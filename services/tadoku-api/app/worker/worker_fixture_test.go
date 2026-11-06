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
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contestevents"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant/alltenants"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

type workerFixture struct {
	database *testpostgres.Database
	db       *pgxpool.Pool
	dsn      string
	key      tenant.Key
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

	return workerFixture{
		database: database,
		db:       database.Pool,
		dsn:      database.DSN,
		key:      key,
	}, nil
}

func (f workerFixture) Close() error { return f.database.Close() }

// runner runs contest for leaderboard.invalidate_contest.v1 under its production policy.
func (f workerFixture) runner(
	t *testing.T,
	contest func(context.Context, jobs.InvalidateContestLeaderboardV1) error,
	shutdown time.Duration,
) *Application {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application, err := NewApplication(
		workerDependencies(t, f.db),
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

	handleLeaderboard(t, application, contest, ignoreJob[jobs.InvalidateOfficialLeaderboardV1])
	return application
}

// handleLeaderboard makes application run contest and official for the leaderboard invalidation types under their
// production policies.
func handleLeaderboard(
	t *testing.T,
	application *Application,
	contest func(context.Context, jobs.InvalidateContestLeaderboardV1) error,
	official func(context.Context, jobs.InvalidateOfficialLeaderboardV1) error,
) {
	t.Helper()

	handlers, err := registrations(contest, official, application.SendEmail, application.DeleteAccount)
	if err != nil {
		t.Fatal(err)
	}
	application.runner.handlers = handlers
	application.runner.metrics.initialize(handlers)
}

// workerDependencies gives email no sender, so the worker holds every message.
func workerDependencies(t *testing.T, pool *pgxpool.Pool) Dependencies {
	t.Helper()

	model, err := permissionmodel.New()
	if err != nil {
		t.Fatal(err)
	}

	return Dependencies{
		DB:               pool,
		JobQueue:         jobqueue.NewService(jobqueue.NewRepository(pool)),
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

// blockingHandler holds each job it runs until release closes, which completes the job, or until the job's context
// ends, which signals canceled when it is set.
type blockingHandler struct {
	started  chan struct{}
	release  <-chan struct{}
	canceled chan struct{}
}

func blockingJob[J jobs.Job](h *blockingHandler) func(context.Context, J) error {
	return func(ctx context.Context, _ J) error {
		select {
		case h.started <- struct{}{}:
		default:
		}

		select {
		case <-h.release:
			return nil
		case <-ctx.Done():
			if h.canceled != nil {
				select {
				case h.canceled <- struct{}{}:
				default:
				}
			}
			return ctx.Err()
		}
	}
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
