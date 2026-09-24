---
title: Database and migrations
description: How Tadoku API ships migrations, which database features it uses, SQL style, sqlc code generation, application-owned transactions and how PostgreSQL errors map to responses.
sidebar_position: 4
---

# Database and migrations

Read this when you write a migration, change a SQL query or open a transaction.

## Migrations

Migrations live in `services/tadoku-api/migrations/` as numbered pairs,
`NNNN_<description>.up.sql` and `NNNN_<description>.down.sql`, using the next
free number. sqlc reads the same directory as its schema.

- Ship every schema or data migration as a standalone change. It lands on
  `main` in its own commit and pull request, separate from application code,
  and is deployed independently before any code that depends on it.
- Never combine a migration and runtime behavior in the same commit, pull
  request or deployment, including within a stacked PR series.
- A migration pull request contains no changes outside
  `services/tadoku-api/migrations/`. CI enforces this; the `migration-move`
  label permits mechanical relocations and logs a warning.
- A migration must stay compatible with the application version currently
  deployed. Additive migrations come before the code that depends on them.
  Destructive or cleanup migrations come later, in their own standalone pull
  request, once the old code is no longer deployed.
- After a migration is merged and deployed, base dependent work on the updated
  `main`.

## Schema design

- Do not create application-defined database functions, stored procedures or
  triggers. Keep business behavior in the application and domain layers.
- Use declarative database features for data integrity and performance:
  `not null`, `check`, unique and foreign-key constraints, and indexes.
- The existing `data.create_contest_round` function, called by the scheduled
  contest jobs described in `services/tadoku-api/migrations/README.md`, predates
  this rule.

## SQL style

Always write SQL keywords in lowercase: `select` and `create table`, not
`SELECT` and `CREATE TABLE`.

## sqlc code generation

Queries live in one package per feature or shared storage repository under `services/tadoku-api/sql/<name>/`:
the feature's `.sql` query file, a `sqlc.yaml` that writes Go code to
`services/tadoku-api/generated/sqlc/<feature>/`, and a `generate.go` that pins the
sqlc version. Add a new package to `SQLC_PACKAGES` in `scripts/generate-sqlc.sh`.
The shared async outbox uses `sql/asyncoutbox/` and generates
`generated/sqlc/asyncoutbox/`, visible only to its repository.

Always regenerate after changing a SQL query. The checked-in generated Go files
must exactly match the query sources. Run the generator from the repository
root:

```sh
./scripts/generate-sqlc.sh
```

- The script downloads and runs the sqlc version pinned in each active package's
  `generate.go` (v1.31.1, generating pgx/v5 code), so `go` does not need to be
  installed or on `PATH`.
- Never edit sqlc-generated files by hand. Never delete, revert or selectively
  omit changes produced by the generator; commit the complete generated diff,
  even when it reveals previously stale output.
- If the generated changes are unexpected, investigate the query inputs and the
  pinned sqlc version, then rerun the generator. Do not discard the output.
- CI runs the same script on every pull request and fails if it changes the
  working tree.

## Transactions

`postgres.RunInTransaction(ctx, pool, callback)` in `services/tadoku-api/infra/postgres/` owns begin,
commit and rollback. The callback receives a context carrying the transaction.
Repositories select the database handle with `postgres.Executor(ctx, pool)` for
each operation and pass it directly to the sqlc queries:

```go
db, err := postgres.Executor(ctx, r.pool)
if err != nil {
    return err
}
return queries.New(db).InsertItem(ctx, params)
```

- The application owns transactions. Open one only when the operation needs
  one.
- Outside a transaction, `Executor` returns the pool; inside one, it returns the
  active transaction.
- Wrong-pool, nested and ended transaction scopes fail. An ended context never
  falls back to the pool. There are no retries or savepoints.
- Finish all work and row iteration before the callback returns. Do not run
  parallel SQL on one transaction or hold it across network or cache
  operations.
- Callback errors and panics keep their identity. Rollback gets an independent
  five-second timeout, so cancellation does not prevent the attempt.
- A commit transport error can leave the persistence outcome uncertain.

The helper's own test setup is described in
`services/tadoku-api/infra/postgres/README.md`.

## PostgreSQL errors

PostgreSQL and pgx errors are not classified by cause. They reach
`services/tadoku-api/transport/http/errors.go` without an `errx` kind unless a
repository operation translates one it expects:

| Error | Response |
| --- | --- |
| `pgx.ErrNoRows` checked by the operation | The operation's own error, usually not found |
| Unique violation checked by the operation with `postgres.IsUniqueViolation` | The operation's already-exists error, `409` |
| Any other PostgreSQL or pgx error, including other constraint violations, begin and commit failures and a closed or unreachable pool | `500`, logged at Error level |

- `postgres.IsUniqueViolation` in `services/tadoku-api/infra/postgres/errors.go`
  is the only SQLSTATE check. Check a SQLSTATE only where the operation gives
  that constraint a meaning.
- Errors wrapping `context.DeadlineExceeded` return `504`, and errors wrapping
  `context.Canceled` after the request was canceled return `499`, whatever
  their source.
- `503` is reserved for Keto, Kratos and Flipt provider failures. No database
  error maps to it; only the `/readyz` probe reports a PostgreSQL failure as
  `503`.
- Failed transactions are not retried or replayed, and wrapped causes remain
  available to `errors.Is` and `errors.As`.
