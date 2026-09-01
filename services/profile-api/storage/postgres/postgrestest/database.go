package postgrestest

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const databaseURLEnvironmentVariable = "PROFILE_TEST_POSTGRES_URL"

var cachedValidatedMigrationDirectory = sync.OnceValues(findValidatedMigrationDirectory)

// OpenMigratedDatabase creates an isolated database and applies every checked-in
// profile migration. The caller receives the schema production reaches from an
// empty database.
func OpenMigratedDatabase(t testing.TB) *sql.DB {
	t.Helper()

	adminURL := os.Getenv(databaseURLEnvironmentVariable)
	if adminURL == "" {
		t.Skipf("%s is not set; skipping PostgreSQL integration test", databaseURLEnvironmentVariable)
	}

	parsedAdminURL, err := url.Parse(adminURL)
	require.NoError(t, err)
	require.Equal(t, "postgres", parsedAdminURL.Scheme)

	adminDB, err := sql.Open("postgres", adminURL)
	require.NoError(t, err)
	require.NoError(t, adminDB.Ping())

	databaseName := "profile_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = adminDB.Exec(`create database "` + databaseName + `"`)
	require.NoError(t, err)

	testDatabaseURL := *parsedAdminURL
	testDatabaseURL.Path = "/" + databaseName
	testDB, err := sql.Open("postgres", testDatabaseURL.String())
	require.NoError(t, err)
	t.Cleanup(func() {
		if testDB != nil {
			require.NoError(t, testDB.Close())
		}
		_, dropErr := adminDB.Exec(`drop database "` + databaseName + `" with (force)`)
		require.NoError(t, dropErr)
		require.NoError(t, adminDB.Close())
	})
	require.NoError(t, testDB.Ping())

	migrationDirectory, err := cachedValidatedMigrationDirectory()
	require.NoError(t, err)
	migrationSourceURL := (&url.URL{Scheme: "file", Path: migrationDirectory}).String()
	migrator, err := migrate.New(migrationSourceURL, testDatabaseURL.String())
	require.NoError(t, err)
	migrationErr := migrator.Up()
	sourceCloseErr, databaseCloseErr := migrator.Close()
	require.NoError(t, migrationErr)
	require.NoError(t, sourceCloseErr)
	require.NoError(t, databaseCloseErr)

	return testDB
}

func findValidatedMigrationDirectory() (string, error) {
	firstMigration, err := bazel.Runfile("services/profile-api/storage/postgres/migrations/0001_init.up.sql")
	if err != nil {
		return "", fmt.Errorf("could not resolve first profile migration: %w", err)
	}
	directory := filepath.Dir(firstMigration)

	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", fmt.Errorf("could not read profile migration directory: %w", err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		paths = append(paths, filepath.Join(directory, entry.Name()))
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return "", fmt.Errorf("profile migration directory contains no up migrations")
	}
	if filepath.Base(paths[0]) != "0001_init.up.sql" {
		return "", fmt.Errorf("profile migration sequence does not start at 0001: %s", paths[0])
	}

	for index, path := range paths {
		expectedPrefix := fmt.Sprintf("%04d_", index+1)
		if !strings.HasPrefix(filepath.Base(path), expectedPrefix) {
			return "", fmt.Errorf("profile migration sequence contains a gap at %s", path)
		}
	}

	return directory, nil
}
