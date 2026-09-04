package migrations_test

import (
	"database/sql"
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
	"github.com/stretchr/testify/require"
)

func TestCanonicalHistoryReachesVersion29AndThenDoesNothing(t *testing.T) {
	adminURL := os.Getenv("IMMERSION_TEST_POSTGRES_URL")
	if adminURL == "" {
		t.Skip("IMMERSION_TEST_POSTGRES_URL is not set; skipping PostgreSQL integration test")
	}

	parsedAdminURL, err := url.Parse(adminURL)
	require.NoError(t, err)
	require.Equal(t, "postgres", parsedAdminURL.Scheme)

	adminDB, err := sql.Open("postgres", adminURL)
	require.NoError(t, err)
	require.NoError(t, adminDB.Ping())

	databaseName := "tadoku_migrations_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = adminDB.Exec(`create database "` + databaseName + `"`)
	require.NoError(t, err)

	testURL := *parsedAdminURL
	testURL.Path = "/" + databaseName
	testDB, err := sql.Open("postgres", testURL.String())
	require.NoError(t, err)
	require.NoError(t, testDB.Ping())
	t.Cleanup(func() {
		require.NoError(t, testDB.Close())
		_, dropErr := adminDB.Exec(`drop database "` + databaseName + `" with (force)`)
		require.NoError(t, dropErr)
		require.NoError(t, adminDB.Close())
	})

	firstMigration, err := bazel.Runfile("services/immersion-api/storage/postgres/migrations/0001_init.up.sql")
	require.NoError(t, err)
	sourceURL := (&url.URL{Scheme: "file", Path: filepath.Dir(firstMigration)}).String()
	migrator, err := migrate.New(sourceURL, testURL.String())
	require.NoError(t, err)
	require.NoError(t, migrator.Up())

	var version int
	var dirty bool
	require.NoError(t, testDB.QueryRow("select version, dirty from schema_migrations").Scan(&version, &dirty))
	require.Equal(t, 29, version)
	require.False(t, dirty)

	for _, table := range []string{"logs", "pages", "profiles", "moderation_audit_log"} {
		var exists bool
		require.NoError(t, testDB.QueryRow("select to_regclass($1) is not null", table).Scan(&exists))
		require.True(t, exists, "%s must exist", table)
	}

	require.ErrorIs(t, migrator.Up(), migrate.ErrNoChange)
	sourceCloseErr, databaseCloseErr := migrator.Close()
	require.NoError(t, sourceCloseErr)
	require.NoError(t, databaseCloseErr)
}
