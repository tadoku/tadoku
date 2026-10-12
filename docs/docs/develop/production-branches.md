---
title: Production branches
description: How the maintainer deploys a runtime branch beside production with dev-cli production mode, which images and hosts it uses, who can use it, how it is torn down and how the deploy label authorizes pull-request deploys.
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

### Branch image cleanup

`.github/workflows/clean-branch-images.yaml` runs
`scripts/clean-branch-images.sh` daily to delete old versions of the six branch
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
copies production's active platform scoring rule set and published CMS pages
into it, gives its Keto object `app:tadoku/<route>` the canonical parent and
creates its Flipt namespace `tadoku_<route>` in the `test` environment.
Selecting the worker sets a worker override so only the branch worker claims
the tenant's jobs. [Configuration](../tadoku-api/configuration.md#tenant-lifecycle-command)
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

## Pull-request deploys

**Not deploying yet.** `.github/workflows/pr-branch.yaml` only authorizes.
Its deploy and teardown jobs are placeholders that re-check the actor and print
their decision. They hold no secrets and change nothing in the cluster.

The `deploy` label asks for a pull request to be deployed beside production.
The workflow runs on `pull_request_target` for pull requests into `main`, so
GitHub always runs the copy on `main`; a pull request that edits the workflow
does not change how it is judged. Pull requests into other branches never
deploy. Its `authorize` job has no checkout, no cache and no action, and its
token has only `pull-requests: write` and `issues: write`. Event values reach
its script only through `env:`, and it never reads the branch name or title.
It outputs `action` (`deploy`, `teardown` or `none`), the pinned `sha` and the
pull request number `pr`.

### Who is eligible

An actor is eligible when the `role_name` returned by
`GET repos/tadoku/tadoku/collaborators/<login>/permission` is `admin` or
`maintain`. Write, triage and read are not, and any API error, including the
404 returned for bots and unknown users, counts as ineligible. A deploy needs
three eligible actors: the pull request's author, the event's sender (who added
the label, pushed or reopened) and `github.triggering_actor`, who differs from
the sender when someone re-runs the workflow.

### Decisions

| Event | Result |
| --- | --- |
| `deploy` added | Deploy, if every check below passes; otherwise refuse |
| Push to, or reopen of, a pull request with `deploy` | Redeploy, if every check passes; otherwise refuse and tear down |
| `deploy` removed, or a pull request with `deploy` closed or merged | Tear down, with no checks |
| Any other label change, push, reopen or close | Nothing, and no comment |

A deploy is refused when:

- the head is in a fork, or its repository was deleted;
- the pull request is closed;
- the head SHA is not 40 lowercase hexadecimal characters;
- the author, sender or triggering actor is not eligible;
- a changed file, or the old path of a renamed file, is under
  `services/tadoku-api/migrations/`. Ship the migration through the normal
  release first;
- the changed files cannot be checked: the files API fails, the pull request
  changes more than the 3000 files that API lists, or the listed files do not
  add up to the pull request's changed-file count.

Every refusal removes the `deploy` label. A refused push or reopen of a
same-repository pull request also tears down, because an earlier commit may be
live; label changes made by the workflow's token start no new run, so the
teardown happens in the same run. A refused label, and any refusal on a fork,
does nothing else. Teardown is always allowed, because it runs only `main`'s
code.

A deploy pins `pull_request.head.sha` from the event, never the branch tip.
If the head has moved by the time the job checks the changed files, the run
does nothing and leaves the decision to the newer push's run.

### The pull-request comment

The workflow keeps one comment per pull request. It begins with
`<!-- tadoku-pr-branch -->` and is edited in place, never reposted. It shows the
state (`deploying`, `tearing down`, `refused` or `refused, tearing down`), the
commit, the reason and a link to the run. Events that do nothing leave it
unchanged.

### Placeholder jobs

The `deploy` job runs only when `action` is `deploy`, in the `prod-branch`
environment, which only `main` can use and which has no reviewers. Its first
step re-checks the role of `github.triggering_actor`, so an ineligible
collaborator who re-runs only that job fails there. The `teardown` job runs
when `action` is `teardown` and needs no environment; on a re-run it re-checks
the triggering actor the same way. Neither job references a secret.
