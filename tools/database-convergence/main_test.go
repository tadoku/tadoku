package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPostgresURLVariable = "IMMERSION_TEST_POSTGRES_URL"

var migrationRunfiles = map[string]string{
	"content": "services/content-api/storage/postgres/migrations/0001_create_pages.up.sql",
	"profile": "services/profile-api/storage/postgres/migrations/0001_init.up.sql",
	"authz":   "services/authz-api/storage/postgres/migrations/0001_init.up.sql",
	"target":  "services/immersion-api/storage/postgres/migrations/0001_init.up.sql",
}

func TestAllowlistIsExact(t *testing.T) {
	assert.Equal(t, []string{"pages", "pages_content", "posts", "posts_content", "announcements"}, tableNames(serviceSpecs["content"].tables))
	assert.Equal(t, []string{"profiles", "account_deletion_requests"}, tableNames(serviceSpecs["profile"].tables))
	assert.Equal(t, []string{"moderation_audit_log"}, tableNames(serviceSpecs["authz"].tables))
}

func TestParseConfigRejectsAuthzReset(t *testing.T) {
	_, err := parseConfig([]string{"--service", "authz", "--source-dsn", "postgres://source", "--target-dsn", "postgres://target", "--reset-target"})
	require.Error(t, err)
	assert.Equal(t, "invalid_arguments", errorCode(err))
}

func TestRedactError(t *testing.T) {
	dsn := "postgres://operator:top-secret@database.example/tadoku"
	message := redactError(fmt.Errorf("connect %s using top-secret", dsn), dsn)
	assert.NotContains(t, message, dsn)
	assert.NotContains(t, message, "top-secret")
}

func TestTwoDeterministicRehearsalsPerService(t *testing.T) {
	adminURL := testPostgresURL(t)
	for _, service := range []string{"content", "profile", "authz"} {
		t.Run(service, func(t *testing.T) {
			var first result
			for rehearsal := 1; rehearsal <= 2; rehearsal++ {
				sourceDB, sourceURL := openMigratedDatabase(t, adminURL, migrationRunfiles[service])
				targetDB, targetURL := openMigratedDatabase(t, adminURL, migrationRunfiles["target"])
				seedSource(t, sourceDB, service)
				if service == "authz" {
					seedTargetAudit(t, targetDB)
				}

				res, err := execute(context.Background(), rehearsalConfig(service, sourceURL, targetURL))
				require.NoError(t, err)
				assert.Equal(t, "imported", res.Status)
				assertValidationMatches(t, res)
				assertSyntheticWriteRollsBack(t, targetDB, service)

				retry, err := execute(context.Background(), rehearsalConfig(service, sourceURL, targetURL))
				require.NoError(t, err)
				assertValidationMatches(t, retry)
				if service == "authz" {
					assert.Zero(t, retry.Audit.Inserted)
				} else {
					assert.Equal(t, "already_current", retry.Status)
				}

				if rehearsal == 1 {
					first = res
				} else {
					assert.Equal(t, first.Tables, res.Tables)
					assert.Equal(t, first.DomainChecks, res.DomainChecks)
					assert.Equal(t, first.Distributions, res.Distributions)
					assert.Equal(t, first.Audit, res.Audit)
				}
			}
		})
	}
}

func TestInterruptedOwnedImportRollsBackAndRetries(t *testing.T) {
	adminURL := testPostgresURL(t)
	sourceDB, sourceURL := openMigratedDatabase(t, adminURL, migrationRunfiles["content"])
	targetDB, targetURL := openMigratedDatabase(t, adminURL, migrationRunfiles["target"])
	seedSource(t, sourceDB, "content")

	cfg := rehearsalConfig("content", sourceURL, targetURL)
	cfg.afterCopy = func(string) error { return errors.New("simulated interruption") }
	_, err := execute(context.Background(), cfg)
	require.Error(t, err)
	assert.Equal(t, "interrupted", errorCode(err))
	assertOwnedTablesEmpty(t, targetDB, serviceSpecs["content"].tables)

	cfg.afterCopy = nil
	res, err := execute(context.Background(), cfg)
	require.NoError(t, err)
	assertValidationMatches(t, res)
}

func TestSourceValidationFailureAbortsBeforeTargetWrites(t *testing.T) {
	adminURL := testPostgresURL(t)
	sourceDB, sourceURL := openMigratedDatabase(t, adminURL, migrationRunfiles["content"])
	targetDB, targetURL := openMigratedDatabase(t, adminURL, migrationRunfiles["target"])
	seedSource(t, sourceDB, "content")
	_, err := sourceDB.Exec(`update pages set current_content_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'`)
	require.NoError(t, err)

	_, err = execute(context.Background(), rehearsalConfig("content", sourceURL, targetURL))
	require.Error(t, err)
	assert.Equal(t, "validation_error", errorCode(err))
	assertOwnedTablesEmpty(t, targetDB, serviceSpecs["content"].tables)
}

func TestMismatchedOwnedTargetRequiresExplicitReset(t *testing.T) {
	adminURL := testPostgresURL(t)
	sourceDB, sourceURL := openMigratedDatabase(t, adminURL, migrationRunfiles["content"])
	targetDB, targetURL := openMigratedDatabase(t, adminURL, migrationRunfiles["target"])
	seedSource(t, sourceDB, "content")
	_, err := targetDB.Exec(`insert into announcements (id, namespace, title, content, starts_at, ends_at) values ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'test', 'stale', 'stale', '2026-01-01', '2026-01-02')`)
	require.NoError(t, err)

	cfg := rehearsalConfig("content", sourceURL, targetURL)
	_, err = execute(context.Background(), cfg)
	require.Error(t, err)
	assert.Equal(t, "target_not_empty", errorCode(err))

	cfg.resetTarget = true
	res, err := execute(context.Background(), cfg)
	require.NoError(t, err)
	assert.True(t, res.ResetTarget)
	assertValidationMatches(t, res)
}

func TestAuthzConflictAbortsWithoutChangingTarget(t *testing.T) {
	adminURL := testPostgresURL(t)
	sourceDB, sourceURL := openMigratedDatabase(t, adminURL, migrationRunfiles["authz"])
	targetDB, targetURL := openMigratedDatabase(t, adminURL, migrationRunfiles["target"])
	seedSource(t, sourceDB, "authz")
	_, err := targetDB.Exec(`insert into moderation_audit_log (id, user_id, action, metadata, description, created_at) values ('30000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000011', 'ban_user', '{"reason":"different"}', 'conflict', '2026-01-03 00:00:00')`)
	require.NoError(t, err)

	_, err = execute(context.Background(), rehearsalConfig("authz", sourceURL, targetURL))
	require.Error(t, err)
	assert.Equal(t, "audit_conflict", errorCode(err))
	var count int
	require.NoError(t, targetDB.QueryRow(`select count(*) from moderation_audit_log`).Scan(&count))
	assert.Equal(t, 1, count)
}

func rehearsalConfig(service, sourceURL, targetURL string) config {
	return config{
		service:          service,
		sourceDSN:        sourceURL,
		targetDSN:        targetURL,
		sourceSnapshot:   "test-fixture",
		timeout:          time.Minute,
		statementTimeout: 30 * time.Second,
		lockTimeout:      5 * time.Second,
	}
}

func assertValidationMatches(t *testing.T, res result) {
	t.Helper()
	for _, table := range res.Tables {
		assert.Equal(t, table.SourceCount, table.TargetCount, table.Table)
		assert.Equal(t, table.SourceChecksum, table.TargetChecksum, table.Table)
		assert.Equal(t, table.SourceKeyChecksum, table.TargetKeyChecksum, table.Table)
		assert.Equal(t, table.SourceNullCounts, table.TargetNullCounts, table.Table)
	}
	for name, count := range res.DomainChecks {
		assert.Zero(t, count, name)
	}
	if res.Service == "authz" {
		require.NotNil(t, res.Audit)
		assert.Zero(t, res.Audit.Conflicts)
	}
}

func assertOwnedTablesEmpty(t *testing.T, db *sql.DB, tables []tableSpec) {
	t.Helper()
	for _, table := range tables {
		var count int
		require.NoError(t, db.QueryRow("select count(*) from "+pq.QuoteIdentifier(table.name)).Scan(&count))
		assert.Zero(t, count, table.name)
	}
}

func assertSyntheticWriteRollsBack(t *testing.T, db *sql.DB, service string) {
	t.Helper()
	tx, err := db.Begin()
	require.NoError(t, err)
	switch service {
	case "content":
		_, err = tx.Exec(`insert into announcements (namespace, title, content, starts_at, ends_at) values ('synthetic', 'synthetic', 'synthetic', '2026-02-01', '2026-02-02')`)
	case "profile":
		_, err = tx.Exec(`insert into profiles (user_id) values ('20000000-0000-0000-0000-000000000099')`)
	case "authz":
		_, err = tx.Exec(`insert into moderation_audit_log (user_id, action) values ('30000000-0000-0000-0000-000000000099', 'ban_user')`)
	}
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
}

func seedSource(t *testing.T, db *sql.DB, service string) {
	t.Helper()
	var statements []string
	switch service {
	case "content":
		statements = []string{
			`insert into pages (id, namespace, slug, current_content_id, published_at, created_at, updated_at) values ('10000000-0000-0000-0000-000000000001', 'main', 'about', '10000000-0000-0000-0000-000000000011', '2026-01-01', '2026-01-01', '2026-01-02')`,
			`insert into pages_content (id, page_id, title, html, created_at) values ('10000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000001', 'About v1', '<p>v1</p>', '2025-12-31'), ('10000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000001', 'About', '<p>current</p>', '2026-01-01')`,
			`insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at) values ('10000000-0000-0000-0000-000000000002', 'blog', 'launch', '10000000-0000-0000-0000-000000000021', '2026-01-02', '2026-01-01', '2026-01-02')`,
			`insert into posts_content (id, post_id, title, content, created_at) values ('10000000-0000-0000-0000-000000000021', '10000000-0000-0000-0000-000000000002', 'Launch', 'Body', '2026-01-02')`,
			`insert into announcements (id, namespace, title, content, style, starts_at, ends_at, created_at, updated_at) values ('10000000-0000-0000-0000-000000000003', 'main', 'Notice', 'Hello', 'info', '2026-01-01', '2026-02-01', '2026-01-01', '2026-01-01')`,
		}
	case "profile":
		statements = []string{
			`insert into profiles (user_id, created_at, updated_at) values ('20000000-0000-0000-0000-000000000001', '2026-01-01', '2026-01-02'), ('20000000-0000-0000-0000-000000000002', '2026-01-02', '2026-01-03')`,
			`insert into account_deletion_requests (id, identity_id, status, accepted_at, created_at, updated_at) values ('20000000-0000-0000-0000-000000000011', '20000000-0000-0000-0000-000000000001', 'receipt_pending', '2026-01-05', '2026-01-05', '2026-01-05')`,
			`insert into account_deletion_requests (id, identity_id, status, accepted_at, discord_channel_id, discord_message_id, queued_at, attempt_count, next_attempt_at, lease_owner, lease_expires_at, lease_generation, created_at, updated_at) values ('20000000-0000-0000-0000-000000000012', '20000000-0000-0000-0000-000000000002', 'queued', '2026-01-06', '123', '456', '2026-01-06', 2, '2026-01-07', '20000000-0000-0000-0000-000000000022', '2026-01-07', 3, '2026-01-06', '2026-01-07')`,
		}
	case "authz":
		statements = []string{
			`insert into moderation_audit_log (id, user_id, action, metadata, description, created_at) values ('30000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000011', 'ban_user', '{"reason":"spam"}', 'ban', '2026-01-03 00:00:00'), ('30000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000012', 'unban_user', '{}', null, '2026-01-04 00:00:00')`,
		}
	}
	for _, statement := range statements {
		_, err := db.Exec(statement)
		require.NoError(t, err)
	}
}

func seedTargetAudit(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`insert into moderation_audit_log (id, user_id, action, metadata, description, created_at) values ('40000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000011', 'feature_role_grant', '{}', 'immersion row', '2025-12-01')`)
	require.NoError(t, err)
}

func openMigratedDatabase(t *testing.T, adminURL, firstMigrationRunfile string) (*sql.DB, string) {
	t.Helper()
	parsedAdminURL, err := url.Parse(adminURL)
	require.NoError(t, err)
	require.Equal(t, "postgres", parsedAdminURL.Scheme)

	adminDB, err := sql.Open("postgres", adminURL)
	require.NoError(t, err)
	require.NoError(t, adminDB.Ping())
	databaseName := "database_convergence_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = adminDB.Exec(`create database ` + pq.QuoteIdentifier(databaseName))
	require.NoError(t, err)

	testURL := *parsedAdminURL
	testURL.Path = "/" + databaseName
	testDB, err := sql.Open("postgres", testURL.String())
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
	migrationURL := (&url.URL{Scheme: "file", Path: filepath.Dir(firstMigration)}).String()
	migrator, err := migrate.New(migrationURL, testURL.String())
	require.NoError(t, err)
	require.NoError(t, migrator.Up())
	sourceCloseErr, databaseCloseErr := migrator.Close()
	require.NoError(t, sourceCloseErr)
	require.NoError(t, databaseCloseErr)
	return testDB, testURL.String()
}

func testPostgresURL(t *testing.T) string {
	t.Helper()
	value := os.Getenv(testPostgresURLVariable)
	if value == "" {
		t.Skipf("%s is not set; skipping PostgreSQL rehearsal", testPostgresURLVariable)
	}
	return value
}

func tableNames(tables []tableSpec) []string {
	names := make([]string, len(tables))
	for index, table := range tables {
		names[index] = table.name
	}
	return names
}
