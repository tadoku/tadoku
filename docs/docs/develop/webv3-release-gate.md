---
title: Webv3 release gate
description: How a Webv3 image reaches the prod tag only after the credentialed browser suite passes on a development branch tenant, its runner and credentials, dry runs and what blocks a release.
---

# Webv3 release gate

Read this when a Webv3 release is blocked, when you run a gate dry run, or when you change `.github/workflows/build-frontend-webv3.yaml`, `scripts/webv3-release-gate.sh` or the gate's runner.


On `main`, `build-frontend-webv3.yaml` tags `prod` only on an image digest
that passed `pnpm --filter webv3 test:e2e` against a real Tadoku API, with
Dev Reader and Dev Admin, on a fresh development tenant. Its jobs run in order:

1. **Publish.** A push builds the image once and pushes the commit tag and
   `latest`. A re-run builds and pushes nothing: it resolves the digest its
   commit already published as `:<commit>`, and fails when there is none. A
   `workflow_dispatch` also pushes nothing and resolves `latest`, the current
   candidate, since `main`'s head is often a commit that publishes no Webv3
   image. The job's output is the digest.
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
   - `test` runs the suite against the branch's hosts, in Chromium and, for the
     specs its WebKit project lists, in WebKit. It refuses to start
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
   cannot move `prod`. Only a push's first attempt builds and pushes. A re-run
   of a push gates the image its commit already published; a dispatch, and a
   re-run of one, resolves `latest` again. It copies the
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
the cause and use **Re-run failed jobs** on the push's run, which tests the
same digest again, or merge a fix. A push's re-run can only re-gate an image
its commit already published: when a push's check or publish job failed before the image was
pushed, that commit needs a new push to `main` that touches one of the
workflow's watched paths.

The suite refuses a committed `test.only`, so a focused test cannot pass the
gate in place of the whole suite. The suite and the branch tenant's API come
from the workflow's revision: the pushed commit, or `main`'s head for a
dispatch, which can be newer than the commit that built `latest`. Production may
still run an older API, so a Webv3 change that needs an API change production
does not have yet can pass the gate.

## Dry runs

A `workflow_dispatch` from `main` gates the current candidate, `latest`, with
the suite from `main`'s head. Its `dry_run` input, checked by default, runs the gate and reports
the decision and the digests in the promote job's summary without logging in
to the registry or moving `prod`. The `force_failure` input, allowed only with
`dry_run`, writes a failing spec into the runner's checkout before the suite
runs, to show a blocked release. A dispatch with `dry_run` unchecked promotes
like a push, promoting `latest` if it passes.

`scripts/webv3-release-gate.sh` also runs by hand from a branch checkout with
the `homelab-talos-dev` context, installed frontend dependencies and a built
`paper-ui`; set `GATE_OWNER` to your own dev-cli owner and
`NODE_EXTRA_CA_CERTS` to the Lab CA.

## Runner and credentials

The gate drives dev-cli on `homelab-talos-dev` and opens its branch hosts,
which only the homelab reaches. The `tadoku-release-gate` runner set, defined
in the homelab repository's runner runbook, scales to zero, admits only this
workflow's `push` and `workflow_dispatch` runs from `main`, runs no Docker
daemon, and reaches only DNS, public HTTPS, the development ingress (which also
serves `registry.dev.lab`) and the `homelab-talos-dev` API server. The gate
needs no Docker, `registry.lab`, Bazel remote cache or production host: Bazel
builds and pushes the branch's images and pulls only public base images. The
runner image provides Chromium's and WebKit's system libraries, so the suite's
WebKit project runs without apt, and trusts the Lab CA in its system bundle,
which WebKit reads, and in the `runner` user's NSS database, because Chromium
does not read `NODE_EXTRA_CA_CERTS`. The job sets that variable to
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
