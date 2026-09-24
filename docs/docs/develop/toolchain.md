---
title: Toolchain and repository
description: The build tools, code generators, package managers and CI workflows in the Tadoku monorepo, with the commands to run each one.
sidebar_position: 1
---

# Toolchain and repository

Read this when you need to build, test or regenerate code, or want to know what CI checks on a pull request.

Run every command from the repository root unless a directory is given.

## Bazel

Bazel builds and tests all Go code. Install [Bazelisk](https://github.com/bazelbuild/bazelisk);
it runs the version pinned in `.bazelversion`. The Go SDK version comes from
`go.mod` through `MODULE.bazel`. Always use `bazel` for backend work, never `go`.

```sh
bazel build //services/...
bazel test //services/...                        # everything
bazel test //services/tadoku-api/e2e:e2e_test    # one target
bazel test //services/tadoku-api/e2e:e2e_test --test_filter=TestAuthentication
```

Tadoku API integration tests need disposable loopback PostgreSQL 17 and
Valkey 9 instances, passed as `TADOKU_TEST_POSTGRES_URL` and
`TADOKU_TEST_VALKEY_URL`. CI uses
`postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable` and
`redis://127.0.0.1:6379`. See [Tadoku API testing](../tadoku-api/testing.md).

Format Go code with `gofmt -w services/` before committing.

### Gazelle

Gazelle generates `BUILD.bazel` files from Go imports. Run it after adding or
removing Go files or changing imports:

```sh
bazel run //:gazelle
```

CI fails when `bazel run //:gazelle -- -mode=diff` reports changes. Gazelle
makes new libraries public; narrow their `visibility` as described in
[Import boundaries](../tadoku-api/import-boundaries.md), then run
`./scripts/check-tadoku-api-visibility.sh` and
`./tools/ci/check_tadoku_api_provider_deps.sh`.

## Code generation

Generated code is checked in and CI fails when it is stale. Commit the complete
generated diff and never edit generated files by hand.

| Generator | Run after | Command | Output |
| --- | --- | --- | --- |
| sqlc | Changing a query in `services/tadoku-api/sql/` | `./scripts/generate-sqlc.sh` | `services/tadoku-api/generated/sqlc/` |
| oapi-codegen | Changing `services/tadoku-api/spec/openapi.yaml` | `./scripts/generate-openapi.sh` | `services/tadoku-api/generated/openapi/` |
| API reference | Changing `services/tadoku-api/spec/openapi.yaml` | `pnpm api:generate` in `docs/` | `docs/docs/api/` |

`./scripts/generate-sqlc.sh` downloads the sqlc version pinned in each query
package's `generate.go`, so it needs only `curl` and `tar`. New query packages
must be added to `SQLC_PACKAGES` in the script. `./scripts/generate-openapi.sh`
runs oapi-codegen through Bazel from `tools/oapi-codegen/`. See
[Tadoku API contract](../tadoku-api/contract.md).

## pnpm workspaces

`frontend/` and `docs/` are separate pnpm projects with their own lockfiles.
Always use `pnpm`, never `npm`. Use pnpm 10.10.0: it is the `packageManager` in
`docs/package.json` and the version CI installs for `frontend/`. CI uses Node 20.

### Frontend checks

From `frontend/`, after `pnpm install --frozen-lockfile`:

```sh
pnpm --filter webv2 exec tsc --noEmit   # typecheck
pnpm --filter webv2 lint
pnpm --filter webv2 test                # when the app defines test
pnpm build                              # every workspace package
pnpm check:paper-boundaries
```

Replace `webv2` with the application you changed. CI runs unit tests only for
admin and the Paper projects, so run the others locally. See
[Frontend overview](../frontend/index.md) for per-application commands.

### Docs site

From `docs/`, after `pnpm install --frozen-lockfile`:

| Command | Does |
| --- | --- |
| `pnpm start` | Serves the site locally with live reload |
| `pnpm build` | Builds the static site; broken links fail the build |
| `pnpm typecheck` | Type-checks the site's TypeScript, such as its configuration and sidebars |
| `pnpm api:generate` | Regenerates the API reference from the Tadoku API contract |
| `pnpm api:check` | Tests the contract views, regenerates and fails if `docs/docs/api` changed |
| `pnpm docs:check` | Fails when a page lacks a `description`, an inline repository path is missing, or a relative link in `AGENTS.md`, a README or a skill is broken |

## CI workflows

All workflows live in `.github/workflows/`.

| Workflow | Runs on | Checks |
| --- | --- | --- |
| `build-bazel.yaml` | PRs and `main` pushes touching Bazel or Go inputs | Frozen lockfile, Gazelle diff, visibility, provider-dependency and `.depolicy.yaml` import checks, OpenAPI generation diff, build, image target coverage, tests and race tests; scans and publishes backend images from `main` |
| `verify-sqlc.yaml` | Every PR | Reruns `./scripts/generate-sqlc.sh` and fails on any change |
| `verify-standalone-migrations.yaml` | Every PR | Fails when migration SQL files change together with other files, unless the PR has the `migration-move` label |
| `verify-docs.yaml` | Every PR | Runs `docs/scripts/docs-check.test.mjs`, so removing code that the docs still name fails |
| `verify-http-error-mapping.yaml` | Every PR | Rejects literal 5xx responses in `services/tadoku-api/transport/http/`; `services/tadoku-api/transport/http/errors.go` maps errors to statuses |
| `build-frontend-webv2.yaml` | Pushes touching webv2, `ui` or the lockfile | Builds webv2; publishes its image from `main` |
| `build-frontend-auth.yaml` | Pushes touching auth, `ui` or the lockfile | Builds auth; publishes its image from `main` |
| `build-frontend-admin.yaml` | Pushes touching admin, `ui` or the lockfile | Tests and builds admin; publishes its image from `main` |
| `build-frontend-styleguide.yaml` | Pushes touching styleguide, `ui` or the lockfile | Builds the `ui` styleguide; publishes its image from `main` |
| `build-frontend-paper-styleguide.yaml` | PRs and `main` pushes touching Paper | Paper boundaries, `paper-ui` and `paper-styleguide` lint, typecheck, test and build, package checks, image smoke test; publishes from `main` |
| `check-paper-playground.yaml` | PRs and `main` pushes touching the playground or `paper-ui` | Paper boundaries and `paper-playground` lint, typecheck, test and build; never published |
| `deploy-docs.yaml` | PRs and `main` pushes touching `docs/` or the API spec | `pnpm api:check`, typecheck, build and the public API boundary (fixed page counts, no internal paths); deploys GitHub Pages from `main` |
| `build-postgres-backup-job.yaml` | Pushes touching `jobs/postgres-backup/` | Builds and pushes that container job image |
