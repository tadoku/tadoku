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

**Always write tests for new backend functionality** — new domain services, repository methods, and HTTP handlers should have corresponding test coverage.

**Always use `testify` for test assertions** — use `assert` for checks and `require` for fatal preconditions (`github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require`). Never use raw `if err != nil { t.Fatal(...) }` patterns.

**SQL style: always use lowercase keywords** (select, create table, not SELECT, CREATE TABLE)

### sqlc code generation

**Always run sqlc code generation after changing a SQL query.** The checked-in generated Go files must exactly match the query sources.

**Never manually edit sqlc-generated files, and never delete, revert, or selectively omit changes produced by sqlc code generation.** Commit the complete generated diff, even when code generation reveals previously stale output. If generated changes are unexpected, investigate the query inputs and pinned sqlc version, then rerun code generation; do not discard the generated changes.

Run the generator for every affected service from the repository root:

```sh
(cd services/immersion-api/storage/postgres && go generate)
(cd services/content-api/storage/postgres && go generate)
```

Each `go generate` command installs the sqlc version pinned in that service's `generate.go` and then runs `sqlc generate`.

CI runs sqlc for both packages on every pull request using the versions pinned in their `generate.go` files, and fails if code generation changes the working tree. Before pushing, commit the complete generated output so this check stays clean.

**Accept narrow interfaces** — define interfaces where they are used, with only the methods that consumer needs. Don't create wide/shared interfaces that bundle many methods together. Concrete implementations can be large, but each consumer should accept the smallest dependency possible. This follows Go's interface segregation principle and makes testing easier.

**Use "Repository" for persistent source-of-truth data, "Store" for everything else** — `Repository` interfaces access the primary database (Postgres) where authoritative data lives. `Store` interfaces access auxiliary storage (e.g. Valkey/Redis) for caches, derived data, pub/sub, coordination state, or any non-authoritative data. Implementations live under `storage/postgres/` and `storage/valkey/` respectively.

**Never call `time.Now()` directly** — always inject `commondomain.Clock` and use `clock.Now()`. This applies to domain services, repository methods, and background workers. The clock is created in `main.go` and threaded through constructors. This makes time-dependent code testable via `mockClock`.

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
# Run the commands in the "sqlc code generation" section above for every affected service.

# 7. Regenerate OpenAPI code (after modifying OpenAPI specs)
cd services/immersion-api/http/rest/openapi && go generate
cd services/content-api/http/rest/openapi && go generate

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

1. **Write a failing test first** — Don't start by trying to fix the bug. Instead, write a test that reproduces the bug and confirms it fails.
2. **Use subagents to fix** — Have subagents attempt to fix the bug and prove the fix by making the test pass.
