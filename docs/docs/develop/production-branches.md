---
title: Production branches
description: How the maintainer deploys a runtime branch beside production with dev-cli production mode, which images and hosts it uses, who can use it and how it is torn down.
---

# Production branches

Read this when the maintainer has deployed, or plans to deploy, a branch beside
production, or asks you to verify one.

dev-cli's production mode runs a pushed branch commit beside the production
workloads, on its own hosts and its own test tenant. Production users and data
stay on the base deployments. Only the maintainer deploys, and only branches
that change runtime code. Agents never run production mode without an explicit
instruction from the maintainer for that branch and command; they may verify a
deployed branch when asked, following
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
| webv2, auth, admin | `//frontend:webv2_branch_push`, `//frontend:auth_branch_push`, `//frontend:admin_branch_push` | `frontend/scripts/publish-image.sh` from `HEAD:frontend` |
| Tenant lifecycle Jobs | `//services/tadoku-api/cmd/tadoku-tenant:branch_push` | `--config=release` |

Each deployable's `dev_deployable` metadata names its `release.imageName` and
`release.pushTarget`; see `services/tadoku-api/BUILD.bazel` and
`.dev/frontend.bzl`. The push targets set no repository or tags: dev-cli
supplies both. [Toolchain](./toolchain.md#release-and-branch-image-entrypoints)
describes the targets and how to run them against a local registry.

### Branch image path

Images are pushed to `ghcr.io/tadoku/tadoku/branches/<image>`, where `<image>`
is `tadoku-api`, `tadoku-worker`, `frontend-webv2`, `frontend-auth`,
`frontend-admin` or `tenant`. The tag is `<route>-<commit12>`, the route
followed by the first 12 characters of the commit. An existing tag is reused
because it names a clean, pushed commit. Workloads are deployed by digest.

The branch packages are public. After each push dev-cli resolves the digest
anonymously, which proves the cluster can pull it without an image pull Secret.

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
| Account host | `<route>.account.preview.tadoku.app` |
| Admin host | `<route>.admin.preview.tadoku.app` |
| API | The branch host's `/api/internal` path |
| Expiry | 4 hours after `dev up`; rerunning `dev up` renews it |

Each branch host serves the branch's overlay for a deployable it deployed and
the production base for every other one. The routing headers described in
[Development environment](./environment.md#routing-headers) are emitted here as
well; `X-Dev-Backend` identifies the upstream that served a request. Kratos
stays on the production account host, and identities are shared with
production. [Frontend conventions](../frontend/conventions.md#cross-app-urls)
describe how the frontends resolve sibling branch hosts.

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
gives its Keto object `app:tadoku/<route>` the canonical parent and creates its
Flipt namespace `tadoku_<route>` in the `test` environment. Selecting the
worker sets a worker override so only the branch worker claims the tenant's
jobs. [Configuration](../tadoku-api/configuration.md#tenant-lifecycle-command)
documents the command, the role and its refusals of `tadoku/prod`.

## Who can use a branch

A test tenant admits only users with `access` on its Keto object; everyone
else, including anonymous visitors, receives an empty 403 from its API.
[Authorization](../architecture/authorization.md#request-pipeline) lists the
admission rules.

- **Production administrators** inherit administrator and access permits on
  the branch through the canonical parent of its Keto object.
- **Provisioned test accounts.** The maintainer's private tooling creates
  three Kratos identities per route, marked with the route, using plus-aliased
  addresses of the project mailbox:

  | Account | Branch permits |
  | --- | --- |
  | reader | `testers` |
  | admin | `testers` and `admins` |
  | outsider | None; used to check that access is refused |

  Their passwords exist only in the maintainer's local accounts file. Nobody
  registers new accounts on a branch.

## Teardown and expiry

The maintainer runs `dev down` to remove a branch. It removes the branch's
routes and workloads and waits for their pods to stop; the worker override is
cleared once the branch worker has stopped. Its teardown hooks then delete the
route's test identities and the tenant's Keto relationships, cache keys, Flipt
namespace and registry row with its cascading data. Production data,
identities and base deployments are untouched.

A branch expires 4 hours after its last `dev up`. A cluster CronJob scales the
Deployments of expired branches to zero, so an abandoned branch stops serving
traffic. It deletes nothing: the tenant data, routes and test identities
remain until the maintainer runs `dev down`.
