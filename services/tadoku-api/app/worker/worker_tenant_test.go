package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestWorkerRetainsTenantAcrossGracefulShutdown(t *testing.T) {
	f, err := newWorkerFixture(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})
	application := f.runner(t, f.client, time.Second, 2*time.Second)
	release := make(chan struct{})
	observed := make(chan error, 1)
	type requestKey struct{}
	handlers, err := newRegistry(handle(func(ctx context.Context, _ jobs.InvalidateOfficialLeaderboardV1) error {
		key, ok := tenant.FromContext(ctx)
		if !ok || key != tenant.Production() || ctx.Value(requestKey{}) != "worker-context" {
			observed <- fmt.Errorf("handler lost tenant/context: tenant=%q valid=%t value=%v", key.String(), ok, ctx.Value(requestKey{}))
		} else {
			executor, err := postgres.Executor(ctx, f.db)
			if err == nil {
				var actual string
				err = executor.QueryRow(ctx, "select coalesce(current_setting('tadoku.tenant', true), '')").Scan(&actual)
				if err == nil && actual != key.String() {
					err = fmt.Errorf("handler SQL tenant=%q, want %q", actual, key.String())
				}
			}
			observed <- err
		}
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}, Policy{Concurrency: 1, Timeout: 5 * time.Second, MaxAttempts: 1}))
	if err != nil {
		t.Fatal(err)
	}
	application.runner.handlers = handlers
	ctx := context.WithValue(t.Context(), requestKey{}, "worker-context")
	id, err := insertJob(tenant.WithKey(ctx, tenant.Production()), f.db, string(jobs.LeaderboardInvalidateOfficialV1), `{"year":2025}`, false)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); _ = application.Run(ctx) }()
	select {
	case err := <-observed:
		if err != nil {
			t.Error(err)
		}
	case <-time.After(5 * time.Second):
		cancel()
		close(release)
		<-done
		t.Fatal("worker did not dispatch the due job")
	}
	cancel()
	close(release)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not finish graceful shutdown")
	}
	var state string
	if err := f.db.QueryRow(t.Context(), "select state from jobs where id = $1", id).Scan(&state); err != nil || state != "completed" {
		t.Errorf("detached completion state=%q error=%v", state, err)
	}
}
