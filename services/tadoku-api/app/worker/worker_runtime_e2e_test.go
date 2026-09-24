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
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/asyncwork"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	"github.com/tadoku/tadoku/services/tadoku-api/storage/postgres/asyncoutbox"
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

func (f workerFixture) runner(client valkeygo.Client, providerTimeout, shutdown time.Duration) *Runner {
	service := leaderboard.NewService(leaderboard.NewRepository(f.db), client, providerTimeout, f.prefix)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRunner(asyncoutbox.NewRepository(f.db), NewApplication(service), service, logger, NewMetrics(prometheus.NewRegistry()), shutdown)
}

func startWorker(t *testing.T, runner *Runner) {
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
	startWorker(t, f.runner(blocked, 8*time.Second, 2*time.Second))
	waitFor(t, func() (bool, error) {
		return serviceCacheReady(t.Context(), f.client, f.prefix+"leaderboard:ready")
	})
	id := insertTask(t, f.db, string(asyncwork.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	select {
	case <-blocked.started:
	case <-time.After(5 * time.Second):
		t.Fatal("contest handler did not start")
	}
	newToken := uuid.New()
	if _, err := f.db.Exec(t.Context(), `update async_outbox set claim_token = $1, lease_expires_at = now() + interval '1 minute'
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
	if err := f.db.QueryRow(t.Context(), `select state, claim_token from async_outbox where id = $1`, id).Scan(&state, &token); err != nil {
		t.Fatal(err)
	}
	if state != "running" || token != newToken {
		t.Fatalf("lost claim was transitioned: state=%s token=%s", state, token)
	}
	releaseAll()
	if _, err := f.db.Exec(t.Context(), `update async_outbox set lease_expires_at = now() - interval '1 second' where id = $1`, id); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() (bool, error) {
		var state string
		var attempts int
		err := f.db.QueryRow(t.Context(), `select state, attempts from async_outbox where id = $1`, id).Scan(&state, &attempts)
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
	runner := f.runner(blocked, 8*time.Second, 2*time.Second)
	var id int64
	err := f.db.QueryRow(t.Context(), `insert into async_outbox (task_type, payload, attempts)
		values ($1, $2::jsonb, 4) returning id`, string(asyncwork.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString())).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	repository := asyncoutbox.NewRepository(f.db)
	tasks, err := repository.Claim(t.Context(), asyncwork.InvalidateContest, 1, time.Second, 5)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("claim exhausted task: tasks=%d error=%v", len(tasks), err)
	}
	spec := policy{typeName: asyncwork.InvalidateContest, limit: 2, timeout: 50 * time.Millisecond, maxRuntime: time.Second, lease: time.Second, maxAttempts: 5}
	if err := runner.process(t.Context(), tasks[0], spec); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("deadline handler error = %v", err)
	}
	var state, code string
	var attempts int
	if err := f.db.QueryRow(t.Context(), `select state, attempts, last_error from async_outbox where id = $1`, id).Scan(&state, &attempts, &code); err != nil {
		t.Fatal(err)
	}
	if state != "failed" || attempts != 5 || code != "deadline_exceeded" {
		t.Fatalf("exhausted task: state=%q attempts=%d code=%q", state, attempts, code)
	}
	replayedID, err := repository.Replay(t.Context(), id, "worker-e2e", "deadline repaired")
	if err != nil {
		t.Fatal(err)
	}
	releaseAll()
	startWorker(t, runner)
	waitFor(t, func() (bool, error) {
		var replayState string
		err := f.db.QueryRow(t.Context(), `select state from async_outbox where id = $1`, replayedID).Scan(&replayState)
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
	runner := f.runner(blocked, 8*time.Second, 100*time.Millisecond)
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
	id := insertTask(t, f.db, string(asyncwork.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
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
	if err := f.db.QueryRow(t.Context(), `select state from async_outbox where id = $1`, id).Scan(&state); err != nil {
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
	runner := f.runner(observed, time.Second, time.Second)
	id := insertTask(t, f.db, string(asyncwork.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	tasks, err := asyncoutbox.NewRepository(f.db).Claim(t.Context(), asyncwork.InvalidateContest, 1, 50*time.Millisecond, 5)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("claim: tasks=%d error=%v", len(tasks), err)
	}
	<-time.After(80 * time.Millisecond)
	spec := policy{typeName: asyncwork.InvalidateContest, limit: 2, timeout: time.Second, maxRuntime: time.Second, lease: 50 * time.Millisecond, maxAttempts: 5}
	if err := runner.process(t.Context(), tasks[0], spec); err == nil {
		t.Error("expired claim was dispatched")
	}
	select {
	case <-observed.started:
		t.Error("Valkey invalidation began after lease expiry")
	default:
	}
	var state string
	if err := f.db.QueryRow(t.Context(), `select state from async_outbox where id = $1`, id).Scan(&state); err != nil {
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
	repository := asyncoutbox.NewRepository(limitedPool)
	runner := NewRunner(repository, NewApplication(service), service, logger, NewMetrics(prometheus.NewRegistry()), time.Second)
	id := insertTask(t, f.db, string(asyncwork.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	tasks, err := repository.Claim(t.Context(), asyncwork.InvalidateContest, 1, 3*time.Second, 5)
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
	spec := policy{typeName: asyncwork.InvalidateContest, limit: 2, timeout: 8 * time.Second, maxRuntime: 8 * time.Second, lease: 3 * time.Second, maxAttempts: 5}
	done := make(chan error, 1)
	go func() { done <- runner.process(t.Context(), tasks[0], spec) }()
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
	if err := f.db.QueryRow(t.Context(), `select state from async_outbox where id = $1`, id).Scan(&state); err != nil {
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
	startWorker(t, f.runner(blocked, 8*time.Second, 2*time.Second))
	waitFor(t, func() (bool, error) {
		return serviceCacheReady(t.Context(), f.client, f.prefix+"leaderboard:ready")
	})

	contestIDs := make([]int64, 3)
	for i := range contestIDs {
		contestIDs[i] = insertTask(t, f.db, string(asyncwork.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false)
	}
	officialID := insertTask(t, f.db, string(asyncwork.InvalidateOfficial), `{"year":2025}`, false)
	for range 2 {
		select {
		case <-blocked.started:
		case <-time.After(5 * time.Second):
			t.Fatal("two contest handlers did not start")
		}
	}
	waitFor(t, func() (bool, error) {
		var state string
		err := f.db.QueryRow(t.Context(), `select state from async_outbox where id = $1`, officialID).Scan(&state)
		return state == "completed", err
	})
	var running, pending int
	if err := f.db.QueryRow(t.Context(), `select count(*) filter (where state = 'running'), count(*) filter (where state = 'pending')
		from async_outbox where id = any($1)`, contestIDs).Scan(&running, &pending); err != nil {
		t.Fatal(err)
	}
	if running != 2 || pending != 1 {
		t.Fatalf("contest claims: running=%d pending=%d; want 2 running, 1 pending", running, pending)
	}
	releaseAll()
	waitFor(t, func() (bool, error) {
		var completed int
		err := f.db.QueryRow(t.Context(), `select count(*) from async_outbox where id = any($1) and state = 'completed'`, contestIDs).Scan(&completed)
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
	startWorker(t, f.runner(blocked, 8*time.Second, 2*time.Second))
	waitFor(t, func() (bool, error) {
		return serviceCacheReady(t.Context(), f.client, f.prefix+"leaderboard:ready")
	})
	ids := make([]int64, 0, 6)
	for range 3 {
		ids = append(ids, insertTask(t, f.db, string(asyncwork.InvalidateContest), fmt.Sprintf(`{"contest_id":%q}`, uuid.NewString()), false))
		ids = append(ids, insertTask(t, f.db, string(asyncwork.InvalidateOfficial), `{"year":2025}`, false))
	}
	for range 4 {
		select {
		case <-blocked.started:
		case <-time.After(5 * time.Second):
			t.Fatal("four handlers did not start")
		}
	}
	var running, pending int
	if err := f.db.QueryRow(t.Context(), `select count(*) filter (where state = 'running'), count(*) filter (where state = 'pending')
		from async_outbox where id = any($1)`, ids).Scan(&running, &pending); err != nil {
		t.Fatal(err)
	}
	if running != 4 || pending != 2 {
		t.Fatalf("global claims: running=%d pending=%d; want 4 running, 2 pending", running, pending)
	}
	releaseAll()
	waitFor(t, func() (bool, error) {
		var completed int
		err := f.db.QueryRow(t.Context(), `select count(*) from async_outbox where id = any($1) and state = 'completed'`, ids).Scan(&completed)
		return completed == len(ids), err
	})
}
