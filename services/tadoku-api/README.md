# Tadoku API

Tadoku API replaces the legacy backend services operation by operation. Native
handlers own migrated operations; operations not yet migrated retain their existing
proxy ownership. A native failure never falls back to a legacy service.
Externally the gateway still adds `/api`.

## Code ownership

```
transport/http -> app -> features/<feature> -> generated/sqlc/<feature>
cmd/tadoku-api constructs and owns the pgx/v5 pool, raw Valkey/Kratos clients and HTTP resources
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
context through `internal/identity`. The shared ban lookup records a confirmed ban
for role-introspection reads so they can report it; every other business route
rejects that identity before its handler runs.

Application operations compose features. Feature services own business decisions;
repositories only query and map rows. `postgres.Executor` lets repositories use
the active app-owned transaction. Open transactions only when the operation needs
one; do not add feature or repository interfaces solely for mocking.

Each feature package groups its operations in `<feature>_service.go` and
`<feature>_repository.go`, with matching `_test.go` files. Keep domain types,
errors and shared validation in `domain.go`; service and repository structs and
constructors stay with their implementations. Declare each feature's domain errors
together in one `var` block and reuse them from services and repositories.
Within feature packages, database tests exercise repositories directly and live
only in `<feature>_repository_test.go`.
Service and validation tests in `<feature>_service_test.go` remain database-free.

Application errors use `internal/errx.Error`, which carries a transport-neutral
`Kind`, a message and an optional cause. Use named constructors such as
`errx.NewInvalidInputError(message)` or `errx.NewUnavailableError(message, cause)`;
the latter accepts `nil` when there is no underlying cause.
`errx.KindOf` reads the outermost typed
error using `errors.As`; the HTTP boundary maps its kind to a status, with unknown
or unclassified errors returning 500. Ordinary `%w` wrapping preserves metadata,
and `Unwrap` preserves causes for `errors.Is`/`errors.As`. Do not encode categories
in error text or wrap category sentinels. Legacy error types and mapping stay
unchanged. Add call-site context only when it contributes useful diagnostics.

Application operations and feature services that need authorization receive a
named `*permissions.Checker` and call `RequireAuthenticated`,
`RequireAuthenticatedAllowingUnknownBan`, `RequireAdmin` or `IsAdmin` explicitly.
Do not enforce administrator access with HTTP middleware.
Construction for a protected slice binds the checker to the request-scoped
`app:tadoku#admins` Keto lookup; the checker derives its subject from the verified
`internal/identity` context and does not cache results. The shared HTTP ban gate
remains separate and must not be duplicated in feature operations. Direct
invocation outside the application router must provide its own equivalent baseline
ban policy. `RequireAuthenticated` and administrator checks fail closed when the
shared ban lookup is inconclusive. `RequireAuthenticatedAllowingUnknownBan` is an
explicit availability opt-out for read-only operations. Operations that mutate
state must never use the fail-open variant.

Response enrichment consumes the concrete shared authorization service from
`services/common/authz/roles`. Batch facts are read for each request and are
never treated as caller authorization; the operation still uses the permission
checker for its own access decision and the shared HTTP ban gate still applies
first. Native feature services accept concrete application collaborators by
default. A narrow interface is reserved for a real provider or layer boundary,
not as a seam for mock-only tests.

Provider-backed read caches reuse the common provider client's cursor support,
preserve provider order, read every cursor page and reject repeated continuation
tokens. Construction performs no provider request. The first operation that needs
a cache loads it with the request context; later operations reuse that snapshot
for five minutes and serialize refreshes. Provider availability is not a startup
or health gate. A provider refresh failure retains the last complete snapshot;
an initial provider failure returns unavailable because no complete snapshot
exists yet.

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
legacy handlers with real authentication and authorization. Native failures must
not fall back to the proxy. Keep exhaustive authentication and infrastructure
failure matrices at their own boundaries instead of duplicating them per endpoint.

## Runtime configuration

In addition to the existing four upstream URLs, startup now requires:

- Individual `API_POSTGRES_HOST`, `PORT` (default 5432), `DATABASE`, `USER`,
  `PASSWORD`, `SSLMODE` fields. `API_POSTGRES_URL` remains rejected.
- `API_POSTGRES_MAX_CONNECTIONS` (default 4, validated range 1–32).
- `API_VALKEY_URL`, one standalone TCP URL accepted by `valkey-go`. URL
  credentials, TLS, databases and client options are preserved except for the
  service's timeout, retry and pipelining policy; Sentinel, multiple-address and
  Unix-socket configurations are rejected.
- `API_VALKEY_TIMEOUT` (default 1s), the positive bound for each connection and
  handshake attempt and the established-connection keepalive/I/O interval.
- `API_JWKS`, the gateway's public signing-key URL.
- `API_MAX_TOKEN_AGE` (default 24h), the maximum accepted age since `iat`.
- `API_JWT_ISSUER`, an optional exact issuer match. Empty leaves issuer unchecked
  for rollout compatibility.
- `API_KETO_READ_URL`, the Keto read API URL. Existing ban and administrator
  checks retain a separate read-only client with a 2s total request timeout.
- `API_KETO_WRITE_URL`, an absolute HTTP(S) base URL for the existing Keto write
  service. Credentials, query strings and fragments are rejected. Path prefixes
  are supported; trailing slashes are removed. Development uses
  `http://keto-write.default:4467`.
- `API_KETO_WRITE_TIMEOUT` (default 2s), a positive total HTTP request timeout
  for the retained raw read/write client, including response-body reads. It shares
  the owned transport's `API_DIAL_TIMEOUT`, `API_RESPONSE_HEADER_TIMEOUT` and
  `API_IDLE_TIMEOUT` bounds.
- `API_KRATOS_ADMIN_URL`, an absolute HTTP(S) base URL for the existing Kratos
  admin service. Credentials, query strings and fragments are rejected.
  Path prefixes are supported; trailing slashes are removed. Development
  uses `http://kratos-admin.default`.
- `API_KRATOS_TIMEOUT` (default 2s), a positive total HTTP request timeout,
  including reading the response body. The owned transport also applies
  `API_DIAL_TIMEOUT`, `API_RESPONSE_HEADER_TIMEOUT` and `API_IDLE_TIMEOUT`.

Startup fetches JWKS within `API_DIAL_TIMEOUT` and pings PostgreSQL before opening
listeners. Either failure aborts startup. Tadoku API also constructs and owns its
raw Valkey client unconditionally. Invalid Valkey configuration or canceled setup
aborts startup; an unavailable standalone server logs a warning and starts in
degraded mode so the retained client can reconnect on a later command. Signing
keys remain cached until restart; there is no periodic refresh or refresh on an
unknown key ID.
Startup also constructs and retains the raw Kratos SDK and shared Keto read/write
clients; constructing them makes no provider request. After the application is
successfully constructed, provider-backed caches remain cold until a request needs
them. Provider availability and cache refresh completion are not startup or
health-check gates; startup and health checks never mutate provider state.
`/readyz` checks PostgreSQL; `/livez` remains independent of dependency health.
Valkey is deliberately not a readiness or liveness gate. Commands use the caller's
context, and blocking commands require an explicit caller deadline. The raw client
preserves `valkey-go` behavior during a lazy connection or reconnect: canceling
before an existing deadline may wait for that deadline or `API_VALKEY_TIMEOUT`
while the handshake or another caller's shared setup finishes. Streaming commands
also retain the upstream client's native cancellation behavior. See
[`infra/valkey`](infra/valkey/) for the direct command pattern.

The existing proxy metrics and Go process metrics remain on the metrics listener
(`API_METRICS_PORT`, default 9090). They describe proxy request volume/latency/errors
and process health. This thin slice adds no native-specific metric family.
Shutdown closes request/metrics listeners, the pool, the Valkey client and idle
HTTP connections, including Kratos and Keto connections. Startup failure closes
the same owned transport. Raw clients have no separate close operation. Shutdown
closes database and provider transport dependencies after request handling stops.
Valkey close follows the upstream client's native per-connection
close allowance rather than `API_VALKEY_TIMEOUT`. The dev deployment uses the
existing disposable development DB role;
secret synchronization and reset scripts include Tadoku API.

Authentication accepts only RS256 bearer JWTs, requires `exp` and `iat`, and
rejects tokens older than `API_MAX_TOKEN_AGE`. It also verifies the existing `nbf`
constraint and, when configured, requires an exact `API_JWT_ISSUER` match. Audience
is not additionally restricted. Subject, email, display name and issued-at time are
propagated; the identity's `CreatedAt` means token issue time, not account creation
time. JWT parsing itself performs no role, ban,
permission or service-audience policy. After authentication, the application router checks
the authenticated subject's direct `app:tadoku#banned` relation once. Missing,
empty and signed `guest` subjects skip Keto. A confirmed ban is returned to
role-introspection reads as request context and returns an empty 403 from every
other business route, including for administrators. Keto ban read errors are logged and deliberately allow
the shared gate to continue, preserving the existing availability policy for
explicitly opted-in read operations. The failed lookup is kept in the request
context so authenticated and administrator checks return unavailable without
another ban query. Unlike legacy role enrichment, this narrow check has no
unrelated administrator lookup whose failure could discard a successful ban
result. Request deadlines bound the provider call.

Missing or malformed bearer headers return the legacy 400 JSON error; extracted
but invalid JWTs return its 401 JSON error. Anonymous gateway traffic supplies a
signed user token with subject `guest`, so it is distinct from a direct request
without credentials. Signed tokens missing `iat` or `exp`, exceeding the configured
maximum age, or failing the configured issuer check return 401. A missing `iat`
avoids the legacy identity middleware's panic/500. Service tokens are unsupported
and return 401; they are never converted into human identities. These cases
intentionally differ from legacy behavior.

**Production activation is not part of this change.** Provision a dedicated
runtime credential with only the grants required by the application, not
a migration/admin DSN. Confirm provider TLS/pooling settings and the coexistence
connection budget: two native replicas default to eight connections, in addition
to the still-running legacy pools. Update production secrets/manifests and complete
the master rollout gate under separate release authorization before deployment.

### Raw Keto relationship primitive

The composition root retains a concrete `*ketoclient.Client` from
`services/common/client/keto` on `application.keto`. Pass concrete clients and
shared application services explicitly into consumers. Existing ban/admin
consumers receive the separate `NewReadClient` instance, which has no configured
write API.

`keto.NewClient(readURL, writeURL, keto.WithHTTPClient(httpClient))` applies options
to both APIs. The caller owns the HTTP client and transport. The two-argument
constructor remains compatible with existing callers and keeps SDK defaults;
Tadoku API supplies the bounded HTTP client described above.

Use the existing primitives with the operation's caller context and a complete
tuple target. Set exactly one of `Subject.ID` or `Subject.Set`:

```go
subject := keto.Subject{ID: subjectID}
err := client.AddRelation(ctx, namespace, object, relation, subject)
// Delete only this namespace/object/relation/subject tuple.
err = client.DeleteRelation(ctx, namespace, object, relation, subject)

group := keto.Subject{Set: &keto.SubjectSet{
    Namespace: groupNamespace,
    Object:    groupID,
    Relation:  membershipRelation,
}}
err = client.AddRelation(ctx, namespace, object, relation, group)
err = client.DeleteRelation(ctx, namespace, object, relation, group)
```

Direct subjects use `subject_id`; subject sets retain their namespace, object and
relation. Delete sends every target component, including every subject-set field.
The existing idempotency rules remain: add treats HTTP 409 as success and delete
treats HTTP 404 as success. Other errors retain their wrapped provider error;
`errors.As` can inspect `*ketoapi.GenericOpenAPIError` and its `Body()`/`Model()`.
`errors.Is` identifies `context.Canceled` and `context.DeadlineExceeded`. These
primitives return an error only, not separate HTTP response metadata.

The total client timeout, any earlier caller deadline and caller cancellation
bound requests, including response-body reads. No mutation retry loop is added;
a timeout or cancellation does not establish whether Keto committed the write.
Future migrations supply authorization, auditing and application consumers.

### Raw Kratos primitive

The composition root keeps a concrete `*kratosapi.APIClient` on its runtime
`application.kratos` field, ready to pass explicitly into a future consumer's
constructor. `services/common/client/kratos.NewAPIClient(baseURL, kratos.WithHTTPClient(httpClient))`
constructs the pinned `github.com/ory/kratos-client-go` v0.11.1 SDK. It does not
validate deployment configuration or own the supplied HTTP client. Tadoku API
validates configuration and supplies a client with a total timeout using its
existing owned HTTP transport. Both constructors accept `WithHTTPClient`;
`NewClient(baseURL, kratos.WithHTTPClient(httpClient))` uses that same client for
SDK operations and cursor pagination. Existing `NewClient(baseURL)` callers keep
their helper behavior and the SDK's default HTTP client.

Call the SDK directly with the operation's caller context:

```go
identity, response, err := kratos.IdentityApi.GetIdentity(ctx, identityID).Execute()
```

`response` can be nil for transport/cancellation failures. When present, its
status and headers remain available even with an error. Use `errors.As` to inspect
`*kratosapi.GenericOpenAPIError`, including its `Body()` and `Model()`, and
`errors.Is` for `context.Canceled` or `context.DeadlineExceeded`.

The raw client returns the SDK's models, response metadata and errors unchanged;
it does not apply legacy domain mapping, trait policy, account-age rules or the
legacy helpers' not-found and idempotent-delete translations. Consuming features
own those decisions and any cache lifecycle. Use the pinned SDK's request builders
and pagination options directly. The total client timeout and any earlier caller
deadline bound requests; caller cancellation also interrupts response-body reads.

## Verification

Prefer real-database repository tests plus HTTP E2Es over testing feature services
in isolation. Repository tests are highly recommended for query behavior, row
mapping, constraints and persistence; E2Es cover feature-service orchestration and
the HTTP contract. Keep parameter validation and pure domain rules in unit tests.
Do not add database-backed feature-service tests or repeat dependency-failure
matrices for each endpoint using an already-tested dependency.

Native integration tests require `TADOKU_TEST_POSTGRES_URL` pointing to an explicit
loopback port, database `postgres`, credentials `postgres:postgres` and exactly
`sslmode=disable`. Missing or unsafe configuration fails before any connection;
new tests never skip. Fixtures create random disposable databases and apply the
complete canonical migration history. Do not point them at shared dev or production.
Relationship scenarios also start a pinned, official Linux x86-64 Keto v25.4.0
SQLite-enabled executable under Bazel. Each helper owns an in-memory, loopback-only
process using `infra/dev/ory/namespaces.keto.ts`; it never accepts an external target.
HTTP E2Es also own a pinned Kratos v26.2.0 SQLite-enabled Linux x86-64 process.
`internal/testkratos` runs its migrations against a private database under writable
tmpfs at `/dev/shm`, then seeds the standard identities from `e2e/testdata/kratos.sql`.
It uses the shared `infra/dev/ory/identity.default.schema.json` and private Unix
sockets, with a bounded HTTP client supplied to the existing raw SDK constructor.
No Docker service, external Kratos URL or Kratos environment variable is required
by these tests. Startup and reset are bounded; failure/shutdown closes connections,
reaps the child and removes the owned database, journals and sockets.

Direct SQL seeding preserves the existing signed subjects, traits and timestamps.
This is SQLite on RAM-backed storage, not Kratos's process-private `dsn: memory`.
The test seed depends on the pinned provider schema and creates identity records
only; it does not create passwords, credential identifiers or sessions. Tests for
those operations must arrange the corresponding provider state explicitly. Guest
and unauthenticated requests have no Kratos identity; roles and bans stay in Keto.

Raw Valkey and application lifecycle tests also require
`TADOKU_TEST_VALKEY_URL` in the exact form `redis://127.0.0.1:<port>` (or
`localhost`) for a disposable Valkey 9 service. They isolate and delete their own
keys and never flush the shared instance.

```sh
bazel test //services/tadoku-api/... --test_output=errors
```

`TestMain` creates one disposable database, applies migrations and constructs the
native production HTTP router and the legacy handlers required by the E2E suite
once, using real JWT verification, ban checks and Keto-backed permissions. There
is no second bypass router or injected administrator identity. Handlers and the
fallback sentinel execute in process, without HTTP listeners.
The suite also retains the raw Kratos SDK through `api.kratos.Client()` for future
explicit dependency injection. Construction adds no business consumer or startup
identity lookup to the running application.
Cover each operation's response mapping, input handling, business rules and relevant
boundaries. Keep shared dependency-failure and route-ownership checks at their own
boundaries instead of repeating them for every operation.

New API request bodies use the generated JSON decoder. Do not add XML/form adapters
or non-JSON request fixtures to reproduce legacy binder behavior. Mark intentional
JSON-decoding differences in the parity test table.

HTTP scenarios run sequentially and call `reset` before each implementation of each
scenario, not between dependent requests. **Kratos is seeded once in `TestMain` and
excluded from that ordinary reset.** Unmarked tests must leave its state unchanged.
For a case that may mutate Kratos, set `resetKratos: true` on each native/legacy
`implementation`. For a journey or direct test, call `resetKratosAfter(t, api)` on
the enclosing test before any mutation. The marker registers `t.Cleanup` to restore
the full seed after the implementation or whole journey, even after `t.Fatal` or
cancellation. It uses a fresh bounded context because `t.Context()` is cancelled
before cleanup. Never mark individual journey steps.

An explicit Kratos reset stops the process, recreates/migrates the database and
restores the original seed. Existing SDK clients keep working through the same
socket path. A reset failure fails the test and subsequent cases reject the
unusable fixture. Neither automatic mutation detection nor response UUID
normalization is used; the existing tokens and goldens remain unchanged.

`internal/testpostgres/cleanup.sql`
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
failure test opens a second pool on the shared DSN and closes it; it does not
create another migrated database. Repository/transaction tests retain their
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

Each operation's HTTP tests use one explicit Go golden-case table, without separate
generated-ID, readback or hand-decoded response tests. Persistence assertions belong
in repository tests. Each row declares
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
Explicit seed IDs and frozen business time make responses deterministic.
Buffered response goldens use `httputil.DumpResponse`: when `ContentLength` is
unknown, fill it from the recorder body's byte length before serialization.
Preserve known declared lengths and explicitly set `Connection` headers. The
resulting `Content-Length` framing treats an omitted length and an explicitly
correct one as equivalent. Apart from completing the unknown length, only HTTP
line endings are normalized. Missing or changed goldens fail the test by default.

To regenerate existing goldens intentionally, run this uncached command from the
repository root:

```sh
TADOKU_GOLDEN_SOURCE_ROOT="$PWD/services/tadoku-api/e2e/testdata" \
  bazel test //services/tadoku-api/e2e:e2e_test \
  --test_env=TADOKU_GOLDEN_SOURCE_ROOT \
  --test_arg=-update-goldens \
  --sandbox_writable_path="$PWD/services/tadoku-api/e2e/testdata" \
  --cache_test_results=no \
  --test_output=all
```

The explicit source root makes the command write checked-in fixtures instead of
runfiles copies. Only the native Tadoku API response records each existing file;
the legacy implementation then compares against it and still fails on divergence.
The update path never creates a missing golden or records a response with an
unexpected status, and it exits before dependency startup when `CI` is set. Review
every rewritten path and the complete Git diff, then explain the intentional
contract change in the PR body. Lifecycle tests stay focused on startup/shutdown,
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

Time-dependent cases give both implementations the same controlled time through
their ordinary clock dependencies. If legacy behavior reads the database clock,
introduce an explicit application-clock input rather than rewriting its SQL in a
test connector. Execute the production queries and row mapping unchanged against
real PostgreSQL; never shift or ignore fixture timestamps or response fields to
make a comparison pass. JWT verification, identity propagation, ban checks and endpoint
permissions run in both implementations. Gateway token issuance/rewriting and
unrelated infrastructure remain outside these in-process comparisons.

```sh
bazel test //services/tadoku-api/e2e:e2e_test --test_output=errors
```

The existing CI E2E and race targets include these comparisons automatically.
Add future migrated operations explicitly; do not add a general service launcher
or duplicate per-implementation fixtures.

Pool-failure, cancellation and proxy-routing checks remain separate Go tests;
they exercise dependency behavior rather than SQL-defined response cases.

Endpoint fixtures contain synthetic signed JWTs and any required Keto relationships.
Administrator scenarios seed explicit admin tuples; public scenarios use signed
guest tokens, as the gateway does. An absent relationship file grants no roles.
No endpoint test bypasses authentication, injects an administrator claim, or uses
an always-allow permission checker. Focused authentication
scenarios register `GET /test/authentication` through the same production `Router`
with a test-only success handler. The comparison handler uses legacy `VerifyJWT` and
`Identity`, without authorization or business endpoints. Both consume the same signed
HTTP requests and goldens. Test-only identity headers prove downstream context
propagation. The transport router test proves all registered application routes
inherit the shared gates. No test endpoint is added to production.

The suite serves a synthetic checked-in public JWKS locally; private keys and live
identity providers are not needed. The HTTP runner temporarily fixes
`jwt/v4.TimeFunc` at the signed fixtures' verification instant and restores it on
return. Business-clock tests may advance `timex` independently of token expiry. These
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

### User journeys

A user journey chains requests against one seeded state, so everything after
its first step comes from the API itself: write and read paths must agree
without a handwritten seed between them, business time can move between steps,
and one user's writes can be observed by another. User journeys exist to keep
the functionality users expect working, so cover every important user journey
in the application. They run only against Tadoku API; parity stays per
operation in the golden-case tables.

All user journeys live in `e2e/user_journeys_test.go`, one explicit Go table
per journey passed to `runJourney`, with fixtures under
`e2e/testdata/journeys/<Journey>/`:

```text
e2e/testdata/journeys/
  cast.json                  # cast member -> bearer token
  relationships.json         # cast Keto tuples, seeded for every journey
  <Journey>/
    setup.sql                # optional, journey-specific
    relationships.json       # optional, journey-specific
    01_<step>/
      request.http           # no Authorization header
      golden.http
    02_<step>/
      verify.sql             # one ordered query over deterministic columns
      verify.json
```

Steps are numbered by their position in the table. A request step names the
cast member that sends `request.http`; the runner injects that member's token,
and `none` sends no credentials. Its optional `others` map replays the same
request as other members before the primary request and checks their status
only. Add replays only where they add information: identities that must be
rejected on mutating steps, `banned` included, and a second user only to
observe limited visibility of a resource. Those replays must not legitimately
mutate state, so an identity that is supposed to succeed at a write gets its
own step. A verify step runs
`verify.sql`, aggregates the rows into one JSON array in query order and
compares the indented result with `verify.json`. Reserve verify steps for
effects no endpoint exposes, such as soft deletes, outbox rows or audit
entries. Verify queries end with `order by` and select only
application-supplied columns; database-defaulted ids and `now()` timestamps are
not deterministic.

The runner resets PostgreSQL and Keto once per journey: cleanup, then the
shared `journeys/setup.sql` and `journeys/relationships.json`, then the
journey's own files. There are no per-step seeds. The JWT clock stays at the
fixture instant for the whole journey while each step's `at` sets its business
instant through `timex`. The runner stops at the first failing step and, like
every scenario, fails if a legacy upstream was contacted. Unknown entries in a
journey or step directory fail the journey.

Cast members are `guest`, `user`, `user2`, `admin` and `banned`, plus the
implicit `none`. `admin` and `banned` hold their `app:tadoku` tuples through the
shared relationships file. Add a member by signing a token with the recipe
above, appending its public key when a new signing key is used, adding the
token to `cast.json` and any tuple to `relationships.json`.

Golden regeneration uses the same `-update-goldens` command and source root as
the golden-case tables and covers `verify.json` too. It never creates a missing
file, so add empty placeholders before recording a new user journey and review the
complete diff.

Asynchronous work will be covered by running a worker's single synchronous pass
as a journey step at that step's business instant; tests never start polling
loops. Migrated workers must expose that single pass as a method returning an
error and let production `Run` loop over it. The job step kind lands with the
first migrated worker.

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
