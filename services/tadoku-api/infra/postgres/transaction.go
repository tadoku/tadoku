package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant/alltenants"
)

var (
	ErrNestedTransaction = errors.New("postgres: nested transaction")
	ErrWrongDatabase     = errors.New("postgres: transaction belongs to another database handle")
	ErrNoTenant          = errors.New("postgres: no tenant")
	ErrTenantMismatch    = errors.New("postgres: tenant mismatch")
)

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type scopeKey struct{}

type scope struct {
	db     *pgxpool.Pool
	tx     pgx.Tx
	tenant tenantScope
	done   atomic.Bool
}

type tenantScope struct {
	key tenant.Key
}

func tenantFromContext(ctx context.Context) (tenantScope, error) {
	key, present := tenant.FromContext(ctx)
	if alltenants.Enabled(ctx) {
		if present {
			return tenantScope{}, ErrTenantMismatch
		}
		return tenantScope{}, nil
	}
	if !present {
		return tenantScope{}, ErrNoTenant
	}
	return tenantScope{key: key}, nil
}

func (scope tenantScope) setting() (string, string) {
	if scope.key == (tenant.Key{}) {
		return "tadoku.all_tenants", "on"
	}
	return "tadoku.tenant", scope.key.String()
}

// Executor must use the operation's context; repositories must pass that same
// context to SQL calls. Do not retain executors or rows past RunInTransaction.
// Supply exactly one tenant or an all-tenants scope. An ended transaction scope
// never falls back to the pool.
func Executor(ctx context.Context, db *pgxpool.Pool) (DBTX, error) {
	requested, err := tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s, ok := ctx.Value(scopeKey{}).(*scope)
	if !ok {
		return poolExecutor{db: db, tenant: requested}, nil
	}

	if s.db != db {
		return nil, ErrWrongDatabase
	}
	if s.done.Load() {
		return nil, pgx.ErrTxClosed
	}
	if s.tenant != requested {
		return nil, ErrTenantMismatch
	}
	return s.tx, nil
}

type poolExecutor struct {
	db     *pgxpool.Pool
	tenant tenantScope
}

func (executor poolExecutor) batch(ctx context.Context, query string, args ...any) (pgx.BatchResults, error) {
	requested, err := tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if requested != executor.tenant {
		return nil, ErrTenantMismatch
	}

	name, value := requested.setting()
	batch := new(pgx.Batch)
	batch.Queue("select set_config($1, $2, true)", name, value)
	batch.Queue(query, args...)

	results := executor.db.SendBatch(ctx, batch)
	if _, err := results.Exec(); err != nil {
		return nil, errors.Join(err, results.Close())
	}
	return results, nil
}

func (executor poolExecutor) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	results, err := executor.batch(ctx, query, args...)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	tag, err := results.Exec()
	return tag, errors.Join(err, results.Close())
}

func (executor poolExecutor) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	results, err := executor.batch(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	rows, err := results.Query()
	if err != nil {
		return nil, errors.Join(err, results.Close())
	}
	return &batchRows{Rows: rows, results: results}, nil
}

func (executor poolExecutor) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	results, err := executor.batch(ctx, query, args...)
	if err != nil {
		return batchRow{err: err}
	}
	return batchRow{Row: results.QueryRow(), results: results}
}

type batchRows struct {
	pgx.Rows
	results  pgx.BatchResults
	closeErr error
	closed   bool
}

func (rows *batchRows) Close() {
	if rows.closed {
		return
	}
	rows.closed = true
	rows.Rows.Close()
	rows.closeErr = rows.results.Close()
}

func (rows *batchRows) Next() bool {
	next := rows.Rows.Next()
	if !next {
		rows.Close()
	}
	return next
}

func (rows *batchRows) Scan(dest ...any) error {
	err := rows.Rows.Scan(dest...)
	if err != nil {
		rows.Close()
	}
	return errors.Join(err, rows.closeErr)
}

func (rows *batchRows) Err() error {
	return errors.Join(rows.Rows.Err(), rows.closeErr)
}

type batchRow struct {
	pgx.Row
	results pgx.BatchResults
	err     error
}

func (row batchRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	err := row.Row.Scan(dest...)
	if closeErr := row.results.Close(); closeErr != nil {
		return errors.Join(err, closeErr)
	}
	return err
}

// All transaction work must finish in the callback. Nested scopes are rejected;
// cross-feature operations share this one outer context. Keep network, cache,
// and asynchronous work outside the callback. Independent transactions may run
// concurrently; SQL within a transaction must not run in parallel.
//
// A commit error can leave persistence uncertain; do not retry the callback.
func RunInTransaction(ctx context.Context, db *pgxpool.Pool, work func(context.Context) error) error {
	if s, ok := ctx.Value(scopeKey{}).(*scope); ok {
		if s.done.Load() {
			return pgx.ErrTxClosed
		}
		return ErrNestedTransaction
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	requested, err := tenantFromContext(ctx)
	if err != nil {
		return err
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres: begin: %w", err)
	}
	s := &scope{db: db, tx: tx, tenant: requested}
	defer func() {
		s.done.Store(true)
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()

	if err := ctx.Err(); err != nil {
		return err
	}
	name, value := requested.setting()
	if _, err := tx.Exec(ctx, "select set_config($1, $2, true)", name, value); err != nil {
		return fmt.Errorf("postgres: set transaction scope: %w", err)
	}

	if err := work(context.WithValue(ctx, scopeKey{}, s)); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres: commit: %w", err)
	}
	return nil
}
