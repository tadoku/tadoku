package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

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
	blocked, cancel := context.WithCancel(ctx)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		result <- postgres.RunInTransaction(blocked, f.db, func(child context.Context) error {
			return f.books.Add(child, 1, "must be canceled")
		})
	}()

	// Test safety: Observe the insert waiting on this fixture's table lock before canceling.
	// Starting a short deadline before Begin races with connection setup on CI.
	waiting, stopWaiting := context.WithTimeout(ctx, 5*time.Second)
	defer stopWaiting()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waitingForLock bool
		err := lock.QueryRow(waiting, `select exists (
			select 1 from pg_locks
			where relation = $1::regclass and mode = 'RowExclusiveLock' and not granted
		)`, f.schema+".book").Scan(&waitingForLock)
		if err != nil {
			t.Fatalf("observe blocked SQL: %v", err)
		}
		if waitingForLock {
			break
		}
		select {
		case err := <-result:
			t.Fatalf("transaction completed before SQL waited for the lock: %v", err)
		case <-waiting.Done():
			t.Fatal("SQL did not wait for the table lock within the setup bound")
		case <-ticker.C:
		}
	}

	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("blocked SQL error=%v, want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("blocked SQL exceeded the cancellation bound")
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

func TestExecutorForwardsCancellationForEveryOperation(t *testing.T) {
	t.Parallel()
	db := openPool(t)
	ctx, cancel := context.WithTimeout(tenant.WithKey(context.Background(), tenant.Production()), 5*time.Second)
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
