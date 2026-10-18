package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

const testDSNVariable = "TADOKU_TEST_POSTGRES_URL"

// Test safety: Use only a disposable local PostgreSQL instance with synthetic credentials.
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

	ctx, cancel := context.WithTimeout(tenant.WithKey(context.Background(), tenant.Production()), 5*time.Second)
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

	ctx, cancel := context.WithTimeout(tenant.WithKey(context.Background(), tenant.Production()), 15*time.Second)
	t.Cleanup(cancel)

	if _, err := db.Exec(ctx, "create schema "+schema); err != nil {
		t.Fatalf("create synthetic schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(tenant.WithKey(context.Background(), tenant.Production()), 5*time.Second)
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

type rejection struct{ reason string }

func (e *rejection) Error() string { return e.reason }
