package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant/alltenants"
)

func oneConnectionPool(t *testing.T, tracer pgx.QueryTracer) *pgxpool.Pool {
	t.Helper()
	config, err := pgxpool.ParseConfig(disposableDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	config.MaxConns = 1
	config.ConnConfig.Tracer = tracer
	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
	return pool
}

func TestTransactionTenantEndsAtCommit(t *testing.T) {
	pool := oneConnectionPool(t, nil)
	key, err := tenant.Parse("e2e/transaction")
	if err != nil {
		t.Fatal(err)
	}
	ctx := tenant.WithKey(t.Context(), key)
	var before, during, after uint32
	if err := pool.QueryRow(ctx, "select pg_backend_pid()").Scan(&before); err != nil {
		t.Fatal(err)
	}
	err = postgres.RunInTransaction(ctx, pool, func(child context.Context) error {
		executor, err := postgres.Executor(child, pool)
		if err != nil {
			return err
		}
		var actual string
		if err := executor.QueryRow(child, "select coalesce(current_setting('tadoku.tenant', true), ''), pg_backend_pid()").Scan(&actual, &during); err != nil {
			return err
		}
		if actual != key.String() {
			return fmt.Errorf("transaction tenant=%q, want %q", actual, key.String())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var remaining string
	if err := pool.QueryRow(ctx, "select coalesce(current_setting('tadoku.tenant', true), ''), pg_backend_pid()").Scan(&remaining, &after); err != nil {
		t.Fatal(err)
	}
	if remaining != "" || before != during || during != after {
		t.Errorf("after commit tenant=%q, backend IDs=%d/%d/%d", remaining, before, during, after)
	}
}

func TestExecutorTenantStatementLifecycle(t *testing.T) {
	pool := oneConnectionPool(t, nil)
	ctx := tenant.WithKey(t.Context(), tenant.Production())
	if _, err := pool.Exec(ctx, "create temporary table tenant_scope_rows (tenant text not null)"); err != nil {
		t.Fatal(err)
	}
	executor, err := postgres.Executor(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	tag, err := executor.Exec(ctx, "insert into tenant_scope_rows select coalesce(current_setting('tadoku.tenant', true), '') from generate_series(1, 3)")
	if err != nil || tag.RowsAffected() != 3 || !tag.Insert() {
		t.Fatalf("insert tag=%s error=%v", tag, err)
	}
	rows, err := executor.Query(ctx, "select tenant, coalesce(current_setting('tadoku.tenant', true), '') from tenant_scope_rows")
	if err != nil {
		t.Fatal(err)
	}
	if !rows.Next() {
		t.Fatalf("first row: %v", rows.Err())
	}
	var stored, current string
	if err := rows.Scan(&stored, &current); err != nil {
		t.Fatal(err)
	}
	if stored != tenant.Production().String() || current != stored {
		t.Errorf("statement stored=%q current=%q", stored, current)
	}
	wait, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	connection, acquireErr := pool.Acquire(wait)
	cancel()
	if connection != nil {
		connection.Release()
	}
	if !errors.Is(acquireErr, context.DeadlineExceeded) {
		t.Errorf("unclosed streaming rows did not retain the connection: %v", acquireErr)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows, err = executor.Query(ctx, "select coalesce(current_setting('tadoku.tenant', true), '') from generate_series(1, 3)")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for rows.Next() {
		if err := rows.Scan(&current); err != nil {
			t.Fatal(err)
		}
		if current != tenant.Production().String() {
			t.Errorf("streamed row tenant=%q", current)
		}
		count++
	}
	if err := rows.Err(); err != nil || count != 3 {
		t.Fatalf("stream count=%d error=%v", count, err)
	}
	if err := executor.QueryRow(ctx, "select 1 where false").Scan(&count); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("empty QueryRow error=%v, want pgx.ErrNoRows", err)
	}
	if err := executor.QueryRow(ctx, "select coalesce(current_setting('tadoku.tenant', true), '')").Scan(&current); err != nil || current != tenant.Production().String() {
		t.Errorf("QueryRow tenant=%q error=%v", current, err)
	}
	if err := pool.QueryRow(ctx, "select coalesce(current_setting('tadoku.tenant', true), '')").Scan(&current); err != nil || current != "" {
		t.Errorf("after standalone statements tenant=%q error=%v", current, err)
	}
	if err := executor.QueryRow(ctx, "select 1").Scan(new(time.Time)); err == nil {
		t.Error("invalid scan accepted")
	}
	if err := pool.Ping(ctx); err != nil {
		t.Errorf("pool unusable after Scan failure: %v", err)
	}
}

type queryCounter struct{ queries atomic.Int64 }

func (counter *queryCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	counter.queries.Add(1)
	return ctx
}

func (*queryCounter) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func TestMissingTenantFailsBeforeSQL(t *testing.T) {
	counter := new(queryCounter)
	pool := oneConnectionPool(t, counter)
	counter.queries.Store(0)
	for _, ctx := range []context.Context{t.Context(), tenant.WithKey(t.Context(), tenant.Key{})} {
		if _, err := postgres.Executor(ctx, pool); err == nil || err.Error() != "postgres: no tenant" {
			t.Errorf("Executor missing tenant error=%v", err)
		}
		called := false
		err := postgres.RunInTransaction(ctx, pool, func(context.Context) error { called = true; return nil })
		if err == nil || err.Error() != "postgres: no tenant" || called {
			t.Errorf("transaction missing tenant error=%v callback=%t", err, called)
		}
	}
	if actual := counter.queries.Load(); actual != 0 {
		t.Errorf("missing tenant sent %d SQL statements", actual)
	}
}

func TestTransactionRejectsTenantChange(t *testing.T) {
	pool := oneConnectionPool(t, nil)
	key, err := tenant.Parse("e2e/other")
	if err != nil {
		t.Fatal(err)
	}
	ctx := tenant.WithKey(t.Context(), tenant.Production())
	err = postgres.RunInTransaction(ctx, pool, func(child context.Context) error {
		changed := tenant.WithKey(child, key)
		if _, err := postgres.Executor(changed, pool); err == nil || err.Error() != "postgres: tenant mismatch" {
			return fmt.Errorf("changed tenant accepted: %v", err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAllTenantsScope(t *testing.T) {
	pool := oneConnectionPool(t, nil)
	ctx := alltenants.With(t.Context())
	check := func(ctx context.Context) error {
		executor, err := postgres.Executor(ctx, pool)
		if err != nil {
			return err
		}
		var key, marker string
		if err := executor.QueryRow(ctx, "select coalesce(current_setting('tadoku.tenant', true), ''), coalesce(current_setting('tadoku.all_tenants', true), '')").Scan(&key, &marker); err != nil {
			return err
		}
		if key != "" || marker != "on" {
			return fmt.Errorf("all-tenants scope tenant=%q marker=%q", key, marker)
		}
		return nil
	}
	if err := check(ctx); err != nil {
		t.Error(err)
	}
	if err := postgres.RunInTransaction(ctx, pool, check); err != nil {
		t.Error(err)
	}
	conflict := tenant.WithKey(ctx, tenant.Production())
	if _, err := postgres.Executor(conflict, pool); err == nil {
		t.Error("Executor accepted both tenant and all-tenants marker")
	}
	if err := postgres.RunInTransaction(conflict, pool, func(context.Context) error { return nil }); err == nil {
		t.Error("transaction accepted both tenant and all-tenants marker")
	}
	var key, marker string
	if err := pool.QueryRow(ctx, "select coalesce(current_setting('tadoku.tenant', true), ''), coalesce(current_setting('tadoku.all_tenants', true), '')").Scan(&key, &marker); err != nil || key != "" || marker != "" {
		t.Errorf("after all-tenants statements tenant=%q marker=%q error=%v", key, marker, err)
	}
}

func TestExecutorInterleavedTenantsOnOneConnection(t *testing.T) {
	pool := oneConnectionPool(t, nil)
	errorsFound := make(chan error, 1000)
	var workers sync.WaitGroup
	for index, raw := range []string{"e2e/one", "e2e/two", "e2e/three"} {
		key, err := tenant.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		workers.Add(1)
		go func() {
			defer workers.Done()
			ctx := tenant.WithKey(t.Context(), key)
			executor, err := postgres.Executor(ctx, pool)
			if err != nil {
				errorsFound <- err
				return
			}
			for iteration := index; iteration < 1000; iteration += 3 {
				var actual string
				err := executor.QueryRow(ctx, "select coalesce(current_setting('tadoku.tenant', true), '')").Scan(&actual)
				if err != nil || actual != raw {
					errorsFound <- fmt.Errorf("statement %d tenant=%q want=%q error=%v", iteration, actual, raw, err)
				}
			}
		}()
	}
	workers.Wait()
	close(errorsFound)
	count := 0
	for err := range errorsFound {
		if count < 5 {
			t.Error(err)
		}
		count++
	}
	if count != 0 {
		t.Errorf("wrong-tenant statements=%d of 1000", count)
	}
}

func TestCanceledTenantBatchLeavesPoolUsable(t *testing.T) {
	pool := oneConnectionPool(t, nil)
	ctx := tenant.WithKey(t.Context(), tenant.Production())
	executor, err := postgres.Executor(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	canceled, stop := context.WithTimeout(ctx, 50*time.Millisecond)
	_, err = executor.Exec(canceled, "select pg_sleep(10)")
	stop()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("canceled batch error=%v", err)
	}
	reuse, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	var actual string
	if err := executor.QueryRow(reuse, "select coalesce(current_setting('tadoku.tenant', true), '')").Scan(&actual); err != nil || actual != tenant.Production().String() {
		t.Errorf("pool reuse tenant=%q error=%v", actual, err)
	}
}

func TestTenantBatchPreservesCommitAndStreamingErrors(t *testing.T) {
	pool := oneConnectionPool(t, nil)
	ctx := tenant.WithKey(t.Context(), tenant.Production())
	if _, err := pool.Exec(ctx, "create temporary table deferred_scope_rows (id integer unique deferrable initially deferred)"); err != nil {
		t.Fatal(err)
	}
	executor, err := postgres.Executor(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := executor.Exec(ctx, "insert into deferred_scope_rows values (1), (1)"); !postgres.IsUniqueViolation(err) {
		t.Errorf("deferred constraint error=%v", err)
	}
	var count int
	if err := pool.QueryRow(ctx, "select count(*) from deferred_scope_rows").Scan(&count); err != nil || count != 0 {
		t.Errorf("implicit transaction rollback count=%d error=%v", count, err)
	}
	rows, err := executor.Query(ctx, "select 12 / (3 - i) from generate_series(1, 3) i")
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			t.Fatal(err)
		}
	}
	if err := rows.Err(); err == nil {
		t.Error("streaming SQL error lost")
	}
	if err := pool.Ping(ctx); err != nil {
		t.Errorf("pool unusable after SQL error: %v", err)
	}
}

func TestExecutorRejectsContextScopeChanges(t *testing.T) {
	pool := oneConnectionPool(t, nil)
	ctx := tenant.WithKey(t.Context(), tenant.Production())
	executor, err := postgres.Executor(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	other, err := tenant.Parse("e2e/changed")
	if err != nil {
		t.Fatal(err)
	}
	for _, changed := range []context.Context{t.Context(), tenant.WithKey(t.Context(), other), alltenants.With(t.Context())} {
		if _, err := executor.Exec(changed, "select 1"); err == nil {
			t.Error("Exec accepted a changed scope")
		}
		if rows, err := executor.Query(changed, "select 1"); err == nil {
			rows.Close()
			t.Error("Query accepted a changed scope")
		}
		if err := executor.QueryRow(changed, "select 1").Scan(new(int)); err == nil {
			t.Error("QueryRow accepted a changed scope")
		}
	}
}
