package postgrestest

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
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

const databaseURLEnvironmentVariable = "IMMERSION_TEST_POSTGRES_URL"

// OpenMigratedDatabase creates an isolated database and applies every checked-in
// immersion migration. The caller receives a database at the same schema version
// production would reach from an empty database.
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

	databaseName := "immersion_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = adminDB.Exec(`create database "` + databaseName + `"`)
	require.NoError(t, err)

	testDatabaseURL := *parsedAdminURL
	testDatabaseURL.Path = "/" + databaseName
	testDB, err := sql.Open("postgres", testDatabaseURL.String())
	t.Cleanup(func() {
		if testDB != nil {
			require.NoError(t, testDB.Close())
		}
		_, dropErr := adminDB.Exec(`drop database "` + databaseName + `" with (force)`)
		require.NoError(t, dropErr)
		require.NoError(t, adminDB.Close())
	})
	require.NoError(t, err)
	require.NoError(t, testDB.Ping())

	migrationDirectory := validatedMigrationDirectory(t)
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

func validatedMigrationDirectory(t testing.TB) string {
	t.Helper()

	firstMigration, err := bazel.Runfile("services/immersion-api/storage/postgres/migrations/0001_init.up.sql")
	require.NoError(t, err)
	directory := filepath.Dir(firstMigration)

	entries, err := os.ReadDir(directory)
	require.NoError(t, err)

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		paths = append(paths, filepath.Join(directory, entry.Name()))
	}
	sort.Strings(paths)
	require.NotEmpty(t, paths)
	require.Equal(t, "0001_init.up.sql", filepath.Base(paths[0]))

	for index, path := range paths {
		expectedPrefix := fmt.Sprintf("%04d_", index+1)
		require.Truef(t, strings.HasPrefix(filepath.Base(path), expectedPrefix), "migration sequence contains a gap at %s", path)
	}

	return directory
}
