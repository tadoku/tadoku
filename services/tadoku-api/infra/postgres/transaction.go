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
)

var (
	ErrNestedTransaction = errors.New("postgres: nested transaction")
	ErrWrongDatabase     = errors.New("postgres: transaction belongs to another database handle")
)

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

// Executor must use the operation's context; repositories must pass that same
// context to SQL calls. Do not retain executors or rows past RunInTransaction.
// An ended scope never falls back to the pool.
func Executor(ctx context.Context, db *pgxpool.Pool) (DBTX, error) {
	s, ok := ctx.Value(scopeKey{}).(*scope)
	if !ok {
		return db, nil
	}
	if s.db != db {
		return nil, ErrWrongDatabase
	}
	if s.done.Load() {
		return nil, pgx.ErrTxClosed
	}
	return s.tx, nil
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
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres: begin: %w", err)
	}
	s := &scope{db: db, tx: tx}
	defer func() {
		s.done.Store(true)
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
