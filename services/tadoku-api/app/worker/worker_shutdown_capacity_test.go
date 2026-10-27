package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

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

	application, err := NewApplication(
		workerDependencies(f.db, leaderboard.NewCache(blocked, 8*time.Second)),
		Config{
			Scope:           baseWorkerScope(t),
			Concurrency:     3,
			ShutdownTimeout: 2 * time.Second,
			Logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
	)
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
	}, policy{Concurrency: 1, Timeout: 50 * time.Millisecond, MaxAttempts: 1}))
	if err != nil {
		t.Fatal(err)
	}

	deps := workerDependencies(f.db, nil)
	runtime := &runner{
		scope:           baseWorkerScope(t),
		queue:           deps.JobQueue,
		email:           deps.Email,
		identity:        deps.IdentityExpiry,
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
