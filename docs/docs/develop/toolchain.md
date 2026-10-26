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

### Safe execution

Bazelisk automatically runs the executable `tools/bazel` hook. It requires
Python 3 and applies to commands from the repository root and its subdirectories.
The separate `tools/oapi-codegen/` workspace links its own hook to the same file,
so OpenAPI generation has the same protection. Calling a downloaded Bazel binary
directly, or setting `BAZELISK_SKIP_WRAPPER`, bypasses this protection; use Bazelisk
as `bazel` on your `PATH`.

On Linux, the hook requires `flock` from util-linux and serializes Bazel clients
across worktrees using one fixed lock per Unix user under `/tmp/` with `flock -o`.
The lock covers the build, and cannot be inherited by the idle server. On macOS,
where `flock` is usually absent, builds still receive the environment and resource
defaults but are not serialized.

The defaults are a 1 GiB JVM heap, a `HOST_RAM*0.25` action memory budget and a
60-second idle timeout. The action budget controls scheduling; it is not a hard
per-process memory limit. Explicit Bazel command-line overrides remain available:

```sh
bazel --host_jvm_args=-Xmx2048m --max_idle_secs=120 build \
  --local_resources=memory=2048 //services/...
```

Only explicitly allowed environment variables reach the Bazel client and server:
the host toolchain, home directory, terminal, locale and certificate paths, Docker
configuration path, and the repository's disposable test-service settings.
Provider token variables are excluded. Do not put secrets in Bazel flags or RC
files: explicit values there can still be recorded in logs. The test-service URL
variables intentionally reach Bazel tests and must contain disposable credentials.

For `bazel run`, a native preflight resolves RC imports and configs without
analyzing, building or executing the target. The hook then builds a native launch
script under the lock, releases the lock, removes its private script, and executes
the target with the caller's runtime environment. Long-running commands and
commands that invoke Bazel do not hold the build lock. Arguments, stdin, exit
status, signals and runfiles remain
native to the generated launch script. Explicit `--script_path` requests remain
build-only, and `--norun` does not execute the target.

CI image publication authenticates through the Docker config file written by
`docker login`, normally under `$HOME/.docker/`, or through `DOCKER_CONFIG`.
`tools/ci/run_bazel_images.sh` invokes the OCI push programs after the build lock
is released. Registry tokens do not need to enter the Bazel server environment.

Run the real execution checks with Python 3, Bazelisk and Linux `flock`; the
publication check additionally needs Docker. Choose new evidence directories
outside the repository:

```sh
python3 tools/ci/check_bazel_wrapper.py --evidence-dir /tmp/tadoku-wrapper-check
python3 tools/ci/check_bazel_publish.py --evidence-dir /tmp/tadoku-publish-check
```

The publication check uses synthetic credentials and a disposable registry bound
to loopback. It runs the same image runner and OCI push rules as CI, rejects an
unauthenticated push, then verifies both Docker config paths and reads the
published manifest back. It does not publish to GHCR. The build workflow retains
the checks' reports and logs as the `bazel-wrapper` artifact for seven days.

Tadoku API integration tests need disposable loopback PostgreSQL 17,
PgBouncer 1.25.2 in transaction mode and Valkey 9 instances, passed as
`TADOKU_TEST_POSTGRES_URL`, `TADOKU_TEST_PGBOUNCER_URL` and
`TADOKU_TEST_VALKEY_URL`. CI uses
`postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable` and
the same URL on port 6432 for PgBouncer, plus `redis://127.0.0.1:6379` for
Valkey. See [Tadoku API testing](../tadoku-api/testing.md).

Format Go code with `gofmt -w services/` before committing. `gofmt` does not
cover layout; follow [Go style](../tadoku-api/go-style.md) as well.
Run `bazel run //tools/ci/commentpolicy` to check handwritten Go comments;
it ignores generated files and accepts tool directives, declaration usage
notes and marked test fixture safety notes.

### Release and branch image entrypoints

Development runtime pulls in `MODULE.bazel` declare a major-version `tag` and
an immutable `digest`. Keep both when updating an image: Bazel resolves the
digest, while dependency tooling uses the tag to distinguish digest refreshes
from runtime-major upgrades.

`bazel build --config=release` pins Linux amd64 and pure Go. CI uses this
configuration for backend image scanning and publication, and prints the API
and worker image digests for comparison with a local build of the same commit.

The API and worker provide separate `branch_push` targets alongside their
fixed release targets. Branch targets have no fixed repository or tags:

```sh
bazel run --config=release //services/tadoku-api:branch_push -- \
  --repository localhost:5000/check/tadoku-api --tag check-0123456789ab
```

Frontend workflows and the `//frontend:webv2_branch_push`,
`//frontend:webv3_branch_push`, `//frontend:authv2_branch_push` and `//frontend:admin_branch_push` targets use
`frontend/scripts/publish-image.sh`. It builds `HEAD:frontend` with `git archive`,
so ignored local environment files and uncommitted edits do not enter images.
Webv3 and authv2 use their standalone `frontend/apps/webv3/Dockerfile` and
`frontend/apps/authv2/Dockerfile`; the other apps use
`frontend/Dockerfile`. The script publishes only explicitly requested tags, on Linux amd64:

```sh
frontend/scripts/publish-image.sh webv2 \
  --repository localhost:5000/check/frontend-webv2 --tag check-0123456789ab
```

Production tag approval remains manual.

### Gazelle

Direct Go dependencies belong in the first `require` block of `go.mod` and in
the `go_deps` `use_repo` list in `MODULE.bazel`. Promoting an existing indirect
dependency preserves its pinned version and checksum; rerun Gazelle to wire its
Bazel target. The Flipt seed converter uses the existing `go.yaml.in/yaml/v3`
dependency with strict known-field decoding.

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
| OpenAPI bundle and oapi-codegen | Changing a fragment in `services/tadoku-api/spec/openapi/` | `./scripts/generate-openapi.sh` | `services/tadoku-api/spec/openapi.yaml` and `services/tadoku-api/generated/openapi/` |
| API reference | Changing the OpenAPI contract, after `./scripts/generate-openapi.sh` | `pnpm api:generate` in `docs/` | `docs/docs/api/` |
| Frontend API client | Changing the OpenAPI contract, after `./scripts/generate-openapi.sh` | `pnpm --filter api-client generate` in `frontend/` | `frontend/packages/api-client/src/generated/` |

`./scripts/generate-sqlc.sh` downloads the sqlc version pinned in each query
package's `generate.go`, so it needs only `curl` and `tar`. New query packages
must be added to `SQLC_PACKAGES` in the script. `./scripts/generate-openapi.sh`
concatenates the contract fragments into `services/tadoku-api/spec/openapi.yaml`,
then runs oapi-codegen through Bazel from `tools/oapi-codegen/`, which pins the same
Bazel version as the repo root via `tools/oapi-codegen/.bazelversion`. See
[Tadoku API contract](../tadoku-api/contract.md).

## pnpm workspaces

`frontend/` and `docs/` are separate pnpm projects with their own lockfiles.
Always use `pnpm`, never `npm`. Use pnpm 10.10.0: it is the `packageManager` in
`docs/package.json` and the version CI installs for `frontend/`. CI uses Node 20.

`frontend/patches/` holds pnpm `patchedDependencies`, declared in
`frontend/package.json`. `react-dom@18.2.0.patch` applies React 19.0.0's fix to
the Node stream encoder in `renderToPipeableStream`, which otherwise writes a NUL
into the HTML when a multi-byte character meets its 2 KB buffer boundary. React 18
has no release with the fix; delete the patch when the frontend moves to React 19.
Image builds and dev-cli copy `frontend/patches/` with the lockfile.

### Frontend checks

From `frontend/`, after `pnpm install --frozen-lockfile`:

```sh
pnpm --filter webv2 exec tsc --noEmit   # typecheck
pnpm --filter webv2 lint
pnpm --filter webv2 test                # when the app defines test
pnpm build                              # every workspace package
pnpm check:paper-boundaries
```

Replace `webv2` with the application you changed. CI runs tests for webv2,
the Paper projects and the generated API client. See
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

When a normal or race test step fails, `build-bazel.yaml` uploads the available HTTP
E2E target's `test.log` and `test.xml` as a workflow artifact for seven days.
Artifact names include the job and run attempt, for example `http-e2e-build-1`.
Download these files from the workflow run when Bazel suppresses oversized
console output; they include assertion output and fixture cleanup errors.

| Workflow | Runs on | Checks |
| --- | --- | --- |
| `build-bazel.yaml` | PRs and `main` pushes touching Bazel or Go inputs | Frozen lockfile, changed forward migration risk check on PRs, Gazelle diff, visibility, provider-dependency, repository, Go comment and `.depolicy.yaml` import checks, OpenAPI bundle and generation diff, build, image target coverage, tests and race tests; scans and publishes backend images from `main` |
| `verify-sqlc.yaml` | Every PR | Reruns `./scripts/generate-sqlc.sh` and fails on any change |
| `verify-frontend-api-client.yaml` | PRs and `main` pushes touching the contract, recorded HTTP golden responses, the client package or the frontend lockfile | Regenerates the [frontend API client](../frontend/api-client.md), fails on any change, then typechecks it and replays the recorded responses through it |
| `verify-standalone-migrations.yaml` | Every PR | Fails when migration SQL files change together with other files, unless the PR has the `migration-move` label |
| `verify-docs.yaml` | Every PR | Runs `docs/scripts/docs-check.test.mjs`, so removing code that the docs still name fails |
| `verify-dev-base.yaml` | PRs and `main` pushes touching development manifests, migration inputs or its workflow | Runs `k8s/dev/base/e2e.cjs` against real published images with disposable Docker fixtures, checks profiles and owner/runtime separation, and uploads its report and logs |
| `verify-file-sizes.yaml` | Every PR | Runs `./tools/ci/check_file_sizes.sh`, which fails when a handwritten text file is over 20,000 bytes; see [Splitting a long file](../architecture/module-design.md#splitting-a-long-file) |
| `verify-http-error-mapping.yaml` | Every PR | Rejects literal 5xx responses in `services/tadoku-api/transport/http/`; `services/tadoku-api/transport/http/errors.go` maps errors to statuses |
| `build-frontend-webv2.yaml` | Pushes touching webv2, `ui` or the lockfile | Builds webv2; publishes its image from `main` |
| `build-frontend-admin.yaml` | Pushes touching admin, `ui` or the lockfile | Builds admin; publishes its image from `main` |
| `build-frontend-styleguide.yaml` | Pushes touching styleguide, `ui` or the lockfile | Builds the `ui` styleguide; publishes its image from `main` |
| `build-frontend-paper-styleguide.yaml` | PRs and `main` pushes touching Paper | Paper boundaries, `paper-ui` and `paper-styleguide` lint, typecheck, test and build, package checks, image smoke test; publishes from `main` |
| `build-frontend-authv2.yaml` | PRs and main pushes touching authv2 or Paper | Node 22 typecheck, lint, boundaries and container build; main publishes commit/latest, owner-approved manual dispatch promotes prod; Playwright remains manual |
| `build-frontend-webv3.yaml` | PRs and `main` pushes touching Webv3 or Paper inputs | Node 22 lint, typecheck, boundaries, standalone container and synthetic read-only browser journeys; publishes commit/`latest` for automatic Webv3 preview updates |
| `check-paper-playground.yaml` | PRs and `main` pushes touching the playground or `paper-ui` | Paper boundaries and `paper-playground` lint, typecheck, test and build; never published |
| `pr-branch.yaml` | `pull_request_target` label, push, reopen and close events | Authorizes the `deploy` label, deploys the pinned commit beside production with dev-cli and tears the branch down when the label is removed or the pull request closes. Its deploy and teardown jobs are the called workflows `pr-branch-called-deploy.yaml` and `pr-branch-called-teardown.yaml`. See [Pull-request deploys](./pull-request-deploys.md) |
| `pr-branch-reconcile.yaml` | Manual dispatch, started every 15 minutes by a homelab CronJob | Tears down pull-request branches whose pull request is closed or unlabelled, or whose branch expired; never runs `dev cleanup`. See [Pull-request deploys](./pull-request-deploys.md#the-reconciler) |
| `clean-branch-images.yaml` | Daily, and manual dispatch as a dry run by default | Deletes old production-branch image versions; see [Production branches](./production-branches.md#branch-image-cleanup) |
| `deploy-docs.yaml` | PRs and `main` pushes touching `docs/` or the API spec | `pnpm api:check`, typecheck, build and the public API boundary (fixed page counts, no internal paths); deploys GitHub Pages from `main` |

Frontend application images from `dev/` branches publish only their immutable commit tag.
They never update `latest` or `prod`. Development manifests can pin those images
without a production release. Main publication updates the commit tag and `latest`
for development and automatic Webv3 preview updates. The existing production
frontend workflows update `prod` only through a manual
workflow dispatch on `main` with `publish_production: true`, after the owner approves
production. This also lets frontend source land on main for future branch overlays
while production remains on its previous image.
