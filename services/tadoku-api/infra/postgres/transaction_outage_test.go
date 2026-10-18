package postgres_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestRunInTransactionFailedBeginNeverCallsWork(t *testing.T) {
	t.Parallel()

	t.Run("already canceled", func(t *testing.T) {
		db := openPool(t)
		ctx, cancel := context.WithCancel(tenant.WithKey(context.Background(), tenant.Production()))
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
		ctx := tenant.WithKey(context.Background(), tenant.Production())

		called := false
		err := postgres.RunInTransaction(ctx, db, func(context.Context) error { called = true; return nil })
		if err == nil || called {
			t.Errorf("closed-pool begin: error=%v, called=%v", err, called)
		}
		if got := errx.KindOf(err); got != errx.Unknown {
			t.Errorf("closed-pool begin kind=%v, want unknown", got)
		}
	})

	t.Run("waiting for pool connection", func(t *testing.T) {
		db := openPool(t)
		ctx, cancel := context.WithTimeout(tenant.WithKey(context.Background(), tenant.Production()), 5*time.Second)
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

func TestDatabaseOutagesAreUnavailable(t *testing.T) {
	t.Parallel()

	t.Run("unreachable server", func(t *testing.T) {
		config, err := pgxpool.ParseConfig(disposableDSN(t))
		if err != nil {
			t.Fatalf("parse disposable PostgreSQL: %v", err)
		}

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("reserve closed port: %v", err)
		}
		address := listener.Addr().(*net.TCPAddr)
		if err := listener.Close(); err != nil {
			t.Fatalf("close reserved port: %v", err)
		}

		config.ConnConfig.Host = address.IP.String()
		config.ConnConfig.Port = uint16(address.Port)
		config.ConnConfig.Fallbacks = nil

		ctx, cancel := context.WithTimeout(tenant.WithKey(context.Background(), tenant.Production()), 5*time.Second)
		defer cancel()
		db, err := pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			t.Fatalf("open unreachable pool: %v", err)
		}
		defer db.Close()

		called := false
		err = postgres.RunInTransaction(ctx, db, func(context.Context) error { called = true; return nil })
		var connectError *pgconn.ConnectError
		if errx.KindOf(err) != errx.Unavailable || !errors.As(err, &connectError) || called {
			t.Errorf("unreachable begin: kind=%v, error=%v, called=%v", errx.KindOf(err), err, called)
		}

		q, err := postgres.Executor(ctx, db)
		if err != nil {
			t.Fatalf("pool executor: %v", err)
		}
		_, err = q.Exec(ctx, "select 1")
		if got := errx.KindOf(err); got != errx.Unavailable {
			t.Errorf("unreachable exec: kind=%v, error=%v", got, err)
		}
		var one int
		err = q.QueryRow(ctx, "select 1").Scan(&one)
		if got := errx.KindOf(err); got != errx.Unavailable {
			t.Errorf("unreachable query row: kind=%v, error=%v", got, err)
		}
	})

	for _, failAt := range []string{"statement", "commit"} {
		t.Run("terminated backend at "+failAt, func(t *testing.T) {
			ctx, f := newFixture(t)

			calls := 0
			err := postgres.RunInTransaction(ctx, f.db, func(child context.Context) error {
				calls++
				if err := f.books.Add(child, 1, "lost with the backend"); err != nil {
					return err
				}

				q, err := postgres.Executor(child, f.db)
				if err != nil {
					return err
				}
				var pid int32
				if err := q.QueryRow(child, "select pg_backend_pid()").Scan(&pid); err != nil {
					return err
				}
				var terminated bool
				err = f.db.QueryRow(ctx, "select pg_terminate_backend($1, 5000)", pid).Scan(&terminated)
				if err != nil || !terminated {
					t.Fatalf("terminate backend: terminated=%v, error=%v", terminated, err)
				}

				if failAt == "statement" {
					return f.books.Add(child, 2, "after termination")
				}
				return nil
			})

			var pgError *pgconn.PgError
			if got := errx.KindOf(err); got != errx.Unavailable {
				t.Errorf("terminated backend: kind=%v, error=%v", got, err)
			}
			if !errors.As(err, &pgError) || pgError.SQLState() != pgerrcode.AdminShutdown {
				t.Errorf("terminated backend error=%v, want SQLSTATE 57P01", err)
			}
			if calls != 1 {
				t.Errorf("callback calls=%d, want 1", calls)
			}
			if count, err := f.books.Count(ctx); err != nil || count != 0 {
				t.Errorf("terminated backend book count=%d, error=%v, want 0", count, err)
			}
		})
	}
}

func TestRunInTransactionClassifiesOnlyOutageSQLSTATEs(t *testing.T) {
	t.Parallel()
	db := openPool(t)
	ctx := tenant.WithKey(context.Background(), tenant.Production())
	pgError := func(code string) error { return &pgconn.PgError{Code: code} }

	for _, test := range []struct {
		name string
		err  error
		want errx.Kind
	}{
		{name: "08006", err: pgError(pgerrcode.ConnectionFailure), want: errx.Unavailable},
		{name: "08001", err: pgError(pgerrcode.SQLClientUnableToEstablishSQLConnection), want: errx.Unavailable},
		{name: "53300", err: pgError(pgerrcode.TooManyConnections), want: errx.Unavailable},
		{name: "53100", err: pgError(pgerrcode.DiskFull), want: errx.Unavailable},
		{name: "57P01", err: pgError(pgerrcode.AdminShutdown), want: errx.Unavailable},
		{name: "57P03", err: pgError(pgerrcode.CannotConnectNow), want: errx.Unavailable},
		{name: "08P01", err: pgError(pgerrcode.ProtocolViolation), want: errx.Unknown},
		{name: "53400", err: pgError(pgerrcode.ConfigurationLimitExceeded), want: errx.Unknown},
		{name: "57014", err: pgError(pgerrcode.QueryCanceled), want: errx.Unknown},
		{name: "23505", err: pgError(pgerrcode.UniqueViolation), want: errx.Unknown},
		{name: "23503", err: pgError(pgerrcode.ForeignKeyViolation), want: errx.Unknown},
		{name: "42601", err: pgError(pgerrcode.SyntaxError), want: errx.Unknown},
		{name: "42501", err: pgError(pgerrcode.InsufficientPrivilege), want: errx.Unknown},
		{name: "22001", err: pgError(pgerrcode.StringDataRightTruncationDataException), want: errx.Unknown},
		{name: "25P02", err: pgError(pgerrcode.InFailedSQLTransaction), want: errx.Unknown},
		{name: "40001", err: pgError(pgerrcode.SerializationFailure), want: errx.Unknown},
		{name: "not found", err: errx.NewNotFoundError("synthetic missing row"), want: errx.NotFound},
		{name: "no rows", err: pgx.ErrNoRows, want: errx.Unknown},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := postgres.RunInTransaction(ctx, db, func(context.Context) error { return test.err })

			if got := errx.KindOf(err); got != test.want {
				t.Errorf("kind=%v, want %v", got, test.want)
			}
			if !errors.Is(err, test.err) {
				t.Errorf("error=%v lost cause %v", err, test.err)
			}
		})
	}
}
