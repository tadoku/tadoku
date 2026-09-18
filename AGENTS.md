# Development Workflow

## Frontend

**Always use `pnpm`, not `npm`.**

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

**Test meaningful behavior and risk, not every change.** Add coverage where a regression would matter, especially for production authentication and endpoint parity, persistence and data integrity, and nontrivial domain rules. There is no automatic test-per-method or test-per-change requirement. No new test is a valid choice for trivial, documentation-only, generated-output-only, and test-removal changes.

**Do not add low-value tests.** Skip tautological assertions that mirror the implementation, trivial constant or schema equality checks, pass-through or error-wrapper assertions without distinct behavior, and coverage already provided at a shared boundary. Do not export internals, broaden visibility, or add dependencies or abstractions solely to enable such assertions. When a test is explicitly removed, do not recreate equivalent coverage elsewhere unless requested.

**Prefer repository tests plus HTTP E2Es over isolated feature-service tests.** Real-database repository tests are highly recommended for query behavior, row mapping, constraints and persistence. Exercise feature-service orchestration through HTTP E2Es; do not add database-backed feature-service tests. Use unit tests for pure parameter validation and domain rules. Cover shared dependency failures at their boundaries once, not again for every endpoint using those dependencies.

**Use Go's standard `testing` functionality for new or rewritten backend tests.** Use ordinary comparisons, `t.Fatalf` for failed prerequisites, `t.Errorf` for independent checks, and `t.Cleanup` for resource cleanup. Compare errors with `errors.Is`/`errors.As`. Do not add Testify, another assertion framework, or a homegrown assertion DSL. Existing tests do not need a bulk rewrite; convert them when their relevant slice is migrated or in a separately scoped mechanical change.

### Native Tadoku API slices

**Document durable conventions for the whole application.** Tadoku API is a full application port. State architecture, compatibility and testing rules for every migrated operation, not as properties of whichever endpoint is being implemented now. Do not document individual endpoint implementations, enumerate their coverage, or add endpoint-status sections to general documentation. Use generic examples. Keep deferred cleanup notes as checkboxes in `services/tadoku-api/MIGRATION_LOG.md`.

**Keep each runtime slice small and reviewable.** Implement the requested operation without bringing along legacy service identities, audience checks, authorization features, or compatibility servers. Keep authentication and shared ban enforcement in separately reviewed changes. Existing proxied routes retain their upstream behavior.

**Run endpoint E2Es through real authentication and authorization.** Use the production JWT and ban middleware, real Keto, and real PostgreSQL for both Tadoku API and legacy endpoint comparisons. Put synthetic signed tokens in `request.http` and required role tuples in each case's `relationships.json`; anonymous gateway traffic uses a signed guest token. Never substitute passthrough middleware, inject an identity/admin claim, or install an always-allow checker in endpoint E2Es. Keep the full invalid-credential/role/ban/provider matrix at the shared boundary instead of repeating it for every endpoint. Provider-protocol tests belong with the provider adapter; do not embed ad hoc Keto emulators in endpoint fixtures or start legacy network servers for response comparisons.

**Require authentication wiring in production construction.** Business route registration must require authentication middleware; startup constructs it from required JWKS configuration with a bounded initial fetch. Keep probes and proxied routes at their existing access boundaries. Test authentication goldens at the middleware boundary with a test-only success handler and downstream identity observation, not a business endpoint. Those focused authentication comparisons use real legacy authentication without its authorization stack; endpoint comparisons use the complete auth stack. Mark intentional differences with `skipParity` in the same test table. Use synthetic signed requests and public JWKS fixtures. JWT clock overrides belong only in scoped, sequential tests and must always be restored.

**Prove migrated endpoint parity with shared goldens.** For every migrated operation, run the same SQL/relationship setup, signed HTTP request and golden response through the native and corresponding legacy production handlers in process, including their real authentication/authorization and domain/repository/query paths. Reset/reseed before each implementation so writes cannot leak between runs. Reuse production route registration and keep legacy assembly test-only. Supply deterministic time through normal clock dependencies; never rewrite SQL inside a test connector. Do not duplicate goldens, normalize away response differences, or add legacy dependencies to the native runtime. Gateway token issuance and unrelated infrastructure remain outside these in-process tests.

**Keep HTTP cases in language-neutral, reviewable files, with an explicit Go test table.** Each case declares `description []string` and `want` as an HTTP status constant. Use `APITestName(operation, want, description...)` for both the subtest name and its fixture path under `testdata/<operation>/<status>_<description>/`. Every case has `request.http` and `golden.http`; add `setup.sql` only when it needs seed data. Shared cleanup still runs when setup is absent; skip only missing seed files, not other read or SQL errors. Do not maintain a separate fixture-name field or discover cases from directories. Load any case SQL, check the response status against `want`, and compare the complete HTTP response with its golden. Compare status, headers and body directly, without decoding into the same generated types used by the handler. Seed deterministic IDs and freeze business time. For buffered HTTP goldens, fill an unknown response ContentLength from the recorder body before serializing with httputil.DumpResponse; preserve known declared lengths and explicitly set Connection headers. This framing convention treats an omitted Content-Length and an explicitly correct one as equivalent. Apart from completing the unknown length, normalize only HTTP line endings. Tests never rewrite goldens by default. Intentional local regeneration requires `-update-goldens`, an explicit source-fixture root, and review of the resulting diff. The native Tadoku API response is the sole recorder; legacy responses remain comparison-only. Update mode must reject CI, unexpected response statuses and missing golden files. Explain reviewed contract changes in the PR body. Keep endpoint assertions out of process-lifecycle tests.

**Share the HTTP E2E router, not scenario state.** Initialize the disposable database, migrations and production router once in `TestMain`. Run HTTP scenarios sequentially, resetting before each scenario and loading starting data from SQL files. Reuse `internal/testpostgres/cleanup.sql`, which explicitly lists mutable tables to truncate with `restart identity`; no table discovery or `cascade`. Preserve migration-seeded static tables and migration bookkeeping. Add newly tested mutable tables to that one file. Requests within a scenario use normal application transactions and real commits, not an outer test transaction. Database helpers return errors and provide explicit cleanup without depending on `testing.TB`. Keep pool-closing failure tests isolated; independently isolated repository/transaction tests may remain parallel.

**Write for readability.** Separate setup, execution, error handling and response mapping with whitespace. Put unrelated struct fields and composite-literal entries on separate lines. Split application and transport operations into files by functionality, and use descriptive operation names. Keep constructors and resource lifecycle code visibly separate from endpoint behavior.

**Keep each operation's HTTP E2Es in one golden-case table.** Do not add separate generated-ID, persistence-readback or hand-decoded response tests beside it. Put persistence assertions in repository tests. Request-body fixtures use JSON only.

**SQL style: always use lowercase keywords** (select, create table, not SELECT, CREATE TABLE)

### sqlc code generation

**Always run sqlc code generation after changing a SQL query.** The checked-in generated Go files must exactly match the query sources.

**Never manually edit sqlc-generated files, and never delete, revert, or selectively omit changes produced by sqlc code generation.** Commit the complete generated diff, even when code generation reveals previously stale output. If generated changes are unexpected, investigate the query inputs and pinned sqlc version, then rerun code generation; do not discard the generated changes.

Run the repository generator from the repository root. It downloads and runs the sqlc versions pinned in each active package's `generate.go`, so it does not require `go` to be installed or available on `PATH`:

```sh
./scripts/generate-sqlc.sh
```

CI runs the same script on every pull request and fails if code generation changes the working tree. Before pushing, commit the complete generated output so this check stays clean.

**Accept narrow interfaces** — define interfaces where they are used, with only the methods that consumer needs. Don't create wide/shared interfaces that bundle many methods together. Concrete implementations can be large, but each consumer should accept the smallest dependency possible. This follows Go's interface segregation principle and makes testing easier.

**Use "Repository" for persistent source-of-truth data, "Store" for everything else** — `Repository` interfaces access the primary database (Postgres) where authoritative data lives. `Store` interfaces access auxiliary storage (e.g. Valkey/Redis) for caches, derived data, pub/sub, coordination state, or any non-authoritative data. Implementations live under `storage/postgres/` and `storage/valkey/` respectively.

**Never call `time.Now()` directly** — always inject `commondomain.Clock` and use `clock.Now()`. This applies to domain services, repository methods, and background workers. The clock is created in `main.go` and threaded through constructors. This makes time-dependent code testable via `mockClock`.

For new native `tadoku-api` code, use `internal/timex.Now()` for business time instead of an injected clock. Its implementation and wall-clock tests may read `time.Now()` directly; real timers and deadlines remain independent. Scope `timex.TheWorld` inside tests that need controlled business time, never around the whole suite in `TestMain`. A test may use separate, non-nested scopes for different times. Tests using `timex.TheWorld`, including their parent tests, must not use `t.Parallel`. Leave legacy clock consumers unchanged until their slice is migrated.

**Domain must not import storage packages** — the `domain` package must never import from `storage/postgres`, `storage/valkey`, or any other storage layer. Define domain types and interfaces in the domain package; the storage layer implements them. Repository methods should convert between sqlc types and domain types internally.

**Use unexported fields for domain-enriched request data** — when a request struct has fields set exclusively by the domain layer (e.g. `UserID` from session, `Year` from clock), make those fields unexported (lowercase) and add getter methods. This prevents the HTTP layer from setting them while still allowing the repository layer to read them. The domain layer (same package) writes directly (`req.userID = ...`), other packages read via getters (`req.UserID()`). Fields that the HTTP layer legitimately sets (e.g. `ContestID`, `LanguageCodes`) remain exported.

```sh
# 1. Make changes

# 2. Compile (fast)
bazel build //services/...

# 3. Run tests
bazel test //services/... # everything
bazel test //services/immersion-api/domain/command:command_test # one test file
bazel test //services/immersion-api/domain/command:command_test --test_filter=TestValidateAndNormalizeTags # specific function

# 4. Format before committing
gofmt -w services/

# 5. Regenerate BUILD.bazel files (after adding/removing Go files or changing deps/imports)
# CI fails if these are stale (it runs `bazel run //:gazelle -- -mode=diff`)
bazel run //:gazelle

# 6. Regenerate sqlc code after modifying SQL queries
./scripts/generate-sqlc.sh

# 7. Regenerate Tadoku API OpenAPI code (after modifying its canonical spec)
./scripts/generate-openapi.sh
# Legacy service OpenAPI output is frozen until those services are retired.

# 8. Before creating PR
bazel build //services/... && bazel test //services/...
```

## Dev Environment

Use `k8s/dev/` as the Tilt entrypoint for the shared and local Kubernetes dev stacks.
Cluster-specific hostnames, registry hosts, and kube access details belong in ignored local config (`tilt_config.json`, `.env.local`); committed files should use placeholder examples.

Dev Postgres is provisioned with the Zalando `postgresql` custom resource. Do not add or reintroduce hand-rolled Postgres Deployments or Helm releases for the dev stack.

Use `make dev-seed` (`scripts/dev/seed-db.sh`) to rerun the idempotent seed data and `make dev-reset` (`scripts/dev/reset-env.sh`) for a destructive database reset; both are also exposed as `dev-seed`/`dev-reset` Tilt resources.

## Commit Guidelines

**Commit in atomic diffs** — each commit should represent one logical change. Don't bundle unrelated changes into a single commit.

**For larger refactors spanning many files**, commit in chunks that make sense — e.g. one commit per page, per service, per domain area, etc.

## Bug Reports

When a bug is reported, follow this process:

1. **Reproduce the bug first** — Use the smallest useful existing test, build, lint check, or manual scenario. Add a regression test when it protects meaningful behavior. If you claim a test fails before the fix, verify that it compiles and fails for the intended behavioral reason.
2. **Use subagents to fix** — Have subagents implement the fix and prove it against the reproduction and affected checks.
