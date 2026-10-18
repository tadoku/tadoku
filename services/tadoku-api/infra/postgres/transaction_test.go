package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestRunInTransactionCommitsBothParticipantsAndReadsWrites(t *testing.T) {
	t.Parallel()
	ctx, f := newFixture(t)
	type requestKey struct{}
	parent := context.WithValue(ctx, requestKey{}, "request-17")
	var retained context.Context
	calls := 0
	err := postgres.RunInTransaction(parent, f.db, func(child context.Context) error {
		calls++
		retained = child
		if child == parent || child.Value(requestKey{}) != "request-17" {
			return errors.New("callback did not receive a child retaining parent values")
		}
		wantDeadline, _ := parent.Deadline()
		gotDeadline, ok := child.Deadline()
		if !ok || !gotDeadline.Equal(wantDeadline) {
			return errors.New("parent deadline was not forwarded")
		}
		if err := f.books.Add(child, 1, "A synthetic book"); err != nil {
			return err
		}
		if err := f.notes.Add(child, 11, 1); err != nil {
			return err
		}
		title, err := f.notes.BookTitle(child, 11)
		if err != nil || title != "A synthetic book" {
			return fmt.Errorf("read own cross-feature writes: title=%q, error=%v", title, err)
		}
		count, err := f.books.Count(parent)
		if err != nil || count != 0 {
			return fmt.Errorf("uncommitted data visible through pool: count=%d, error=%v", count, err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("transaction: %v", err)
	}
	if calls != 1 {
		t.Errorf("callback calls=%d, want 1", calls)
	}
	if title, err := f.notes.BookTitle(ctx, 11); err != nil || title != "A synthetic book" {
		t.Errorf("committed participants: title=%q, error=%v", title, err)
	}
	if err := f.books.Add(retained, 2, "must not escape"); !errors.Is(err, pgx.ErrTxClosed) {
		t.Errorf("ended successful context error=%v, want pgx.ErrTxClosed", err)
	}
	if count, err := f.books.Count(ctx); err != nil || count != 1 {
		t.Errorf("committed book count=%d, error=%v, want 1", count, err)
	}
}

func TestRunInTransactionCallbackErrorRollsBackAndPreservesIdentity(t *testing.T) {
	t.Parallel()
	ctx, f := newFixture(t)
	denied := &rejection{"synthetic refusal"}
	primary := fmt.Errorf("business operation: %w", denied)
	var retained context.Context
	err := postgres.RunInTransaction(ctx, f.db, func(child context.Context) error {
		retained = child
		if err := f.books.Add(child, 1, "rolled back"); err != nil {
			return err
		}
		if err := f.notes.Add(child, 11, 1); err != nil {
			return err
		}
		return primary
	})
	var got *rejection
	if !errors.Is(err, primary) || !errors.Is(err, denied) || !errors.As(err, &got) || got != denied {
		t.Errorf("callback error identity lost: %v", err)
	}
	if count, err := f.books.Count(ctx); err != nil || count != 0 {
		t.Errorf("rollback book count=%d, error=%v, want 0", count, err)
	}
	if _, err := f.notes.BookTitle(ctx, 11); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("rollback note read error=%v, want pgx.ErrNoRows", err)
	}
	if count, err := f.notes.Count(ctx); err != nil || count != 0 {
		t.Errorf("rollback note count=%d, error=%v, want 0", count, err)
	}
	if err := f.books.Add(retained, 2, "must not escape"); !errors.Is(err, pgx.ErrTxClosed) {
		t.Errorf("ended failed context error=%v, want pgx.ErrTxClosed", err)
	}
	if err := f.books.Add(ctx, 1, "pool reused"); err != nil {
		t.Errorf("reuse pool after callback failure: %v", err)
	}
}

func TestRunInTransactionPanicRollsBackAndRepanicsOriginalValue(t *testing.T) {
	t.Parallel()
	ctx, f := newFixture(t)
	panicValue := &struct{ label string }{"original panic"}
	var retained context.Context
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_ = postgres.RunInTransaction(ctx, f.db, func(child context.Context) error {
			retained = child
			if err := f.books.Add(child, 1, "panic rollback"); err != nil {
				t.Fatalf("write before panic: %v", err)
			}
			if err := f.notes.Add(child, 11, 1); err != nil {
				t.Fatalf("second write before panic: %v", err)
			}
			panic(panicValue)
		})
	}()
	if recovered != panicValue {
		t.Errorf("recovered=%v, want original pointer %v", recovered, panicValue)
	}
	if count, err := f.books.Count(ctx); err != nil || count != 0 {
		t.Errorf("panic rollback count=%d, error=%v, want 0", count, err)
	}
	if count, err := f.notes.Count(ctx); err != nil || count != 0 {
		t.Errorf("panic rollback note count=%d, error=%v, want 0", count, err)
	}
	if _, err := postgres.Executor(retained, f.db); !errors.Is(err, pgx.ErrTxClosed) {
		t.Errorf("ended panicked context error=%v, want pgx.ErrTxClosed", err)
	}
	if err := f.books.Add(ctx, 1, "pool reused"); err != nil {
		t.Errorf("reuse pool after panic: %v", err)
	}
}

func TestRunInTransactionDeferredConstraintFailsAtCommitWithoutReplay(t *testing.T) {
	t.Parallel()
	ctx, f := newFixture(t)
	calls := 0
	var retained context.Context
	err := postgres.RunInTransaction(ctx, f.db, func(child context.Context) error {
		calls++
		retained = child
		if err := f.books.Add(child, 1, "commit will fail"); err != nil {
			return err
		}
		if err := f.notes.Add(child, 11, 999); err != nil {
			t.Fatalf("deferred constraint failed before commit: %v", err)
		}
		return nil
	})
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.SQLState() != "23503" {
		t.Fatalf("commit error=%v, want PostgreSQL foreign-key violation", err)
	}
	if !errors.Is(err, pgError) {
		t.Errorf("commit error cause was not preserved: %v", err)
	}
	if got := errx.KindOf(err); got != errx.Unknown {
		t.Errorf("commit error kind=%v, want unknown", got)
	}
	if calls != 1 {
		t.Errorf("callback replayed: calls=%d", calls)
	}
	if count, err := f.books.Count(ctx); err != nil || count != 0 {
		t.Errorf("failed-commit book count=%d, error=%v, want 0", count, err)
	}
	if count, err := f.notes.Count(ctx); err != nil || count != 0 {
		t.Errorf("failed-commit note count=%d, error=%v, want 0", count, err)
	}
	if _, err := postgres.Executor(retained, f.db); !errors.Is(err, pgx.ErrTxClosed) {
		t.Errorf("ended commit-failure context error=%v, want pgx.ErrTxClosed", err)
	}
	if err := postgres.RunInTransaction(ctx, f.db, func(child context.Context) error {
		if err := f.books.Add(child, 1, "pool reused"); err != nil {
			return err
		}
		return f.notes.Add(child, 11, 1)
	}); err != nil {
		t.Errorf("reuse pool after commit failure: %v", err)
	}
}

func TestRunInTransactionRejectsNestedAndWrongDatabaseWithoutEscapedWrites(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"nested same pool", "nested other pool", "wrong repository"} {
		t.Run(mode, func(t *testing.T) {
			ctx, f := newFixture(t)
			other := openPool(t)
			otherBooks := bookRepository{other, f.schema}
			nestedCalled := false
			err := postgres.RunInTransaction(ctx, f.db, func(child context.Context) error {
				if err := f.books.Add(child, 1, "outer must roll back"); err != nil {
					return err
				}
				if mode == "wrong repository" {
					return otherBooks.Add(child, 2, "must not escape")
				}
				nestedPool := f.db
				if mode == "nested other pool" {
					nestedPool = other
				}
				return postgres.RunInTransaction(child, nestedPool, func(nested context.Context) error {
					nestedCalled = true
					return (bookRepository{nestedPool, f.schema}).Add(nested, 2, "must not escape")
				})
			})
			want := postgres.ErrNestedTransaction
			if mode == "wrong repository" {
				want = postgres.ErrWrongDatabase
			}
			if !errors.Is(err, want) || nestedCalled {
				t.Errorf("rejection error=%v, want %v; nested callback=%v", err, want, nestedCalled)
			}
			if count, err := f.books.Count(ctx); err != nil || count != 0 {
				t.Errorf("escaped writes: count=%d, error=%v", count, err)
			}
		})
	}
}

func TestExecutorSupportsPoolAndTransactionSQLSurface(t *testing.T) {
	t.Parallel()
	var _ postgres.DBTX = (*pgxpool.Pool)(nil)
	var _ postgres.DBTX = (pgx.Tx)(nil)
	ctx, f := newFixture(t)
	for _, transactional := range []bool{false, true} {
		work := func(ctx context.Context) error {
			q, err := postgres.Executor(ctx, f.db)
			if err != nil {
				return err
			}
			if transactional {
				if _, ok := q.(pgx.Tx); !ok {
					return errors.New("transaction executor is not a native transaction")
				}
			}
			var value int
			if err := q.QueryRow(ctx, "select $1::integer", 17).Scan(&value); err != nil || value != 17 {
				return fmt.Errorf("single row value=%d, error=%v", value, err)
			}
			rows, err := q.Query(ctx, "select $1::integer union all select $2::integer", 21, 34)
			if err != nil {
				return err
			}
			defer rows.Close()
			sum := 0
			for rows.Next() {
				if err := rows.Scan(&value); err != nil {
					return err
				}
				sum += value
			}
			if err := rows.Err(); err != nil || sum != 55 {
				return fmt.Errorf("row iteration sum=%d, error=%v", sum, err)
			}
			return nil
		}
		var err error
		if transactional {
			err = postgres.RunInTransaction(ctx, f.db, work)
		} else {
			err = work(ctx)
		}
		if err != nil {
			t.Errorf("SQL surface transactional=%v: %v", transactional, err)
		}
	}
	if err := f.books.Add(ctx, 1, "ordinary pool write"); err != nil {
		t.Fatalf("pool write: %v", err)
	}
	if count, err := f.books.Count(ctx); err != nil || count != 1 {
		t.Errorf("pool read count=%d, error=%v", count, err)
	}
}

func TestRunInTransactionRejectsRestartingEndedContext(t *testing.T) {
	t.Parallel()
	ctx, f := newFixture(t)
	var retained context.Context
	if err := postgres.RunInTransaction(ctx, f.db, func(child context.Context) error {
		retained = child
		return nil
	}); err != nil {
		t.Fatalf("initial transaction: %v", err)
	}
	called := false
	err := postgres.RunInTransaction(retained, f.db, func(context.Context) error { called = true; return nil })
	if !errors.Is(err, pgx.ErrTxClosed) || called {
		t.Errorf("restart ended context: error=%v, called=%v", err, called)
	}
}

func TestIndependentConcurrentContextsRemainIsolated(t *testing.T) {
	t.Parallel()
	ctx, f := newFixture(t)
	ready := make(chan int, 2)
	release := make(chan struct{})
	type outcome struct {
		id  int
		err error
	}
	results := make(chan outcome, 2)
	rolledBack := errors.New("rollback only this context")
	for id := 1; id <= 2; id++ {
		go func(id int) {
			err := postgres.RunInTransaction(ctx, f.db, func(child context.Context) error {
				if err := f.books.Add(child, id, fmt.Sprintf("concurrent-%d", id)); err != nil {
					return err
				}
				count, err := f.books.Count(child)
				if err != nil || count != 1 {
					return fmt.Errorf("context %d saw count=%d, error=%v", id, count, err)
				}
				ready <- id
				select {
				case <-release:
				case <-child.Done():
					return child.Err()
				}
				if id == 2 {
					return rolledBack
				}
				return nil
			})
			results <- outcome{id, err}
		}(id)
	}
	for range 2 {
		select {
		case <-ready:
		case result := <-results:
			t.Fatalf("context %d failed before barrier: %v", result.id, result.err)
		case <-ctx.Done():
			t.Fatal("concurrent transactions did not reach barrier")
		}
	}
	if count, err := f.books.Count(ctx); err != nil || count != 0 {
		t.Errorf("outside pool saw uncommitted data: count=%d, error=%v", count, err)
	}
	close(release)
	for range 2 {
		select {
		case result := <-results:
			if result.id == 1 && result.err != nil {
				t.Errorf("committing context: %v", result.err)
			}
			if result.id == 2 && !errors.Is(result.err, rolledBack) {
				t.Errorf("rolling-back context: %v", result.err)
			}
		case <-ctx.Done():
			t.Fatal("concurrent transactions did not finish")
		}
	}
	var id int
	if err := f.db.QueryRow(ctx, "select id from "+f.schema+".book").Scan(&id); err != nil || id != 1 {
		t.Errorf("final independent commit: id=%d, error=%v, want 1", id, err)
	}
	if count, err := f.books.Count(ctx); err != nil || count != 1 {
		t.Errorf("final independent count=%d, error=%v, want 1", count, err)
	}
}
