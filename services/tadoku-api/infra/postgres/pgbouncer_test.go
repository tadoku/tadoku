package postgres_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestPgBouncerTransactionTenantIsolation(t *testing.T) {
	u, err := url.Parse(os.Getenv("TADOKU_TEST_PGBOUNCER_URL"))
	if err != nil || u == nil || u.User == nil {
		t.Fatal("TADOKU_TEST_PGBOUNCER_URL must be a disposable loopback PostgreSQL DSN")
	}
	password, _ := u.User.Password()
	if u.Scheme != "postgres" ||
		(u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1") ||
		u.Port() == "" || u.Path != "/postgres" || u.RawPath != "" || u.Opaque != "" ||
		u.User.Username() != "postgres" || password != "postgres" ||
		u.RawQuery != "sslmode=disable" || u.Fragment != "" {
		t.Fatal("TADOKU_TEST_PGBOUNCER_URL must use postgres:postgres on loopback with an explicit port, " +
			"/postgres and sslmode=disable")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	db, err := testpostgres.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.Pool.Exec(ctx, `create table tenant_transport_rows (
		expected text not null,
		actual text not null default coalesce(current_setting('tadoku.tenant', true), '')
	)`); err != nil {
		t.Fatal(err)
	}
	disposable, err := url.Parse(db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = disposable.Path
	config, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		t.Fatal(err)
	}
	config.MaxConns = 30
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	const currentTenantSQL = "select coalesce(current_setting('tadoku.tenant', true), '')"
	keys := make([]tenant.Key, 3)
	for index, raw := range []string{"e2e/pgbouncer-one", "e2e/pgbouncer-two", "e2e/pgbouncer-three"} {
		keys[index], err = tenant.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
	}
	var reads, writes, wrongReads atomic.Int64
	errors := make(chan error, 300)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for worker := range 300 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			key := keys[worker%len(keys)]
			tenantCtx := tenant.WithKey(ctx, key)
			executor, err := postgres.Executor(tenantCtx, pool)
			if err != nil {
				errors <- err
				return
			}
			for iteration := range 20 {
				var actual string
				if err := executor.QueryRow(tenantCtx, currentTenantSQL).Scan(&actual); err != nil {
					errors <- err
					return
				}
				reads.Add(1)
				if actual != key.String() {
					wrongReads.Add(1)
				}
				err := postgres.RunInTransaction(tenantCtx, pool, func(child context.Context) error {
					tx, err := postgres.Executor(child, pool)
					if err != nil {
						return err
					}
					if err := tx.QueryRow(child, currentTenantSQL).Scan(&actual); err != nil {
						return err
					}
					reads.Add(1)
					if actual != key.String() {
						wrongReads.Add(1)
					}
					if iteration%2 == 1 {
						tag, err := tx.Exec(child, "insert into tenant_transport_rows (expected) values ($1)", key.String())
						if err != nil {
							return err
						}
						if tag.RowsAffected() != 1 {
							return fmt.Errorf("transaction insert affected %d rows", tag.RowsAffected())
						}
					}
					return nil
				})
				if err != nil {
					errors <- err
					return
				}
				if iteration%2 == 0 {
					tag, err := executor.Exec(tenantCtx, "insert into tenant_transport_rows (expected) values ($1)", key.String())
					if err != nil {
						errors <- err
						return
					}
					if tag.RowsAffected() != 1 {
						errors <- fmt.Errorf("standalone insert affected %d rows", tag.RowsAffected())
						return
					}
				}
				writes.Add(1)
			}
		}()
	}
	close(start)
	workers.Wait()
	close(errors)
	failed := 0
	for err := range errors {
		failed++
		if failed <= 5 {
			t.Error(err)
		}
	}
	var stored, misfiled int64
	err = db.Pool.QueryRow(ctx, `select count(*), count(*) filter (where actual <> expected)
		from tenant_transport_rows`).Scan(&stored, &misfiled)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(
		"goroutines=300 tenants=3 reads=%d writes=%d stored=%d wrong-tenant reads=%d misfiled writes=%d failed workers=%d",
		reads.Load(), writes.Load(), stored, wrongReads.Load(), misfiled, failed,
	)
	if failed != 0 ||
		reads.Load() != 12000 ||
		writes.Load() != 6000 ||
		stored != 6000 ||
		wrongReads.Load() != 0 ||
		misfiled != 0 {
		t.Fatal("PgBouncer transaction workload did not preserve every tenant read and write")
	}

	first, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Release()
	second, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Release()
	const (
		setSessionTenantSQL           = "select set_config('tadoku.tenant', $1, false)"
		setSessionTenantAndBackendSQL = "select set_config('tadoku.tenant', $1, false), pg_backend_pid()"
		tenantAndBackendSQL           = "select coalesce(current_setting('tadoku.tenant', true), ''), pg_backend_pid()"
	)
	leaks := 0
	for range 100 {
		var actual string
		var backend int64
		// Test safety: session settings are confined to this disposable database's
		// negative control, with two distinct PgBouncer clients sharing backends.
		if err := first.QueryRow(ctx, setSessionTenantSQL, keys[0].String()).Scan(&actual); err != nil {
			t.Fatal(err)
		}
		if err := second.QueryRow(ctx, setSessionTenantAndBackendSQL, keys[1].String()).Scan(&actual, &backend); err != nil {
			t.Fatal(err)
		}
		var observedBackend int64
		if err := first.QueryRow(ctx, tenantAndBackendSQL).Scan(&actual, &observedBackend); err != nil {
			t.Fatal(err)
		}
		if actual == keys[1].String() && observedBackend == backend {
			leaks++
		}
	}
	t.Logf("negative control: 100 reads through separate clients; verified cross-client session leaks=%d", leaks)
	if leaks == 0 {
		t.Fatal("session-level negative control detected no cross-client leak")
	}
}
