---
title: HTTP end-to-end tests
description: How Tadoku API HTTP golden cases are laid out, seeded, compared and regenerated, and how E2Es authenticate with real JWTs, Keto and the Kratos fixture.
sidebar_position: 6
---

# HTTP end-to-end tests

Read this when you add or change an operation's HTTP golden cases, fixture
tokens or relationships, or anything that touches the shared Kratos fixture.
The general test principles are in [Testing](./testing.md).

## HTTP E2E suite

- `TestMain` creates one disposable database, applies migrations and constructs
  the production HTTP router once, with real JWT verification, ban checks and
  Keto-backed permissions. There is no second bypass router or injected
  administrator identity. Handlers run in process, without HTTP listeners.
- Scenarios run sequentially. Call `reset` before each independent scenario,
  never between dependent requests.
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
- Lifecycle tests stay focused on startup and shutdown; keep endpoint
  assertions out of them.
- Gateway token issuance and unrelated infrastructure stay outside these
  in-process tests.

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
  relationships.json     # optional operation-level Keto tuples
  <status>_<description>/
    setup.sql            # optional
    relationships.json   # optional Keto relation-tuple array
    request.http
    golden.http
```

- Every case has `request.http` and `golden.http`. Request bodies are JSON
  only; do not add non-JSON request fixtures.
- `golden.http` contains the request label and the complete expected response.

### Seed resolution

Seed files (`setup.sql`, `relationships.json`) resolve per file name, one level
deep:

- A non-empty case file is the case's only seed for that name. It replaces the
  operation file rather than extending it.
- A zero-byte case file opts the case out: it loads no seed for that name, even
  when the operation has one. Keep these files; they are intentional.
- A missing case file inherits `testdata/<operation>/<name>`. There is no
  fallback beyond the operation directory.

Put seed data shared by most cases in the operation file, add a case file only
when a case needs different data, and add a zero-byte case file when a case
must start without the operation's data. Shared PostgreSQL and Keto cleanup
always runs, even without setup. A resolved seed file that does not exist is
skipped; other read, SQL, decode and provider errors fail the scenario.

### Comparison

- Tests parse `request.http` with `net/http`, run the production handler at the
  operation's minimum required access level, check the status against `want`,
  and compare the entire response, including status, headers and body, with the
  golden.
- Compare directly; do not decode the response into the generated types the
  handler uses.
- Explicit seed IDs and frozen business time make responses deterministic.
  Supply business time through `timex`. Never rewrite SQL inside a test
  connector or rewrite response fields.
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

- Endpoint E2Es run through the production JWT and ban middleware, real Keto and
  real PostgreSQL. Never substitute passthrough middleware, inject an identity
  or administrator claim, or install an always-allow permission checker.
- Put synthetic signed tokens in `request.http` and required role tuples in the
  case's `relationships.json`. Administrator scenarios seed explicit admin
  tuples; an absent relationship file grants no roles. Public scenarios use
  signed `guest` tokens, as the gateway does.
- Keep the full invalid-credential, role, ban and provider matrix at the shared
  middleware boundary instead of repeating it for every endpoint.
  Provider-protocol tests belong with the provider adapter; do not embed ad hoc
  Keto emulators in endpoint fixtures.
- Test authentication goldens at the middleware boundary, not on a business
  endpoint. They register `GET /test/authentication` on the same production
  router with a test-only success handler, and test-only identity headers prove
  downstream identity propagation. Ban-policy scenarios register
  `GET /test/banned` the same way, using the same real Keto fixture. Provider
  fail-open behavior and deadlines are tested at that boundary. No test
  endpoint is added to production.
- The transport router test proves that every registered application route
  inherits the shared gates.
- The suite serves a synthetic, checked-in public JWKS
  (`e2e/testdata/authentication.jwks.json`) locally; no private keys or live
  identity providers are needed.

### Time in tests

- The HTTP runner fixes `jwt.TimeFunc` at the signed fixtures' verification
  instant, `2026-09-12T12:00:00Z` (`1789214400`), and restores it on return. JWT
  clock overrides belong only in such scoped, sequential tests and must always
  be restored.
- Control business time with `timex.TheWorld` inside the tests that need it,
  never around the whole suite in `TestMain`. A test may use separate,
  non-nested scopes for different times, and business time can move
  independently of token expiry. No test-only request-context wrapper is
  needed.
- Tests that use `timex.TheWorld` or override the JWT clock, and their parent
  tests, must not call `t.Parallel`.

### Signing a new fixture token

Reuse an existing signed request when only its Keto relationship tuples change.
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

## Kratos fixture

- HTTP E2Es share one pinned Kratos v26.2.0 SQLite-enabled Linux x86-64 process,
  owned by `internal/testkratos`. It runs its migrations against a private
  SQLite database on RAM-backed tmpfs at `/dev/shm` (not Kratos's
  process-private `dsn: memory`). It uses the shared
  `infra/dev/ory/identity.default.schema.json` and private Unix sockets, and
  supplies a bounded HTTP client to the raw SDK constructor.
- `TestMain` seeds it once from `e2e/testdata/kratos.sql` by direct SQL, with
  fixed identities that match the signed fixture subjects, traits and
  timestamps. The seed depends on the pinned provider schema and creates
  identity records only: no passwords, credential identifiers or sessions.
  Tests of those operations arrange the provider state explicitly.
- Guest and unauthenticated requests have no Kratos identity; roles and bans
  stay in Keto.
- Kratos is excluded from the ordinary per-case reset. Tests that are not marked
  for a Kratos reset must leave its state unchanged.
- A test that may mutate Kratos calls `resetKratosAfter(t, api)` on the
  enclosing test before any mutation. The helper registers `t.Cleanup` to
  restore the full seed after the test or whole journey, even after `t.Fatal`
  or cancellation. It uses a fresh bounded context because `t.Context()` is
  cancelled before cleanup. Never mark individual journey steps.
- A reset stops the process, recreates and migrates the database and restores
  the seed. Existing SDK clients keep working through the same socket path. A
  failed reset fails the test, and later cases reject the unusable fixture.
- Startup and reset are bounded. Failure or shutdown closes connections, reaps
  the child process and removes the owned database, journals and sockets.
- Tests sharing the provider stay sequential. Mutations are not detected
  automatically and response UUIDs are not normalized; goldens remain exact
  comparisons.
- Keep provider-schema SQL and process ownership in test-only helpers.
