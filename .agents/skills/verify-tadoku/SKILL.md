---
name: verify-tadoku
description: Verify Tadoku application changes in the shared development environment using DevCLI branch overlays and real browser journeys. Use for frontend/backend live edits, routing checks, screenshots and verification handoffs; not production deployment or environment bootstrap.
---

# Verify Tadoku

Run commands from the repository root unless stated otherwise. No global skill
or conversation history is required. If automatic skill discovery is unavailable,
read this file directly. This skill guides verification; it does not grant access,
authorize unrelated mutations, or replace the repository's `AGENTS.md`.

## Choose the proof

Read the [feature index](references/features/README.md), then the relevant area.
Identify the user action, expected visible result, persistence or downstream
effect, and the failure/empty/permission state the change could break. For a bug,
reproduce before changing it. A successful build or HTTP 200 alone is not proof.

Use the existing DevCLI, browser automation, and repository tests. Do not invent
another router, wrapper CLI, fixture service or authentication bypass. Preserve
the real login → frontend → gateway → API path. Choose meaningful affected
journeys rather than running the entire inventory for every edit.

## Establish the environment

1. Inspect `git status --short`, `git branch --show-current` and `git rev-parse HEAD`.
   Preserve other work. Fetch `origin/main` when available; record a stale base if
   fetching fails. Use a task branch and a unique stable owner for this checkout.
   Check for an existing `dev up` loop belonging to this exact checkout before
   starting another; reuse only your own matching loop.
2. Run `dev version` and `dev doctor`. The supported baseline is v0.4.0; installation,
   private Git access, PATH and overrides are in [the runbook](../../../.dev/README.md).
   Required local tools include Git, Bazel, kubectl and pnpm for frontend checks.
   Local container-based E2Es additionally need Docker. Lab networking, DNS, CA
   trust and existing cluster credentials must already work. Doctor is a read-only
   prerequisite check, not an E2E.
3. Read [`.dev/config.yaml`](../../../.dev/config.yaml). It targets `homelab-dev`
   (`https://192.168.1.190:6443`, node `ct190`), not production. Always specify
   `--context homelab-dev` on kubectl commands. If access or the base is broken,
   report the failing check; do not bootstrap credentials, restart shared services,
   change Argo resources, or substitute production. Operators have a separate
   [base runbook](../../../k8s/dev/base/README.md).

Synthetic account details and fixture caveats live in the [feature index](references/features/README.md).
Do not run `make dev-seed` automatically: it mutates shared identities and base
fixtures. Use it only when missing fixtures need an authorized shared setup.

## Start the changed code

Make the intended edits **before** starting the loop. Example owner: replace
`agent-my-task` consistently, including in other terminals.

```sh
dev up --owner agent-my-task --task migrate --task seed
```

Keep this process running in a durable terminal/session; record its handle and
working directory. Tasks prepare a branch application database before API startup.
For frontend-only **read-only** checks, omit the tasks. For writes, ensure there
is a branch API even if only frontend code changed:

```sh
dev up --owner agent-my-task --service tadoku-api --task migrate --task seed
```

Bazel selects affected deployables from the merge base with `origin/main`, including
uncommitted edits. `--base <ref>` changes that comparison. `--service` **adds** to
the affected set, not filters it. Unknown paths can select all services; discovery
happens once at startup, so restart your loop when edits introduce another service.
Do not use `--no-watch` for live-update verification. Do not hand-maintain a service
catalog from this feature map.

In another terminal, in the same checkout:

```sh
dev status --owner agent-my-task
dev url --owner agent-my-task '/'
dev url --owner agent-my-task --host account.tadoku.dev.lab '/login'
dev url --owner agent-my-task --host admin.tadoku.dev.lab '/'
dev logs --owner agent-my-task tadoku-api
```

Use the **printed URLs**, including for deep links. Flags precede positional
paths/service names. The `dev-branch` query sets a host-only cookie; it remains
in the URL and wins over an older cookie. Browser routing headers cannot override
the cookie. A printed URL does not create an overlay or prove it is ready.

Select every host the journey uses in the **same browser context**. In particular,
admin → API needs both the admin link and the main-host link. Tabs share cookies;
use separate contexts/profiles for base and other owners. Authentication cookies
and branch selection are different things. `/kratos` always uses shared Kratos.

## Drive and observe

- Open the selected link with a browser tool or Playwright. Use labels/roles and
  visible navigation. Inspect the rendered page and relevant network responses;
  do not invoke internal React handlers or mock the API to claim an E2E pass.
- Before writes, confirm the API response's `X-Dev-Backend` identifies your overlay,
  not base. `X-Dev-Selected` is only intent; `X-Dev-Proxy-Backend` can describe outer
  Oathkeeper rather than the final API. Check the document response for frontend
  changes and actual API responses for backend changes. Unselected base routes
  need not emit these headers. Cold-start convergence may briefly serve base.
- Exercise the relevant feature-map steps, including a reload or a second view
  for persisted changes. Frontend-only overlays still use base data unless an API
  overlay is also selected. Keep writes in your branch DB. Kratos, Keto and feature
  providers remain shared: use owned disposable identities for account/role tests,
  never change another developer's account or shared flag policy for convenience.
- For live frontend edits, keep the page mounted, edit again and prove HMR without
  navigation; record unchanged Pod UID/image. For Go edits, observe rebuild,
  supervised restart and readiness, then exercise the changed behavior. Compile
  errors leave the previous process running; successful restarts can briefly 503.
  A replacement Pod is resynced only while the local loop runs; without it, the
  image baseline returns. More routing/sync gates: [acceptance](../../../.dev/acceptance.md).
- Use pnpm frontend checks and Bazel backend checks from `AGENTS.md`. Backend
  destructive reset/reseed suites use disposable local test databases, never the
  shared development server. Existing [auth browser tests](../../../frontend/apps/auth/e2e/README.md)
  exercise registration/recovery; they do not select a branch automatically and
  their Playwright config bypasses TLS verification. Report that limitation and
  any skipped tests instead of claiming full coverage.

If the wrong backend, auth failure, missing data or sync error blocks proof,
capture status and scoped logs first. Fix within the assigned change or report the
exact unmet prerequisite; don't broaden access or erase data to make a check pass.

## Evidence and cleanup

Retain repeatable evidence outside tracked source: tested SHA and uncommitted diff,
CLI version, commands, owner/branch/base, selected links, actual backend identities,
fixture IDs, browser actions, expected/actual results, screenshot or trace and
relevant sanitized logs. Record test skips, TLS bypasses and unverified boundaries.
Do not save cookies, tokens, Secret contents or authenticated browser storage in
Git or publicly shared artifacts. One-off browser scripts/screenshots stay outside
source commits. Capture trigger and outcome, not merely a loaded page.

For PR verification, follow [Publishing PR evidence](references/pr-evidence.md):
attach useful screenshots and workflow recordings directly to the authorized PR,
alongside the tested revision, steps, assertions and limitations. Inspect media
before uploading and verify the posted attachments. Don't substitute local file
paths for delivered evidence, or claim an attachment was posted when upload failed.

When finished, remove your overlays from the same branch/checkout/owner:

```sh
dev down --owner agent-my-task
dev url --owner agent-my-task --clear '/'
dev url --owner agent-my-task --host account.tadoku.dev.lab --clear '/'
dev url --owner agent-my-task --host admin.tadoku.dev.lab --clear '/'
```

Visit the clear links if keeping that browser context; don't revisit stale selected
links. Confirm owned overlay removal and base availability. `dev down` stops the
local loop; Ctrl-C alone leaves overlays. Branch databases are intentionally
retained. Don't drop databases/PVCs or run namespace-wide deletion/cleanup.
`dev status` and `dev up` can maintain expired overlays; `dev cleanup` is broader
than the current owner. Stop only port-forwards/browser processes you started.

If the user wants the environment left running, hand off the links, owner, branch,
worktree, process handle, sync health, expiry and exact cleanup command instead.
Report residual fixture data. Update the relevant map in the same PR when a route,
precondition or behavior changes; distinguish observed behavior from untested checks.
