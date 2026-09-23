// Package testpostgres assembles disposable, fully migrated test databases.
// It is test-only in Bazel and never accepts production credentials.
package testpostgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
	DSN  string

	admin *pgxpool.Pool
	name  string
}

func New(ctx context.Context) (_ *Database, err error) {
	raw := os.Getenv("TADOKU_TEST_POSTGRES_URL")
	u, err := url.Parse(raw)
	if err != nil || u == nil || u.User == nil {
		return nil, errors.New("TADOKU_TEST_POSTGRES_URL must be a disposable loopback PostgreSQL DSN")
	}

	password, _ := u.User.Password()
	if u.Scheme != "postgres" ||
		(u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1") ||
		u.Port() == "" ||
		u.Path != "/postgres" ||
		u.RawPath != "" ||
		u.Opaque != "" ||
		u.User.Username() != "postgres" ||
		password != "postgres" ||
		u.RawQuery != "sslmode=disable" ||
		u.Fragment != "" {
		return nil, errors.New("TADOKU_TEST_POSTGRES_URL must use postgres:postgres on loopback with an explicit port, /postgres and sslmode=disable")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	admin, err := pgxpool.New(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("open test admin: %w", err)
	}
	db := &Database{admin: admin}
	defer func() {
		if err != nil {
			err = errors.Join(err, db.Close())
		}
	}()

	name := "tadoku_native_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, `create database "`+name+`"`); err != nil {
		return nil, fmt.Errorf("create disposable database: %w", err)
	}
	db.name = name

	u.Path = "/" + name
	db.DSN = u.String()

	first, err := bazel.Runfile("services/tadoku-api/migrations/0001_init.up.sql")
	if err != nil {
		return nil, fmt.Errorf("resolve canonical migrations: %w", err)
	}

	migrator, err := migrate.New((&url.URL{Scheme: "file", Path: filepath.Dir(first)}).String(), db.DSN)
	if err != nil {
		return nil, fmt.Errorf("open canonical migrations: %w", err)
	}

	migrationErr := migrator.Up()
	sourceErr, databaseErr := migrator.Close()
	if err := errors.Join(migrationErr, sourceErr, databaseErr); err != nil {
		return nil, fmt.Errorf("migrate test database: %w", err)
	}

	cfg, err := pgxpool.ParseConfig(db.DSN)
	if err != nil {
		return nil, fmt.Errorf("configure test pool: %w", err)
	}
	cfg.MaxConns = 4

	db.Pool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open test pool: %w", err)
	}
	if err := db.Pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping test pool: %w", err)
	}
	if _, err := db.Pool.Exec(ctx, "create table tadoku_test_language_baseline as select code, name from languages"); err != nil {
		return nil, fmt.Errorf("snapshot language baseline: %w", err)
	}
	if _, err := db.Pool.Exec(ctx, "create table tadoku_test_scoring_rule_sets_baseline as select * from scoring_rule_sets"); err != nil {
		return nil, fmt.Errorf("snapshot scoring rule sets baseline: %w", err)
	}
	if _, err := db.Pool.Exec(ctx, "create table tadoku_test_scoring_rules_baseline as select * from scoring_rules"); err != nil {
		return nil, fmt.Errorf("snapshot scoring rules baseline: %w", err)
	}
	if _, err := db.Pool.Exec(ctx, "create table tadoku_test_platform_scoring_config_baseline as select * from platform_scoring_config"); err != nil {
		return nil, fmt.Errorf("snapshot platform scoring config baseline: %w", err)
	}
	return db, nil
}

// Close releases the pool and drops only the database created by New. Cleanup
// has its own deadline so canceled tests can still release their resources.
func (d *Database) Close() error {
	if d.Pool != nil {
		d.Pool.Close()
	}
	defer d.admin.Close()
	if d.name != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := d.admin.Exec(ctx, `drop database "`+d.name+`" with (force)`); err != nil {
			return fmt.Errorf("drop disposable database: %w", err)
		}
	}
	return nil
}

//go:embed cleanup.sql
var cleanupSQL string

// Reset clears the explicitly listed mutable tables and loads existing SQL fixtures.
// Missing seed files mean no setup is needed; other read and SQL errors fail reset.
// Setup commits before requests run; it never encloses application transactions.
// Call only between sequential scenarios, after their database work has finished.
func (d *Database) Reset(ctx context.Context, seedFiles ...string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin test reset: %w", err)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if rollbackErr := tx.Rollback(cleanupCtx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			err = errors.Join(err, fmt.Errorf("rollback test reset: %w", rollbackErr))
		}
	}()

	if _, err := tx.Exec(ctx, cleanupSQL); err != nil {
		return fmt.Errorf("execute cleanup.sql: %w", err)
	}
	for _, path := range seedFiles {
		seed, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read seed %s: %w", path, err)
		}
		if _, err := tx.Exec(ctx, string(seed)); err != nil {
			return fmt.Errorf("execute seed %s: %w", path, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit test reset: %w", err)
	}
	return nil
}
