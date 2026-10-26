---
title: Production branches
description: How a runtime branch is deployed beside production with dev-cli production mode, which images and hosts it uses, who can use it and how it is torn down and expires.
---

# Production branches

Read this when a branch is, or will be, deployed beside production, by hand or
with a pull request's `deploy` label, or when you are asked to verify one.

dev-cli's production mode runs a pushed branch commit beside the production
workloads, on its own hosts and its own test tenant. Production users and data
stay on the base deployments. Only the maintainer deploys a branch by hand;
a repository admin or maintainer deploys a pull request by adding the `deploy`
label (see [Pull-request deploys](./pull-request-deploys.md)). Only branches that
change runtime code are deployed. Agents never run production mode without an
explicit instruction from the maintainer for that branch and command; they may
verify a deployed branch when asked, following
`.agents/skills/verify-tadoku/SKILL.md`.

The production configuration, its workload templates and the operator runbook
live in the maintainer's private infrastructure repository. This page covers
what the Tadoku repository provides and what a deployed branch looks like.

## What can be deployed

- **Runtime branches only.** A branch that changes anything under
  `services/tadoku-api/migrations/` is refused. Ship the migration through the
  normal release first, as described in [Database](../tadoku-api/database.md),
  then deploy the dependent branch.
- **Pushed, clean commits.** dev-cli refuses a dirty working tree and a `HEAD`
  that no `origin` branch contains, so every image names a commit anyone can
  check out.
- **Affected deployables.** As in development, `dev up` deploys only the
  deployables affected by the branch's diff against `origin/main`;
  `--service <name>` adds an unchanged one. Selecting the API or the worker
  deploys both.
- **No live edits.** `dev up` builds, deploys, waits for every rollout and
  route, prints the route, commit, image digests, branch URLs and expiry, then
  exits. There is no sync loop, hot reload or supervised restart. Push a new
  commit and rerun `dev up` to change the branch.

## Release-identical images

Branch images are built the same way as release images, so a branch runs what
`main` would publish for the same commit.

| Deployable | Push target | Build |
| --- | --- | --- |
| Tadoku API | `//services/tadoku-api:branch_push` | `--config=release` |
| Tadoku worker | `//services/tadoku-api/cmd/tadoku-worker:branch_push` | `--config=release` |
| webv2, webv3, authv2, admin | `//frontend:webv2_branch_push`, `//frontend:webv3_branch_push`, `//frontend:authv2_branch_push`, `//frontend:admin_branch_push` | `frontend/scripts/publish-image.sh` from `HEAD:frontend` |
| Tenant lifecycle Jobs | `//services/tadoku-api/cmd/tadoku-tenant:branch_push` | `--config=release` |

Each deployable's `dev_deployable` metadata names its `release.imageName` and
`release.pushTarget`; see `services/tadoku-api/BUILD.bazel` and
`.dev/frontend.bzl`. The push targets set no repository or tags: dev-cli
supplies both. [Toolchain](./toolchain.md#release-and-branch-image-entrypoints)
describes the targets and how to run them against a local registry.

### Branch image path

Images are pushed to `ghcr.io/tadoku/tadoku/branches/<image>`, where `<image>`
is `tadoku-api`, `tadoku-worker`, `frontend-webv2`, `frontend-admin`,
`frontend-webv3`, `frontend-authv2` or `tenant`. The tag is `<route>-<commit12>`, the route
followed by the first 12 characters of the commit. An existing tag is reused
because it names a clean, pushed commit. Workloads are deployed by digest.

Every branch package must be public. After each push dev-cli resolves the
digest anonymously, which proves the cluster can pull it without an image pull
Secret, and refuses a digest it cannot resolve. GitHub creates a package that
does not exist yet as private on its first push; a maintainer then makes it
public once in the package settings and reruns `dev up`.

### Branch image cleanup

`.github/workflows/clean-branch-images.yaml` runs
`scripts/clean-branch-images.sh` daily to delete old versions of the eight branch
packages. It names only packages under `branches/`; release images are never
touched. A version is deleted when all of these hold:

- it was last updated more than 14 days ago;
- its package has received no push in the last 24 hours, because tagging an
  existing digest for a new route updates only the package's timestamp, not the
  version's;
- it is not the package's newest tagged version, nor a platform or attestation
  manifest of a version that is kept.

Branches expire 4 hours after their last `dev up`, so these windows leave a
wide margin. Reusing an existing tag pushes nothing, though: a `dev up` that
reuses a version close to 14 days old can see it deleted while the branch is
live. Rerun `dev up`; dev-cli pushes a deleted tag again whenever it needs
one.

A manual run of the workflow only lists what it would delete unless `dry_run`
is unchecked. Locally, `./scripts/clean-branch-images.sh` lists the same with
your `gh` credentials and deletes only with `DRY_RUN=false`. Deleting from the
workflow needs the `tadoku/tadoku` repository to hold the Admin role in each
branch package's Manage Actions access settings.

### Why `latest` and `prod` are never pushed

Production rolls out images tagged `prod`, and development follows `latest`.
A branch pushing either tag would replace a base deployment instead of running
beside it. dev-cli therefore accepts only `<route>-<commit12>` tags, only a
single image directly under the branch registry path, and only push targets
whose Bazel definition sets neither a repository nor fixed tags. A release
target with fixed `latest` or `prod` tags cannot be used for a branch.

## Routes, tenants and hosts

A route identifies one branch deployment. It has the form
`<owner-prefix>-<branch-slug>-<8 hex>`, for example
`974796-anton-prod-dry-run-c66ea12d`.

| Item | Value |
| --- | --- |
| Test tenant | `tadoku/<route>`, for example `tadoku/974796-anton-prod-dry-run-c66ea12d` |
| Production tenant | `tadoku/prod`, never used by a branch's lifecycle Jobs |
| App host | `<route>.preview.tadoku.app` |
| Webv3 host | `<route>.next.preview.tadoku.app` |
| Account host | `<route>.account.preview.tadoku.app` (authv2) |
| Admin host | `<route>.admin.preview.tadoku.app` |
| API | The branch host's `/api/internal` path |
| Expiry | 4 hours after `dev up`; rerunning `dev up` renews it |

Each branch host serves the branch's overlay for a deployable it deployed and
the production base for every other one. The routing headers described in
[Development environment](./environment.md#routing-headers) are emitted here as
well; `X-Dev-Backend` identifies the upstream that served a request.
[Frontend conventions](../frontend/conventions.md#cross-app-urls)
describe how the frontends resolve sibling branch hosts.

A branch has its own native accounts and sessions in its test tenant. Every
branch host, whether it serves an overlay or the production base, reads the
tenant's own session cookie, and production hosts ignore it; see
[Sessions and cookies](../tadoku-api/identity.md#sessions-and-cookies).

The branch API and worker run with `API_BRANCH` and `WORKER_BRANCH` set to the
branch tenant. The API rejects a request signed for any other tenant with HTTP
421, and the worker claims only that tenant's jobs. Each branch host route
sets `x-dev-branch` to the route, and Oathkeeper signs `tadoku/<route>` from
it; requests on production hosts, whose routes remove the header, carry
`tadoku/prod`. See
[Authorization](../architecture/authorization.md#request-pipeline). A branch without an API overlay calls the production API deployment, which
serves the request under the branch tenant.

### Tenant lifecycle

dev-cli's lifecycle hooks run `tadoku-tenant` as Jobs before startup and after
teardown, connecting as the scoped `tadoku_tenant_lifecycle` database role
rather than the owner. Provisioning creates the test tenant's registry row,
copies production's active platform scoring rule set and published CMS pages
into it and creates its Flipt namespace `tadoku_<route>` in the `test`
environment.
Selecting the worker sets a worker override so only the branch worker claims
the tenant's jobs. [Tenant lifecycle command](../tadoku-api/tenant-lifecycle.md)
documents the command, the role and its refusals of `tadoku/prod`.

## Who can use a branch

Anyone can use a branch. Its pages load for anonymous visitors, and its API
admits guests and every signed-in user without any grant. Pages that need a
session, such as `/logs/new`, redirect to the branch's login page on its
account host. Only a ban stored in the branch tenant refuses a signed-in user.
[Authorization](../architecture/authorization.md#request-pipeline) lists the
admission rules.

- **Administrators.** A branch tenant has none, production administrators
  included, until one is granted with the
  [operator SQL](../architecture/authorization.md#granting-an-administrator)
  for `tadoku/<route>`.
- **Test accounts.** Pull-request deploys create none. For a branch the
  maintainer deploys by hand, the maintainer's private tooling registers
  three accounts in the branch tenant through the branch's own identity API,
  marked with the route, using plus-aliased addresses of the project mailbox:
  reader, admin and outsider. Only admin holds a role, `admins`, granted with
  the operator SQL. Their passwords
  exist only in the maintainer's local accounts file.
- **Disposable accounts.** To test a signed-in journey on a branch with an API
  overlay, register an account on its account host with a synthetic address,
  such as one under `example.com`. The account exists only in the branch
  tenant, its email is held rather than sent, and teardown deletes it with the
  tenant.

## Teardown and expiry

The maintainer runs `dev down` to remove a branch. It removes the branch's
routes and workloads and waits for their pods to stop; the worker override is
cleared once the branch worker has stopped. Its teardown hooks then delete the
tenant's cache keys, Flipt namespace and registry row with its cascading data,
including its test accounts, sessions, access tokens and permission
relationships. Production data, accounts and base deployments are untouched.

A branch expires 4 hours after its last `dev up`. A cluster CronJob scales the
Deployments of expired branches to zero, so an abandoned branch stops serving
traffic. It deletes nothing. A maintainer branch keeps its tenant data,
including test accounts, and routes until the maintainer runs `dev down`, or
until a later production `dev up`, which first tears down expired routes of
every owner. A pull-request branch is torn down in full by the first
[reconciler](./pull-request-deploys.md#the-reconciler) run after it expires,
normally within 15 minutes; a run waits while the runner is busy.
