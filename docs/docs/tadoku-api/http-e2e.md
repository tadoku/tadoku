---
title: HTTP end-to-end tests
description: How Tadoku API HTTP golden cases are laid out, seeded, compared and regenerated, and how E2Es authenticate with real JWTs, PostgreSQL permission relationships and local identity accounts.
sidebar_position: 6
---

# HTTP end-to-end tests

Read this when you add or change an operation's HTTP golden cases, fixture
tokens or relationships, or anything that touches the shared identity fixtures.
The general test principles are in [Testing](./testing.md).

## HTTP E2E suite

- Each Bazel shard's `TestMain` creates its own disposable database, applies
  migrations and constructs the production HTTP router once, with real JWT verification, ban checks and
  PostgreSQL-backed permissions. There is no second bypass router or injected
  administrator identity. Handlers run in process, without HTTP listeners.
- The production router, profile-service resets, alternate scoring/dependency
  routers and real worker journeys all use the restricted `AppPool`.
  Fixture reset and verification SQL use the owner `Pool`; an owner-pool
  application router would bypass the policies the suite must exercise.
- Bazel uses native rules_go process sharding. Each process owns its database,
  application role, router, JWT verification clock and leased Valkey logical
  database; each golden case and journey owns its business clock and UUID
  stream. Serial golden cases and journeys run one after another within a
  process, each after `reset`, never between dependent requests. Cases and
  journeys that run as
  [scenarios](./scenario-fixtures.md#golden-cases-and-journeys-as-scenarios)
  own private tenants instead of a reset, and run in parallel after every
  serial test has finished.
- The router uses the default deployed API request budget of 30 seconds,
  including cost-12 password hashing. The complete E2E target uses an explicit
  `timeout = "long"` (900-second) budget so race instrumentation can cover real
  bcrypt work without reducing production-cost controls or skipping throttle
  requests.
  Provider fixtures retain their own bounded timeouts; request-deadline behavior
  is tested at the transport boundary.
- `reset` runs `internal/testpostgres/cleanup.sql`, which explicitly lists the
  mutable tables to truncate with `restart identity`. Add newly tested mutable
  tables to that one file. Do not discover tables automatically or use
  `cascade`.
- Migration-seeded static tables and `schema_migrations` are preserved. When a
  preserved table's foreign key makes PostgreSQL reject `truncate`, use
  `delete` instead and document the exception in `cleanup.sql`.
- Cases load their starting data from their own SQL and relationship files;
  fixtures are not generated in Go.
- Reset and seeding commit before requests run, so requests use normal
  application transactions and real commits. There is no outer rollback
  transaction, and tests do not change `RunInTransaction`.
- Pool-failure, cancellation and routing checks are separate Go tests; they
  exercise dependency behavior rather than response cases.
- Request-body byte limits are proved the same way, in
  `services/tadoku-api/e2e/request_body_limit_test.go`. It builds oversized and
  allowed bodies in code against existing fixture tokens, so no oversized
  request fixture is checked in.
- Lifecycle tests stay focused on startup and shutdown; keep endpoint
  assertions out of them.
- The [email gateway fixture](#email-gateway-fixture) checks cookie and bearer
  routing with real local providers; it supplements the live branch-host gate.
- Unrelated gateway infrastructure stays outside these in-process tests.

## Tenant feature-flag isolation

`services/tadoku-api/e2e/flipt_tenant_test.go` compares a complete sequence of
public decision HTTP responses under `testdata/FliptTenantIsolation/` for the
same account in canonical and test tenants. The test sets up a test grant, a
canonical revoke and a test revoke through the real Flipt management adapter
against the provider fixture; the application has no management endpoint.
It freezes the normal fixture clock, uses signed tenant JWTs, real permission
relationships and the non-owner application pool, and preserves the opposite
tenant's decisions.
The Flipt fixture is an explicit provider simulation: it stores independent
members and revisions per environment/namespace and uses the production
`Targets` mapping. `SeedTenant` provisions another parsed test tenant for a
journey after reset. Separate Flipt SDK integration and operator verification
exercise the provider HTTP boundary.

## HTTP golden cases

For every operation, each golden case runs SQL and relationship setup plus a
signed HTTP request through the production router and compares the complete
response with its golden.

### Case table

- Each operation's HTTP tests use one explicit Go golden-case table. Do not add
  separate generated-ID, persistence-readback or hand-decoded response tests
  beside it; persistence assertions belong in repository tests.
- Each row declares `description []string` and `want` as an HTTP status
  constant. `APITestName(operation, want, description...)` names both the
  subtest and its fixture directory, so there is no separate fixture-name field
  to keep in sync.
- Add a case by adding a descriptive table row and its request and golden files.
  Do not discover cases from directories.
- Cover each operation's response mapping, input handling, business rules and
  relevant boundaries.

### Fixture layout

Case files are language-neutral and reviewable:

```text
e2e/testdata/<operation>/
  setup.sql              # optional operation-level seed
  setup.<role>.sql       # optional seed for a scenario role's tenant
  relationships.json     # optional operation-level permission relationships
  <status>_<description>/
    setup.sql            # optional
    finalize.sql         # optional final state after the account cast
    relationships.json   # optional permission relationship array
    request.http
    golden.http
```

- Every case has `request.http` and `golden.http`. Request bodies are JSON
  only; do not add non-JSON request fixtures.
- `golden.http` contains the request label and the complete expected response.

### Seed resolution

Seed files (`setup.sql`, `finalize.sql`, `relationships.json`) resolve per file name, one level
deep:

- A non-empty case file is the case's only seed for that name. It replaces the
  operation file rather than extending it.
- A zero-byte case file opts the case out: it loads no seed for that name, even
  when the operation has one. Keep these files; they are intentional.
- A missing case file inherits `testdata/<operation>/<name>`. There is no
  fallback beyond the operation directory.

Put seed data shared by most cases in the operation file, add a case file only
when a case needs different data, and add a zero-byte case file when a case must
start without the operation's data. Shared PostgreSQL cleanup, including
`permission_relationships`, always runs, even without setup. A resolved seed
file that does not exist is skipped; other read, SQL, decode and provider errors
fail the scenario.

### Relationship files

`relationships.json` is an array of permission relationships; unknown fields
fail decoding:

```json
[
  {
    "tenant": "tadoku/prod",
    "object": "app:root",
    "relation": "admins",
    "subject": "user:11111111-1111-4111-8111-111111111111"
  }
]
```

The serial harness runs the resolved `setup.sql`, `testdata/accounts.sql` and the
resolved `finalize.sql`, then replaces every stored relationship with the
[production baseline](./scenario-fixtures.md#production-baseline) and the
resolved file's rows. A row's tenant must be registered by then: for a test
tenant, insert its registry row in the case's `setup.sql` with
`on conflict (key) do nothing`. A scenario case seeds the same files into its
own tenants instead, giving each row to the tenant that stands in for the row's
`tenant`. Its SQL for a test tenant goes in the operation's `setup.<role>.sql`
instead of `setup.sql`; see
[Golden cases and journeys as scenarios](./scenario-fixtures.md#golden-cases-and-journeys-as-scenarios).

### Comparison

- Tests parse `request.http` with `net/http`, run the production handler at the
  operation's minimum required access level, check the status against `want`,
  and compare the entire response, including status, headers and body, with the
  golden.
- Compare directly; do not decode the response into the generated types the
  handler uses.
- Explicit seed IDs, a per-case UUID source and frozen business time make
  responses deterministic; see [Time in tests](#time-in-tests). Never rewrite
  SQL inside a test connector or rewrite response fields. The one exception is
  a scenario case's
  [tenant alias](./scenario-fixtures.md#golden-cases-and-journeys-as-scenarios),
  which names each generated test tenant by its role.
- Buffered response goldens are serialized with `httputil.DumpResponse`. When
  `ContentLength` is unknown, fill it from the recorder body's byte length
  before serialization. Preserve known declared lengths and explicitly set
  `Connection` headers.
- This framing treats an omitted `Content-Length` and an explicitly correct one
  as equivalent. Apart from completing the unknown length, only HTTP line
  endings are normalized.
- Missing or changed goldens fail the test.

### Regenerating goldens

Tests never rewrite goldens by default. To regenerate existing goldens
intentionally, run this uncached command from the repository root:

```sh
TADOKU_GOLDEN_SOURCE_ROOT="$PWD/services/tadoku-api/e2e/testdata" \
  bazel test //services/tadoku-api/e2e:e2e_test \
  --test_env=TADOKU_GOLDEN_SOURCE_ROOT \
  --test_arg=-update-goldens \
  --sandbox_writable_path="$PWD/services/tadoku-api/e2e/testdata" \
  --cache_test_results=no \
  --test_output=all
```

- The explicit source root makes the command write the checked-in fixtures
  instead of runfiles copies. The actual response is recorded into each
  existing golden file.
- `-update-goldens` rewrites only golden files. It never creates, removes or
  changes seed files, never creates a missing golden and never records a
  response with an unexpected status. It exits before starting dependencies
  when `CI` is set.
- Review every rewritten path and the complete Git diff, then explain the
  intentional contract change in the PR body.

## Authentication in E2Es

- Endpoint E2Es run through the production JWT and ban middleware and the real
  permission checker on the restricted `AppPool`. Never substitute passthrough
  middleware, inject an identity or administrator claim, or install an
  always-allow permission checker.
- Put synthetic signed tokens in `request.http` and required role
  relationships in the case's `relationships.json`. Administrator scenarios
  seed explicit `admins` relationships; an absent relationship file grants no
  roles. Public scenarios use signed `guest` tokens, as the gateway does.
- Keep the full invalid-credential, role, ban and permission-store matrix at
  the shared middleware boundary instead of repeating it for every endpoint.
  Simulate a permission-store outage with `withPermissionsPool` in
  `services/tadoku-api/e2e/router_options_test.go`, which gives the checker a
  closed pool. Role reads and ban writes use the router's transaction pool;
  make their writes fail inside the database instead.
- Test authentication goldens at the middleware boundary, not on a business
  endpoint. They register `GET /test/authentication` on the same production
  router with a test-only success handler, and test-only identity headers prove
  downstream identity propagation. `X-Test-Tenant` proves the parsed signed
  tenant also reached the handler. `AuthenticationBranch` goldens construct the
  same production router with an exact `e2e/branch-golden` deployment and prove
  the own-tenant success and other-tenant `421` boundary. The successful case
  seeds no relationships: admission needs no grant on any tenant. Ban-policy
  scenarios register `GET /test/banned` the same way and cover guests,
  ungranted users and bans on both canonical and test tenants. Permission-store
  fail-open behavior and deadlines are tested at that boundary. No test endpoint
  is added to production.
- The transport router test proves that every registered application route
  inherits the shared gates.
- The suite serves a synthetic, checked-in public JWKS
  (`e2e/testdata/authentication.jwks.json`) locally, plus the public key of
  the per-process [scenario fixture](./scenario-fixtures.md) signing key. No checked-in
  private keys or live identity providers are needed.

### Time in tests

- `TestMain` fixes `jwt.TimeFunc` at the signed fixtures' verification
  instant, `2026-09-12T12:00:00Z` (`1789214400`), before any test runs. Tests
  do not write it, except one whose tokens are signed at another time, such as
  the Oathkeeper-minted tokens of the email gateway fixture; it restores the
  fixture instant in cleanup.
- The golden runner gives each case its own `timex.Clock` at the case's
  business time and its own UUID stream, and attaches both to the request
  context with `timex.WithClock` and `uuidx.WithSource`. The default stream
  starts with sixteen `0x11` bytes; a case table may set a seed instead.
  Business time moves independently of token expiry, and production, which
  attaches neither, keeps real time and random IDs.
- A test outside the runners that needs fixed business time either attaches a
  clock to its requests' contexts or wraps them in `timex.TheWorld`
  (`atFixtureInstant` for the fixture instant), never the whole suite in
  `TestMain`. `TheWorld` applies only to contexts without a clock; a test may
  use separate, non-nested scopes.
- To let business time pass while a request is in flight, call `Advance` on
  its clock, or `timex.Advance` inside a `TheWorld` scope. Release whatever the
  request waits on and wait for it to finish before the scope returns, on
  failure paths too.
- Tests that use `timex.TheWorld` or `uuid.SetRand` or override
  `jwt.TimeFunc`, and their parent tests, must not call `t.Parallel`.

Latency-padding checks use `testing/synctest` with controlled processing delays
below and above the floor. Keep the real router, restricted database role,
admission store, rendering and transactional writes. Virtual time checks timer
policy, not production I/O latency; retain real-clock minimum-delay and
production-cost password checks. Short wall-clock sample means cannot establish
constant-time behavior under host stalls. Minimum padding does not conceal work
that exceeds its floor.

Create and close bubble clients and pools inside that bubble. Use synchronous
Valkey (`DisableAutoPipelining`): background socket reads prevent virtual time
from advancing. Finish the local JWKS exchange and close its server before
measuring timers; finish the scenario before key refresh is due. Never reuse HTTP
keep-alive connections across bubbles.

### Signing a new fixture token

`e2e/testdata/TenantIsolation/` contains synthetic signed canonical and
`e2e/isolation-0123abcd` tokens for the same seeded account subject. Its test
registry entry exists only in the suite's disposable database. A case that
needs an administrator or a ban on the test tenant adds a relationship for that
tenant. Changing a tenant never grants an administrator identity or bypasses
the JWT and ban gates. Seed rows
name their tenant explicitly, and owner teardown removes only the fixture's
test entry. Do not insert test tenants in the shared development base or
production database.

Reuse an existing signed request when only its relationships change.
When a scenario needs different JWT claims, generate a new synthetic key and
token locally with Node's built-in cryptography, then append the printed public
JWK to the existing `keys` array without removing any checked-in keys. Run this
from the repository root with Node.js and `jq` installed:

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
  tenant: "tadoku/prod",
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

- Review the `.new` JWKS before replacing the fixture, and paste the printed
  token into the new `request.http`.
- The private key exists only inside that Node process. Editing claims in a
  token by hand invalidates its signature, so rerun the recipe instead.
- Never use production signing keys or tokens in fixtures.

## Local account fixtures

`e2e/testdata/accounts.sql` seeds the fixed five-user cast in `tadoku/prod` and
`e2e/alpha-0000000a`, matching the signed cast's tenants and subjects. Its
synthetic bcrypt cost-4 hash keeps fixture login cheap; production registration
and the dedicated password-policy workflows retain cost 12. Sessions come from
real API login.

Ordinary native-identity fixtures also use cost-4 hashes. The login timing and
throttle control explicitly restores cost-12 reader hashes and checks their
parsed bcrypt cost before exercising known and unknown accounts.

The IP-throttle fixture seeds wrong-password cost-4 accounts for repeated login
attempts, while keeping an unknown-account cost-12 dummy comparison in each
tenant and the cost-12 successful fail-open login. All throttle requests still
use the real router, PostgreSQL and Valkey. Registration, credential-storage,
password-policy and real-clock timing controls keep production cost 12.

Ordinary golden and journey resets load case SQL before the account seed.
Explicit user/account primary-key conflicts preserve existing rows, including
case-specific names, timestamps, inactive accounts and deleted or locked users.
Repeating the seed is safe; other constraint failures fail setup. Optional
`finalize.sql` uses the same case/operation resolution and runs after the cast
in the reset transaction. Put final absence mutations there so the default cast
cannot resurrect a user that a scenario deliberately removes. Focused scenarios
that define a different account population reset PostgreSQL with their own
explicit seed.

Profile, account-age and role-update requests use the production identity Reader
and restricted application pool. The signed cast workflow verifies each unique
tenant/subject pair through the real JWT router, rejects a foreign session under
the target tenant's cookie name, and checks unavailable states and reset. Roles
remain real permission relationships.

## Email gateway fixture

[HTTP email gateway fixture](./http-email-gateway.md) covers cookie and bearer
routing through checksum-pinned Oathkeeper, including local checks of rendered
production gateway rules.
