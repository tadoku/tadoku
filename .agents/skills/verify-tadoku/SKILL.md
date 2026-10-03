---
name: verify-tadoku
description: Verify Tadoku application changes in the shared development environment using dev-cli branch overlays and real browser journeys. Use for frontend/backend live edits, routing checks, screenshots and verification handoffs; not production deployment or environment bootstrap.
---

# Verify Tadoku

Run commands from the repository root unless stated otherwise. No global skill
or conversation history is required. If automatic skill discovery is unavailable,
read this file directly. This skill guides verification; it does not grant access,
authorize unrelated mutations, or replace the repository's `AGENTS.md`.

[Development environment](../../../docs/docs/develop/environment.md) is the
canonical reference for installing dev-cli, starting, opening and inspecting a
branch, routing headers, branch tenants, fixture accounts and cleanup. This
skill adds only what verification needs on top of it.

## Choose the proof

Read the [feature index](references/features/README.md), then the relevant area.
Identify the user action, expected visible result, persistence or downstream
effect, and the failure/empty/permission state the change could break. For a bug,
reproduce before changing it. A successful build or HTTP 200 alone is not proof.

Use the existing dev-cli, browser automation, and repository tests. Do not invent
another router, wrapper CLI, fixture service or authentication bypass. Preserve
the real login → frontend → gateway → API path. Choose meaningful affected
journeys rather than running the entire inventory for every edit.

## Establish the environment

1. Inspect `git status --short`, `git branch --show-current` and `git rev-parse HEAD`.
   Preserve other work. Fetch `origin/main` when available; record a stale base if
   fetching fails. Use a task branch and a unique stable owner for this checkout.
   Check for an existing `dev up` loop belonging to this exact checkout before
   starting another; reuse only your own matching loop.
2. Run `dev version` (v0.7.0 or newer) and `dev doctor`. Installation, prerequisites and overrides
   are in [Development environment](../../../docs/docs/develop/environment.md).
   Lab networking, DNS, CA trust and existing cluster credentials must already
   work. Doctor is a read-only prerequisite check, not an E2E.
3. Read [`.dev/config.yaml`](../../../.dev/config.yaml). It targets `homelab-talos-dev`.
   Always specify `--context homelab-talos-dev` on kubectl commands. If access or the base is broken,
   report the failing check; do not bootstrap credentials, restart shared services,
   change Argo resources, or substitute production. Operators have a separate
   [Development base](../../../docs/docs/operations/development-base.md) runbook.

Verification fixtures and their caveats live in the [feature index](references/features/README.md).
Do not run `make dev-seed` automatically: it mutates shared identities and base
fixtures. Use it only when missing fixtures need an authorized shared setup.

## Start the changed code

Make the intended edits **before** starting the loop, then start it as described
in [Start a branch](../../../docs/docs/develop/environment.md#start-a-branch). Use
an owner such as `agent-my-task` consistently, including in other terminals.

- Keep the loop running in a durable terminal/session; record its handle and
  working directory.
- Run `dev up --owner <owner>` without migration or seed task flags. Its
  lifecycle hook provisions and seeds `tadoku/<route>`, including frontend-only
  branches. Frontend-only writes use that tenant through the base API. Add an
  API overlay only when the changed API code is part of the proof.
- Discovery happens once at startup: restart your loop when edits introduce
  another service. Do not use `--no-watch` for live-update verification. Do not
  hand-maintain a service catalog from the feature map.
- Migration paths automatically select `isolated-database` and its migrated
  `tadoku-<route>` database. For migration-only changes add `--service tadoku-api`
  to select the paired API/worker. Follow the migration gates in `.dev/acceptance.md`:
  prove the scratch version, owner/marker, API persistence in that database,
  refusal to switch an active profile and down using its recorded database.
  Keep the retained database; no removal approval is implied by verification.

Use the **printed `dev url` links**, including for deep links; a printed link does
not create an overlay or prove it is ready. Select every host the journey uses in
the **same browser context**. In particular, admin → API needs both the admin
link and the main-host link. Tabs share cookies; use separate contexts/profiles
for base and other owners. Authentication cookies and branch selection are
different things.

## Drive and observe

- Open the selected link with the product's browser tool where available. Use labels/roles and
  visible navigation. Inspect the rendered page and relevant network responses;
  do not invoke internal React handlers or mock the API to claim an E2E pass.
- Before writes, confirm `X-Dev-Selected` on the API response identifies your
  route, and select the main hostname even for auth or admin work. Require an
  overlay `X-Dev-Backend` when verifying changed API code; a frontend-only branch
  intentionally uses the base API. `X-Dev-Selected` alone does not prove persisted
  isolation: verify the result's tenant when testing routing. `X-Dev-Proxy-Backend` can describe outer
  Oathkeeper rather than the final API. Check the document response for frontend
  changes and actual API responses for backend changes. Unselected base routes
  need not emit these headers. Cold-start convergence may briefly serve base.
- Exercise the relevant feature-map steps, including a reload or a second view
  for persisted changes. Keep writes in your branch tenant. Kratos identities
  remain shared: use fixture accounts or owned disposable identities for
  account/role tests without changing shared credentials or canonical roles.
  Restrict Keto and Flipt changes to your tenant object and
  `test/tadoku_<route>` namespace. Discover branch fixture IDs from its UI or
  API; canonical fixture UUIDs do not identify branch contests or logs.
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
put one evidence block in the authorized PR description, with a results table,
visible coverage warning and collapsed media and reproduction details. Attach
useful screenshots and workflow recordings directly to that PR. Inspect media
before uploading and verify the posted attachments. Don't substitute local file
paths for delivered evidence, or claim an attachment was posted when upload failed.

When finished, from the same branch/checkout/owner, run `dev down` and clear the
selection on every host you selected, as described in
[Clean up](../../../docs/docs/develop/environment.md#clean-up). Visit the clear
links if keeping that browser context; don't revisit stale selected links.
Confirm owned overlay removal, tenant/provider cleanup and base availability.
Down stops pods, clears worker overrides and removes branch data through the
recorded lifecycle hooks. If it fails, retain its marker, inspect the task logs
and retry. Ctrl-C alone leaves overlays. Don't drop databases/PVCs or run namespace-wide deletion/cleanup;
`dev cleanup` is broader than the current owner. Stop only port-forwards/browser
processes you started.

If the user wants the environment left running, hand off the links, owner, branch,
worktree, process handle, sync health, expiry and exact cleanup command instead.
Report residual fixture data. Update the relevant map in the same PR when a route,
precondition or behavior changes; distinguish observed behavior from untested checks.
