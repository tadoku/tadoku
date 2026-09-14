# Tadoku API

Tadoku API replaces the legacy backend services operation by operation. Native
handlers own migrated operations; operations not yet migrated retain their existing
proxy ownership. A native failure never falls back to a legacy service.
Externally the gateway still adds `/api`.

**Current native coverage:** `GET /content/announcements/{namespace}/active`.
HEAD and OPTIONS on that path remain proxied. This read is deliberately
**unprotected**: it does not validate JWTs, call Keto, check bans or interpret
service audiences. Authentication and shared authorization
will be separate reviewed increments. Existing proxied routes are unchanged.

## Code ownership

```
transport/http -> app -> features/<feature> -> generated/sqlc/<feature>
cmd/tadoku-api constructs and closes the shared pgx/v5 pool and HTTP resources
```

`transport/http/router.go` constructs the application router with standard
method/path registrations, request deadlines and its own health checks. It does
not depend on upstream URLs, proxy transports or proxy metrics. Startup separately
calls `RegisterProxyRoutes` to attach the temporary legacy routes. That call and
`proxy.go` can be removed when the migration is complete without changing the
application router or announcement handler. The temporary HEAD override lives
in `proxy.go`; the application route is an ordinary GET registration.

For example, `app/announcements.go` exposes `ListActiveAnnouncements`.
Content service chooses one `timex.Now()` cutoff and a limit of ten. Its concrete
repository only queries and maps rows. `postgres.Executor` makes the repository
usable inside a future app-owned transaction; this read does not open an
unnecessary transaction. There are no new feature/repository interfaces.

## Contract and compatibility

`spec/openapi.yaml` is the one canonical contract: 72 public operations plus nine
retained internal/callback operations. Source-prefixed operation IDs and component
names avoid collisions. Equivalent Content page/post templates share one canonical
path; their old parameter names are recorded for compatibility. This changes no
wire URLs. Original upstream server/path metadata preserves the inventory of
direct internal callers; the merge does not make those routes public.

Run `./scripts/generate-openapi.sh` for the single framework-free DTO package.
The generator is the existing oapi-codegen v1.12.4 pinned in `go.mod`, now runnable
through Bazel. Its Bazel binary has no Go module build-info header; the pin is in
`go.mod`, not the generated header. Run `./scripts/generate-sqlc.sh` for SQL output.
The native query uses sqlc v1.31.1/pgx-v5; legacy generators are unchanged.

Documentation builds filtered public views from the canonical contract. They
retain the four existing documentation sections/URLs, not four independent API
contract sources. Legacy specs are frozen build and contract-comparison inputs.
The contract tests compare all 81 operations and their component/security shapes.
Manually registered health and metrics endpoints are separately inventoried in
the contract's `x-tadoku-operational-surfaces` metadata, not rendered as product API.

Every migrated operation must preserve its existing API inputs, outputs and
business behavior, including response status, headers, field shapes, nullability,
empty results, filtering, ordering and limits where applicable. Prove parity by
running the same request/response golden cases against the native and corresponding
legacy handlers. Native failures must not fall back to the proxy. Test shared
authentication and infrastructure middleware at their own boundaries,
not duplicated in each endpoint's parity cases.

## Runtime configuration

In addition to the existing four upstream URLs, startup now requires:

- Individual `API_POSTGRES_HOST`, `PORT` (default 5432), `DATABASE`, `USER`,
  `PASSWORD`, `SSLMODE` fields. `API_POSTGRES_URL` remains rejected.
- `API_POSTGRES_MAX_CONNECTIONS` (default 4, validated range 1–32).

Startup pings PostgreSQL before opening listeners. No JWKS or Keto config is needed.
`/readyz` checks PostgreSQL; `/livez` remains independent of dependency health.
The existing proxy metrics and Go process metrics remain on the metrics listener
(`API_METRICS_PORT`, default 9090). They describe proxy request volume/latency/errors
and process health. This thin slice adds no native-specific metric family.
Shutdown closes request/metrics listeners, the pool and idle HTTP
connections. The dev deployment uses the existing disposable development DB role;
secret synchronization and reset scripts include Tadoku API.

**Production activation is not part of this change.** Provision a dedicated
runtime credential with connect/schema-usage/select-on-announcements grants, not
a migration/admin DSN. Confirm provider TLS/pooling settings and the coexistence
connection budget: two native replicas default to eight connections, in addition
to the still-running legacy pools. Update production secrets/manifests and complete
the master rollout gate under separate release authorization before deployment.
No schema migration or write-owner handoff is required for this read.

## Verification

Native integration tests require `TADOKU_TEST_POSTGRES_URL` pointing to an explicit
loopback port, database `postgres`, credentials `postgres:postgres` and exactly
`sslmode=disable`. Missing or unsafe configuration fails before any connection;
new tests never skip. Fixtures create random disposable databases and apply the
complete canonical migration history. Do not point them at shared dev or production.

```sh
bazel test //services/tadoku-api/... --test_output=errors
```

`TestMain` creates one disposable database, applies migrations and constructs the
native production HTTP router and the legacy handlers required by the E2E suite
once. Handlers and the fallback sentinel execute in process, without HTTP listeners.
Cover each operation's response mapping, input handling, business rules and relevant
boundaries; keep dependency-failure and route-ownership checks alongside those cases.

HTTP scenarios run sequentially and call `reset` before each implementation of each
scenario, not between dependent requests. `internal/testpostgres/cleanup.sql`
explicitly lists mutable tables to truncate with `restart identity`. Add tables
there as their slices gain tests; do not discover tables automatically or use
`cascade`. Static data from migrations and `schema_migrations` are preserved.
Each case has its own SQL setup; fixtures are not generated in Go.
Reset and seeding commit before requests run,
so application transactions commit normally.
There is no outer rollback transaction and no change to `RunInTransaction`.

Database helpers take contexts and return errors, with explicit `Close` cleanup
instead of `testing.TB`. Suite teardown preserves test failures and reports cleanup
failures; partial setup also cleans up. Freeze business time inside the tests that
need it, not in `TestMain`; a scenario can use separate `timex.TheWorld` scopes for
different times. No extra request-context wrapper is needed. The pool-closing failure test
has isolated dependencies. Repository/transaction tests retain their independent
databases and may run in parallel.

HTTP cases derive their name from the operation, expected status and description.
For example, `APITestName("ListActiveAnnouncements", http.StatusOK, "without", "auth")`
produces `ListActiveAnnouncements/200_without_auth`, used for both the subtest and
its fixture directory:

```text
e2e/testdata/ListActiveAnnouncements/200_without_auth/
  setup.sql
  request.http
  golden.http
```

Each operation's tests use an explicit Go table. Each row declares
`description []string` and `want` as an HTTP status constant; there is no separate
fixture-name field to keep in sync. Add a case by adding a descriptive table row and
its three fixture files; do not discover cases from directories. Each case owns its seed,
including an explicit comment-only `setup.sql` for an empty database. Shared
cleanup runs before that SQL. The `golden.http` file contains the request label
and complete expected response. Tests parse the request files with `net/http`,
execute the production handler at the operation's minimum required access level,
check the HTTP status against the table's `want`, and compare the entire response,
including status, headers and body, with the golden.
Explicit seed IDs and frozen business time make responses deterministic; only
HTTP line endings are normalized. Missing or changed goldens fail the test.
There is no automatic recording mode: edit and review the expected files for an
intentional contract change. Lifecycle tests stay focused on startup/shutdown,
not a growing list of endpoint assertions.

### Legacy parity

For every migrated operation, run each case against the native API and its
corresponding legacy API in separately named subtests. Both consume the same
`setup.sql`, `request.http` and `golden.http`; each must independently match the
full response. Use production route registration, handlers, domain operations,
repositories and generated queries, mounting the legacy routes at the matching
API prefix. Initialize only the dependencies the tested operations need. Confine
legacy assembly to test targets; do not add legacy service or framework dependencies
to the native runtime.

Time-dependent cases must give both implementations the same controlled time input.
When a legacy query reads the database clock, bind that input explicitly in test-only
code. Reject unsupported query shapes instead of silently applying a generic SQL
rewrite. Apart from clock binding, execute production queries and row mapping
unchanged against real PostgreSQL. Do not shift or ignore fixture timestamps or
response fields to make a comparison pass. Such bindings do not verify the legacy
choice of clock. Gateway rewriting, authentication and infrastructure middleware
also remain outside endpoint parity; test them at their respective boundaries.

**Current coverage and clock-binding limit:** the six cases in
`e2e/announcements_test.go` run against Tadoku API and Content API.
`e2e/legacy_content_test.go` binds the active-announcements query's two `now()` calls
to `timex.Now()`, accepting only that query name and its original namespace argument.
Other legacy queries require their own reviewed clock handling where needed.

```sh
bazel test //services/tadoku-api/e2e:e2e_test --test_output=errors
```

The existing CI E2E and race targets include these comparisons automatically.
Add future migrated operations explicitly; do not add a general service launcher
or duplicate per-implementation fixtures.

Pool-failure, cancellation and proxy-routing checks remain separate Go tests;
they exercise dependency behavior rather than SQL-defined response cases.

**Testing decision:** endpoint tests prove the minimum access level and business
behavior. When auth middleware is added, test its credential/role/ban/failure matrix
once at that boundary. Do not repeat that matrix or emulate Keto in every endpoint
fixture. This rule is also recorded in `AGENTS.md` for future agents.

CI checks Depolicy, OpenAPI/sqlc generation and the database suites. The standalone
Echo and Testify graph checks have been removed. Bazel visibility remains in place.

### Import policies

```sh
bazel run //tools/ci/depolicy
```

Depolicy is pinned to `d754cd9f261c92d7422a34c7f4b721044ea0b8c3` in `go.mod`.
Installation was explicitly approved on 13 September 2026 after review of the
upstream license-file absence. This records project approval, not a change to
upstream licensing.

The Bazel runner validates the root `.depolicy.yaml` and `go.mod`, then invokes
the unmodified upstream analyzer on every Go file in this subtree. Tests,
generated code and inactive build-tag files are included. It uses Bazel's pinned
SDK sources to classify standard-library imports, with no system Go installation
or package downloads at runtime. Missing/invalid or nested configuration, an
empty scope, uncovered/ambiguous packages and denied imports fail the check.
Compilation/type checking remains in the ordinary Bazel build.

Policies describe layer direction: transport uses application operations and
HTTP contract types; application code composes features; features use their own
domain and generated SQL plus infrastructure; infrastructure and domain code
cannot import application or transport code. Shared internal utilities sit below
these layers. Startup and integration tests are assembly boundaries.

Rules use layer/feature patterns, not lists of utility, database-driver or provider
packages. Standard-library and third-party imports are outside this local-layer
check; dependency choices still follow the repository's development guidelines.
Same-package and external-package tests both use their directory's layer policy,
without synthetic package names or per-feature test exceptions. Fixture libraries
retain Bazel's `testonly` restrictions.

Depolicy replaces the temporary graph checks. It checks direct imports, not
transitive dependencies.

Same-package service/repository responsibilities and business signatures still
require review; import rules do not enforce those conventions.
