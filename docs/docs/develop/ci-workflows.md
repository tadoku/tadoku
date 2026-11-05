---
title: CI workflows
description: What each GitHub Actions workflow checks and when it runs, how backend CI uses the private runner pool and Bazel caches, and where failing test logs are uploaded.
---

# CI workflows

Read this when you need to know what CI checks on a pull request, why a
workflow ran or how backend CI caches its work. The commands each workflow runs
are explained in [Toolchain and repository](./toolchain.md).

All workflows live in `.github/workflows/`.

Backend CI runs two jobs on the repository-scoped `tadoku-ci` Kubernetes runner
pool. The build job uses `bazel test --nobuild_tests_only` to build the full
target set and run its tests together against disposable services; Bazel runs
up to eight actions, tests included, at once. Development-only images listed in
`tools/ci/dev_image_targets` are excluded there and built by the parallel
`dev-images` job, so tests do not wait on their base image pulls. The build job
turns off PostgreSQL `fsync`, full-page writes and synchronous commit on its
disposable service before the tests. Release image commands follow the tests so their
distinct configuration does not invalidate normal analysis between the build and
tests. Main publication runs only after the tests and image scanning succeed.
Image scanning uses the Trivy that the runner image puts on `PATH`;
GitHub-hosted runs install the same version before scanning. The scan step
asserts the version in both environments.
Race tests run separately and never gate publication; see [Race tests](./race-tests.md).
ARC admits at most five Tadoku build runners, and each backend run takes two;
resource reservations may keep excess jobs queued.

The runners use the private Bazel HTTP action cache and Lab CA. Pull requests
read existing results without uploading; trusted non-PR runs populate the cache.
The normal job also uses Bazel's remote repository contents cache for repository
rules that declare reproducible outputs. Bazel validates their declared and
recorded inputs before reuse. Changed inputs and cache lookup misses or errors
fall back to repository evaluation. Missing or corrupt content discovered later
during lazy downloads is not guaranteed to recover within that command. The
upload policy applies to all Bazel commands, including dependency resolution
and queries. The first trusted
run pays the repository upload cost; later jobs can reuse those results without
repeating repository generation. This is separate from compiled action caching.
Each runner Pod prefetches trusted public `main` dependencies while becoming
idle: frozen module metadata, Gazelle and the OpenAPI generator, plus the
workflow's digest-pinned service images. Native repository archives and extracted
source trees use the Pod's private Bazelisk `emptyDir` under
`/home/runner/bazelisk/repositories`; output roots and Bazel servers remain
separate and disposable. Source-cache hits differ from HTTP action-cache hits.
Prewarm failure falls back to ordinary cold setup. The `bazel-prewarm.json`
receipt in `bazel-performance` records the seed revision, duration and size, or
reports `cold` when no successful source seed receipt was available. A `ready`
receipt means repository fetches completed; service-image pulls are best effort.

Local action outputs and service data are disposable per runner. Fork pull
requests retain GitHub-hosted isolation and cannot use the private runner pool.
Manual dispatch can select the hosted control for an equivalent-workload
comparison. Both scheduled and manual builds execute the wrapper, publication
authentication and OpenAPI checks even when their successful receipts exist. `bazel-performance` artifacts retain build profiles and build-event
JSON; include queue/setup time and publication when comparing complete CI runs.

See [CI setup verification](./ci-setup.md) for exact-input receipt keys, runtime
invalidation, cache scope and the fingerprint check command.

When the test step fails, `build-bazel.yaml` uploads the available HTTP
E2E target's `test.log` and `test.xml` as a workflow artifact for seven days.
Artifact names include the job and run attempt, for example `http-e2e-build-1`;
`race-bazel.yaml` uploads the same files as `http-e2e-race-<attempt>`.
Download these files from the workflow run when Bazel suppresses oversized
console output; they include assertion output and fixture cleanup errors.

| Workflow | Runs on | Checks |
| --- | --- | --- |
| `build-bazel.yaml` | PRs and `main` pushes touching Bazel or Go inputs | Frozen lockfile, changed forward migration risk check on PRs (the **Migration safety** check, which an approval of the `migration-safety-override` environment can override for one run), Gazelle diff, visibility, provider-dependency, repository, Go comment and `.depolicy.yaml` import checks, OpenAPI bundle and generation diff, build, image target coverage, tests and, in a parallel job, development images; scans and publishes backend images from `main` |
| `race-bazel.yaml` | Daily, manual dispatch and PRs changing it | Race-instrumented Tadoku API tests; alerts through Notify when a scheduled or manual run does not succeed. See [Race tests](./race-tests.md) |
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
| `build-frontend-authv2.yaml` | PRs and main pushes touching authv2 or Paper | Node 22 typecheck, lint, boundaries and container build; main publishes commit, `latest` and `prod`; Playwright remains manual |
| `build-frontend-webv3.yaml` | PRs and `main` pushes touching Webv3 or Paper inputs | Node 22 lint, typecheck, boundaries, standalone container and synthetic read-only browser journeys; publishes commit, `latest` and `prod` from `main` |
| `check-paper-playground.yaml` | PRs and `main` pushes touching the playground or `paper-ui` | Paper boundaries and `paper-playground` lint, typecheck, test and build; never published |
| `pr-branch.yaml` | `pull_request_target` label, push, reopen and close events | Authorizes the `deploy` label, deploys the pinned commit beside production with dev-cli and tears the branch down when the label is removed or the pull request closes. Its deploy and teardown jobs are the called workflows `pr-branch-called-deploy.yaml` and `pr-branch-called-teardown.yaml`. See [Pull-request deploys](./pull-request-deploys.md) |
| `pr-branch-reconcile.yaml` | Manual dispatch, started every 15 minutes by a homelab CronJob | Tears down pull-request branches whose pull request is closed or unlabelled, or whose branch expired; never runs `dev cleanup`. See [Pull-request deploys](./pull-request-deploys.md#the-reconciler) |
| `clean-branch-images.yaml` | Daily, and manual dispatch as a dry run by default | Deletes old production-branch image versions; see [Production branches](./production-branches.md#branch-image-cleanup) |
| `deploy-docs.yaml` | PRs and `main` pushes touching `docs/` or the API spec | `pnpm api:check`, typecheck, build and the public API boundary (`pnpm api:check-build` and no internal paths); deploys GitHub Pages from `main` |

Frontend application images from `dev/` branches publish only their immutable commit tag.
They never update `latest` or `prod`. Development manifests can pin those images
without a production release. Main publication updates the commit tag, `latest`
and `prod`. Development follows `latest` and production follows `prod`, so merging
frontend source to `main` releases each affected frontend to production, as it
does for the backend images.
