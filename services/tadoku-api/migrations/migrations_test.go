package migrations_test

import (
	"database/sql"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func TestCanonicalHistoryReachesVersion29AndThenDoesNothing(t *testing.T) {
	adminURL := os.Getenv("IMMERSION_TEST_POSTGRES_URL")
	if adminURL == "" {
		t.Skip("IMMERSION_TEST_POSTGRES_URL is not set; skipping PostgreSQL integration test")
	}

	parsedAdminURL, err := url.Parse(adminURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsedAdminURL.Scheme != "postgres" {
		t.Fatalf("got %v, want %v", parsedAdminURL.Scheme, "postgres")
	}

	adminDB, err := sql.Open("postgres", adminURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := adminDB.Ping(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	databaseName := "tadoku_migrations_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = adminDB.Exec(`create database "` + databaseName + `"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testURL := *parsedAdminURL
	testURL.Path = "/" + databaseName
	testDB, err := sql.Open("postgres", testURL.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := testDB.Ping(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := testDB.Close(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, dropErr := adminDB.Exec(`drop database "` + databaseName + `" with (force)`)
		if err := dropErr; err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := adminDB.Close(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	firstMigration, err := bazel.Runfile("services/immersion-api/storage/postgres/migrations/0001_init.up.sql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sourceURL := (&url.URL{Scheme: "file", Path: filepath.Dir(firstMigration)}).String()
	migrator, err := migrate.New(sourceURL, testURL.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := migrator.Up(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var version int
	var dirty bool
	if err := testDB.QueryRow("select version, dirty from schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != 29 {
		t.Fatalf("got %v, want %v", version, 29)
	}
	if dirty {
		t.Fatalf("unexpected boolean: %v", dirty)
	}

	for _, table := range []string{"logs", "pages", "profiles", "moderation_audit_log"} {
		var exists bool
		if err := testDB.QueryRow("select to_regclass($1) is not null", table).Scan(&exists); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Fatalf("unexpected boolean: %v", exists)
		}
	}

	if err := migrator.Up(); !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("error = %v, want %v", err, migrate.ErrNoChange)
	}
	sourceCloseErr, databaseCloseErr := migrator.Close()
	if err := sourceCloseErr; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := databaseCloseErr; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
