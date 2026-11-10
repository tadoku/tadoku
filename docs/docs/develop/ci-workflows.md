---
title: CI workflows
description: What each GitHub Actions workflow checks and when it runs, how backend CI uses the private runner pool and Bazel caches, where failing test logs are uploaded and how the Webv3 release gate decides when to tag prod.
---

# CI workflows

Read this when you need to know what CI checks on a pull request, why a
workflow ran, how backend CI caches its work or how a Webv3 image reaches
`prod`. The commands each workflow runs
are explained in [Toolchain and repository](./toolchain.md).

All workflows live in `.github/workflows/`.

Backend CI runs two jobs on the repository-scoped `tadoku-ci` Kubernetes runner
pool. The build job uses `bazel test --nobuild_tests_only` to build the full
target set and run its tests together against disposable services; Bazel runs
up to eight actions, tests included, at once. Development-only images listed in
`tools/ci/dev_image_targets` are excluded there and built by the parallel
`dev-images` job, so tests do not wait on their base image pulls. The build job
turns off PostgreSQL `fsync`, full-page writes and synchronous commit on its
disposable service before the tests. Test results are cached until their inputs or the job's service
definitions change; see [Running tests](../tadoku-api/testing.md#running-tests). Release image commands follow the tests so their
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
| `verify-config-contract.yaml` | Every PR | Runs the [production config contract](#production-config-contract) self-check and checker |
| `verify-file-sizes.yaml` | Every PR | Runs `./tools/ci/check_file_sizes.sh`, which fails when a handwritten text file is over 20,000 bytes; see [Splitting a long file](../architecture/module-design.md#splitting-a-long-file) |
| `verify-http-error-mapping.yaml` | Every PR | Rejects literal 5xx responses in `services/tadoku-api/transport/http/`; `services/tadoku-api/transport/http/errors.go` maps errors to statuses |
| `build-frontend-webv2.yaml` | Pushes touching webv2, `ui` or the lockfile | Builds webv2; publishes its image from `main` |
| `build-frontend-admin.yaml` | Pushes touching admin, `ui` or the lockfile | Builds admin; publishes its image from `main` |
| `build-frontend-styleguide.yaml` | Pushes touching styleguide, `ui` or the lockfile | Builds the `ui` styleguide; publishes its image from `main` |
| `build-frontend-paper-styleguide.yaml` | PRs and `main` pushes touching Paper | Paper boundaries, `paper-ui` and `paper-styleguide` lint, typecheck, test and build, package checks, image smoke test; publishes from `main` |
| `build-frontend-authv2.yaml` | PRs and main pushes touching authv2 or Paper | Node 22 typecheck, lint, boundaries and container build; main publishes commit, `latest` and `prod`; Playwright remains manual |
| `build-frontend-webv3.yaml` | PRs and `main` pushes touching an input of the Webv3 image or its release gate | Node 22 lint, typecheck, boundaries, standalone container and synthetic read-only browser journeys; from `main`, publishes the commit tag and `latest`, then tags `prod` only after the [release gate](#webv3-release-gate) passes on that digest |
| `webv3-production-canary.yaml` | Hourly and manual dispatch from `main`, once `WEBV3_CANARY_ENABLED` is `true` | Signs in on `next.tadoku.app`, walks a read-only journey and publishes green or red; see [Webv3 production canary](../frontend/webv3-production-canary.md) |
| `check-paper-playground.yaml` | PRs and `main` pushes touching the playground or `paper-ui` | Paper boundaries and `paper-playground` lint, typecheck, test and build; never published |
| `pr-branch.yaml` | `pull_request_target` label, push, reopen and close events | Authorizes the `deploy` label, deploys the pinned commit beside production with dev-cli and tears the branch down when the label is removed or the pull request closes. Its deploy and teardown jobs are the called workflows `pr-branch-called-deploy.yaml` and `pr-branch-called-teardown.yaml`. See [Pull-request deploys](./pull-request-deploys.md) |
| `pr-branch-reconcile.yaml` | Manual dispatch, started every 15 minutes by a homelab CronJob | Tears down pull-request branches whose pull request is closed or unlabelled, or whose branch expired; never runs `dev cleanup`. See [Pull-request deploys](./pull-request-deploys.md#the-reconciler) |
| `clean-branch-images.yaml` | Daily, and manual dispatch as a dry run by default | Deletes old production-branch image versions; see [Production branches](./production-branches.md#branch-image-cleanup) |
| `deploy-docs.yaml` | PRs and `main` pushes touching `docs/` or the API spec | `pnpm api:check`, typecheck, build and the public API boundary (`pnpm api:check-build` and no internal paths); deploys GitHub Pages from `main` |

Frontend application images from `dev/` branches publish only their immutable commit tag.
They never update `latest` or `prod`. Development manifests can pin those images
without a production release. Main publication updates the commit tag, `latest`
and `prod`, except that Webv3's `prod` waits for its release gate. Development
follows `latest` and production follows `prod`, so merging frontend source to
`main` releases each affected frontend to production, as it does for the backend
images.

## Webv3 release gate

On `main`, `build-frontend-webv3.yaml` tags `prod` only on an image digest
that passed `pnpm --filter webv3 test:e2e` against a real Tadoku API, with
Dev Reader and Dev Admin, on a fresh development tenant. Its jobs run in order:

1. **Publish.** A push builds the image once and pushes the commit tag and
   `latest`. A `workflow_dispatch` builds and pushes nothing: it resolves the
   digest its commit already published as `:<commit>`, and fails when there is
   none. The job's output is the digest.
2. **Release gate** runs `scripts/webv3-release-gate.sh` on the
   `tadoku-release-gate` runner set:
   - `deploy` waits, for up to 20 minutes, until the development base's
     `frontend-webv3` Deployment has finished rolling out and every pod runs the
     digest. Development follows `latest`, so Image Updater rolls the base to a
     pushed candidate; a bad candidate therefore runs on the shared development
     base before the gate has tested it. The step then runs
     `dev up --owner gha --base HEAD --no-watch --service tadoku-api` on the
     local branch `webv3-gate-<run>-<attempt>`, whose `tenant` hook provisions
     and seeds a new tenant (see
     [Branch tenants](./development-data.md#branch-tenants)). Comparing with
     `HEAD` selects no deployable from the commit's changes. `tadoku-api`, which
     also starts the worker, is the only deployable that is not a frontend:
     dev-cli copies a frontend overlay's sources in with `kubectl exec`, which
     the gate's ServiceAccount may not do. The branch's Webv3 host is served by
     the base, and the step fails unless it answers with the base's
     `X-Dev-Backend` and the route's `X-Dev-Selected`. Last, it records the
     base Deployment's generation and each pod's UID and digest in
     `base-before.json`.
   - `test` runs the suite against the branch's hosts. It refuses to start
     without the four fixture variables, reads Playwright's JSON report, and
     fails on any skipped test outside the allowlist at the top of the script,
     which holds only tests that need an environment no branch tenant has. It
     writes `suite.json` with the skips and counts.
   - `verify` reads the base again and fails unless the Deployment's generation
     and every pod's UID and digest equal `base-before.json`, so a rollout that
     came and went during the suite cannot pass. Only then does the job output
     the tested digest.
   - `down` runs `dev down` whatever happened. A failed teardown does not block
     the release: the branch expires after the development TTL and the next
     `dev up` on the cluster tears it down.

   The artifact `webv3-release-gate-<attempt>` keeps the Playwright report and
   results, the `dev up` and `dev down` logs, both base snapshots and
   `suite.json` for 14 days. It holds no kubeconfig. The report records typed
   values and the repository is public, so the fixture file holds only
   development fixture identities.
3. **Promote** runs after every gate result except in a cancelled run. It
   refuses unless the gate passed on the published digest and that digest is
   still `latest`, so an older run, or a re-run of one after a newer push,
   cannot move `prod`. Only a push's first attempt builds and pushes; any
   re-run, like a dispatch, gates the image its commit already published. It copies the
   digest to `prod` with `docker buildx imagetools create`, which pushes the
   same manifest without rebuilding, then reads `prod` back and fails unless it
   names that digest.

| Gate result | Tested digest | `latest` | Result |
| --- | --- | --- | --- |
| Succeeded | The published digest | The published digest | `prod` moves to the published digest |
| Succeeded | The published digest | Any other | Blocked: a newer candidate exists |
| Succeeded | Any other, or none | Any | Blocked |
| Failed, cancelled or timed out | Any | Any | Blocked |

A blocked release fails the promote job and leaves production on its current
`prod` digest. Every way the gate can break, such as a misconfigured runner, a
base that never rolls out the digest or a hung suite, ends in a failed or
timed-out job, so it blocks the release. One failure is quiet: when no runner
picks the job up, GitHub keeps it queued for up to 24 hours before failing it,
and the workflow's later runs wait behind it. To release a blocked commit, fix
the cause and use **Re-run failed jobs**, which tests the same digest again,
or merge a fix. A re-run can only re-gate an image its commit already
published: when a push's check or publish job failed before the image was
pushed, that commit needs a new push to `main`.

The suite refuses a committed `test.only`, so a focused test cannot pass the
gate in place of the whole suite. The branch tenant runs the candidate commit's
API, while production may still run an older one, so a Webv3 change that needs
an API change production does not have yet can pass the gate.

### Dry runs

A `workflow_dispatch` from `main` gates the image the dispatched commit
published. Its `dry_run` input, checked by default, runs the gate and reports
the decision and the digests in the promote job's summary without logging in
to the registry or moving `prod`. The `force_failure` input, allowed only with
`dry_run`, writes a failing spec into the runner's checkout before the suite
runs, to show a blocked release. A dispatch with `dry_run` unchecked promotes
like a push. The development base runs only `latest`, so a dispatch for a
commit whose image is no longer `latest` waits out the 20 minutes and blocks.

`scripts/webv3-release-gate.sh` also runs by hand from a branch checkout with
the `homelab-talos-dev` context, installed frontend dependencies and a built
`paper-ui`; set `GATE_OWNER` to your own dev-cli owner and
`NODE_EXTRA_CA_CERTS` to the Lab CA.

### Runner and credentials

The gate drives dev-cli on `homelab-talos-dev` and opens its branch hosts,
which only the homelab reaches. The `tadoku-release-gate` runner set, defined
in the homelab repository's runner runbook, scales to zero, admits only this
workflow's `push` and `workflow_dispatch` runs from `main`, runs no Docker
daemon, and reaches only DNS, public HTTPS, the development ingress (which also
serves `registry.dev.lab`) and the `homelab-talos-dev` API server. The gate
needs no Docker, `registry.lab`, Bazel remote cache or production host: Bazel
builds and pushes the branch's images and pulls only public base images. The
runner image provides Chromium's system libraries and trusts the Lab CA in its
system bundle and in the `runner` user's NSS database, because Chromium does
not read `NODE_EXTRA_CA_CERTS`. The job sets that variable to
`/usr/local/share/ca-certificates/lab-root-ca.crt` for Node.

The runner set mounts its credentials only into its own pods, at
`/run/secrets/tadoku-release-gate/`. The repository holds no secret for the
gate.

| File | Contents | Given to |
| --- | --- | --- |
| `kubeconfig` | One context, `homelab-talos-dev`, with the token of a ServiceAccount bound by RoleBindings in the namespaces of `.dev/config.yaml` | Set as `KUBECONFIG` only in the steps that run `kubectl` or dev-cli |
| `fixtures.env` | `E2E_READER_EMAIL`, `E2E_READER_PASSWORD`, `E2E_SECOND_EMAIL` and `E2E_SECOND_PASSWORD` for Dev Reader and Dev Admin as the `tenant` hook seeds them | The suite step only |

The files are mounted for the whole job, so any step could read them; scoping `KUBECONFIG` only prevents accidental use. The real limits are that only reviewed code on `main` runs on this runner and that the ServiceAccount is namespaced, with no Secrets and no `pods/exec`.

The job fails before it touches the cluster when either file is missing or
the kubeconfig does not authenticate as a ServiceAccount. dev-cli is pinned in
the workflow to `v0.12.0`.

## Production config contract

Production manifests depend on names and paths that this repository defines:
runtime environment variables, ingress hosts, probe paths, the auth cookie
domain and feature flag keys.
`tools/ci/configcontract/contract.json` lists each item with the files that
consume it and the text that must appear there.
`tools/ci/configcontract/check.py` fails and names the item and file when that
text or file is missing. The text must not be part of a longer name: a key does
not match a longer key, `HOST=` does not match `APP_HOST=` and `tadoku.app` does
not match `tadoku.app.example`. When an entry's text ends in punctuation, such
as a template literal, include its closing delimiter so a suffix cannot extend
it. Matching searches the raw file text, comments and strings included. It
catches a missing file and an ordinary rename but does not prove the consumer is
still active, so when you rename a consumer, also update comments that quote its
old name. An equivalent rewrite of the consuming code, such as
`process.env['X']` for `process.env.X`, fails until its entry is updated.
`verify-config-contract.yaml` runs it on every pull request, so it can be a
required check.

```sh
python3 tools/ci/configcontract/check.py
python3 tools/ci/configcontract/self_check.py
```

When you add, rename or remove an item that production depends on, change the
production configuration and the contract in step: add one entry line per file
that consumes the item, pointing at the code that reads or serves it rather
than at a copy of the value. Renaming an item breaks the deployed manifests
until they are updated to match.

The production manifests live outside this repository and are not checked here:
the Webv3 and Webv2 Deployments' environment values and probes, the HTTPRoute
hostnames, the Tadoku API's `API_IDENTITY_COOKIE_DOMAIN` (production relies on
its default) and the production Flipt flag and segment state. Webv3's optional
runtime `NODE_EXTRA_CA_CERTS` is read by Node itself and is not in the contract.
