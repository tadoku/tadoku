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

func noRelease() {}

// DBTX is the native pgx/sqlc execution surface shared by *pgxpool.Pool and pgx.Tx.
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

// Executor selects the active transaction for db, or acquires a connection outside a scope.
// Repositories must resolve it with the operation's context and pass that same
// context to SQL calls, then defer release. Do not retain executors or rows past
// release or RunInTransaction.
// An ended scope returns pgx.ErrTxClosed; it never falls back to the pool.
func Executor(ctx context.Context, db *pgxpool.Pool) (DBTX, func(), error) {
	s, ok := ctx.Value(scopeKey{}).(*scope)
	if !ok {
		return acquire(ctx, db)
	}
	if s.db != db {
		return nil, noRelease, ErrWrongDatabase
	}
	if s.done.Load() {
		return nil, noRelease, pgx.ErrTxClosed
	}
	return s.tx, noRelease, nil
}

func acquire(ctx context.Context, db *pgxpool.Pool) (*pgxpool.Conn, func(), error) {
	acquireCtx, cancel := context.WithTimeout(ctx, acquisitionTimeout)
	connection, err := db.Acquire(acquireCtx)
	acquireErr := acquireCtx.Err()
	cancel()
	if err != nil {
		if parentErr := ctx.Err(); parentErr != nil {
			return nil, noRelease, parentErr
		}
		if errors.Is(acquireErr, context.DeadlineExceeded) {
			return nil, noRelease, errx.NewUnavailableError("acquire postgres connection", errAcquisitionTimeout)
		}
		return nil, noRelease, fmt.Errorf("postgres: acquire: %w", err)
	}
	return connection, connection.Release, nil
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
	connection, release, err := acquire(ctx, db)
	if err != nil {
		return err
	}
	defer release()
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
