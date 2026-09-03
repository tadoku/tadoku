package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

const (
	resultSchemaVersion = 1
	toolVersion         = "db-convergence-v1"
	targetLedgerVersion = 29
)

type tableSpec struct {
	name       string
	key        string
	columns    []string
	nullable   []string
	deleteRank int
}

type serviceSpec struct {
	sourceLedgerVersion int
	tables              []tableSpec
}

var serviceSpecs = map[string]serviceSpec{
	"content": {
		sourceLedgerVersion: 3,
		tables: []tableSpec{
			{name: "pages", key: "id", columns: []string{"id", "namespace", "slug", "current_content_id", "published_at", "created_at", "updated_at", "deleted_at"}, nullable: []string{"published_at", "deleted_at"}, deleteRank: 5},
			{name: "pages_content", key: "id", columns: []string{"id", "page_id", "title", "html", "created_at"}, deleteRank: 4},
			{name: "posts", key: "id", columns: []string{"id", "namespace", "slug", "current_content_id", "published_at", "created_at", "updated_at", "deleted_at"}, nullable: []string{"published_at", "deleted_at"}, deleteRank: 3},
			{name: "posts_content", key: "id", columns: []string{"id", "post_id", "title", "content", "created_at"}, deleteRank: 2},
			{name: "announcements", key: "id", columns: []string{"id", "namespace", "title", "content", "style", "href", "starts_at", "ends_at", "created_at", "updated_at", "deleted_at"}, nullable: []string{"href", "deleted_at"}, deleteRank: 1},
		},
	},
	"profile": {
		sourceLedgerVersion: 2,
		tables: []tableSpec{
			{name: "profiles", key: "user_id", columns: []string{"user_id", "created_at", "updated_at"}, deleteRank: 2},
			{name: "account_deletion_requests", key: "id", columns: []string{
				"id", "identity_id", "status", "resume_status", "accepted_at", "discord_channel_id", "discord_message_id",
				"queued_at", "access_locked_at", "immersion_scrubbed_at", "caches_reconciled_at", "authorization_removed_at",
				"identity_deleted_at", "completed_at", "attempt_count", "next_attempt_at", "last_error_code", "lease_owner",
				"lease_expires_at", "lease_generation", "manual_attention_at", "remediation_due_at", "created_at", "updated_at",
			}, nullable: []string{
				"resume_status", "discord_channel_id", "discord_message_id", "queued_at", "access_locked_at", "immersion_scrubbed_at",
				"caches_reconciled_at", "authorization_removed_at", "identity_deleted_at", "completed_at", "next_attempt_at",
				"last_error_code", "lease_owner", "lease_expires_at", "manual_attention_at", "remediation_due_at",
			}, deleteRank: 1},
		},
	},
	"authz": {
		sourceLedgerVersion: 2,
		tables: []tableSpec{
			{name: "moderation_audit_log", key: "id", columns: []string{"id", "user_id", "action", "metadata", "description", "created_at"}, nullable: []string{"description"}},
		},
	},
}

type config struct {
	service          string
	sourceDSN        string
	targetDSN        string
	sourceSnapshot   string
	timeout          time.Duration
	statementTimeout time.Duration
	lockTimeout      time.Duration
	resetTarget      bool
	afterCopy        func(string) error
}

type ledgerResult struct {
	Version int  `json:"version"`
	Dirty   bool `json:"dirty"`
}

type endpointResult struct {
	Database string       `json:"database"`
	Schema   string       `json:"schema"`
	Ledger   ledgerResult `json:"ledger"`
}

type tableResult struct {
	Table             string         `json:"table"`
	SourceCount       int64          `json:"source_count"`
	TargetCount       int64          `json:"target_count"`
	SourceChecksum    string         `json:"source_checksum"`
	TargetChecksum    string         `json:"target_checksum"`
	SourceKeyChecksum string         `json:"source_key_checksum"`
	TargetKeyChecksum string         `json:"target_key_checksum"`
	SourceNullCounts  map[string]int `json:"source_null_counts,omitempty"`
	TargetNullCounts  map[string]int `json:"target_null_counts,omitempty"`
}

type auditResult struct {
	TargetCountBefore int64            `json:"target_count_before"`
	TargetCountAfter  int64            `json:"target_count_after"`
	Inserted          int64            `json:"inserted"`
	Conflicts         int64            `json:"conflicts"`
	SourceActions     map[string]int64 `json:"source_actions"`
}

type resultError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type result struct {
	SchemaVersion  int              `json:"schema_version"`
	ToolVersion    string           `json:"tool_version"`
	Service        string           `json:"service"`
	SourceSnapshot string           `json:"source_snapshot,omitempty"`
	Status         string           `json:"status"`
	Source         endpointResult   `json:"source"`
	Target         endpointResult   `json:"target"`
	ResetTarget    bool             `json:"reset_target,omitempty"`
	Tables         []tableResult    `json:"tables,omitempty"`
	DomainChecks   map[string]int64 `json:"domain_checks,omitempty"`
	Distributions  map[string]int64 `json:"distributions,omitempty"`
	Audit          *auditResult     `json:"audit,omitempty"`
	Error          *resultError     `json:"error,omitempty"`
}

type tableSummary struct {
	count       int64
	checksum    string
	keyChecksum string
	nullCounts  map[string]int
}

type codedError struct {
	code string
	err  error
}

func (e *codedError) Error() string { return e.err.Error() }
func (e *codedError) Unwrap() error { return e.err }

func failure(code, format string, args ...any) error {
	return &codedError{code: code, err: fmt.Errorf(format, args...)}
}

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout))
}

func runCLI(args []string, output io.Writer) int {
	cfg, err := parseConfig(args)
	if err != nil {
		writeResult(output, failedResult(config{}, err))
		return 2
	}

	res, err := execute(context.Background(), cfg)
	if err != nil {
		res.Status = "failed"
		res.Error = &resultError{Code: errorCode(err), Message: redactError(err, cfg.sourceDSN, cfg.targetDSN)}
		writeResult(output, res)
		return 1
	}

	writeResult(output, res)
	return 0
}

func parseConfig(args []string) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("database-convergence", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.service, "service", "", "content, profile, or authz")
	flags.StringVar(&cfg.sourceDSN, "source-dsn", os.Getenv("DATABASE_CONVERGENCE_SOURCE_DSN"), "source PostgreSQL DSN")
	flags.StringVar(&cfg.targetDSN, "target-dsn", os.Getenv("DATABASE_CONVERGENCE_TARGET_DSN"), "target PostgreSQL DSN")
	flags.StringVar(&cfg.sourceSnapshot, "source-snapshot", "", "backup or snapshot identifier")
	flags.DurationVar(&cfg.timeout, "timeout", 10*time.Minute, "overall operation timeout")
	flags.DurationVar(&cfg.statementTimeout, "statement-timeout", 2*time.Minute, "PostgreSQL statement timeout")
	flags.DurationVar(&cfg.lockTimeout, "lock-timeout", 5*time.Second, "PostgreSQL lock timeout")
	flags.BoolVar(&cfg.resetTarget, "reset-target", false, "clear only the service allowlist before import")
	if err := flags.Parse(args); err != nil {
		return config{}, failure("invalid_arguments", "invalid arguments: %v", err)
	}
	if flags.NArg() != 0 {
		return config{}, failure("invalid_arguments", "unexpected positional arguments")
	}
	if _, ok := serviceSpecs[cfg.service]; !ok {
		return config{}, failure("invalid_arguments", "service must be content, profile, or authz")
	}
	if cfg.sourceDSN == "" || cfg.targetDSN == "" {
		return config{}, failure("invalid_arguments", "source and target DSNs are required; prefer DATABASE_CONVERGENCE_SOURCE_DSN and DATABASE_CONVERGENCE_TARGET_DSN")
	}
	if cfg.timeout <= 0 || cfg.statementTimeout <= 0 || cfg.lockTimeout <= 0 {
		return config{}, failure("invalid_arguments", "timeouts must be positive")
	}
	if cfg.service == "authz" && cfg.resetTarget {
		return config{}, failure("invalid_arguments", "authz never clears the shared target audit table")
	}
	return cfg, nil
}

func execute(parent context.Context, cfg config) (res result, err error) {
	res = result{
		SchemaVersion:  resultSchemaVersion,
		ToolVersion:    toolVersion,
		Service:        cfg.service,
		SourceSnapshot: cfg.sourceSnapshot,
		Status:         "running",
		ResetTarget:    cfg.resetTarget,
	}

	spec, ok := serviceSpecs[cfg.service]
	if !ok {
		return res, failure("invalid_arguments", "unknown service %q", cfg.service)
	}
	ctx, cancel := context.WithTimeout(parent, cfg.timeout)
	defer cancel()

	sourceDB, err := openDatabase(ctx, cfg.sourceDSN)
	if err != nil {
		return res, failure("database_error", "open source database: %w", err)
	}
	defer sourceDB.Close()
	targetDB, err := openDatabase(ctx, cfg.targetDSN)
	if err != nil {
		return res, failure("database_error", "open target database: %w", err)
	}
	defer targetDB.Close()

	sourceTx, err := sourceDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return res, failure("database_error", "begin source snapshot: %w", err)
	}
	defer sourceTx.Rollback()
	if err := configureTransaction(ctx, sourceTx, cfg); err != nil {
		return res, failure("database_error", "configure source snapshot: %w", err)
	}
	res.Source, err = inspectEndpoint(ctx, sourceTx, spec.sourceLedgerVersion)
	if err != nil {
		return res, err
	}

	targetTx, err := targetDB.BeginTx(ctx, nil)
	if err != nil {
		return res, failure("database_error", "begin target transaction: %w", err)
	}
	defer targetTx.Rollback()
	if err := configureTransaction(ctx, targetTx, cfg); err != nil {
		return res, failure("database_error", "configure target transaction: %w", err)
	}
	res.Target, err = inspectEndpoint(ctx, targetTx, targetLedgerVersion)
	if err != nil {
		return res, err
	}
	if res.Source.Database == res.Target.Database {
		return res, failure("unsafe_target", "source and target resolve to the same database")
	}

	if cfg.service == "authz" {
		err = importAuthz(ctx, sourceTx, targetTx, spec, &res)
	} else {
		err = importOwnedTables(ctx, sourceTx, targetTx, spec, cfg, &res)
	}
	if err != nil {
		return res, err
	}
	if err := targetTx.Commit(); err != nil {
		return res, failure("database_error", "commit target transaction: %w", err)
	}
	if res.Status == "running" {
		res.Status = "imported"
	}
	return res, nil
}

func openDatabase(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func configureTransaction(ctx context.Context, tx *sql.Tx, cfg config) error {
	_, err := tx.ExecContext(ctx, `select
		set_config('statement_timeout', $1, true),
		set_config('lock_timeout', $2, true),
		set_config('TimeZone', 'UTC', true),
		set_config('DateStyle', 'ISO, YMD', true)`,
		strconv.FormatInt(cfg.statementTimeout.Milliseconds(), 10),
		strconv.FormatInt(cfg.lockTimeout.Milliseconds(), 10),
	)
	return err
}

func inspectEndpoint(ctx context.Context, tx *sql.Tx, expectedVersion int) (endpointResult, error) {
	var endpoint endpointResult
	if err := tx.QueryRowContext(ctx, `select current_database(), current_schema()`).Scan(&endpoint.Database, &endpoint.Schema); err != nil {
		return endpoint, failure("database_error", "inspect database identity: %w", err)
	}
	if endpoint.Schema == "" {
		return endpoint, failure("schema_error", "connection has no current schema")
	}
	if err := tx.QueryRowContext(ctx, `select version, dirty from schema_migrations`).Scan(&endpoint.Ledger.Version, &endpoint.Ledger.Dirty); err != nil {
		return endpoint, failure("ledger_error", "read migration ledger: %w", err)
	}
	if endpoint.Ledger.Dirty || endpoint.Ledger.Version != expectedVersion {
		return endpoint, failure("ledger_error", "expected clean migration version %d, got version %d dirty=%t", expectedVersion, endpoint.Ledger.Version, endpoint.Ledger.Dirty)
	}
	return endpoint, nil
}

func importOwnedTables(ctx context.Context, source, target *sql.Tx, spec serviceSpec, cfg config, res *result) error {
	for _, table := range spec.tables {
		if _, err := target.ExecContext(ctx, "lock table "+qualified(res.Target.Schema, table.name)+" in exclusive mode"); err != nil {
			return failure("database_error", "lock target table %s: %v", table.name, err)
		}
	}

	sourceChecks, err := domainChecks(ctx, source, res.Source.Schema, cfg.service)
	if err != nil {
		return err
	}
	if err := requireZeroChecks("source", sourceChecks); err != nil {
		return err
	}
	sourceSummaries, err := summarizeTables(ctx, source, res.Source.Schema, spec.tables)
	if err != nil {
		return err
	}
	targetSummaries, err := summarizeTables(ctx, target, res.Target.Schema, spec.tables)
	if err != nil {
		return err
	}

	if summariesEqual(sourceSummaries, targetSummaries, spec.tables) {
		res.Tables = buildTableResults(sourceSummaries, targetSummaries, spec.tables)
		res.DomainChecks = sourceChecks
		if cfg.service == "profile" {
			res.Distributions, err = distribution(ctx, source, res.Source.Schema, "account_deletion_requests", "status")
			if err != nil {
				return err
			}
		}
		res.Status = "already_current"
		return nil
	}

	if !allEmpty(targetSummaries) {
		if !cfg.resetTarget {
			return failure("target_not_empty", "target allowlist is non-empty and does not exactly match the source; use --reset-target only before target writes are authoritative")
		}
		if err := clearOwnedTables(ctx, target, res.Target.Schema, spec.tables); err != nil {
			return err
		}
	}

	for _, table := range spec.tables {
		if err := copyTable(ctx, source, res.Source.Schema, target, res.Target.Schema, table, table.name); err != nil {
			return err
		}
		if cfg.afterCopy != nil {
			if err := cfg.afterCopy(table.name); err != nil {
				return failure("interrupted", "after copying %s: %v", table.name, err)
			}
		}
	}

	targetSummaries, err = summarizeTables(ctx, target, res.Target.Schema, spec.tables)
	if err != nil {
		return err
	}
	if !summariesEqual(sourceSummaries, targetSummaries, spec.tables) {
		return failure("validation_error", "target table validation differs from the source snapshot")
	}
	targetChecks, err := domainChecks(ctx, target, res.Target.Schema, cfg.service)
	if err != nil {
		return err
	}
	if err := requireZeroChecks("target", targetChecks); err != nil {
		return err
	}
	if !mapsEqual(sourceChecks, targetChecks) {
		return failure("validation_error", "target domain validation differs from the source snapshot")
	}

	res.Tables = buildTableResults(sourceSummaries, targetSummaries, spec.tables)
	res.DomainChecks = targetChecks
	if cfg.service == "profile" {
		res.Distributions, err = distribution(ctx, target, res.Target.Schema, "account_deletion_requests", "status")
		if err != nil {
			return err
		}
	}
	return nil
}

func importAuthz(ctx context.Context, source, target *sql.Tx, spec serviceSpec, res *result) error {
	table := spec.tables[0]
	sourceSummary, err := summarizeTable(ctx, source, res.Source.Schema, table, "")
	if err != nil {
		return err
	}
	sourceActions, err := distribution(ctx, source, res.Source.Schema, table.name, "action")
	if err != nil {
		return err
	}
	var countBefore int64
	if err := target.QueryRowContext(ctx, "select count(*) from "+qualified(res.Target.Schema, table.name)).Scan(&countBefore); err != nil {
		return failure("database_error", "count target audit rows: %v", err)
	}

	const stage = "moderation_audit_log_stage"
	if _, err := target.ExecContext(ctx, "create temporary table "+pq.QuoteIdentifier(stage)+" as select "+columnList(table.columns)+" from "+qualified(res.Target.Schema, table.name)+" with no data"); err != nil {
		return failure("database_error", "create audit staging table: %v", err)
	}
	if err := copyTable(ctx, source, res.Source.Schema, target, "", table, stage); err != nil {
		return err
	}

	conflictQuery := fmt.Sprintf(`select count(*) from %s as source
		join %s as target using (id)
		where row(target.user_id, target.action, target.metadata, target.description, target.created_at)
			is distinct from row(source.user_id, source.action, source.metadata, source.description, source.created_at)`,
		pq.QuoteIdentifier(stage), qualified(res.Target.Schema, table.name))
	var conflicts int64
	if err := target.QueryRowContext(ctx, conflictQuery).Scan(&conflicts); err != nil {
		return failure("database_error", "compare staged audit rows: %v", err)
	}
	if conflicts != 0 {
		return failure("audit_conflict", "%d source audit UUIDs match target rows with different content", conflicts)
	}

	insertQuery := fmt.Sprintf(`insert into %s (%s)
		select %s from %s
		on conflict (id) do nothing`,
		qualified(res.Target.Schema, table.name), columnList(table.columns), columnList(table.columns), pq.QuoteIdentifier(stage))
	insertedResult, err := target.ExecContext(ctx, insertQuery)
	if err != nil {
		return failure("database_error", "merge staged audit rows: %v", err)
	}
	inserted, err := insertedResult.RowsAffected()
	if err != nil {
		return failure("database_error", "read inserted audit count: %v", err)
	}

	stageSummary, err := summarizeTable(ctx, target, "", tableSpec{name: stage, key: table.key, columns: table.columns, nullable: table.nullable}, "")
	if err != nil {
		return err
	}
	targetScopedSummary, err := summarizeTable(ctx, target, res.Target.Schema, table,
		"where "+pq.QuoteIdentifier(table.key)+" in (select "+pq.QuoteIdentifier(table.key)+" from "+pq.QuoteIdentifier(stage)+")")
	if err != nil {
		return err
	}
	if !summaryEqual(sourceSummary, stageSummary) || !summaryEqual(sourceSummary, targetScopedSummary) {
		return failure("validation_error", "merged target audit rows differ from the source snapshot")
	}
	var countAfter int64
	if err := target.QueryRowContext(ctx, "select count(*) from "+qualified(res.Target.Schema, table.name)).Scan(&countAfter); err != nil {
		return failure("database_error", "count merged target audit rows: %v", err)
	}
	if countAfter != countBefore+inserted {
		return failure("validation_error", "target audit count changed unexpectedly: before=%d inserted=%d after=%d", countBefore, inserted, countAfter)
	}

	res.Tables = buildTableResults(map[string]tableSummary{table.name: sourceSummary}, map[string]tableSummary{table.name: targetScopedSummary}, spec.tables)
	res.Audit = &auditResult{TargetCountBefore: countBefore, TargetCountAfter: countAfter, Inserted: inserted, Conflicts: conflicts, SourceActions: sourceActions}
	return nil
}

func copyTable(ctx context.Context, source *sql.Tx, sourceSchema string, target *sql.Tx, targetSchema string, table tableSpec, targetTable string) error {
	query := "select " + castColumnList(table.columns) + " from " + qualified(sourceSchema, table.name) + " order by " + pq.QuoteIdentifier(table.key)
	rows, err := source.QueryContext(ctx, query)
	if err != nil {
		return failure("database_error", "read source table %s: %v", table.name, err)
	}
	defer rows.Close()

	copyStatement := pq.CopyIn(targetTable, table.columns...)
	if targetSchema != "" {
		copyStatement = pq.CopyInSchema(targetSchema, targetTable, table.columns...)
	}
	statement, err := target.PrepareContext(ctx, copyStatement)
	if err != nil {
		return failure("database_error", "prepare target copy for %s: %v", table.name, err)
	}
	defer statement.Close()

	for rows.Next() {
		values, destinations := scanBuffers(len(table.columns))
		if err := rows.Scan(destinations...); err != nil {
			return failure("database_error", "scan source table %s: %v", table.name, err)
		}
		if _, err := statement.ExecContext(ctx, sqlValues(values)...); err != nil {
			return failure("database_error", "copy source table %s: %v", table.name, err)
		}
	}
	if err := rows.Err(); err != nil {
		return failure("database_error", "stream source table %s: %v", table.name, err)
	}
	if _, err := statement.ExecContext(ctx); err != nil {
		return failure("database_error", "finish target copy for %s: %v", table.name, err)
	}
	if err := statement.Close(); err != nil {
		return failure("database_error", "close target copy for %s: %v", table.name, err)
	}
	return nil
}

func summarizeTables(ctx context.Context, tx *sql.Tx, schema string, tables []tableSpec) (map[string]tableSummary, error) {
	summaries := make(map[string]tableSummary, len(tables))
	for _, table := range tables {
		summary, err := summarizeTable(ctx, tx, schema, table, "")
		if err != nil {
			return nil, err
		}
		summaries[table.name] = summary
	}
	return summaries, nil
}

func summarizeTable(ctx context.Context, tx *sql.Tx, schema string, table tableSpec, filter string) (tableSummary, error) {
	query := "select " + castColumnList(table.columns) + " from " + qualified(schema, table.name)
	if filter != "" {
		query += " " + filter
	}
	query += " order by " + pq.QuoteIdentifier(table.key)
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return tableSummary{}, failure("database_error", "summarize table %s: %v", table.name, err)
	}
	defer rows.Close()

	rowHash := sha256.New()
	keyHash := sha256.New()
	nulls := make(map[string]int, len(table.nullable))
	nullable := make(map[string]bool, len(table.nullable))
	for _, column := range table.nullable {
		nullable[column] = true
		nulls[column] = 0
	}
	keyIndex := indexOf(table.columns, table.key)
	var count int64
	for rows.Next() {
		values, destinations := scanBuffers(len(table.columns))
		if err := rows.Scan(destinations...); err != nil {
			return tableSummary{}, failure("database_error", "scan table %s for validation: %v", table.name, err)
		}
		for index, value := range values {
			writeHashValue(rowHash, value)
			if nullable[table.columns[index]] && !value.Valid {
				nulls[table.columns[index]]++
			}
		}
		writeHashValue(keyHash, values[keyIndex])
		count++
	}
	if err := rows.Err(); err != nil {
		return tableSummary{}, failure("database_error", "validate table %s: %v", table.name, err)
	}
	return tableSummary{
		count:       count,
		checksum:    "sha256:" + hex.EncodeToString(rowHash.Sum(nil)),
		keyChecksum: "sha256:" + hex.EncodeToString(keyHash.Sum(nil)),
		nullCounts:  nulls,
	}, nil
}

func writeHashValue(writer io.Writer, value sql.NullString) {
	if !value.Valid {
		_, _ = writer.Write([]byte{0})
		return
	}
	_, _ = writer.Write([]byte{1})
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value.String)))
	_, _ = writer.Write(length[:])
	_, _ = io.WriteString(writer, value.String)
}

func domainChecks(ctx context.Context, tx *sql.Tx, schema, service string) (map[string]int64, error) {
	queries := map[string]string{}
	switch service {
	case "content":
		pages := qualified(schema, "pages")
		pageContent := qualified(schema, "pages_content")
		posts := qualified(schema, "posts")
		postContent := qualified(schema, "posts_content")
		queries = map[string]string{
			"page_current_pointer_mismatches": `select count(*) from ` + pages + ` p left join ` + pageContent + ` c on c.id = p.current_content_id and c.page_id = p.id where c.id is null`,
			"page_version_orphans":            `select count(*) from ` + pageContent + ` c left join ` + pages + ` p on p.id = c.page_id where p.id is null`,
			"post_current_pointer_mismatches": `select count(*) from ` + posts + ` p left join ` + postContent + ` c on c.id = p.current_content_id and c.post_id = p.id where c.id is null`,
			"post_version_orphans":            `select count(*) from ` + postContent + ` c left join ` + posts + ` p on p.id = c.post_id where p.id is null`,
		}
	case "profile":
		requests := qualified(schema, "account_deletion_requests")
		queries = map[string]string{
			"attempt_count_violations":    `select count(*) from ` + requests + ` where attempt_count < 0`,
			"lease_generation_violations": `select count(*) from ` + requests + ` where lease_generation < 0`,
			"lease_pair_violations":       `select count(*) from ` + requests + ` where (lease_owner is null) <> (lease_expires_at is null)`,
			"manual_attention_violations": `select count(*) from ` + requests + ` where (status = 'manual_attention' and (manual_attention_at is null or remediation_due_at is distinct from manual_attention_at + interval '7 days')) or (status <> 'manual_attention' and (manual_attention_at is not null or remediation_due_at is not null))`,
			"receipt_id_pair_violations":  `select count(*) from ` + requests + ` where (discord_channel_id is null) <> (discord_message_id is null)`,
			"receipt_state_violations":    `select count(*) from ` + requests + ` where status <> 'receipt_pending' and (discord_channel_id is null or discord_message_id is null)`,
			"resume_status_violations":    `select count(*) from ` + requests + ` where (status = 'manual_attention' and (resume_status is null or resume_status not in ('queued','access_locked','immersion_scrubbed','caches_reconciled','authorization_removed','identity_deleted'))) or (status <> 'manual_attention' and resume_status is not null)`,
		}
	}

	checks := make(map[string]int64, len(queries))
	names := make([]string, 0, len(queries))
	for name := range queries {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		var count int64
		if err := tx.QueryRowContext(ctx, queries[name]).Scan(&count); err != nil {
			return nil, failure("database_error", "run domain check %s: %v", name, err)
		}
		checks[name] = count
	}
	return checks, nil
}

func distribution(ctx context.Context, tx *sql.Tx, schema, table, column string) (map[string]int64, error) {
	query := "select " + pq.QuoteIdentifier(column) + "::text, count(*) from " + qualified(schema, table) + " group by 1 order by 1"
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, failure("database_error", "read %s distribution: %v", column, err)
	}
	defer rows.Close()
	values := map[string]int64{}
	for rows.Next() {
		var value string
		var count int64
		if err := rows.Scan(&value, &count); err != nil {
			return nil, failure("database_error", "scan %s distribution: %v", column, err)
		}
		values[value] = count
	}
	if err := rows.Err(); err != nil {
		return nil, failure("database_error", "read %s distribution: %v", column, err)
	}
	return values, nil
}

func clearOwnedTables(ctx context.Context, tx *sql.Tx, schema string, tables []tableSpec) error {
	ordered := append([]tableSpec(nil), tables...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].deleteRank < ordered[j].deleteRank })
	for _, table := range ordered {
		if _, err := tx.ExecContext(ctx, "delete from "+qualified(schema, table.name)); err != nil {
			return failure("database_error", "clear allowlisted target table %s: %v", table.name, err)
		}
	}
	return nil
}

func requireZeroChecks(side string, checks map[string]int64) error {
	for name, count := range checks {
		if count != 0 {
			return failure("validation_error", "%s domain check %s found %d violations", side, name, count)
		}
	}
	return nil
}

func buildTableResults(source, target map[string]tableSummary, tables []tableSpec) []tableResult {
	results := make([]tableResult, 0, len(tables))
	for _, table := range tables {
		sourceSummary := source[table.name]
		targetSummary := target[table.name]
		results = append(results, tableResult{
			Table:             table.name,
			SourceCount:       sourceSummary.count,
			TargetCount:       targetSummary.count,
			SourceChecksum:    sourceSummary.checksum,
			TargetChecksum:    targetSummary.checksum,
			SourceKeyChecksum: sourceSummary.keyChecksum,
			TargetKeyChecksum: targetSummary.keyChecksum,
			SourceNullCounts:  sourceSummary.nullCounts,
			TargetNullCounts:  targetSummary.nullCounts,
		})
	}
	return results
}

func summariesEqual(source, target map[string]tableSummary, tables []tableSpec) bool {
	for _, table := range tables {
		if !summaryEqual(source[table.name], target[table.name]) {
			return false
		}
	}
	return true
}

func summaryEqual(left, right tableSummary) bool {
	return left.count == right.count &&
		left.checksum == right.checksum &&
		left.keyChecksum == right.keyChecksum &&
		mapsEqual(left.nullCounts, right.nullCounts)
}

func allEmpty(summaries map[string]tableSummary) bool {
	for _, summary := range summaries {
		if summary.count != 0 {
			return false
		}
	}
	return true
}

func mapsEqual[K comparable, V comparable](left, right map[K]V) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func scanBuffers(count int) ([]sql.NullString, []any) {
	values := make([]sql.NullString, count)
	destinations := make([]any, count)
	for index := range values {
		destinations[index] = &values[index]
	}
	return values, destinations
}

func sqlValues(values []sql.NullString) []any {
	result := make([]any, len(values))
	for index, value := range values {
		if value.Valid {
			result[index] = value.String
		}
	}
	return result
}

func qualified(schema, table string) string {
	if schema == "" {
		return pq.QuoteIdentifier(table)
	}
	return pq.QuoteIdentifier(schema) + "." + pq.QuoteIdentifier(table)
}

func columnList(columns []string) string {
	quoted := make([]string, len(columns))
	for index, column := range columns {
		quoted[index] = pq.QuoteIdentifier(column)
	}
	return strings.Join(quoted, ", ")
}

func castColumnList(columns []string) string {
	cast := make([]string, len(columns))
	for index, column := range columns {
		cast[index] = pq.QuoteIdentifier(column) + "::text"
	}
	return strings.Join(cast, ", ")
}

func indexOf(values []string, wanted string) int {
	for index, value := range values {
		if value == wanted {
			return index
		}
	}
	panic("missing key column " + wanted)
}

func failedResult(cfg config, err error) result {
	return result{
		SchemaVersion: resultSchemaVersion,
		ToolVersion:   toolVersion,
		Service:       cfg.service,
		Status:        "failed",
		Error:         &resultError{Code: errorCode(err), Message: redactError(err, cfg.sourceDSN, cfg.targetDSN)},
	}
}

func errorCode(err error) string {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "timeout"
	}
	var coded *codedError
	if errors.As(err, &coded) {
		return coded.code
	}
	return "internal_error"
}

func redactError(err error, secrets ...string) string {
	message := err.Error()
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		message = strings.ReplaceAll(message, secret, "[redacted-dsn]")
		parsed, parseErr := url.Parse(secret)
		if parseErr == nil && parsed.User != nil {
			if password, ok := parsed.User.Password(); ok && password != "" {
				message = strings.ReplaceAll(message, password, "[redacted]")
			}
		}
	}
	return message
}

func writeResult(output io.Writer, value result) {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}
