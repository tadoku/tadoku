# Transactions

`RunInTransaction(ctx, pool, callback)` owns begin, commit, and rollback.
The callback receives a context carrying the transaction. Repositories select
the database handle with `Executor(ctx, pool)` for each operation, then pass it
directly to native pgx-compatible sqlc queries:

```go
db, err := postgres.Executor(ctx, r.pool)
if err != nil {
    return err
}
return queries.New(db).InsertItem(ctx, params)
```

Outside a transaction, `Executor` returns the pool. Inside one, it returns the
active transaction. Wrong-pool, nested, and ended transaction scopes fail; an
ended context never falls back to the pool. There are no retries or savepoints.
All work and row iteration must finish before the callback returns. Do not run
parallel SQL on one transaction or hold it across network/cache operations.

Callback errors and panics retain their identity. Cleanup gets an independent
five-second timeout so cancellation does not prevent the rollback attempt.
A commit transport error can leave an uncertain persistence outcome.

## Tests

The suite requires real disposable PostgreSQL. Missing configuration fails;
tests are never silently skipped. Only loopback hosts, an explicit port, the
`postgres` database, synthetic `postgres:postgres` credentials, and
`sslmode=disable` are accepted. Never use a forwarded/shared/application DB.
Each test cleans up only its randomly named synthetic schema.

Start a disposable instance (no volume or persistent data):

```sh
docker run --rm -d --name tadoku-helper-tests \
  --tmpfs /var/lib/postgresql/data \
  -e POSTGRES_PASSWORD=postgres -p 127.0.0.1:15432:5432 postgres:17
docker exec tadoku-helper-tests pg_isready -U postgres
```

Wait until `pg_isready` reports accepting connections, then from the repo root:

```sh
export TADOKU_TEST_POSTGRES_URL='postgres://postgres:postgres@127.0.0.1:15432/postgres?sslmode=disable'
bazel test --@rules_go//go/config:race \
  //services/tadoku-api/infra/postgres:postgres_test \
  //services/tadoku-api/internal/timex:timex_test
docker stop tadoku-helper-tests
```

CI reuses its disposable PostgreSQL service and runs both normal and race tests.
The transaction test target disables result caching and remote execution.
Coverage includes cross-repository commit/rollback, panic, cancellation before
and after writes, deferred-constraint commit failure, pool reuse, nested/wrong
pool/ended context rejection, concurrent independent transactions, and direct
sqlc compatibility. Secondary cleanup transport failures are not simulated.

`testdata/sqlc` contains only synthetic query-generation inputs, not application
migrations. `internal/pgxcompat` is generated and Bazel-test-only. Regenerate with
`./scripts/generate-sqlc.sh`; it uses this fixture's separate sqlc v1.31.1 pin
without upgrading the legacy generators. Never edit generated Go by hand.
