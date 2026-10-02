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

Tenant coverage uses a one-connection pool:

- Local settings end at commit or standalone statement completion.
- All SQL methods see the context tenant.
- Streaming, scan and commit errors release their batch.
- Missing, changed and conflicting scopes fail closed.
- 1,000 interleaved statements for three tenants see the correct key.
- Canceled batches leave the pool reusable.

## PgBouncer transaction pooling

The `pgbouncer_test` target requires `TADOKU_TEST_POSTGRES_URL` and
`TADOKU_TEST_PGBOUNCER_URL`; missing or unsafe configuration fails. Both URLs
accept only loopback, an explicit port, the `postgres` database, synthetic
`postgres:postgres` credentials and exactly `sslmode=disable`. The fixture
creates and drops its own migrated database through the direct PostgreSQL URL.
Gazelle excludes this source from its default test grouping; maintain the
dedicated target's dependencies when its imports change.

With the disposable PostgreSQL instance above running, start PgBouncer on a
Linux host:

```sh
docker run --rm -d --name tadoku-helper-pgbouncer --network host \
  -e DB_HOST=127.0.0.1 -e DB_PORT=15432 \
  -e DB_USER=postgres -e DB_PASSWORD=postgres \
  -e AUTH_TYPE=scram-sha-256 -e POOL_MODE=transaction \
  -e LISTEN_ADDR=127.0.0.1 -e LISTEN_PORT=6432 \
  -e MAX_CLIENT_CONN=1000 -e DEFAULT_POOL_SIZE=5 \
  -e MAX_PREPARED_STATEMENTS=200 \
  edoburu/pgbouncer:v1.25.2-p0@sha256:7d7a27d9e90985cab5cf42256f5c13a3120baa4b055b69df37beb272b89b2340
docker exec tadoku-helper-pgbouncer pg_isready -h 127.0.0.1 -p 6432 -U postgres
```

After readiness succeeds:

```sh
export TADOKU_TEST_PGBOUNCER_URL='postgres://postgres:postgres@127.0.0.1:6432/postgres?sslmode=disable'
bazel test --test_output=all //services/tadoku-api/infra/postgres:pgbouncer_test
bazel test --test_output=all --@rules_go//go/config:race \
  //services/tadoku-api/infra/postgres:pgbouncer_test
docker stop tadoku-helper-pgbouncer
```

The pinned image generates a wildcard database mapping when `DB_NAME` is unset.
Prepared statement tracking remains enabled for pgx's normal protocol. The test
uses 300 goroutines for three tenants, verifies all 12,000 reads and 6,000
writes, then proves the unsafe session-setting control leaks between separate
clients that share a backend. Session settings occur only inside that
disposable test control, never application code. CI prints the normal and race
test logs.

`testdata/sqlc` contains only synthetic query-generation inputs, not application
migrations. `internal/pgxcompat` is generated and Bazel-test-only. Regenerate with
`./scripts/generate-sqlc.sh`; it uses this fixture's separate sqlc v1.31.1 pin. Never edit generated Go by hand.
