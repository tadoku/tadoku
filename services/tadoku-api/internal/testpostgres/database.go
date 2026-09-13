// Package testpostgres assembles disposable, fully migrated test databases.
// It is test-only in Bazel and never accepts production credentials.
package testpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type Database struct {
	Pool *pgxpool.Pool
	SQL  *sql.DB // only the legacy comparison/migration runner uses database/sql
	DSN  string
}

func New(t testing.TB) *Database {
	t.Helper()
	raw := os.Getenv("TADOKU_TEST_POSTGRES_URL")
	u, err := url.Parse(raw)
	if err != nil || u == nil {
		t.Fatal("TADOKU_TEST_POSTGRES_URL must be a disposable loopback PostgreSQL DSN")
	}
	password, _ := u.User.Password()
	if u.Scheme != "postgres" || (u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1") || u.Port() == "" || u.Path != "/postgres" || u.User.Username() != "postgres" || password != "postgres" || u.RawQuery != "sslmode=disable" || u.Fragment != "" {
		t.Fatal("TADOKU_TEST_POSTGRES_URL must use postgres:postgres on loopback with an explicit port, /postgres and sslmode=disable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, raw)
	if err != nil {
		t.Fatalf("open test admin: %v", err)
	}
	t.Cleanup(admin.Close)
	name := "tadoku_native_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, `create database "`+name+`"`); err != nil {
		t.Fatalf("create disposable database: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, `drop database "`+name+`" with (force)`); err != nil {
			t.Errorf("drop disposable database: %v", err)
		}
	})
	u.Path = "/" + name
	dsn := u.String()
	first, err := bazel.Runfile("services/immersion-api/storage/postgres/migrations/0001_init.up.sql")
	if err != nil {
		t.Fatalf("resolve canonical migrations: %v", err)
	}
	migrator, err := migrate.New((&url.URL{Scheme: "file", Path: filepath.Dir(first)}).String(), dsn)
	if err != nil {
		t.Fatalf("open canonical migrations: %v", err)
	}
	migrationErr := migrator.Up()
	sourceErr, databaseErr := migrator.Close()
	if migrationErr != nil {
		t.Fatalf("migrate test database: %v", migrationErr)
	}
	if sourceErr != nil || databaseErr != nil {
		t.Fatalf("close migrator: %v, %v", sourceErr, databaseErr)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.MaxConns = 4
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	legacy, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	legacy.SetMaxOpenConns(4)
	t.Cleanup(func() {
		if err := legacy.Close(); err != nil {
			t.Errorf("close legacy pool: %v", err)
		}
	})
	return &Database{Pool: pool, SQL: legacy, DSN: dsn}
}

// SeedAnnouncement only writes fixture data; the caller chooses publication times.
func (d *Database) SeedAnnouncement(t testing.TB, namespace, title string, start, end time.Time, deleted bool) string {
	t.Helper()
	id := uuid.NewString()
	var deletion *time.Time
	if deleted {
		deletion = &start
	}
	_, err := d.Pool.Exec(context.Background(), `insert into announcements
		(id, namespace, title, content, style, href, starts_at, ends_at, created_at, updated_at, deleted_at)
		values ($1,$2,$3,$4,'info',null,$5,$6,$5,$5,$7)`, id, namespace, title, fmt.Sprintf("<p>%s</p>", title), start, end, deletion)
	if err != nil {
		t.Fatalf("seed announcement: %v", err)
	}
	return id
}
