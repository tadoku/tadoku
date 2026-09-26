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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	valkeygo "github.com/valkey-io/valkey-go"
)

type workerFixture struct {
	db     *pgxpool.Pool
	dsn    string
	client valkeygo.Client
	prefix string
}

func newWorkerFixture(t *testing.T) workerFixture {
	t.Helper()
	database, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
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
	client, err := valkeygo.NewClient(option)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	prefix := "test:" + uuid.NewString() + ":"
	t.Cleanup(func() {
		ctx, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		var cursor uint64
		for {
			page, err := client.Do(ctx, client.B().Scan().Cursor(cursor).Match(prefix+"*").Count(100).Build()).AsScanEntry()
			if err != nil {
				t.Error(err)
				return
			}
			if len(page.Elements) > 0 {
				if err := client.Do(ctx, client.B().Del().Key(page.Elements...).Build()).Error(); err != nil {
					t.Error(err)
					return
				}
			}
			if page.Cursor == 0 {
				return
			}
			cursor = page.Cursor
		}
	})
	return workerFixture{db: database.Pool, dsn: database.DSN, client: client, prefix: prefix}
}

func (f workerFixture) runner(t *testing.T, client valkeygo.Client, providerTimeout, shutdown time.Duration) *Application {
	service := leaderboard.NewService(leaderboard.NewRepository(f.db), client, providerTimeout, f.prefix)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application, err := NewApplication(jobqueue.NewService(jobqueue.NewRepository(f.db)), service, Config{Logger: logger, Metrics: NewMetrics(prometheus.NewRegistry()), ShutdownTimeout: shutdown})
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

func TestWorkerCancelsHandlerAfterLostLeaseAndReclaims(t *testing.T) {
	f := newWorkerFixture(t)
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
	startWorker(t, f.runner(t, blocked, 8*time.Second, 2*time.Second))
	waitFor(t, func() (bool, error) {
		return serviceCacheReady(t.Context(), f.client, f.prefix+"leaderboard:ready")
	})
	id := insertTask(t, f.db, string(jobs.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	select {
	case <-blocked.started:
	case <-time.After(5 * time.Second):
		t.Fatal("contest handler did not start")
	}
	newToken := uuid.New()
	if _, err := f.db.Exec(t.Context(), `update jobs set claim_token = $1, lease_expires_at = now() + interval '1 minute'
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
	if err := f.db.QueryRow(t.Context(), `select state, claim_token from jobs where id = $1`, id).Scan(&state, &token); err != nil {
		t.Fatal(err)
	}
	if state != "running" || token != newToken {
		t.Fatalf("lost claim was transitioned: state=%s token=%s", state, token)
	}
	releaseAll()
	if _, err := f.db.Exec(t.Context(), `update jobs set lease_expires_at = now() - interval '1 second' where id = $1`, id); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() (bool, error) {
		var state string
		var attempts int
		err := f.db.QueryRow(t.Context(), `select state, attempts from jobs where id = $1`, id).Scan(&state, &attempts)
		return state == "completed" && attempts == 2, err
	})
}

func TestWorkerDeadlineExhaustionAndReplay(t *testing.T) {
	f := newWorkerFixture(t)
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
	err := f.db.QueryRow(t.Context(), `insert into jobs (task_type, payload, attempts)
		values ($1, $2::jsonb, 4) returning id`, string(jobs.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString())).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	repository := jobqueue.NewRepository(f.db)
	tasks, err := repository.Claim(t.Context(), jobs.InvalidateContest, 1, time.Second, 5)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("claim exhausted task: tasks=%d error=%v", len(tasks), err)
	}
	spec := policy{typeName: jobs.InvalidateContest, limit: 2, timeout: 50 * time.Millisecond, lease: time.Second, maxAttempts: 5}
	if err := runner.runner.process(t.Context(), tasks[0], spec); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("deadline handler error = %v", err)
	}
	var state, code string
	var attempts int
	if err := f.db.QueryRow(t.Context(), `select state, attempts, last_error from jobs where id = $1`, id).Scan(&state, &attempts, &code); err != nil {
		t.Fatal(err)
	}
	if state != "failed" || attempts != 5 || code != "deadline_exceeded" {
		t.Fatalf("exhausted task: state=%q attempts=%d code=%q", state, attempts, code)
	}
	replayedID, err := Replay(t.Context(), jobqueue.NewService(repository), id, "worker-e2e", "deadline repaired")
	if err != nil {
		t.Fatal(err)
	}
	releaseAll()
	startWorker(t, runner)
	waitFor(t, func() (bool, error) {
		var replayState string
		err := f.db.QueryRow(t.Context(), `select state from jobs where id = $1`, replayedID).Scan(&replayState)
		return replayState == "completed", err
	})
}

func TestWorkerShutdownCancelsActiveTaskAndRevokesReadiness(t *testing.T) {
	f := newWorkerFixture(t)
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
	waitFor(t, func() (bool, error) {
		return serviceCacheReady(t.Context(), f.client, f.prefix+"leaderboard:ready")
	})
	id := insertTask(t, f.db, string(jobs.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
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
	if err := f.db.QueryRow(t.Context(), `select state from jobs where id = $1`, id).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state == "completed" {
		t.Error("canceled task was completed")
	}
	ready, err := serviceCacheReady(t.Context(), f.client, f.prefix+"leaderboard:ready")
	if err != nil || ready {
		t.Errorf("cache readiness after shutdown = %t, %v; want false", ready, err)
	}
}

func TestWorkerDoesNotDispatchAfterClaimLeaseExpires(t *testing.T) {
	f := newWorkerFixture(t)
	release := make(chan struct{})
	close(release)
	observed := &blockingValkey{
		Client:  f.client,
		prefix:  f.prefix + "leaderboard:contest:",
		started: make(chan struct{}, 1),
		release: release,
	}
	runner := f.runner(t, observed, time.Second, time.Second)
	id := insertTask(t, f.db, string(jobs.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	tasks, err := jobqueue.NewRepository(f.db).Claim(t.Context(), jobs.InvalidateContest, 1, 50*time.Millisecond, 5)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("claim: tasks=%d error=%v", len(tasks), err)
	}
	<-time.After(80 * time.Millisecond)
	spec := policy{typeName: jobs.InvalidateContest, limit: 2, timeout: time.Second, lease: 50 * time.Millisecond, maxAttempts: 5}
	if err := runner.runner.process(t.Context(), tasks[0], spec); err == nil {
		t.Error("expired claim was dispatched")
	}
	select {
	case <-observed.started:
		t.Error("Valkey invalidation began after lease expiry")
	default:
	}
	var state string
	if err := f.db.QueryRow(t.Context(), `select state from jobs where id = $1`, id).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "running" {
		t.Errorf("expired claim state = %s; want running for recovery", state)
	}
}

func TestWorkerCompletesWhenRenewalIsCanceledByFinishedHandler(t *testing.T) {
	f := newWorkerFixture(t)
	poolConfig, err := pgxpool.ParseConfig(f.dsn)
	if err != nil {
		t.Fatal(err)
	}
	poolConfig.MaxConns = 1
	limitedPool, err := pgxpool.NewWithConfig(t.Context(), poolConfig)
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
	service := leaderboard.NewService(leaderboard.NewRepository(limitedPool), blocked, 8*time.Second, f.prefix)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repository := jobqueue.NewRepository(limitedPool)
	runner, err := NewApplication(jobqueue.NewService(repository), service, Config{Logger: logger, Metrics: NewMetrics(prometheus.NewRegistry()), ShutdownTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	id := insertTask(t, f.db, string(jobs.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	tasks, err := repository.Claim(t.Context(), jobs.InvalidateContest, 1, 3*time.Second, 5)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("claim: tasks=%d error=%v", len(tasks), err)
	}
	conn, err := limitedPool.Acquire(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var returned sync.Once
	returnConn := func() { returned.Do(conn.Release) }
	defer returnConn()
	baseline := limitedPool.Stat().CanceledAcquireCount()
	spec := policy{typeName: jobs.InvalidateContest, limit: 2, timeout: 8 * time.Second, lease: 3 * time.Second, maxAttempts: 5}
	done := make(chan error, 1)
	go func() { done <- runner.runner.process(t.Context(), tasks[0], spec) }()
	select {
	case <-blocked.started:
	case <-time.After(3 * time.Second):
		t.Fatal("handler did not begin")
	}
	<-time.After(1200 * time.Millisecond)
	releaseAll()
	<-time.After(50 * time.Millisecond)
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
	if err := f.db.QueryRow(t.Context(), `select state from jobs where id = $1`, id).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "completed" {
		t.Errorf("task state = %s; want completed", state)
	}
}

func TestWorkerFairClaimsWithoutPrefetch(t *testing.T) {
	f := newWorkerFixture(t)
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
	startWorker(t, f.runner(t, blocked, 8*time.Second, 2*time.Second))
	waitFor(t, func() (bool, error) {
		return serviceCacheReady(t.Context(), f.client, f.prefix+"leaderboard:ready")
	})

	contestIDs := make([]int64, 3)
	for i := range contestIDs {
		contestIDs[i] = insertTask(t, f.db, string(jobs.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	}
	officialID := insertTask(t, f.db, string(jobs.InvalidateOfficial), `{"year":2025}`, false)
	for range 2 {
		select {
		case <-blocked.started:
		case <-time.After(5 * time.Second):
			t.Fatal("two contest handlers did not start")
		}
	}
	waitFor(t, func() (bool, error) {
		var state string
		err := f.db.QueryRow(t.Context(), `select state from jobs where id = $1`, officialID).Scan(&state)
		return state == "completed", err
	})
	var running, pending int
	if err := f.db.QueryRow(t.Context(), `select count(*) filter (where state = 'running'), count(*) filter (where state = 'pending')
		from jobs where id = any($1)`, contestIDs).Scan(&running, &pending); err != nil {
		t.Fatal(err)
	}
	if running != 2 || pending != 1 {
		t.Fatalf("contest claims: running=%d pending=%d; want 2 running, 1 pending", running, pending)
	}
	releaseAll()
	waitFor(t, func() (bool, error) {
		var completed int
		err := f.db.QueryRow(t.Context(), `select count(*) from jobs where id = any($1) and state = 'completed'`, contestIDs).Scan(&completed)
		return completed == 3, err
	})
}

func TestWorkerGlobalLimitLeavesDueRowsUnclaimed(t *testing.T) {
	f := newWorkerFixture(t)
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
	service := leaderboard.NewService(leaderboard.NewRepository(f.db), blocked, 8*time.Second, f.prefix)
	application, err := NewApplication(jobqueue.NewService(jobqueue.NewRepository(f.db)), service, Config{Concurrency: 3, ShutdownTimeout: 2 * time.Second, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	if err != nil {
		t.Fatal(err)
	}
	startWorker(t, application)
	waitFor(t, func() (bool, error) {
		return serviceCacheReady(t.Context(), f.client, f.prefix+"leaderboard:ready")
	})
	ids := make([]int64, 0, 6)
	for range 3 {
		ids = append(ids, insertTask(t, f.db, string(jobs.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false))
		ids = append(ids, insertTask(t, f.db, string(jobs.InvalidateOfficial), `{"year":2025}`, false))
	}
	for range 3 {
		select {
		case <-blocked.started:
		case <-time.After(5 * time.Second):
			t.Fatal("three handlers did not start")
		}
	}
	var running, pending int
	if err := f.db.QueryRow(t.Context(), `select count(*) filter (where state = 'running'), count(*) filter (where state = 'pending')
		from jobs where id = any($1)`, ids).Scan(&running, &pending); err != nil {
		t.Fatal(err)
	}
	if running != 3 || pending != 3 {
		t.Fatalf("global claims: running=%d pending=%d; want 3 running, 3 pending", running, pending)
	}
	releaseAll()
	waitFor(t, func() (bool, error) {
		var completed int
		err := f.db.QueryRow(t.Context(), `select count(*) from jobs where id = any($1) and state = 'completed'`, ids).Scan(&completed)
		return completed == len(ids), err
	})
}

func TestWorkerRetainsSlotUntilCanceledHandlerReturns(t *testing.T) {
	f := newWorkerFixture(t)
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
	first := insertTask(t, f.db, string(jobs.InvalidateOfficial), `{"year":2025}`, false)
	second := insertTask(t, f.db, string(jobs.InvalidateOfficial), `{"year":2026}`, false)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() { defer close(done); runtime.run(ctx, func(context.Context, bool, bool) {}) }()
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
	select {
	case <-started:
		t.Fatal("worker reused the canceled handler's occupied slot")
	case <-time.After(700 * time.Millisecond):
	}
	var running, pending int
	if err := f.db.QueryRow(t.Context(), `select count(*) filter (where state = 'running'), count(*) filter (where state = 'pending') from jobs where id = any($1)`, []int64{first, second}).Scan(&running, &pending); err != nil {
		t.Fatal(err)
	}
	if running != 1 || pending != 1 {
		t.Fatalf("canceled handler claims: running=%d pending=%d", running, pending)
	}
	releaseAll()
	waitFor(t, func() (bool, error) {
		var failed int
		err := f.db.QueryRow(t.Context(), `select count(*) from jobs where id = any($1) and state = 'failed' and last_error = 'deadline_exceeded'`, []int64{first, second}).Scan(&failed)
		return failed == 2, err
	})
}

func TestWorkerRenewedDeadlineSchedulesRetry(t *testing.T) {
	f := newWorkerFixture(t)
	handlers, err := newRegistry(handle(func(ctx context.Context, _ jobs.InvalidateOfficialLeaderboardV1) error {
		<-ctx.Done()
		return ctx.Err()
	}, Policy{Concurrency: 1, Timeout: 400 * time.Millisecond, MaxAttempts: 2}))
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
	id := insertTask(t, f.db, string(jobs.InvalidateOfficial), `{"year":2025}`, false)
	spec := policy{typeName: jobs.InvalidateOfficial, limit: 1, timeout: 400 * time.Millisecond, lease: 300 * time.Millisecond, maxAttempts: 2}
	claims, err := repository.Claim(t.Context(), spec.typeName, 1, spec.lease, spec.maxAttempts)
	if err != nil || len(claims) != 1 {
		t.Fatalf("claim deadline job: count=%d error=%v", len(claims), err)
	}

	if err := runtime.process(t.Context(), claims[0], spec); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("deadline handler error = %v", err)
	}

	var state, code string
	if err := f.db.QueryRow(t.Context(), `select state, coalesce(last_error, '') from jobs where id = $1`, id).Scan(&state, &code); err != nil {
		t.Fatal(err)
	}
	if state != "pending" || code != "deadline_exceeded" {
		t.Fatalf("renewed deadline did not schedule retry: state=%q code=%q", state, code)
	}
}

func TestWorkerCleanupRetainsThreeMonthsAndFailures(t *testing.T) {
	f := newWorkerFixture(t)
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
		if err := f.db.QueryRow(t.Context(), `insert into jobs (task_type, payload, state, created_at, completed_at)
   values ($1, '{"year":2025}', 'completed', $2, $2) returning id`, string(jobs.InvalidateOfficial), item.completed).Scan(item.id); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.db.QueryRow(t.Context(), `insert into jobs (task_type, payload, state, created_at, failed_at)
  values ($1, '{"year":2025}', 'failed', $2, $2) returning id`, string(jobs.InvalidateOfficial), old).Scan(&failedID); err != nil {
		t.Fatal(err)
	}
	if err := f.db.QueryRow(t.Context(), `insert into jobs (task_type, payload, state, created_at, completed_at, replay_of_id, replay_actor, replay_reason)
  values ($1, '{"year":2025}', 'completed', $2, $2, $3, 'retention-test', 'repaired') returning id`, string(jobs.InvalidateOfficial), old, failedID).Scan(&replayID); err != nil {
		t.Fatal(err)
	}
	if err := f.db.QueryRow(t.Context(), `insert into jobs (task_type, payload, created_at)
  values ($1, '{"year":2025}', $2) returning id`, string(jobs.InvalidateOfficial), old).Scan(&pendingID); err != nil {
		t.Fatal(err)
	}
	if err := f.db.QueryRow(t.Context(), `insert into jobs (task_type, payload, state, created_at, claim_token, lease_expires_at)
  values ($1, '{"year":2025}', 'running', $2, $3, now() + interval '1 hour') returning id`, string(jobs.InvalidateOfficial), old, uuid.New()).Scan(&runningID); err != nil {
		t.Fatal(err)
	}

	timex.TheWorld(now, func() { runtime.cleanupCompleted(t.Context()) })

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
		if err := f.db.QueryRow(t.Context(), `select exists(select 1 from jobs where id = $1)`, item.id).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists != item.retain {
			t.Errorf("%s retained=%t; want %t", item.name, exists, item.retain)
		}
	}
}
