# Tadoku API

Tadoku API replaces the legacy backend services operation by operation. Native
handlers own migrated operations; operations not yet migrated retain their existing
proxy ownership. A native failure never falls back to a legacy service.
Externally the gateway still adds `/api`.

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
application router or migrated handlers.

The router requires authentication middleware for application business handlers.
Startup always constructs it from the configured gateway JWKS; there is no opt-out
or feature flag. Register application routes through the router's `Handle` or
`HandleFunc` methods during construction; every such route inherits the shared
request deadline, authentication and ban check. Health probes and temporary proxy
registrations keep their existing behavior. Verified user claims travel in request
context through `internal/identity`.

Application operations compose features. Feature services own business decisions;
repositories only query and map rows. `postgres.Executor` lets repositories use
the active app-owned transaction. Open transactions only when the operation needs
one; do not add feature or repository interfaces solely for mocking.

Application operations and feature services that need authorization receive a
named `*permissions.Checker` and call `RequireAuthenticated`, `RequireAdmin` or
`IsAdmin` explicitly. Do not enforce administrator access with HTTP middleware.
Construction for a protected slice binds the checker to the request-scoped
`app:tadoku#admins` Keto lookup; the checker derives its subject from the verified
`internal/identity` context and does not cache results. The shared HTTP ban gate
remains separate and must not be duplicated in feature operations. Direct
invocation outside the application router must provide its own equivalent baseline
ban policy.

## Contract and compatibility

`spec/openapi.yaml` is the one canonical contract: 72 public operations plus nine
retained internal/callback operations. Source-prefixed operation IDs and component
names avoid collisions. Equivalent Content page/post templates share one canonical
path; their old parameter names are recorded for compatibility. This changes no
wire URLs. Original upstream server/path metadata preserves the inventory of
direct internal callers; the merge does not make those routes public.

Run `./scripts/generate-openapi.sh` for the shared DTOs and standard-library
strict-server bindings. The isolated oapi-codegen tool module under
`tools/oapi-codegen` generates every canonical component schema while
`spec/server-codegen.yaml` limits server registration to operations owned by this
application. Generation runs through Bazel and writes one checked-in output file.
The Bazel binary has no Go module build-info header, so read its pin from the tool
module's `go.mod` rather than the generated header. Generated routes register
through the application `Router` as their base router, preserving its shared
deadline, authentication and ban checks. Run `./scripts/generate-sqlc.sh` for SQL
output. Legacy service OpenAPI output remains frozen on v1.12.4 until those services
are retired. The native query uses sqlc v1.31.1/pgx-v5.

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
- `API_JWKS`, the gateway's public signing-key URL.
- `API_KETO_READ_URL`, the Keto read API URL. Tadoku API receives no Keto write
  URL or credential.

Startup fetches JWKS within `API_DIAL_TIMEOUT` and pings PostgreSQL before opening
listeners. Either failure aborts startup. Signing keys remain cached until restart;
there is no periodic refresh or refresh on an unknown key ID.
`/readyz` checks PostgreSQL; `/livez` remains independent of dependency health.
The existing proxy metrics and Go process metrics remain on the metrics listener
(`API_METRICS_PORT`, default 9090). They describe proxy request volume/latency/errors
and process health. This thin slice adds no native-specific metric family.
Shutdown closes request/metrics listeners, the pool and idle HTTP
connections. The dev deployment uses the existing disposable development DB role;
secret synchronization and reset scripts include Tadoku API.

Authentication verifies bearer JWT signatures and the existing `exp`, `nbf` and
`iat` time constraints. `exp` remains optional, and `iat` has no maximum age.
Issuer and audience are not additionally restricted. Subject, email, display name
and issued-at time are propagated; the identity's `CreatedAt` means token issue
time, not account creation time. JWT parsing itself performs no role, ban,
permission or service-audience policy. After authentication, the application router checks
the authenticated subject's direct `app:tadoku#banned` relation once. Missing,
empty and signed `guest` subjects skip Keto. A ban returns an empty 403, including
for administrators. Keto ban read errors are logged and deliberately allow
unprivileged handlers to continue, preserving the existing availability policy.
The failed lookup is kept in the request context so a later administrator check
returns unavailable without another ban query. Unlike legacy role enrichment,
this narrow check has no unrelated administrator lookup whose failure could
discard a successful ban result. Request deadlines bound the provider call.

Missing or malformed bearer headers return the legacy 400 JSON error; extracted
but invalid JWTs return its 401 JSON error. Anonymous gateway traffic supplies a
signed user token with subject `guest`, so it is distinct from a direct request
without credentials. Signed tokens missing `iat` now return 401 instead of the
legacy identity middleware's panic/500. Service tokens are unsupported and return
401; they are never converted into human identities. These two cases intentionally
differ from legacy behavior.

**Production activation is not part of this change.** Provision a dedicated
runtime credential with only the grants required by the application, not
a migration/admin DSN. Confirm provider TLS/pooling settings and the coexistence
connection budget: two native replicas default to eight connections, in addition
to the still-running legacy pools. Update production secrets/manifests and complete
the master rollout gate under separate release authorization before deployment.

## Verification

Native integration tests require `TADOKU_TEST_POSTGRES_URL` pointing to an explicit
loopback port, database `postgres`, credentials `postgres:postgres` and exactly
`sslmode=disable`. Missing or unsafe configuration fails before any connection;
new tests never skip. Fixtures create random disposable databases and apply the
complete canonical migration history. Do not point them at shared dev or production.
Relationship scenarios also start a pinned, official Linux x86-64 Keto v25.4.0
SQLite-enabled executable under Bazel. Each helper owns an in-memory, loopback-only
process using `infra/dev/ory/namespaces.keto.ts`; it never accepts an external target.

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
Cases needing seed data have their own SQL and/or relationship setup; fixtures are
not generated in Go.
Reset and seeding commit before requests run,
so application transactions commit normally.
There is no outer rollback transaction and no change to `RunInTransaction`.

Database helpers take contexts and return errors, with explicit `Close` cleanup
instead of `testing.TB`. Suite teardown preserves test failures and reports cleanup
failures; partial setup also cleans up. Freeze business time inside the tests that
need it, not in `TestMain`; a scenario can use separate `timex.TheWorld` scopes for
different times. No test-only request-context wrapper is needed. The pool-closing
failure test has isolated dependencies. Repository/transaction tests retain their
independent databases and may run in parallel.

HTTP cases derive their name from the operation, expected status and description
using `APITestName(operation, status, description...)`. Use the same name for the
subtest and its fixture directory:

```text
e2e/testdata/<operation>/<status>_<description>/
  setup.sql       # optional
  relationships.json # optional Keto relation-tuple array
  request.http
  golden.http
```

Each operation's tests use an explicit Go table. Each row declares
`description []string` and `want` as an HTTP status constant; there is no separate
fixture-name field to keep in sync. Add a case by adding a descriptive table row and
its request and golden files; do not discover cases from directories. Add `setup.sql`
or `relationships.json` only when the case needs that seed data. Shared PostgreSQL
and Keto cleanup always runs; absent setup files are skipped, while other read,
decode and provider errors fail the scenario.
The `golden.http` file contains the request label
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
optional `setup.sql` and `relationships.json` plus required `request.http` and
`golden.http`; each must independently match the
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

```sh
bazel test //services/tadoku-api/e2e:e2e_test --test_output=errors
```

The existing CI E2E and race targets include these comparisons automatically.
Add future migrated operations explicitly; do not add a general service launcher
or duplicate per-implementation fixtures.

Pool-failure, cancellation and proxy-routing checks remain separate Go tests;
they exercise dependency behavior rather than SQL-defined response cases.

Endpoint contract tests explicitly supply passthrough authentication so their
credential-free request fixtures remain focused on business behavior. Authentication
scenarios register `GET /test/authentication` through the real production `Router`
with a test-only success handler. The comparison handler uses legacy `VerifyJWT` and
`Identity`, without authorization or business endpoints. Both consume the same signed
HTTP requests and goldens. Test-only identity headers prove downstream context
propagation. A routing regression verifies these fixture handlers retain normal
method/path dispatch, while the transport router test proves all registered application
routes inherit the shared gates. No test endpoint is added to production.

The suite serves a synthetic checked-in public JWKS locally; private keys and live
identity providers are not needed. Each authentication scenario temporarily binds
`jwt/v4.TimeFunc` to the scoped application clock and restores it on return. These
scenarios and their parents must not run in parallel. Intentional compatibility
differences use `skipParity` in the same table and run only against Tadoku API.
Ban-policy scenarios register `GET /test/banned` on the same production router and
compare it with legacy `VerifyJWT`, `Identity`,
`RolesFromKeto` and `RejectBannedUsers` using the same real Keto fixture. Provider
fail-open behavior and deadlines are tested at the narrow middleware boundary;
authentication matrices are not repeated for every operation.

Authentication goldens run at the fixed instant `2026-09-12T12:00:00Z`
(`1789214400`). Reuse an existing signed user request when only its Keto relationship
tuples change. When a scenario needs different JWT claims, generate a new synthetic
key and token locally with Node's built-in cryptography, then append the printed public
JWK to the existing `keys` array without removing any checked-in keys. Run this from
the repository root with Node.js and `jq` installed:

```sh
fixture_output=$(mktemp)
node <<'NODE' > "$fixture_output"
const { generateKeyPairSync, randomUUID, sign } = require("node:crypto");

const { privateKey, publicKey } = generateKeyPairSync("rsa", { modulusLength: 2048 });
const kid = `tadoku-fixture-${randomUUID()}`;
const publicJWK = {
  ...publicKey.export({ format: "jwk" }),
  kid,
  use: "sig",
  alg: "RS256",
};
const claims = {
  iss: "http://oathkeeper-api/",
  sub: "22222222-2222-4222-8222-222222222222",
  iat: 1789214400,
  nbf: 1789214400,
  exp: 1789218000,
  type: "user",
  session: {
    identity: {
      traits: {
        display_name: "Fixture User",
        email: "fixture@example.test",
      },
    },
  },
};
const encode = value => Buffer.from(JSON.stringify(value)).toString("base64url");
const signingInput = `${encode({ alg: "RS256", typ: "JWT", kid })}.${encode(claims)}`;
const signature = sign("RSA-SHA256", Buffer.from(signingInput), privateKey).toString("base64url");

process.stdout.write(JSON.stringify({ publicJWK, token: `${signingInput}.${signature}` }, null, 2));
NODE

jq --slurpfile fixture "$fixture_output" \
  '.keys += [$fixture[0].publicJWK]' \
  services/tadoku-api/e2e/testdata/authentication.jwks.json \
  > services/tadoku-api/e2e/testdata/authentication.jwks.json.new
jq -r .token "$fixture_output"
```

Review the `.new` JWKS before replacing the fixture and paste the printed token into
the new `request.http`. The private key exists only inside that Node process. Editing
claims in a token by hand invalidates its signature, so rerun the recipe instead.
Never use production signing keys or tokens in fixtures.

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

## Migration notes

Track deferred cleanup in the [migration log](MIGRATION_LOG.md).
