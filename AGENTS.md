# Development Workflow

**Make impossible states impossible to represent.** Model data structures so only valid combinations can be constructed. Use distinct types for different operations and explicit variants with required payloads for mutually exclusive states, rather than independent flags and optional fields that permit contradictory combinations. Parse and validate external input at the boundary, then pass the resulting domain values through the application.

## CMS content

**Page content managed by the CMS must only be changed in the admin CMS.** Public pages such as Contact and About store their HTML in the CMS (`pages` / `pages_content`, namespace `tadoku`, slug such as `contact`).

**Do not ship CMS copy as a code change.** Do not open a pull request, hardcode the copy in the frontend, or write a database migration or other SQL rewrite that updates CMS page HTML or equivalent content rows.

**Stop as soon as the request is CMS-managed page content.** When the change is page copy rather than application UI or code, stop the coding work immediately. Do not open a pull request, write a migration, or invent a code workaround. Report that the edit belongs in the admin CMS and wait for that edit.

## Frontend

**Always use `pnpm`, not `npm`.**

### Tadoku Paper

For Paper applications, use `paper-ui` rather than the legacy `ui` package. Follow the [Paper composition guide](docs/wip/tadoku-paper/composition.md): controls live in `paper-ui`, product patterns describe reusable Tadoku task sections, and application screens own routes, data, permissions, forms, and mutations. Start pattern code beside its screen; a catalogue pattern does not automatically become a shared component. Keep application services and routers out of `paper-ui`.

**Use the `ui` package design system** - never write custom button/form styles. Use:
- Buttons: `className="btn"` with variants `primary`, `secondary`, `danger`, `ghost`
- Forms: Import `Input`, `Select`, `Checkbox`, etc. from `ui/components/Form`
- Components: Import from `ui` package (Modal, Flash, Navbar, etc.)

**Always use `react-hook-form` for form handling** — never use plain `useState` for form fields. Use:
- `useForm()` + `<FormProvider>` to set up form context
- `<Input>`, `<Select>`, `<TextArea>` from the `ui` package (they use `useFormContext()` internally)
- `useController()` for custom/non-standard form components (e.g. CodeEditor)
- `methods.handleSubmit()` for form submission with built-in validation
- `methods.watch()` for reactive field values (e.g. live previews)
- `methods.reset()` to populate forms with existing data

```sh
# 1. Make changes

# 2. Typecheck (fast)
cd frontend && pnpm --filter webv2 exec tsc --noEmit

# 3. Lint before committing
cd frontend && pnpm --filter webv2 lint

# 4. Before creating PR
cd frontend && pnpm build
```

## Backend

**Always use `bazel`, not `go`.**

### Database migrations

**Ship every schema or data migration as a standalone change.** A migration must land on `main` in its own commit and pull request, separate from application code, and must be deployed independently before any code that depends on it. Do not combine migrations and runtime behavior in the same commit, pull request, or deployment, including within a stacked PR series.

Migration PRs must remain compatible with the application version currently deployed. Additive migrations come before dependent code; destructive or cleanup migrations come in a later standalone PR only after the old code is no longer deployed. After a migration is merged and deployed, base dependent work on the updated `main` branch.

**Do not create application-defined database functions, stored procedures, or triggers.** Keep business behavior in the application and domain layers. Use declarative database features such as `not null`, `check`, unique and foreign-key constraints, and indexes for data integrity and performance.

**Test meaningful behavior and risk, not every change.** Add coverage where a regression would matter, especially for production authentication and HTTP contracts, persistence and data integrity, and nontrivial domain rules. There is no automatic test-per-method or test-per-change requirement. No new test is a valid choice for trivial, documentation-only, generated-output-only, and test-removal changes.

**Do not add low-value tests.** Skip tautological assertions that mirror the implementation, trivial constant or schema equality checks, pass-through or error-wrapper assertions without distinct behavior, and coverage already provided at a shared boundary. Do not export internals, broaden visibility, or add dependencies or abstractions solely to enable such assertions. When a test is explicitly removed, do not recreate equivalent coverage elsewhere unless requested.

**Prefer repository tests plus HTTP E2Es over isolated feature-service tests.** Real-database repository tests are highly recommended for query behavior, row mapping, constraints and persistence. Exercise feature-service orchestration through HTTP E2Es; do not add database-backed feature-service tests. Use unit tests for pure parameter validation and domain rules. Cover shared dependency failures at their boundaries once, not again for every endpoint using those dependencies.

**Use Go's standard `testing` functionality for new or rewritten backend tests.** Use ordinary comparisons, `t.Fatalf` for failed prerequisites, `t.Errorf` for independent checks, and `t.Cleanup` for resource cleanup. Compare errors with `errors.Is`/`errors.As`. Do not add Testify, another assertion framework, or a homegrown assertion DSL. Existing tests do not need a bulk rewrite; convert them when their relevant slice is migrated or in a separately scoped mechanical change.

### Tadoku API

**Maintain Tadoku API import boundaries in Bazel.** Read the [architecture guide](services/tadoku-api/README.md#import-policies) before adding an import, package or feature. Bazel target `visibility` and the [package groups](services/tadoku-api/BUILD.bazel) define the local boundaries. Tests follow their package's layer policy; startup and E2E packages are assembly boundaries.

**Keep new packages inside the architecture.** Give each new or moved `go_library` the narrowest matching visibility. A feature's generated sqlc package must be visible only to that feature. When an import boundary intentionally changes, update the relevant target visibility or package group and the architecture guide in the same change; explain the new dependency direction in the PR. Do not widen visibility to `public` merely to make a build pass.

**Verify import-boundary changes** with `bazel run //:gazelle -- -mode=diff` and `bazel build //services/tadoku-api/...`. The legacy depolicy CI check is temporary and does not define the preferred package structure.

**Document durable conventions for the whole application.** State architecture, compatibility and testing rules for all operations. Do not document individual endpoint implementations, enumerate their coverage, or add endpoint-status sections to general documentation. Use generic examples. Keep deferred cleanup notes as checkboxes in `services/tadoku-api/MIGRATION_LOG.md`.

**Keep each runtime slice small and reviewable.** Implement the requested operation without unrelated service identities, audience checks, or authorization features. Keep authentication and shared ban enforcement in separately reviewed changes.

**Keep endpoint caller authorization in the application layer.** Application operations own full caller-access checks such as authenticated-user and administrator requirements. A feature service may inspect authorization facts only when they expand behavior inside an operation the application has already authorized. Compose independent features in the application layer; features must not import or call sibling features. The shared HTTP JWT and ban boundary remains separate and applies before application operations.

**Keep feature-owned write workflows in feature services.** The application authorizes callers, coordinates locks and transactions across features, and composes their results. A feature service validates or normalizes its own inputs and sequences its repository operations, including related rows and outbox writes. Do not add repository pass-through methods to a feature service just so an application operation can assemble that feature's write. Put fixed business values shared across features in the native domain catalog instead of embedding them in a write operation.

**Run endpoint E2Es through real authentication and authorization.** Use the production JWT and ban middleware, real Keto, and real PostgreSQL. Put synthetic signed tokens in `request.http` and required role tuples in each case's `relationships.json`; anonymous gateway traffic uses a signed guest token. Never substitute passthrough middleware, inject an identity/admin claim, or install an always-allow checker in endpoint E2Es. Keep the full invalid-credential/role/ban/provider matrix at the shared boundary instead of repeating it for every endpoint. Provider-protocol tests belong with the provider adapter; do not embed ad hoc Keto emulators in endpoint fixtures.

**Keep generated request decoding.** Do not add endpoint-specific middleware, replace request bodies or otherwise bypass the generated decoder to alter empty-body behavior.

**Require authentication wiring in production construction.** Business route registration must require authentication middleware; startup constructs it from required JWKS configuration with a bounded initial fetch. Keep probes at their existing access boundaries. Test authentication goldens at the middleware boundary with a test-only success handler and downstream identity observation, not a business endpoint. Use synthetic signed requests and public JWKS fixtures. JWT clock overrides belong only in scoped, sequential tests and must always be restored.

**Prove endpoint behavior with HTTP goldens.** For every operation, run SQL/relationship setup and a signed HTTP request through the production router, then compare the complete response with its golden. Reset/reseed before each case. Supply deterministic business time through normal clock dependencies; never rewrite SQL inside a test connector. Gateway token issuance and unrelated infrastructure remain outside these in-process tests.

**Keep HTTP cases in language-neutral, reviewable files, with an explicit Go test table.** Each case declares `description []string` and `want` as an HTTP status constant. Use `APITestName(operation, want, description...)` for both the subtest name and its fixture path under `testdata/<operation>/<status>_<description>/`. Every case has `request.http` and `golden.http`; a missing case seed file inherits the operation-level file of the same name, a non-empty case file replaces it, and a zero-byte case file opts out of it. Shared cleanup still runs when setup is absent; skip only missing seed files, not other read or SQL errors. Do not maintain a separate fixture-name field or discover cases from directories. Load any case SQL, check the response status against `want`, and compare the complete HTTP response with its golden. Compare status, headers and body directly, without decoding into the same generated types used by the handler. Seed deterministic IDs and freeze business time. For buffered HTTP goldens, fill an unknown response ContentLength from the recorder body before serializing with httputil.DumpResponse; preserve known declared lengths and explicitly set Connection headers. This framing convention treats an omitted Content-Length and an explicitly correct one as equivalent. Apart from completing the unknown length, normalize only HTTP line endings. Tests never rewrite goldens by default. Intentional local regeneration requires `-update-goldens`, an explicit source-fixture root, and review of the resulting diff. Update mode must reject CI, unexpected response statuses and missing golden files. Explain reviewed contract changes in the PR body. Keep endpoint assertions out of process-lifecycle tests.

**Share the HTTP E2E router, not scenario state.** Initialize the disposable database, migrations and production router once in `TestMain`. Run HTTP scenarios sequentially, resetting before each scenario and loading starting data from SQL files. Reuse `internal/testpostgres/cleanup.sql`, which explicitly lists mutable tables to truncate with `restart identity`; no table discovery or `cascade`. Preserve migration-seeded static tables and migration bookkeeping. Add newly tested mutable tables to that one file. Requests within a scenario use normal application transactions and real commits, not an outer test transaction. Database helpers return errors and provide explicit cleanup without depending on `testing.TB`. Keep pool-closing failure tests isolated; independently isolated repository/transaction tests may remain parallel.

**Seed Kratos once; reset only after tests that mutate it.** HTTP E2Es share a pinned Kratos process and RAM-backed SQLite database, with fixed identities matching the signed fixture subjects. Keep Kratos out of ordinary per-case resets. Call `resetKratosAfter(t, api)` before any mutation on the enclosing test. Cleanup restores the full seed after the test, including failure/cancellation, using a fresh bounded context. Never reset between journey steps. Shared-provider tests stay sequential; a failed reset makes the fixture unusable. Keep provider-schema SQL and process ownership in test-only helpers, and retain exact golden comparisons.

**Write for readability.** Separate setup, execution, error handling and response mapping with whitespace. Within a function, put a blank line between coherent phases such as authorization, input extraction, persistence and response mapping. Put unrelated struct fields and composite-literal entries on separate lines. Split application and transport operations into files by functionality, and use descriptive operation names. Keep constructors and resource lifecycle code visibly separate from endpoint behavior.

**Put shared business concepts in the native domain layer.** Use `services/tadoku-api/domain/<concept>` for business values and pure rules used by multiple features. Application operations and features may import these packages. Shared domain packages must not depend on application, feature, transport, storage, generated, infrastructure or `internal` packages; keep services, repositories and provider APIs out of them. Keep feature-specific types, errors and validation in that feature's `domain.go`. Technical support concerns stay under `internal`. Follow the import-boundary maintenance rule above when changing layer boundaries; features still must not import or call sibling features.

**Group feature packages by responsibility.** Use `<feature>_service.go`, `<feature>_service_test.go`, `<feature>_repository.go`, `<feature>_repository_test.go` and `domain.go`, rather than separate feature files for each operation. Keep domain types, errors and shared validation in `domain.go`; service and repository structs and constructors stay with their implementations. Declare each feature's domain errors together in one `var` block and reuse them from services and repositories. Within feature packages, database tests must exercise repositories directly and live only in the repository test file. Keep service and validation tests database-free; exercise service orchestration through HTTP E2Es.

**Return specific validation errors.** Prefer one validation condition per branch with a direct call to the shared invalid-input error constructor and a message that identifies the invalid field or rule. Do not create per-validation sentinel errors or error types. Combine conditions only when they intentionally represent the same validation message.

**Name validation and normalization operations precisely.** Use `Validate` for checks that leave request data unchanged and `Normalize` for canonicalization performed on local service or persistence input. Do not use vague names such as `Prepare` for input validation or normalization, and do not make validation functions return a rewritten request.

**Keep each operation's HTTP E2Es in one golden-case table.** Do not add separate generated-ID, persistence-readback or hand-decoded response tests beside it. Put persistence assertions in repository tests. Request-body fixtures use JSON only.

**Cover every important user journey.** User journeys in `e2e/user_journeys_test.go`, with fixtures under `e2e/testdata/journeys/`, chain requests against one reset so every state after the first step comes from the API: write-then-read agreement, business time moving between steps, and one user observing another's writes. They exist to ensure the important functionality users expect keeps working, so every important user journey in the application belongs in that single file. User journeys run only against Tadoku API, never reset or seed between steps, and name their cast member per step in the Go table instead of embedding tokens in `request.http`. Replay a request as other cast members only where it adds information: identities that must be rejected on mutating steps, `banned` included, and a second user only to observe limited visibility of a resource. Any identity that legitimately writes gets its own step. Reserve verify steps for effects no endpoint exposes, such as soft deletes, outbox rows and audit entries; API-observable persistence still belongs in the next request or a repository test.

**SQL style: always use lowercase keywords** (select, create table, not SELECT, CREATE TABLE)

### sqlc code generation

**Always run sqlc code generation after changing a SQL query.** The checked-in generated Go files must exactly match the query sources.

**Never manually edit sqlc-generated files, and never delete, revert, or selectively omit changes produced by sqlc code generation.** Commit the complete generated diff, even when code generation reveals previously stale output. If generated changes are unexpected, investigate the query inputs and pinned sqlc version, then rerun code generation; do not discard the generated changes.

Run the repository generator from the repository root. It downloads and runs the sqlc versions pinned in each active package's `generate.go`, so it does not require `go` to be installed or available on `PATH`:

```sh
./scripts/generate-sqlc.sh
```

CI runs the same script on every pull request and fails if code generation changes the working tree. Before pushing, commit the complete generated output so this check stays clean.

**Prefer concrete dependencies in native Tadoku API** — pass concrete application-owned collaborators through constructors. Introduce a narrow consumer-owned interface only for a genuine provider or layer boundary with multiple implementations; do not add one solely to inject mocks. Exercise concrete providers through their boundary tests and HTTP E2Es.

**Use "Repository" for persistent source-of-truth data, "Store" for everything else** — `Repository` interfaces access the primary database (Postgres) where authoritative data lives. `Store` interfaces access auxiliary storage (e.g. Valkey/Redis) for caches, derived data, pub/sub, coordination state, or any non-authoritative data. Implementations live under `storage/postgres/` and `storage/valkey/` respectively.

**Keep each repository method to one SQL statement.** A coherent join or CTE still counts as one statement and is appropriate when the data needs one database snapshot. Compose independent repository reads and writes in the feature service or application layer, using an application-owned transaction when the operation must commit atomically.

**Never call `time.Now()` directly** — always inject `commondomain.Clock` and use `clock.Now()`. This applies to domain services, repository methods, and background workers. The clock is created in `main.go` and threaded through constructors. This makes time-dependent code testable via `mockClock`.

For new `tadoku-api` code, use `internal/timex.Now()` for business time instead of an injected clock. Its implementation and wall-clock tests may read `time.Now()` directly; real timers and deadlines remain independent. Scope `timex.TheWorld` inside tests that need controlled business time, never around the whole suite in `TestMain`. A test may use separate, non-nested scopes for different times. Tests using `timex.TheWorld`, including their parent tests, must not use `t.Parallel`.

**Domain must not import storage packages** — the `domain` package must never import from `storage/postgres`, `storage/valkey`, or any other storage layer. Define domain types and interfaces in the domain package; the storage layer implements them. Repository methods should convert between sqlc types and domain types internally.

**Use unexported fields for domain-enriched request data** — when a request struct has fields set exclusively by the domain layer (e.g. `UserID` from session, `Year` from clock), make those fields unexported (lowercase) and add getter methods. This prevents the HTTP layer from setting them while still allowing the repository layer to read them. The domain layer (same package) writes directly (`req.userID = ...`), other packages read via getters (`req.UserID()`). Fields that the HTTP layer legitimately sets (e.g. `ContestID`, `LanguageCodes`) remain exported.

```sh
# 1. Make changes

# 2. Compile (fast)
bazel build //services/...

# 3. Run tests
bazel test //services/... # everything
bazel test //services/tadoku-api/e2e:e2e_test # one test target
bazel test //services/tadoku-api/e2e:e2e_test --test_filter=TestAuthentication # specific function

# 4. Format before committing
gofmt -w services/

# 5. Regenerate BUILD.bazel files (after adding/removing Go files or changing deps/imports)
# CI fails if these are stale (it runs `bazel run //:gazelle -- -mode=diff`)
bazel run //:gazelle

# 6. Regenerate sqlc code after modifying SQL queries
./scripts/generate-sqlc.sh

# 7. Regenerate Tadoku API OpenAPI code (after modifying its canonical spec)
./scripts/generate-openapi.sh

# 8. Before creating PR
bazel build //services/... && bazel test //services/...
```

## Dev Environment

For verifying application changes, read the checked-in
[verify-tadoku skill](.agents/skills/verify-tadoku/SKILL.md), then only the relevant
[feature-map sections](.agents/skills/verify-tadoku/references/features/README.md).
This works without a globally installed skill: open those files directly if your
agent does not discover `.agents/skills`. Update the relevant map in the same PR
when navigation, prerequisites or observable behavior changes. The map is a
verification aid, not a second deployment catalog; Bazel owns service discovery.

Use DevCLI for frontend/native API development; `.dev/README.md` owns its workflow.
The real, non-secret Homelab configuration is committed in `.dev/config.yaml` and
`k8s/dev/base/`. This is development-only GitOps, not the production deployment.
Keep credentials, private keys and kube access details outside
Git; manifests may reference existing Secrets and include the public Lab CA.
Use an ignored local file with `dev <command> --config <path>` for overrides.
Follow `k8s/dev/base/README.md` for activation, automatic
migrations, credential bootstrap and explicitly approved old-stack cleanup.

Dev Postgres is provisioned with the Zalando `postgresql` custom resource. Do not add or reintroduce hand-rolled Postgres Deployments or Helm releases for the dev stack.

Use `make dev-seed` (`scripts/dev/seed-db.sh`) to rerun idempotent seed data. `make dev-reset` is a fail-closed guard. Use `dev down` for overlays; any database deletion requires an explicitly approved, scoped runbook.

`infra/dev/ory/` retains the Kratos schema and Keto namespace fixtures used by Bazel backend tests; these are not obsolete deployment files. Preserve the shared SQL fixtures in `scripts/dev/seed/` for base and branch seeding.

## Commit Guidelines

**Commit in atomic diffs** — each commit should represent one logical change. Don't bundle unrelated changes into a single commit.

**For larger refactors spanning many files**, commit in chunks that make sense — e.g. one commit per page, per service, per domain area, etc.

## Bug Reports

When a bug is reported, follow this process:

1. **Reproduce the bug first** — Use the smallest useful existing test, build, lint check, or manual scenario. Add a regression test when it protects meaningful behavior. If you claim a test fails before the fix, verify that it compiles and fails for the intended behavioral reason.
2. **Use subagents to fix** — Have subagents implement the fix and prove it against the reproduction and affected checks.
