# Agent guide

This is the entry point for work in the Tadoku monorepo. It holds the rules for
every change and names the page that holds each area's rules. Before you change
files in an area, read its page; those rules are as binding as this file.

## Repository map

| Path | Contains |
| --- | --- |
| `services/tadoku-api/` | Tadoku API, the only backend (Go, built with Bazel) |
| `services/` | Supporting services and shared packages in `services/common/` |
| `frontend/` | pnpm workspace: webv2, auth, admin, the Paper apps and the `ui` and `paper-ui` packages |
| `docs/docs/` | Developer documentation, published at https://tadoku.github.io/tadoku/ |
| `.dev/`, `k8s/dev/base/` | Development environment: dev-cli configuration and the Argo CD base |
| `scripts/`, `tools/` | Code generation, development seeding and CI checks |
| `.agents/skills/` | Repository `dev-cli` and `verify-tadoku` skills |

## Rules for every change

- **Make impossible states impossible to represent.** Model data so only valid
  combinations can be constructed: distinct types for different operations and
  explicit variants with required payloads for mutually exclusive states, not
  independent flags and optional fields. Parse and validate external input at
  the boundary, then pass domain values through the application.
- **CMS-managed page content is edited in the admin CMS only.** Public pages such
  as Contact and About store their HTML in the CMS. When a request is page copy
  rather than application code, stop immediately: do not open a pull request,
  hardcode the copy, write a migration or SQL rewrite, or invent a workaround.
  Report that the edit belongs in the admin CMS.
- **Use `pnpm`, never `npm`. Use `bazel`, never `go`.**
- **Go comments give usage instructions, not code narration.** Keep `//go:`
  directives. Add prose only at a declaration when callers need a non-obvious
  constraint; test fixture safety comments begin `// Test safety:`. Do not use
  comments to suppress lint, justify a defect or restate code. Summaries of
  return values, ownership and constructor defaults are restatements when the
  code already shows them. Run
  `bazel run //tools/ci/commentpolicy` after changing Go comments.
- **Ship every database migration as a standalone change.** Put it in its own
  commit and pull request, separate from runtime code, and deploy it
  independently while it remains compatible with the currently deployed
  application. Dependent runtime pull requests may be opened in a review stack
  before the migration is deployed. Before merging or deploying dependent work,
  the migration must be merged and deployed, and the dependent stack must be
  rebased onto the updated `main`. Never combine a migration and runtime
  behavior in one commit, pull request or deployment, including within a
  stacked series.
- **Commit atomic diffs.** Split larger refactors into coherent chunks, such as
  one commit per page, service or domain area.
- **Keep plans, work-in-progress notes and verification evidence out of the
  repository.** Attach evidence to the pull request.
- **Never commit credentials, private keys or kubeconfigs.**
- **Fix bugs from a reproduction.** Reproduce first with the smallest useful
  test, build, lint check or manual scenario. Add a regression test when it
  protects meaningful behavior, and confirm it fails for the intended reason
  before the fix. Have subagents implement the fix and prove it against the
  reproduction and affected checks.
- **Keep the docs current.** Update the page that documents a behavior,
  convention or workflow in the same pull request that changes it. Durable docs
  describe the current state only, without task history or issue numbers.

## Before you change … read

| Area | Read first |
| --- | --- |
| Tadoku API code, the first time | `docs/docs/tadoku-api/index.md` |
| An application operation, feature, repository or domain package | `docs/docs/tadoku-api/conventions.md` |
| `services/tadoku-api/spec/openapi.yaml` or a request or response shape | `docs/docs/tadoku-api/contract.md` |
| A migration, SQL query or transaction | `docs/docs/tadoku-api/database.md` |
| Any Go test | `docs/docs/tadoku-api/testing.md` |
| HTTP golden cases, fixture tokens or relationships | `docs/docs/tadoku-api/http-e2e.md` |
| A user journey | `docs/docs/tadoku-api/user-journeys.md` |
| An import, package, feature or Bazel visibility | `docs/docs/tadoku-api/import-boundaries.md` |
| Startup, environment variables, provider clients or the leaderboard worker | `docs/docs/tadoku-api/configuration.md` |
| Access rules, roles or bans | `docs/docs/architecture/authorization.md` |
| Feature flags | `docs/docs/architecture/feature-flags.md` |
| webv2, auth, admin or the `ui` package | `docs/docs/frontend/conventions.md` |
| A Paper application or `paper-ui` | `docs/docs/frontend/paper-composition.md` |
| Build tooling, code generation or CI | `docs/docs/develop/toolchain.md` |
| Running your branch on the development cluster | `docs/docs/develop/environment.md` |
| Using dev-cli for live branch edits | `.agents/skills/dev-cli/SKILL.md` |
| Proving a change works | `.agents/skills/verify-tadoku/SKILL.md`, then its feature map |
| `k8s/dev/base/` or the development base | `docs/docs/operations/development-base.md` |
| A docs page | `docs/docs/index.md` |

## Checks before a pull request

- Backend: `gofmt -w services/`, `bazel run //:gazelle`,
  `bazel run //tools/ci/commentpolicy`, then
  `bazel build //services/... && bazel test //services/...`.
- Frontend, from `frontend/`: `pnpm --filter <app> exec tsc --noEmit`,
  `pnpm --filter <app> lint`, then `pnpm build`.
- Docs, from `docs/`: `pnpm build` and `pnpm docs:check`.

`docs/docs/develop/verifying-changes.md` lists the checks for each kind of
change, including code generation.

## How the docs are organized

- Folders in `docs/docs/` match the site sidebar: `develop/`, `architecture/`,
  `tadoku-api/`, `frontend/`, `operations/` and the generated `api/` reference.
- Every page's frontmatter `description` says what it covers, and its first line
  says when to read it. Check those before reading a page in full.
- Repository paths in pages are inline code relative to the repository root.
- `pnpm docs:check` fails when a page lacks a description, an inline repository
  path does not exist, or a relative link in `AGENTS.md`, a README or a skill is
  broken.
