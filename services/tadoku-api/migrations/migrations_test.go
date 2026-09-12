package migrations_test

import (
	"context"
	"errors"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestCanonicalHistoryReachesVersion29AndThenDoesNothing(t *testing.T) {
	t.Parallel()
	db := testpostgres.New(t)
	var version int
	var dirty bool
	if err := db.Pool.QueryRow(context.Background(), "select version, dirty from schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatal(err)
	}
	if version != 29 || dirty {
		t.Fatalf("migration state: version=%d dirty=%v", version, dirty)
	}
	for _, table := range []string{"logs", "pages", "profiles", "moderation_audit_log"} {
		var exists bool
		if err := db.Pool.QueryRow(context.Background(), "select to_regclass($1) is not null", table).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Errorf("missing table %s", table)
		}
	}
	first, err := bazel.Runfile("services/immersion-api/storage/postgres/migrations/0001_init.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	migrator, err := migrate.New((&url.URL{Scheme: "file", Path: filepath.Dir(first)}).String(), db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sourceErr, databaseErr := migrator.Close()
		if sourceErr != nil || databaseErr != nil {
			t.Errorf("close migrator: %v, %v", sourceErr, databaseErr)
		}
	})
	if err := migrator.Up(); !errors.Is(err, migrate.ErrNoChange) {
		t.Errorf("second migration=%v want ErrNoChange", err)
	}
}
