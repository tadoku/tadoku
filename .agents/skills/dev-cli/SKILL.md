---
name: dev-cli
description: Run Tadoku branch overlays with DevCLI on homelab-dev. Use for live service edits, branch tasks, links, status, logs and owner-scoped cleanup; use verify-tadoku for browser journeys and PR evidence, not production deployment.
---

# Develop with DevCLI

Run commands from the repository root. Read [Development environment](../../../docs/docs/develop/environment.md) for installation, configuration, branch databases, routing and cleanup. Read [`.dev/config.yaml`](../../../.dev/config.yaml) before using the shared cluster. This skill does not grant cluster access or authorize changes to the shared base.

## Start the branch

1. Inspect Git status, branch and revision. Preserve existing work and check for a `dev up` loop belonging to this exact checkout. Fetch `origin/main` when available; record a stale base if fetching fails. Use a stable owner unique to this checkout, consistently with `--owner` or `DEV_OWNER`.
2. Check `dev version` (v0.4.0 or newer) and `dev doctor`. Doctor checks prerequisites only. Follow the environment guide if DevCLI needs installation; do not expose credentials or kubeconfigs.
3. Make the intended service edits before `dev up`. DevCLI selects Bazel deployables once at startup from changes against the merge base, including uncommitted edits. Unknown paths select all deployables. If a later edit adds a service, restart the loop. `--service` adds a service; `--no-watch` does not provide live updates.
4. Run `dev up --owner <owner>` in a durable terminal. For API work, add `--task migrate --task seed` to prepare the branch database. For frontend work that writes data, add `--service tadoku-api --task migrate --task seed` so writes use a branch database. Frontend-only read-only checks can omit those tasks. The seed task requires shared fixture identities; do not run `make dev-seed` automatically because it changes shared identities and base data.

Keep the loop, owner, checkout and terminal handle together. Do not start another loop for the same checkout and route. Frontend source sync drives HMR; backend edits rebuild the affected Bazel binary and restart its supervised process. Read loop errors, `dev status --owner <owner>` and `dev logs --owner <owner> <service>` before diagnosing a failed update.

## Open and inspect

Use the actual links printed by `dev up` or `dev url --owner <owner> '/desired/path'`. Add `--host account.tadoku.dev.lab` or `--host admin.tadoku.dev.lab` for those hosts. Quote paths containing `?`, `&` or `#`; flags precede positional paths. Open a selection link on every host the journey uses. Selection is a cookie per hostname and is shared across tabs in one browser profile; use separate profiles for simultaneous branches.

A printed URL does not prove the overlay is ready. Check status, then confirm actual frontend and API responses. `X-Dev-Selected` describes intent; `X-Dev-Backend` identifies the upstream that served the request. Cold-start convergence can briefly serve base. For browser journeys, persisted results and PR evidence, follow [verify-tadoku](../verify-tadoku/SKILL.md) and its feature map. Do not claim HMR from a source transfer or backend identity from an HTTP 200 alone.

## Hand off or clean up

If asked to leave the environment running, return its clickable links, owner, branch, checkout, loop handle, sync health, expiry and exact cleanup command. Otherwise run `dev down --owner <owner>` from the same checkout and clear each selected host with `dev url --owner <owner> --clear '/'` (adding `--host` as needed). Ctrl-C alone leaves overlays. `dev cleanup` reaches expired environments across owners; branch databases remain after `dev down`. Do not delete shared resources or branch databases as routine cleanup.
