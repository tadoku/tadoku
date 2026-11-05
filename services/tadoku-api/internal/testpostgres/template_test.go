package testpostgres

import (
	"context"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestEnsureTemplateBuildsOneTemplateForConcurrentCallers(t *testing.T) {
	t.Parallel()

	admin, dir, name := templateFixture(t)

	errs := make(chan error, 4)
	for range 4 {
		go func() { errs <- ensureTemplate(t.Context(), admin, dir, name) }()
	}
	for range 4 {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}

	requireMigratedTemplate(t, admin, dir, name)
}

func TestEnsureTemplateRebuildsAnInterruptedBuild(t *testing.T) {
	t.Parallel()

	admin, dir, name := templateFixture(t)
	if _, err := admin.Exec(t.Context(), "create database "+pgx.Identifier{name}.Sanitize()); err != nil {
		t.Fatal(err)
	}

	if err := ensureTemplate(t.Context(), admin, dir, name); err != nil {
		t.Fatal(err)
	}

	requireMigratedTemplate(t, admin, dir, name)
}

func templateFixture(t *testing.T) (*pgxpool.Pool, string, string) {
	t.Helper()

	config, err := adminConfig()
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)

	dir, err := migrationsDir()
	if err != nil {
		t.Fatal(err)
	}
	name := "tadoku_test_template_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	t.Cleanup(func() { dropTemplate(t, admin, name) })
	return admin, dir, name
}

func requireMigratedTemplate(t *testing.T, admin *pgxpool.Pool, dir, name string) {
	t.Helper()

	var isTemplate, allowsConnections bool
	err := admin.QueryRow(t.Context(), "select datistemplate, datallowconn from pg_database where datname = $1", name).
		Scan(&isTemplate, &allowsConnections)
	if err != nil {
		t.Fatal(err)
	}
	if !isTemplate || allowsConnections {
		t.Fatalf("template state: is_template=%v allow_connections=%v", isTemplate, allowsConnections)
	}

	clone := name + "_clone"
	cloneSQL := pgx.Identifier{clone}.Sanitize()
	if _, err := admin.Exec(t.Context(), "create database "+cloneSQL+" template "+pgx.Identifier{name}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "drop database "+cloneSQL+" with (force)"); err != nil {
			t.Error(err)
		}
	})

	config := admin.Config().ConnConfig.Copy()
	config.Database = clone
	conn, err := pgx.ConnectConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())
	var version int
	var dirty bool
	if err := conn.QueryRow(t.Context(), "select version, dirty from schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatal(err)
	}
	if want := latestMigration(t, dir); version != want || dirty {
		t.Errorf("clone migration state: version=%d dirty=%v, want version=%d clean", version, dirty, want)
	}
}

func latestMigration(t *testing.T, dir string) int {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("list migrations: %v", err)
	}
	versions := make([]int, 0, len(files))
	for _, file := range files {
		version, err := strconv.Atoi(strings.SplitN(filepath.Base(file), "_", 2)[0])
		if err != nil {
			t.Fatal(err)
		}
		versions = append(versions, version)
	}
	return slices.Max(versions)
}

func dropTemplate(t *testing.T, admin *pgxpool.Pool, name string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	nameSQL := pgx.Identifier{name}.Sanitize()
	if _, err := admin.Exec(ctx, "alter database "+nameSQL+" with is_template false"); err != nil {
		t.Errorf("unmark template: %v", err)
	}
	if _, err := admin.Exec(ctx, "drop database "+nameSQL+" with (force)"); err != nil {
		t.Errorf("drop template: %v", err)
	}
}
