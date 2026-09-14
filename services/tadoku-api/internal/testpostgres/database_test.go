package testpostgres

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestResetClearsWritesAndPreservesStaticData(t *testing.T) {
	t.Parallel()
	db, err := New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	// Announcements have UUID IDs. Add a test-only owned identity to verify that
	// the same cleanup statement also restarts sequences for future slices.
	if _, err := db.Pool.Exec(t.Context(), "alter table announcements add column test_identity bigint generated always as identity"); err != nil {
		t.Fatal(err)
	}
	const staticState = `select jsonb_build_object(
		'languages', (select jsonb_agg(to_jsonb(l) order by code) from languages l),
		'scoring_rules', (select jsonb_agg(to_jsonb(s) order by id) from scoring_rules s),
		'migration', (select to_jsonb(m) from schema_migrations m)
	)::text`
	var before string
	if err := db.Pool.QueryRow(t.Context(), staticState).Scan(&before); err != nil {
		t.Fatal(err)
	}
	missingSeed := filepath.Join(t.TempDir(), "setup.sql")
	if err := db.Reset(t.Context(), missingSeed, "testdata/announcements.sql", missingSeed); err != nil {
		t.Fatal(err)
	}

	// A separate connection must see committed setup and can commit its own work.
	reader, err := pgx.Connect(t.Context(), db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := reader.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	var count int
	if err := reader.QueryRow(t.Context(), "select count(*) from announcements").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("committed seed rows=%d, want 2", count)
	}
	if _, err := reader.Exec(t.Context(), `
		update announcements set title = 'changed';
		delete from announcements where namespace = 'other';
		insert into announcements (id, namespace, title, content, starts_at, ends_at)
		values ('33333333-3333-4333-8333-333333333333', 'extra', 'inserted', '', '2026-09-12', '2026-09-13');
	`); err != nil {
		t.Fatal(err)
	}
	if err := db.Reset(t.Context(), missingSeed); err != nil {
		t.Fatal(err)
	}
	if err := reader.QueryRow(t.Context(), "select count(*) from announcements").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("rows after cleanup=%d, want 0", count)
	}
	if err := db.Reset(t.Context(), "testdata/announcements.sql"); err != nil {
		t.Fatal(err)
	}
	var title string
	var identity int64
	if err := reader.QueryRow(t.Context(), "select title, test_identity from announcements where namespace = 'main'").Scan(&title, &identity); err != nil {
		t.Fatal(err)
	}
	if title != "original" || identity != 1 {
		t.Errorf("reseeded row: title=%q identity=%d, want original/1", title, identity)
	}
	var after string
	if err := db.Pool.QueryRow(t.Context(), staticState).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Error("cleanup changed migration-seeded reference data or migration bookkeeping")
	}

	if _, err := reader.Exec(t.Context(), "update announcements set title = 'keep me' where namespace = 'main'"); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"testdata", "testdata/invalid.sql"} {
		t.Run(file, func(t *testing.T) {
			err := db.Reset(t.Context(), "testdata/announcements.sql", file)
			if err == nil {
				t.Fatal("accepted an unreadable or invalid seed")
			}
			var pathError *os.PathError
			if file == "testdata" && !errors.As(err, &pathError) {
				t.Errorf("seed read error was not preserved: %v", err)
			}
			if err := reader.QueryRow(t.Context(), "select count(*), min(title) filter (where namespace = 'main') from announcements").Scan(&count, &title); err != nil {
				t.Fatal(err)
			}
			if count != 2 || title != "keep me" {
				t.Errorf("failed setup changed prior state: rows=%d title=%q", count, title)
			}
		})
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := db.Reset(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("canceled reset=%v, want context.Canceled", err)
	}
	if err := db.Reset(t.Context()); err != nil {
		t.Fatalf("reset after failures: %v", err)
	}
}

func TestNewRejectsUnsafeConfigurationAndCanceledSetup(t *testing.T) {
	valid := os.Getenv("TADOKU_TEST_POSTGRES_URL")
	for _, dsn := range []string{
		"", "%", "postgres://127.0.0.1:1/postgres?sslmode=disable",
		"postgres://postgres:postgres@remote.test:5432/postgres?sslmode=disable",
		"postgres://postgres:postgres@127.0.0.1:5432/immersion?sslmode=disable",
		"postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable&options=x",
	} {
		t.Setenv("TADOKU_TEST_POSTGRES_URL", dsn)
		if db, err := New(t.Context()); err == nil || db != nil {
			t.Fatal("accepted unsafe test database configuration")
		}
	}
	t.Setenv("TADOKU_TEST_POSTGRES_URL", valid)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if db, err := New(ctx); !errors.Is(err, context.Canceled) || db != nil {
		t.Fatalf("canceled setup: db=%v error=%v", db, err)
	}
}
