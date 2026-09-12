package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

const testDSNVariable = "TADOKU_TEST_POSTGRES_URL"

// Use only a disposable local PostgreSQL instance with synthetic credentials.
// An absent DSN is a failure, and remote/application databases are refused.
func disposableDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv(testDSNVariable)
	if dsn == "" {
		t.Fatalf("%s is required for real PostgreSQL integration tests", testDSNVariable)
	}
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme != "postgres" ||
		(u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") || u.Port() == "" ||
		u.Path != "/postgres" || u.RawPath != "" ||
		u.RawQuery != "sslmode=disable" || u.Fragment != "" || u.Opaque != "" ||
		u.User == nil || u.User.Username() != "postgres" {
		t.Fatal("refusing DSN outside the declared disposable PostgreSQL endpoint")
	}
	if password, ok := u.User.Password(); !ok || password != "postgres" {
		t.Fatal("the declared synthetic PostgreSQL credentials are required")
	}
	return dsn
}

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	config, err := pgxpool.ParseConfig(disposableDSN(t))
	if err != nil {
		t.Fatalf("parse disposable PostgreSQL: %v", err)
	}
	config.MaxConns = 6
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("open disposable PostgreSQL: %v", err)
	}
	t.Cleanup(db.Close)
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("ping disposable PostgreSQL: %v", err)
	}
	return db
}

type fixture struct {
	db     *pgxpool.Pool
	schema string
	books  bookRepository
	notes  noteRepository
}

func newFixture(t *testing.T) (context.Context, fixture) {
	t.Helper()
	db := openPool(t)
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatalf("random schema suffix: %v", err)
	}
	schema := "tadoku_tx_test_" + hex.EncodeToString(random[:])
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	if _, err := db.Exec(ctx, "create schema "+schema); err != nil {
		t.Fatalf("create synthetic schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := db.Exec(cleanupCtx, "drop schema "+schema+" cascade"); err != nil {
			t.Errorf("drop synthetic schema: %v", err)
		}
	})
	_, err := db.Exec(ctx, "create table "+schema+`.book (
		id integer primary key,
		title text not null
	);
	create table `+schema+`.note (
		id integer primary key,
		book_id integer not null references `+schema+`.book(id) deferrable initially deferred
	)`)
	if err != nil {
		t.Fatalf("create synthetic tables: %v", err)
	}
	return ctx, fixture{db, schema, bookRepository{db, schema}, noteRepository{db, schema}}
}

// These independent concrete participants stand in for two feature repositories.
// Their business methods accept context and business values, with SQL plumbing
// confined to the repository implementation.
type bookRepository struct {
	db     *pgxpool.Pool
	schema string
}

func (r bookRepository) Add(ctx context.Context, id int, title string) error {
	q, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, "insert into "+r.schema+".book(id, title) values ($1, $2)", id, title)
	return err
}

func (r bookRepository) Count(ctx context.Context) (int, error) {
	q, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var count int
	err = q.QueryRow(ctx, "select count(*) from "+r.schema+".book").Scan(&count)
	return count, err
}

type noteRepository struct {
	db     *pgxpool.Pool
	schema string
}

func (r noteRepository) Add(ctx context.Context, id, bookID int) error {
	q, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, "insert into "+r.schema+".note(id, book_id) values ($1, $2)", id, bookID)
	return err
}

func (r noteRepository) BookTitle(ctx context.Context, id int) (string, error) {
	q, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return "", err
	}
	var title string
	err = q.QueryRow(ctx, "select b.title from "+r.schema+".note n join "+r.schema+".book b on b.id = n.book_id where n.id = $1", id).Scan(&title)
	return title, err
}

func (r noteRepository) Count(ctx context.Context) (int, error) {
	q, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var count int
	err = q.QueryRow(ctx, "select count(*) from "+r.schema+".note").Scan(&count)
	return count, err
}

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

type rejection struct{ reason string }

func (e *rejection) Error() string { return e.reason }

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

func TestRunInTransactionFailedBeginNeverCallsWork(t *testing.T) {
	t.Parallel()
	t.Run("already canceled", func(t *testing.T) {
		db := openPool(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		called := false
		err := postgres.RunInTransaction(ctx, db, func(context.Context) error { called = true; return nil })
		if !errors.Is(err, context.Canceled) || called {
			t.Errorf("canceled begin: error=%v, called=%v", err, called)
		}
	})
	t.Run("closed pool", func(t *testing.T) {
		db := openPool(t)
		db.Close()
		called := false
		err := postgres.RunInTransaction(context.Background(), db, func(context.Context) error { called = true; return nil })
		if err == nil || called {
			t.Errorf("closed-pool begin: error=%v, called=%v", err, called)
		}
	})
	t.Run("waiting for pool connection", func(t *testing.T) {
		db := openPool(t)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for range db.Config().MaxConns {
			connection, err := db.Acquire(ctx)
			if err != nil {
				t.Fatalf("reserve pool connection: %v", err)
			}
			defer connection.Release()
		}
		blocked, stop := context.WithTimeout(ctx, 150*time.Millisecond)
		defer stop()
		called := false
		err := postgres.RunInTransaction(blocked, db, func(context.Context) error { called = true; return nil })
		if !errors.Is(err, context.DeadlineExceeded) || called {
			t.Errorf("blocked begin: error=%v, called=%v", err, called)
		}
	})
}

func TestRunInTransactionCancellationBoundsBlockedSQLAndLeavesPoolUsable(t *testing.T) {
	t.Parallel()
	ctx, f := newFixture(t)
	lock, err := f.db.Begin(ctx)
	if err != nil {
		t.Fatalf("begin independent lock holder: %v", err)
	}
	defer lock.Rollback(ctx)
	if _, err := lock.Exec(ctx, "lock table "+f.schema+".book in access exclusive mode"); err != nil {
		t.Fatalf("lock synthetic table: %v", err)
	}
	blocked, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	result := make(chan error, 1)
	entered := make(chan struct{})
	go func() {
		result <- postgres.RunInTransaction(blocked, f.db, func(child context.Context) error {
			close(entered)
			return f.books.Add(child, 1, "must be canceled")
		})
	}()
	select {
	case err := <-result:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("blocked SQL error=%v, want context.DeadlineExceeded", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("blocked SQL exceeded the cancellation bound")
	}
	select {
	case <-entered:
	default:
		t.Error("deadline expired during Begin instead of reaching blocked SQL")
	}
	if err := lock.Rollback(ctx); err != nil {
		t.Fatalf("release table lock: %v", err)
	}
	if count, err := f.books.Count(ctx); err != nil || count != 0 {
		t.Errorf("canceled write escaped: count=%d, error=%v", count, err)
	}
	if err := f.books.Add(ctx, 1, "pool reused"); err != nil {
		t.Errorf("reuse pool after cancellation: %v", err)
	}
}

func TestRunInTransactionCancellationAfterWritesCleansUpWithoutReplacingOutcome(t *testing.T) {
	t.Parallel()
	for _, outcome := range []string{"nil callback error", "callback error", "panic"} {
		t.Run(outcome, func(t *testing.T) {
			ctx, f := newFixture(t)
			parent, cancel := context.WithCancel(ctx)
			defer cancel()
			primary := &rejection{"callback failed after cancellation"}
			panicValue := &struct{ label string }{"panic after cancellation"}
			var retained context.Context
			var recovered any
			var transactionPID uint32
			var err error
			calls := 0
			func() {
				defer func() { recovered = recover() }()
				err = postgres.RunInTransaction(parent, f.db, func(child context.Context) error {
					calls++
					retained = child
					q, err := postgres.Executor(child, f.db)
					if err != nil {
						return err
					}
					if err := q.QueryRow(child, "select pg_backend_pid()").Scan(&transactionPID); err != nil {
						return err
					}
					if err := f.books.Add(child, 1, "cancellation rollback"); err != nil {
						return err
					}
					if err := f.notes.Add(child, 11, 1); err != nil {
						return err
					}
					cancel()
					switch outcome {
					case "callback error":
						return primary
					case "panic":
						panic(panicValue)
					default:
						return nil
					}
				})
			}()
			if calls != 1 {
				t.Errorf("callback calls=%d, want 1", calls)
			}
			switch outcome {
			case "nil callback error":
				if !errors.Is(err, context.Canceled) || recovered != nil {
					t.Errorf("canceled success: error=%v, panic=%v", err, recovered)
				}
			case "callback error":
				var got *rejection
				if !errors.Is(err, primary) || !errors.As(err, &got) || got != primary || recovered != nil {
					t.Errorf("callback outcome replaced: error=%v, panic=%v", err, recovered)
				}
			case "panic":
				if recovered != panicValue {
					t.Errorf("recovered=%v, want original panic %v", recovered, panicValue)
				}
			}
			if acquired := f.db.Stat().AcquiredConns(); acquired != 0 {
				t.Errorf("cleanup retained %d acquired connections", acquired)
			}
			// No concurrent pool work occurs in this test. Reusing the backend
			// proves rollback succeeded with a live context: canceled rollback
			// would make native pgx discard the connection instead.
			var reusedPID uint32
			if err := f.db.QueryRow(ctx, "select pg_backend_pid()").Scan(&reusedPID); err != nil {
				t.Fatalf("query backend after cleanup: %v", err)
			}
			if reusedPID != transactionPID {
				t.Errorf("cleanup discarded the connection: before=%d, after=%d", transactionPID, reusedPID)
			}
			if count, err := f.books.Count(ctx); err != nil || count != 0 {
				t.Errorf("canceled book writes escaped: count=%d, error=%v", count, err)
			}
			if count, err := f.notes.Count(ctx); err != nil || count != 0 {
				t.Errorf("canceled note writes escaped: count=%d, error=%v", count, err)
			}
			if _, err := postgres.Executor(retained, f.db); !errors.Is(err, pgx.ErrTxClosed) {
				t.Errorf("ended canceled context error=%v, want pgx.ErrTxClosed", err)
			}
			if err := f.books.Add(ctx, 1, "pool reused"); err != nil {
				t.Errorf("reuse pool after canceled callback: %v", err)
			}
		})
	}
}

func TestRunInTransactionRejectsNestedAndWrongDatabaseWithoutEscapedWrites(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"nested same pool", "nested other pool", "wrong repository"} {
		t.Run(mode, func(t *testing.T) {
			ctx, f := newFixture(t)
			other := openPool(t) // Same endpoint and schema, deliberately a different handle.
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
			if !transactional && q != f.db {
				return errors.New("ordinary executor is not the native pool")
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

func TestExecutorForwardsCancellationForEveryOperation(t *testing.T) {
	t.Parallel()
	db := openPool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, transactional := range []bool{false, true} {
		for _, operation := range []string{"exec", "query", "query row"} {
			work := func(ctx context.Context) error {
				q, err := postgres.Executor(ctx, db)
				if err != nil {
					return err
				}
				canceled, stop := context.WithCancel(ctx)
				stop()
				switch operation {
				case "exec":
					_, err = q.Exec(canceled, "select $1::integer", 1)
				case "query":
					var rows pgx.Rows
					rows, err = q.Query(canceled, "select $1::integer", 1)
					if rows != nil {
						rows.Close()
					}
				case "query row":
					var value int
					err = q.QueryRow(canceled, "select $1::integer", 1).Scan(&value)
				}
				return err
			}
			var err error
			if transactional {
				err = postgres.RunInTransaction(ctx, db, work)
			} else {
				err = work(ctx)
			}
			if !errors.Is(err, context.Canceled) {
				t.Errorf("%s transactional=%v: error=%v, want context.Canceled", operation, transactional, err)
			}
		}
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
