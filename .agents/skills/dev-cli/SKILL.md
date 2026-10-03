---
name: dev-cli
description: Run Tadoku branch overlays with dev-cli on homelab-talos-dev. Use for live service edits, tenant lifecycle tasks, links, status, logs and owner-scoped cleanup; use verify-tadoku for browser journeys and PR evidence, not production deployment.
---

# Develop with dev-cli

Run commands from the repository root. Read [Development environment](../../../docs/docs/develop/environment.md) for installation, configuration, branch tenants, routing and cleanup. Read [`.dev/config.yaml`](../../../.dev/config.yaml) before using the shared cluster. Use `--context homelab-talos-dev` on kubectl commands. This skill does not grant cluster access or authorize changes to the shared base.

## Start the branch

1. Inspect Git status, branch and revision. Preserve existing work and check for a `dev up` loop belonging to this exact checkout. Fetch `origin/main` when available; record a stale base if fetching fails. Use a stable owner unique to this checkout, consistently with `--owner` or `DEV_OWNER`.
2. Check `dev version` (v0.7.0 or newer for lifecycle hooks and configuration variables) and `dev doctor`. Doctor checks prerequisites only. Follow the environment guide if dev-cli needs installation; do not expose credentials or kubeconfigs.
3. Make the intended service edits before `dev up`. dev-cli selects Bazel deployables once at startup from changes against the merge base, including uncommitted edits. Unknown paths select all deployables. If a later edit adds a service, restart the loop. `--service` adds a service; `--no-watch` does not provide live updates.
4. Run `dev up --owner <owner>` in a durable terminal without migration or seed task flags. The `tenant` hook provisions and seeds `tadoku/<route>` before startup; frontend-only writes use that tenant through the base API. Add `--service tadoku-api` only when changed API code needs an overlay. Shared fixture identities must already exist; do not run `make dev-seed` automatically because it changes shared identities and base data.

Changes under `services/tadoku-api/migrations/` automatically select
`isolated-database`, which runs migrate before tenant provisioning against
`tadoku-<route>`. A migration-only change selects no deployable; run
`dev up --owner <owner> --service tadoku-api` to start its paired API/worker.
Follow [Migration branches](../../../docs/docs/develop/environment.md#migration-branches).
A live profile switch is refused until down; after rebase, down still uses the
recorded isolated database, removes its tenant and retains the marked database.
Legacy-owned retained databases fail closed and require approved inventory/removal.

Keep the loop, owner, checkout and terminal handle together. Do not start another loop for the same checkout and route. Frontend source sync drives HMR; backend edits rebuild the affected Bazel binary and restart its supervised process. Read loop errors, `dev status --owner <owner>` and `dev logs --owner <owner> <service>` before diagnosing a failed update.

## Open and inspect

Use the actual links printed by `dev up` or `dev url --owner <owner> '/desired/path'`. Add `--host account.tadoku.dev.lab` or `--host admin.tadoku.dev.lab` for those hosts. Quote paths containing `?`, `&` or `#`; flags precede positional paths. Open a selection link on every host the journey uses. Selection is a cookie per hostname and is shared across tabs in one browser profile; use separate profiles for simultaneous branches.

A printed URL does not prove the overlay is ready. Check status, then confirm actual frontend and API responses. `X-Dev-Selected` describes intent; `X-Dev-Backend` identifies the upstream that served the request. Cold-start convergence can briefly serve base. For browser journeys, persisted results and PR evidence, follow [verify-tadoku](../verify-tadoku/SKILL.md) and its feature map. Do not claim HMR from a source transfer or backend identity from an HTTP 200 alone.

Select the main-host link before admin or auth API writes, as well as that app's
link. Verify `X-Dev-Selected` on the API response: host selection determines the
signed tenant even when a frontend-only branch uses the base API. Branch data
normally uses shared `tadoku`; migration branches use `tadoku-<route>`. Both use
tenant-scoped Valkey keys and Flipt namespace
`test/tadoku_<route>`. Reseed only your tenant with
`dev task --owner <owner> tenant`; shared Kratos identities stay unchanged.

Selecting either API or worker starts both. `worker-override-set` runs before
the branch worker starts and `worker-override-clear` after its pods stop. Do not
manually clear an override while its worker is running.

## Hand off or clean up

If asked to leave the environment running, return its clickable links, owner, branch, checkout, loop handle, sync health, expiry and exact cleanup command. Otherwise run `dev down --owner <owner>` from the same checkout and clear each selected host with `dev url --owner <owner> --clear '/'` (adding `--host` as needed). Ctrl-C alone leaves overlays. Down removes the branch's provider state and cascading tenant data after its pods stop; it preserves the Argo CD base and retained databases, including the current migration branch's database.

`dev cleanup` reaches expired environments across owners and replays their
recorded hooks and variables. `dev status` runs no lifecycle hooks and reports
marked expired routes as `teardownPending`. A failed teardown retains its
marker; inspect task logs and retry down or cleanup. Missing recorded tasks
cause cleanup to skip the route. Use compatible configuration and never delete
the marker or task Lease to force cleanup. Do not delete shared resources or
retained databases as routine cleanup.
