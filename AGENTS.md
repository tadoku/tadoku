# Development Workflow

**Make impossible states impossible to represent.** Model data structures so only valid combinations can be constructed. Use distinct types for different operations and explicit variants with required payloads for mutually exclusive states, rather than independent flags and optional fields that permit contradictory combinations. Parse and validate external input at the boundary, then pass the resulting domain values through the application.

## CMS content

**Page content managed by the CMS must only be changed in the admin CMS.** Public pages such as Contact and About store their HTML in the CMS (`pages` / `pages_content`, namespace `tadoku`, slug such as `contact`).

**Do not ship CMS copy as a code change.** Do not open a pull request, hardcode the copy in the frontend, or write a database migration or other SQL rewrite that updates CMS page HTML or equivalent content rows.

**Stop as soon as the request is CMS-managed page content.** When the change is page copy rather than application UI or code, stop the coding work immediately. Do not open a pull request, write a migration, or invent a code workaround. Report that the edit belongs in the admin CMS and wait for that edit.

## Frontend

**Always use `pnpm`, not `npm`.**

### Tadoku Paper

For Paper applications, use `paper-ui` rather than the legacy `ui` package. Follow the [Paper composition guide](docs/docs/frontend/paper-composition.md): controls live in `paper-ui`, product patterns describe reusable Tadoku task sections, and application screens own routes, data, permissions, forms, and mutations. Start pattern code beside its screen; a catalogue pattern does not automatically become a shared component. Keep application services and routers out of `paper-ui`.

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

**Always use `bazel`, not `go`.** Run `gofmt -w services/` before committing and `bazel run //:gazelle` after adding or removing Go files or changing imports.

**Ship every migration as a standalone change**: its own commit, pull request and deployment, before any code that depends on it.

**SQL style: always use lowercase keywords** (`select`, `create table`, not `SELECT`, `CREATE TABLE`).

**Always run `./scripts/generate-sqlc.sh` after changing a SQL query** and commit its complete output. Never edit generated files by hand.

**Never call `time.Now()` for business time**; use `internal/timex.Now()`.

| Before changing … | Read |
| --- | --- |
| Tadoku API code for the first time | `docs/docs/tadoku-api/index.md` |
| An application operation, feature, repository or domain package | `docs/docs/tadoku-api/conventions.md` |
| `spec/openapi.yaml` or a request or response shape | `docs/docs/tadoku-api/contract.md` |
| A migration, SQL query or transaction | `docs/docs/tadoku-api/database.md` |
| Any test | `docs/docs/tadoku-api/testing.md` |
| HTTP golden cases, fixture tokens or relationships | `docs/docs/tadoku-api/http-e2e.md` |
| A user journey | `docs/docs/tadoku-api/user-journeys.md` |
| An import, package, feature or Bazel visibility | `docs/docs/tadoku-api/import-boundaries.md` |
| Startup, environment variables or provider clients | `docs/docs/tadoku-api/configuration.md` |
| An operation's access rules | `docs/docs/architecture/authorization.md` |

```sh
bazel build //services/...                           # compile
bazel test //services/...                            # all tests
bazel test //services/tadoku-api/e2e:e2e_test --test_filter=TestAuthentication
./scripts/generate-openapi.sh                        # after changing the Tadoku API spec
bazel build //services/... && bazel test //services/...  # before creating a PR
```

## Dev Environment

For verifying application changes, read the checked-in
[verify-tadoku skill](.agents/skills/verify-tadoku/SKILL.md), then only the relevant
[feature-map sections](.agents/skills/verify-tadoku/references/features/README.md).
This works without a globally installed skill: open those files directly if your
agent does not discover `.agents/skills`. Update the relevant map in the same PR
when navigation, prerequisites or observable behavior changes. The map is a
verification aid, not a second deployment catalog; Bazel owns service discovery.

Use DevCLI for frontend and Tadoku API development; [Development environment](docs/docs/develop/environment.md) owns its workflow.
The real, non-secret Homelab configuration is committed in `.dev/config.yaml` and
`k8s/dev/base/`. This is development-only GitOps, not the production deployment.
Keep credentials, private keys and kube access details outside
Git; manifests may reference existing Secrets and include the public Lab CA.
Use an ignored local file with `dev <command> --config <path>` for overrides.
Follow [Development base](docs/docs/operations/development-base.md) for activation, automatic
migrations, credential bootstrap and recovery.

Dev Postgres is provisioned with the Zalando `postgresql` custom resource. Do not add or reintroduce hand-rolled Postgres Deployments or Helm releases for the dev stack.

Use `make dev-seed` (`scripts/dev/seed-db.sh`) to rerun idempotent seed data. `make dev-reset` is a fail-closed guard. Use `dev down` for overlays; any database deletion requires an explicitly approved, scoped runbook.

`infra/dev/ory/` retains the Kratos schema and Keto namespace fixtures used by Bazel backend tests; these are not obsolete deployment files. Preserve the shared SQL fixtures in `scripts/dev/seed/` for base and branch seeding.

## Commit Guidelines

**Commit in atomic diffs** — each commit should represent one logical change. Don't bundle unrelated changes into a single commit.

**For larger refactors spanning many files**, commit in chunks that make sense — e.g. one commit per page, per service, per domain area, etc.

**Keep plans, work-in-progress notes and verification evidence out of the repository.** Planning documents, research notes, screenshots and reports do not belong in Git; attach evidence to the pull request.

## Bug Reports

When a bug is reported, follow this process:

1. **Reproduce the bug first** — Use the smallest useful existing test, build, lint check, or manual scenario. Add a regression test when it protects meaningful behavior. If you claim a test fails before the fix, verify that it compiles and fails for the intended behavioral reason.
2. **Use subagents to fix** — Have subagents implement the fix and prove it against the reproduction and affected checks.
