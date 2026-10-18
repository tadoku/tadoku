package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	valkeygo "github.com/valkey-io/valkey-go"
)

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
		leaderboard.NewCache(blocked, 8*time.Second),
	)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repository := jobqueue.NewRepository(limitedPool)
	runner, err := NewApplication(
		jobqueue.NewService(repository),
		service,
		Config{
			Scope:           baseWorkerScope(t),
			Concurrency:     4,
			Logger:          logger,
			Metrics:         NewMetrics(prometheus.NewRegistry()),
			ShutdownTimeout: time.Second,
		},
	)
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
	}, policy{Concurrency: 1, Timeout: 4 * time.Second, MaxAttempts: 2}))
	if err != nil {
		t.Fatal(err)
	}

	repository := jobqueue.NewRepository(f.db)
	runtime := &runner{
		scope:    baseWorkerScope(t),
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
