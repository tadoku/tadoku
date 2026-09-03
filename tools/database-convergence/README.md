# Database convergence importer

`//tools/database-convergence` is the versioned DB-2 importer and validator for moving Content, Profile, and Authz data into the canonical Immersion PostgreSQL database. Version `db-convergence-v1` accepts source ledgers Content 3, Profile 2, or Authz 2 and requires a clean target ledger at Immersion 29.

The command never changes a source database. It reads the source in a repeatable-read, read-only transaction and writes the target in one transaction with an overall timeout, statement timeout, and lock timeout. Output is one JSON document; DSNs and passwords are redacted from errors.

## Build and verify

```sh
bazel build //tools/database-convergence
bazel test //tools/database-convergence:database-convergence_test
```

Set `IMMERSION_TEST_POSTGRES_URL` to an administrative disposable PostgreSQL URL to run the integration rehearsals instead of skipping them. The tests create uniquely named databases and drop them after testing.

## Run

Keep DSNs out of shell history and process arguments:

```sh
export DATABASE_CONVERGENCE_SOURCE_DSN='postgres://...'
export DATABASE_CONVERGENCE_TARGET_DSN='postgres://...'

bazel run //tools/database-convergence -- \
  --service content \
  --source-snapshot 'backup-or-snapshot-id' \
  > content-db-convergence-v1.json
```

Use `--service profile` or `--service authz` for the other datasets. The defaults are a 10-minute operation timeout, 2-minute statement timeout, and 5-second lock timeout; override them with `--timeout`, `--statement-timeout`, and `--lock-timeout` only from rehearsal evidence.

Content and Profile succeed without writing when their target allowlist already exactly matches the source. A non-empty mismatch fails. Before target writes become authoritative, an operator may explicitly use `--reset-target`; it deletes only the selected service's hard-coded allowlist inside the import transaction. Authz rejects this option and never clears the shared audit table.

## Fixed allowlists

| Service | Tables | Behavior |
| --- | --- | --- |
| Content | `pages`, `pages_content`, `posts`, `posts_content`, `announcements` | Copy to empty owned tables in dependency-safe order; validate counts, ordered row/key SHA-256 checksums, null counts, current pointers, and parent/version relationships. |
| Profile | `profiles`, `account_deletion_requests` | Copy to empty owned tables; validate counts, ordered row/key SHA-256 checksums, null counts, status distribution, receipt, lease/generation, retry, and manual-attention invariants. |
| Authz | `moderation_audit_log` | Copy source rows into a temporary target staging table, fail if a matching UUID has different content, and insert only missing exact rows. Existing Immersion rows stay untouched. |

All columns are explicit. The command has no dynamic table-name or arbitrary-SQL option.

## Operator gate

Complete these steps for each service, in Content → Profile → Authz order:

- [ ] Record the protected source backup or snapshot ID and its checksum.
- [ ] Confirm the source and target DSNs point to different databases and use the expected `data` schema.
- [ ] Confirm the source and target migration versions shown above, with `dirty=false`.
- [ ] Confirm the legacy service init migrator is absent while its runtime DSN still selects the source database.
- [ ] Name the pause operator, database operator, deploy operator, and observer; use separate people where the environment requires it.
- [ ] Pause and drain every service writer outside this command. The importer does not manipulate routes, workloads, or workers.
- [ ] Run the command and retain its JSON output beside the source backup record.
- [ ] Require `status` to be `imported` or `already_current`, every source/target count and checksum pair to match, every domain check to be zero, and Authz conflicts to be zero.
- [ ] Run the service's repository/contract checks and rollback-only synthetic write against the target DSN.
- [ ] Observe readiness, error rate, p95 latency, PostgreSQL connections/saturation, locks, statement timeouts, and attempted source connections for the approved window.
- [ ] Switch and resume only after the validation, application, and monitoring owners sign off.

Example machine gate:

```sh
jq -e '
  (.status == "imported" or .status == "already_current") and
  all(.tables[];
    .source_count == .target_count and
    .source_checksum == .target_checksum and
    .source_key_checksum == .target_key_checksum and
    .source_null_counts == .target_null_counts) and
  all((.domain_checks // {})[]; . == 0) and
  ((.audit // {conflicts: 0}).conflicts == 0)
' content-db-convergence-v1.json
```

## Abort boundary

Before target writes resume, any non-zero command exit, failed JSON gate, contract failure, synthetic-write failure, timeout, lock wait, or monitoring regression is an abort. Keep or restore ordinary source routing; the source remains authoritative. Interrupted and failed imports roll back their target transaction. A previously committed Content/Profile rehearsal may be cleared only with the explicit `--reset-target` option while those target tables are confirmed unused.

After target writes resume, do not use `--reset-target` and do not switch back casually. Repair forward, or obtain explicit approval for a separately designed reverse synchronization or restore.

The importer does not pause traffic, alter DSNs, change grants, take backups, run application contracts, or approve an observation window. Those stay explicit operator actions.

The development manifests retain clean-cluster support through separate `*-api-source-schema` Jobs using the frozen Content 3, Profile 2, and Authz 2 migration directories. API Deployments have no migration init container, and the dev Jobs point only to the legacy source database names.
