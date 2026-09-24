# PostgreSQL helpers

`RunInTransaction` and `Executor` implement Tadoku API transactions. Their rules
are documented in [Database and migrations](../../../../docs/docs/tadoku-api/database.md#transactions).

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
Its top-level tests run in parallel, capped at four by the Bazel target. Each
test creates its own pool and fixture before issuing SQL; subtests remain
sequential. Tests using `timex.TheWorld` remain sequential in their own package.
Coverage includes cross-repository commit/rollback, panic, cancellation before
and after writes, deferred-constraint commit failure, pool reuse, nested/wrong
pool/ended context rejection, concurrent independent transactions, and direct
sqlc compatibility. Secondary cleanup transport failures are not simulated.

`testdata/sqlc` contains only synthetic query-generation inputs, not application
migrations. `internal/pgxcompat` is generated and Bazel-test-only. Regenerate with
`./scripts/generate-sqlc.sh`; it uses this fixture's separate sqlc v1.31.1 pin. Never edit generated Go by hand.
