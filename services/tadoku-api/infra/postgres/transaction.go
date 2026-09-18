// Package postgres routes concrete repository SQL through an app-owned transaction.
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
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var (
	ErrNestedTransaction  = errors.New("postgres: nested transaction")
	ErrWrongDatabase      = errors.New("postgres: transaction belongs to another database handle")
	errAcquisitionTimeout = errors.New("postgres: connection acquisition timed out")
)

const acquisitionTimeout = 5 * time.Second

// DBTX is the native pgx/sqlc execution surface shared by the bounded pool and pgx.Tx.
type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type scopeKey struct{}

type scope struct {
	db   *pgxpool.Pool
	tx   pgx.Tx
	done atomic.Bool
}

type boundedPool struct {
	db *pgxpool.Pool
}

type releasingRows struct {
	pgx.Rows
	connection *pgxpool.Conn
}

func (r *releasingRows) Close() {
	r.Rows.Close()
	r.release()
}

func (r *releasingRows) Next() bool {
	next := r.Rows.Next()
	if !next {
		r.Close()
	}
	return next
}

func (r *releasingRows) Scan(dest ...any) error {
	err := r.Rows.Scan(dest...)
	if err != nil {
		r.Close()
	}
	return err
}

func (r *releasingRows) Values() ([]any, error) {
	values, err := r.Rows.Values()
	if err != nil {
		r.Close()
	}
	return values, err
}

func (r *releasingRows) release() {
	if r.connection != nil {
		r.connection.Release()
		r.connection = nil
	}
}

type releasingRow struct {
	pgx.Row
	connection *pgxpool.Conn
}

func (r *releasingRow) Scan(dest ...any) error {
	defer r.release()
	return r.Row.Scan(dest...)
}

func (r *releasingRow) release() {
	if r.connection != nil {
		r.connection.Release()
		r.connection = nil
	}
}

type errorRow struct {
	err error
}

func (r errorRow) Scan(...any) error {
	return r.err
}

func (p *boundedPool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	connection, err := acquire(ctx, p.db)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	defer connection.Release()

	return connection.Exec(ctx, sql, arguments...)
}

func (p *boundedPool) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	connection, err := acquire(ctx, p.db)
	if err != nil {
		return nil, err
	}

	rows, err := connection.Query(ctx, sql, arguments...)
	if err != nil {
		connection.Release()
		return nil, err
	}

	return &releasingRows{
		Rows:       rows,
		connection: connection,
	}, nil
}

func (p *boundedPool) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	connection, err := acquire(ctx, p.db)
	if err != nil {
		return errorRow{err: err}
	}

	return &releasingRow{
		Row:        connection.QueryRow(ctx, sql, arguments...),
		connection: connection,
	}
}

// Executor selects the active transaction for db, or a bounded pool executor outside a scope.
// Repositories must resolve it with the operation's context and pass that same
// context to SQL calls. Finish row iteration in the repository method, and do
// not retain transactional executors or rows past RunInTransaction.
// An ended scope returns pgx.ErrTxClosed; it never falls back to the pool.
func Executor(ctx context.Context, db *pgxpool.Pool) (DBTX, error) {
	s, ok := ctx.Value(scopeKey{}).(*scope)
	if !ok {
		return &boundedPool{db: db}, nil
	}
	if s.db != db {
		return nil, ErrWrongDatabase
	}
	if s.done.Load() {
		return nil, pgx.ErrTxClosed
	}
	return s.tx, nil
}

func acquire(ctx context.Context, db *pgxpool.Pool) (*pgxpool.Conn, error) {
	acquireCtx, cancel := context.WithTimeout(ctx, acquisitionTimeout)
	connection, err := db.Acquire(acquireCtx)
	acquireErr := acquireCtx.Err()
	cancel()
	if err != nil {
		if parentErr := ctx.Err(); parentErr != nil {
			return nil, parentErr
		}
		if errors.Is(acquireErr, context.DeadlineExceeded) {
			return nil, errx.NewUnavailableError("acquire postgres connection", errAcquisitionTimeout)
		}
		return nil, fmt.Errorf("postgres: acquire: %w", err)
	}
	return connection, nil
}

// RunInTransaction calls work once with a child context and commits only when it succeeds.
// All transaction work must finish in the callback. Nested scopes are rejected;
// cross-feature operations share this one outer context. Keep network, cache,
// and asynchronous work outside the callback. Independent transactions may run
// concurrently; SQL within a transaction must not run in parallel.
//
// Callback errors retain their identity. Begin and commit failures wrap their
// causes. Rollback is attempted on every exit, including panic; cleanup cannot
// replace the primary error or panic. A commit error can have an unknown
// persistence outcome, so RunInTransaction never retries or replays work.
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
	connection, err := acquire(ctx, db)
	if err != nil {
		return err
	}
	defer connection.Release()
	tx, err := connection.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres: begin: %w", err)
	}
	s := &scope{db: db, tx: tx}
	defer func() {
		s.done.Store(true)
		// Native pgx does not roll back when the request context is canceled.
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	if err := ctx.Err(); err != nil {
		return err
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
