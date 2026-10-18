package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant/alltenants"
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
			observed <- fmt.Errorf(
				"handler lost tenant/context: tenant=%q valid=%t value=%v",
				key.String(), ok, ctx.Value(requestKey{}),
			)
		} else {
			executor, err := postgres.Executor(ctx, f.db)
			if err == nil {
				var actual string
				err = executor.QueryRow(
					ctx,
					"select coalesce(current_setting('tadoku.tenant', true), '')",
				).Scan(&actual)
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
	}, policy{Concurrency: 1, Timeout: 5 * time.Second, MaxAttempts: 1}))
	if err != nil {
		t.Fatal(err)
	}
	application.runner.handlers = handlers

	ctx := context.WithValue(t.Context(), requestKey{}, "worker-context")
	id, err := insertJob(
		tenant.WithKey(ctx, tenant.Production()),
		f.db,
		string(jobs.LeaderboardInvalidateOfficialV1),
		`{"year":2025}`,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = application.Run(ctx)
	}()

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
	if err := f.db.QueryRow(
		t.Context(),
		"select state from jobs where id = $1",
		id,
	).Scan(&state); err != nil || state != "completed" {
		t.Errorf("detached completion state=%q error=%v", state, err)
	}
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
			}, policy{Concurrency: 1, Timeout: time.Second, MaxAttempts: 3}))
			if err != nil {
				t.Fatal(err)
			}

			runtime := &runner{
				scope:    baseWorkerScope(t),
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
