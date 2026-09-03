package migrations_test

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const postgresTestURLVariable = "IMMERSION_TEST_POSTGRES_URL"

func TestContentSchemaMatchesCanonicalSource(t *testing.T) {
	target := openMigratedDatabase(t, "services/immersion-api/storage/postgres/migrations/0001_init.up.sql")
	source := openMigratedDatabase(t, "services/content-api/storage/postgres/migrations/0001_create_pages.up.sql")
	tables := []string{"announcements", "pages", "pages_content", "posts", "posts_content"}

	require.Equal(t, schemaFingerprint(t, source, tables), schemaFingerprint(t, target, tables))
	for _, table := range tables {
		var count int64
		err := target.QueryRow("select count(*) from " + pq.QuoteIdentifier(table)).Scan(&count)
		require.NoError(t, err)
		require.Zero(t, count, "canonical target table %s must start empty", table)
	}

	var version int
	var dirty bool
	require.NoError(t, target.QueryRow("select version, dirty from schema_migrations").Scan(&version, &dirty))
	require.Equal(t, 28, version)
	require.False(t, dirty)
}

func openMigratedDatabase(t *testing.T, firstMigrationRunfile string) *sql.DB {
	t.Helper()

	adminURL := os.Getenv(postgresTestURLVariable)
	if adminURL == "" {
		t.Skipf("%s is not set; skipping PostgreSQL integration test", postgresTestURLVariable)
	}

	parsedAdminURL, err := url.Parse(adminURL)
	require.NoError(t, err)
	require.Equal(t, "postgres", parsedAdminURL.Scheme)

	adminDB, err := sql.Open("postgres", adminURL)
	require.NoError(t, err)
	require.NoError(t, adminDB.Ping())

	databaseName := "schema_convergence_test_" + uuid.NewString()
	_, err = adminDB.Exec(`create database ` + pq.QuoteIdentifier(databaseName))
	require.NoError(t, err)

	testDatabaseURL := *parsedAdminURL
	testDatabaseURL.Path = "/" + databaseName
	testDB, err := sql.Open("postgres", testDatabaseURL.String())
	require.NoError(t, err)
	require.NoError(t, testDB.Ping())
	t.Cleanup(func() {
		require.NoError(t, testDB.Close())
		_, dropErr := adminDB.Exec(`drop database ` + pq.QuoteIdentifier(databaseName) + ` with (force)`)
		require.NoError(t, dropErr)
		require.NoError(t, adminDB.Close())
	})

	firstMigration, err := bazel.Runfile(firstMigrationRunfile)
	require.NoError(t, err)
	migrationSourceURL := (&url.URL{Scheme: "file", Path: filepath.Dir(firstMigration)}).String()
	migrator, err := migrate.New(migrationSourceURL, testDatabaseURL.String())
	require.NoError(t, err)
	require.NoError(t, migrator.Up())
	sourceCloseErr, databaseCloseErr := migrator.Close()
	require.NoError(t, sourceCloseErr)
	require.NoError(t, databaseCloseErr)

	return testDB
}

func schemaFingerprint(t *testing.T, db *sql.DB, tables []string) []string {
	t.Helper()

	rows, err := db.Query(`
		select concat_ws('|', 'column', table_name, ordinal_position, column_name,
			data_type, coalesce(character_maximum_length::text, ''), is_nullable,
			coalesce(column_default, ''))
		from information_schema.columns
		where table_schema = current_schema()
		  and table_name = any($1)
		union all
		select concat_ws('|', 'constraint', relation.relname, constraint_name.conname,
			constraint_name.contype, pg_get_constraintdef(constraint_name.oid, true),
			constraint_name.convalidated)
		from pg_constraint as constraint_name
		join pg_class as relation on relation.oid = constraint_name.conrelid
		join pg_namespace as namespace on namespace.oid = relation.relnamespace
		where namespace.nspname = current_schema()
		  and relation.relname = any($1)
		union all
		select concat_ws('|', 'index', tablename, indexname, indexdef)
		from pg_indexes
		where schemaname = current_schema()
		  and tablename = any($1)
		order by 1
	`, pq.Array(tables))
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()

	var fingerprint []string
	for rows.Next() {
		var entry string
		require.NoError(t, rows.Scan(&entry))
		fingerprint = append(fingerprint, entry)
	}
	require.NoError(t, rows.Err())
	require.NotEmpty(t, fingerprint, fmt.Sprintf("schema fingerprint for %v", tables))
	return fingerprint
}
