---
title: Pull-request deploys
description: How the deploy label deploys a pull request beside production and tears it down from GitHub Actions, who may use it, the deploy and teardown jobs, the reconciler and the environment secrets.
---

# Pull-request deploys

Read this when a pull request has, or will get, the `deploy` label, when you
change `.github/workflows/pr-branch.yaml` or
`.github/workflows/pr-branch-reconcile.yaml`, or when you are asked to verify a
pull-request branch. [Production branches](./production-branches.md) describes
the images, hosts, tenant and access every production branch shares.

The `deploy` label asks for a pull request to be deployed beside production.
`.github/workflows/pr-branch.yaml` authorizes the request on a GitHub-hosted
runner, then deploys the pinned commit, or tears the branch down, with dev-cli
on the maintainer's homelab runner. Removing the label, or closing or merging
the pull request, tears the branch down.
`.github/workflows/pr-branch-reconcile.yaml` tears down every pull-request
branch that is left behind: one whose pull request is closed or has no
`deploy` label, and one that expired.

Every push by an eligible actor to a labelled pull request redeploys it with no
manual step, and every deploy renews the 4-hour expiry. To renew a live branch
without a new commit, re-run the pull request's newest deploy run: `dev up`
reuses the same images, so pods that did not change keep running. Removing the
label and adding it again tears the branch down first, then deploys it with a
newly provisioned tenant.

The workflow runs on `pull_request_target` for pull requests into `main`, so
GitHub always runs the copy on `main`; a pull request that edits the workflow
does not change how it is judged. Its `deploy` and `teardown` jobs are defined
in the reusable workflows `.github/workflows/pr-branch-called-deploy.yaml` and
`.github/workflows/pr-branch-called-teardown.yaml`, which GitHub reads from the
same `main` commit. Pull requests into other branches never
deploy. Its `authorize` job has no checkout, no cache and no action, and its
token has only `pull-requests: write` and `issues: write`. Event values reach
its script only through `env:`, and it never reads the branch name or title.
It outputs `action` (`deploy`, `teardown` or `none`), the pinned `sha` and the
pull request number `pr`.

## Who is eligible

An actor is eligible when the `role_name` returned by
`GET repos/tadoku/tadoku/collaborators/<login>/permission` is `admin` or
`maintain`. Write, triage and read are not, and any API error, including the
404 returned for bots and unknown users, counts as ineligible. A deploy needs
three eligible actors: the pull request's author, the event's sender (who added
the label, pushed or reopened) and `github.triggering_actor`, who differs from
the sender when someone re-runs the workflow.

## Decisions

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
does nothing else; if an earlier deploy of the pull request is still live, the
reconciler tears it down because the label is gone. Teardown is always
allowed, because it runs only `main`'s code.

A deploy pins `pull_request.head.sha` from the event, never the branch tip.
If the head has moved by the time the job checks the changed files, the run
does nothing and leaves the decision to the newer push's run.

## The pull-request comment

The workflows keep one comment per pull request. It begins with
`<!-- tadoku-pr-branch -->` and is edited in place, never reposted. The
`authorize` job sets the state `deploying`, `tearing down`, `refused` or
`refused, tearing down`, with the commit, the reason and a link to the run.
The `deploy` job then replaces it with its result, and the `teardown` job and
the reconciler with theirs. Events that do nothing and superseded deploy runs
leave it unchanged. A comment that already reports `torn down` or
`torn down by containment` is kept by a `tearing down` decision and by a
`teardown` job that finds no branch, because any deploy since that teardown
would have replaced it. A `refused, tearing down` decision still replaces it
with the refusal's reason.

## The deploy job

The `deploy` job runs only when `action` is `deploy`, in the `prod-branch`
environment, which only `main` can use and which has no reviewers. It shares
the concurrency group `pr-branch-<number>` with the `teardown` job and never
cancels a running job; GitHub keeps only the newest pending run of a group.

It runs on `tadoku-prod-branch`, an ephemeral self-hosted runner set in the
maintainer's homelab, because the cluster API server accepts only allowlisted
addresses and the homelab's is one of them. The runner refuses every job that
does not come from `pr-branch.yaml` or `pr-branch-reconcile.yaml` on `main`,
and it runs one job at a time. Its network allows only DNS, public HTTPS and
public TCP/6379, the port of production's Valkey that the isolation check
reads. The runner set and its image, which provides Bazel through bazelisk,
`kubectl`, `gh`, `gcc` and `g++`, are defined in the homelab repository.

Its steps, in order:

1. Re-check the role of `github.triggering_actor`, before any secret is
   referenced, so an ineligible collaborator who re-runs only this job fails
   here.
2. Read the pull request. Unless it is open, still has `deploy` and its head is
   still the pinned commit, the run stops as **superseded**: it succeeds,
   writes only its job summary and leaves the comment and status to the run of
   the later event.
3. Check out the pinned commit with full history into `tadoku/` as the local
   branch `pr-<number>`, and clone `main` of the private infrastructure
   repository into `argocd/` with its deploy key, over GitHub's SSH endpoint on
   port 443 because the runner cannot reach port 22. Neither checkout keeps
   its credentials.
4. Install dev-cli with `go install` at the `DEV_CLI_VERSION` release tag.
5. Set up Bazel with the bazelisk, repository and `build-bazel` disk caches
   restored only. Caches saved from
   `pull_request_target` would be restored by `main`, so the workflow never
   saves one, and it uploads no artifact. Frontend images build inside Docker
   through `frontend/scripts/publish-image.sh`, so the runner needs no Node or
   pnpm.
6. Write a private kubeconfig whose context is `lke20948-ctx`, and stop unless
   the token authenticates as
   `system:serviceaccount:tdk-prod-branches:pr-branch-deployer`. Log in to
   `ghcr.io` with the job token, which has `packages: write`.
7. Derive the route with `dev url --owner gh`. The owner is always `gh` and the
   branch `pr-<number>`, so the route `gh-pr-<number>-<8 hex>` stays the same
   across pushes and contains nothing from the pull request.
8. Record the isolation baseline with the infrastructure repository's
   isolation check, `argocd/scripts/branch-isolation-check.sh`. When the
   baseline reports FAIL, isolation cannot be verified, so the job deploys
   nothing and reports `failed`.
9. Run `dev up --owner gh --config argocd/dev-cli/production/config.yaml` from
    `tadoku/`. As for the maintainer, it first removes expired routes of every
    owner. `dev doctor` is not run: it checks create rights the ServiceAccount
    deliberately lacks, and `dev up` runs its own preflight.
10. Run the isolation `compare` against the baseline, even when `dev up`
    failed.
11. When the compare reports FAIL, contain it: `dev down --owner gh`, the
    `after-teardown` check and removal of the `deploy` label.
12. Update the comment, set the commit status `prod-branch/deploy` on the
    pinned commit with a link to the run, and write the job summary.

| Outcome | State | Status | What remains |
| --- | --- | --- | --- |
| `dev up` succeeds and the compare reports PASS or only UNVERIFIED checks | `live` | success | The deployment, until the next run or expiry |
| `dev up` finds no affected deployable | `nothing to deploy` | success | Whatever was already deployed |
| `dev up`, the isolation baseline or an earlier step fails | `failed` | failure | The label and anything already applied, Pending pods included, until the next run or expiry |
| The compare reports FAIL | `torn down by containment` | failure | Nothing; the comment opens with a caution |

There is no fixed limit on how many branches run at once; the cluster's free
capacity is the limit. A failed `dev up` tears nothing down: a build, push,
hook, rollout or route wait error, including a pod that cannot be scheduled,
leaves the branch as it is. Push a fix or relabel to retry. An isolation FAIL
is the only failure that tears down. UNVERIFIED checks are a warning.

The comment and the job summary show the state, the pinned commit with a
compare link, the dev-cli version, the cluster identity and the route,
the App, Webv3, Account and Admin branch URLs, the `dev up` lines with the route, commit, affected
overlays, image digests, gateway checks and expiry (and the end of its output
when it failed), every isolation verdict with its report lines and a link to
the run. The cluster's API server address is replaced with `<cluster>` in
both, because unlike the run log they are not masked.

The job holds production credentials and runs the pull request's own Bazel
files and build scripts, which is why only admins and maintainers can start
it. Its job token can push to any package the repository can write, release
packages included; the isolation `images` check reports a changed `:prod`
digest.

## The teardown job

The `teardown` job runs only when `action` is `teardown`, in the same
environment, runner set and concurrency group as the `deploy` job. It runs
only `main`'s code and never checks out the pull request. Its token has
`contents: read` and `pull-requests`, `issues` and `statuses` write; it cannot
write packages.

Its steps, in order:

1. On a re-run, re-check the role of `github.triggering_actor`. A first
   attempt needs no eligible actor, because teardown is always allowed.
2. Write the kubeconfig and check its identity, as the `deploy` job does.
3. Find the branch's lifecycle marker: the ConfigMap in `tdk-prod-branches`
   labelled `dev-cli.io/lifecycle-marker=true` and `dev-cli.io/owner=gh` whose
   `dev-cli.io/owner-original` annotation is `gh` and whose
   `dev-cli.io/branch-original` annotation is `pr-<number>`. dev-cli writes the
   marker before it deploys anything and deletes it last, so without one there
   is nothing to remove: the job reports `nothing to tear down` and succeeds
   without setting up any tool.
4. Check out `main` into `tadoku/` as the local branch `pr-<number>`. dev-cli
   derives the route from the owner and the branch name, so this checkout
   addresses the same route as the deploy. Clone the infrastructure
   repository, install dev-cli and set up Bazel with restored caches, as the
   `deploy` job does. `dev down` queries the tenant Jobs' Bazel push target
   before it reuses the tenant image the deploy published. The job never logs
   in to a registry, so when that image is gone the teardown fails rather than
   publishing `main`'s tenant code under the pull request's tag.
5. Check that `dev url` prints the marker's route, then run
   `dev down --owner gh --config argocd/dev-cli/production/config.yaml` from
   `tadoku/`. It removes the routes and workloads and runs the teardown hooks
   described in [Teardown and expiry](./production-branches.md#teardown-and-expiry).
6. Run the isolation check's `after-teardown` mode, which checks that no
   object, permission relationship, Valkey key, Flipt namespace or test identity remains for
   the route.
7. Update the comment, set the `prod-branch/deploy` status on the deployed
   commit, which the marker's `dev-cli.io/source-revision` annotation names,
   and write the job summary.

| Outcome | State | Status | Job |
| --- | --- | --- | --- |
| No lifecycle marker | `nothing to tear down` | Unchanged | Succeeds |
| No lifecycle marker, and the comment already reports `torn down` or `torn down by containment` | Comment kept, with a notice in the log | Unchanged | Succeeds |
| `dev down` succeeds and `after-teardown` reports PASS or only UNVERIFIED checks | `torn down` | success | Succeeds; UNVERIFIED checks are a warning |
| `dev down` fails, `after-teardown` reports FAIL, or an earlier step fails | `teardown failed` | failure | Fails |

A failed teardown keeps its lifecycle marker, and the pull request is closed
or unlabelled by then, so the reconciler retries it on its next run.

The comment and job summary show the state, the reason from `authorize`, the
route and deployed commit, the end of the `dev down` output, the
`after-teardown` verdict with its report lines and a link to the run, with the
cluster's API server address replaced by `<cluster>`.

## The reconciler

`.github/workflows/pr-branch-reconcile.yaml` has no GitHub schedule, because
GitHub runs this repository's scheduled workflows hours late. The homelab
`tadoku-pr-branch-reconcile` CronJob starts it through `workflow_dispatch`
every 15 minutes, and a maintainer can dispatch it by hand. It runs from
`main`, on the `tadoku-prod-branch` runner set in the
`prod-branch` environment. It sets `deployment: false`, so its runs add no
records to the environment's deployment history. Its concurrency group queues
a run behind the previous one and never cancels it, and because the runner set
runs one job at a time, it never overlaps a deploy or teardown. Its token has
the same permissions as the `teardown` job's.

It lists the lifecycle markers labelled `dev-cli.io/owner=gh` whose
`dev-cli.io/owner-original` annotation is `gh` and whose branch is
`pr-<number>`. It never touches a marker of any other owner, the maintainer's
branches included. It reads each marker's pull request and decides:

| Pull request | Result |
| --- | --- |
| Closed or merged | Torn down |
| Open, without `deploy` | Torn down |
| Open, with `deploy`, and the marker's `dev-cli.io/expires-at` has passed | Torn down |
| Open, with `deploy`, and not expired | Kept |
| Does not exist | Torn down, with no comment or status |
| Cannot be read | Kept until the next run, with a warning |

When no marker exists, or every branch is kept, the run stops before it checks
anything out or sets up a tool. Otherwise it first removes the `deploy` label
from each pull request that still has it, so a label added during the run
starts a deploy that runs after the teardown. Then, for each branch, it
follows the `teardown` job's steps 4 to 7 on `main` checked out as
`pr-<number>`: the route check, `dev down`, `after-teardown`, the comment
with the reason, and the status. A failed teardown does not stop the others;
the run fails at the end, and the next run retries it.

The reconciler never runs `dev cleanup`, which tears down expired routes of
every owner.

## Environment secrets and variables

The `prod-branch` environment holds every credential the jobs use. Each secret
reaches only the steps that need it.

| Name | Kind | Value | Used by |
| --- | --- | --- | --- |
| `KUBE_SERVER` | Secret | The cluster API server's `https://` URL | Kubeconfig steps |
| `KUBE_CA` | Secret | The cluster CA as base64-encoded PEM: the `ca.crt` data of the ServiceAccount's token Secret, as stored | Kubeconfig steps |
| `KUBE_TOKEN` | Secret | The `pr-branch-deployer` ServiceAccount token | Kubeconfig steps |
| `ARGOCD_DEPLOY_KEY` | Secret | The private half of a read-only SSH deploy key for the infrastructure repository | Its checkouts |
| `FEATURE_FLAGS_TOKEN` | Secret | A fine-grained token with read access to the contents of `tadoku/feature-flags` | The `deploy` job's isolation checks, as `GH_TOKEN` |
| `DEV_CLI_VERSION` | Variable | A dev-cli release tag, such as `v0.11.1` | `go install` |

The ServiceAccount, its permissions and token rotation are documented in the
infrastructure repository's production runbook.
