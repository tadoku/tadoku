package testpostgres

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
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
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type Database struct {
	Pool    *pgxpool.Pool
	AppPool *pgxpool.Pool
	DSN     string

	admin    *pgxpool.Pool
	seedPool *pgxpool.Pool
	name     string
	appRole  string
	seedRole string
}

func New(ctx context.Context) (_ *Database, err error) {
	adminConfig, err := adminConfig()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	admin, err := pgxpool.NewWithConfig(ctx, adminConfig)
	if err != nil {
		return nil, fmt.Errorf("open test admin: %w", err)
	}
	db := &Database{admin: admin}
	defer func() {
		if err != nil {
			err = errors.Join(err, db.Close())
		}
	}()

	dir, err := migrationsDir()
	if err != nil {
		return nil, err
	}
	template, err := templateName(dir)
	if err != nil {
		return nil, err
	}
	if err := ensureTemplate(ctx, admin, dir, template); err != nil {
		return nil, err
	}

	name := "tadoku_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	createSQL := "create database " + pgx.Identifier{name}.Sanitize() + " template " + pgx.Identifier{template}.Sanitize()
	if _, err := admin.Exec(ctx, createSQL); err != nil {
		return nil, fmt.Errorf("create disposable database: %w", err)
	}
	db.name = name

	u, err := url.Parse(adminConfig.ConnString())
	if err != nil {
		return nil, fmt.Errorf("parse test admin DSN: %w", err)
	}
	u.Path = "/" + name
	db.DSN = u.String()

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

	role := name + "_app"
	roleSQL := pgx.Identifier{role}.Sanitize()
	password := uuid.NewString()
	if _, err := admin.Exec(ctx, "create role "+roleSQL+" login nosuperuser nobypassrls password '"+password+"'"); err != nil {
		return nil, fmt.Errorf("create test application role: %w", err)
	}
	db.appRole = role

	grants := "grant usage on schema public to " + roleSQL + ";" +
		"grant select, insert, update, delete on all tables in schema public to " + roleSQL + ";" +
		"grant usage, select on all sequences in schema public to " + roleSQL + ";" +
		"revoke insert, update, delete on schema_migrations from " + roleSQL
	if _, err := db.Pool.Exec(ctx, grants); err != nil {
		return nil, fmt.Errorf("grant test application privileges: %w", err)
	}

	appCfg := cfg.Copy()
	appCfg.ConnConfig.User = role
	appCfg.ConnConfig.Password = password
	db.AppPool, err = pgxpool.NewWithConfig(ctx, appCfg)
	if err != nil {
		return nil, fmt.Errorf("open test application pool: %w", err)
	}
	if err := db.AppPool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping test application pool: %w", err)
	}

	if err := db.openSeedPool(ctx, cfg, name+"_seed"); err != nil {
		return nil, err
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

func adminConfig() (*pgxpool.Config, error) {
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

	config, err := pgxpool.ParseConfig(raw)
	if err != nil {
		return nil, fmt.Errorf("configure test admin: %w", err)
	}
	config.MaxConns = 2
	return config, nil
}

func migrationsDir() (string, error) {
	first, err := bazel.Runfile("services/tadoku-api/migrations/0001_init.up.sql")
	if err != nil {
		return "", fmt.Errorf("resolve canonical migrations: %w", err)
	}
	return filepath.Dir(first), nil
}

func templateName(dir string) (string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return "", fmt.Errorf("list canonical migrations: %w", err)
	}
	if len(files) == 0 {
		return "", errors.New("no canonical up migrations")
	}

	hash := sha256.New()
	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read canonical migration: %w", err)
		}
		fmt.Fprintf(hash, "%s\x00%d\x00", filepath.Base(file), len(contents))
		hash.Write(contents)
	}
	return "tadoku_test_template_" + hex.EncodeToString(hash.Sum(nil))[:16], nil
}

func ensureTemplate(ctx context.Context, admin *pgxpool.Pool, dir, name string) (err error) {
	conn, err := admin.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire template connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "select pg_advisory_lock(hashtextextended($1, 0))", name); err != nil {
		return fmt.Errorf("lock migrated template: %w", err)
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, unlockErr := conn.Exec(unlockCtx, "select pg_advisory_unlock(hashtextextended($1, 0))", name); unlockErr != nil {
			err = errors.Join(err, fmt.Errorf("unlock migrated template: %w", unlockErr), conn.Conn().Close(unlockCtx))
		}
	}()

	var ready bool
	err = conn.QueryRow(ctx, "select datistemplate from pg_database where datname = $1", name).Scan(&ready)
	if err == nil && ready {
		return nil
	}
	nameSQL := pgx.Identifier{name}.Sanitize()
	switch {
	case errors.Is(err, pgx.ErrNoRows):
	case err != nil:
		return fmt.Errorf("inspect migrated template: %w", err)
	default:
		if _, err := conn.Exec(ctx, "drop database "+nameSQL+" with (force)"); err != nil {
			return fmt.Errorf("drop interrupted template: %w", err)
		}
	}

	if _, err := conn.Exec(ctx, "create database "+nameSQL); err != nil {
		return fmt.Errorf("create migrated template: %w", err)
	}
	u, err := url.Parse(admin.Config().ConnString())
	if err != nil {
		return fmt.Errorf("parse test admin DSN: %w", err)
	}
	u.Path = "/" + name

	migrator, err := migrate.New((&url.URL{Scheme: "file", Path: dir}).String(), u.String())
	if err != nil {
		return fmt.Errorf("open canonical migrations: %w", err)
	}
	migrationErr := migrator.Up()
	sourceErr, databaseErr := migrator.Close()
	if err := errors.Join(migrationErr, sourceErr, databaseErr); err != nil {
		return fmt.Errorf("migrate template: %w", err)
	}

	if _, err := conn.Exec(ctx, "alter database "+nameSQL+" with is_template true allow_connections false"); err != nil {
		return fmt.Errorf("mark migrated template: %w", err)
	}
	return nil
}

func (d *Database) Close() error {
	if d.seedPool != nil {
		d.seedPool.Close()
	}
	if d.AppPool != nil {
		d.AppPool.Close()
	}
	if d.Pool != nil {
		d.Pool.Close()
	}

	defer d.admin.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cleanupErr error
	if d.name != "" {
		if _, err := d.admin.Exec(ctx, "drop database "+pgx.Identifier{d.name}.Sanitize()+" with (force)"); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("drop disposable database: %w", err))
		}
	}
	if d.appRole != "" {
		if _, err := d.admin.Exec(ctx, "drop role "+pgx.Identifier{d.appRole}.Sanitize()); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("drop test application role: %w", err))
		}
	}
	if d.seedRole != "" {
		if _, err := d.admin.Exec(ctx, "drop role "+pgx.Identifier{d.seedRole}.Sanitize()); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("drop test seeding role: %w", err))
		}
	}
	return cleanupErr
}

//go:embed cleanup.sql
var cleanupSQL string

// Reset must run only between sequential scenarios after their database work
// finishes; setup commits before the next request runs.
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

	if _, err := tx.Exec(ctx, "select set_config('tadoku.tenant', $1, true)", tenant.Production().String()); err != nil {
		return fmt.Errorf("set test reset tenant: %w", err)
	}
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
