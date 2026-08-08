# Phase 1 deployment gate

Date: 2026-08-08  
Status: passed

## Outcome

The Paper foundation and its independent catalogue are live at
`https://paper.tadoku.app`. Argo CD owns the workload, the certificate is
trusted, the running pod matches the immutable release digest, and the legacy
catalogue remains independently available at `https://ui.tadoku.app`.

The first production observation found that nginx's SPA fallback returned the
harmless application shell for scanner probes such as `/.env`. This failed the
delivery gate. A test that reproduced the response was added first, nginx was
hardened, and the corrected image was published and deployed before Phase 1
was accepted.

## Release record

| Field | Value |
| --- | --- |
| Target | Paper catalogue |
| Tadoku source | `6daba6a8267c30fb9df285b933175bed7d27fae7` |
| Source PRs | `tadoku/tadoku#780`, hardening `tadoku/tadoku#781` |
| Incoming image | `ghcr.io/tadoku/tadoku/frontend-paper-styleguide@sha256:6a67d1f77c1b124f79fa1b6e9fe283443517da41de5d40322c444cf0e5e0d0a3` |
| Outgoing Paper image | `ghcr.io/tadoku/tadoku/frontend-paper-styleguide@sha256:c496b05831d63bdb59459a9fcec3aaed305113ac2673dff49a9c450280d0c176` |
| GitOps introduction | `416a489f42074c12fa48cb956cd9078ebaae3138` (`antonve/tadoku-argocd#81`) |
| GitOps deployed revision | `c548d7c369e588223ccc07e4748d8b6ca0c7d20e` |
| Argo CD Application | `tadoku-paper-styleguide` |
| Namespace / Deployment | `tdk-prod-paper-styleguide` / `tadoku-paper-styleguide` |
| Image Updater identity | `tadoku-paper-styleguide` / `paper-styleguide` |
| Rollback | Pause the Paper updater mapping and pin the outgoing digest, following `rollback-runbook.md` |
| Operator / time | Codex, 2026-08-08 UTC |

## Production acceptance

| Check | Evidence | Result |
| --- | --- | --- |
| Reconciliation | Argo CD `Synced Healthy` at `c548d7c` | Pass |
| Runtime identity | One ready pod, zero restarts, exact incoming `imageID` | Pass |
| TLS | Let's Encrypt YR2 certificate for `paper.tadoku.app`, valid through 2026-11-06 | Pass |
| Canonical routes | `/`, `/foundations/color`, `/governance/contributing`, and `/components/actions/button` returned `200` | Pass |
| Sensitive paths | `/.env`, `/.git/config`, `/api/.env`, and `/config.env` returned `404` | Pass |
| Static behavior | HTML uses `no-cache`; hashed JavaScript is gzip-compressed and immutable | Pass |
| Health and assets | `/healthz` healthy; missing assets return `404` | Pass |
| Coexistence | `ui.tadoku.app` returned `200`; its ready pod retained the legacy image and zero restarts | Pass |

The legacy catalogue rollback identity observed at acceptance was
`ghcr.io/tadoku/tadoku/frontend-styleguide@sha256:8f93282ca20960a0fc4b3e82e69407424689c1af50c61e7e6cfee7129fc9b762`.
No legacy host, route, or application migration was performed.

## Production resource snapshot

The corrected pod used 4,866,048 bytes current memory and 8,228,864 bytes
peak memory. All memory event counters were zero. Its cgroup reported 107,943µs
CPU use across 19 periods with zero throttled periods. The pod was ready with
zero restarts under the selected `10m`/`250m` CPU and `32Mi`/`64Mi` memory
request/limit envelope.

This snapshot confirms the initial envelope; Phase 4 repeats the longer idle,
sustained, burst, and recovery protocol against the complete catalogue.

## Gate decision

Phase 1 passes. Phase 2 may implement the end-to-end Button, Input, Modal, and
ActionMenu slice. Production application migration remains prohibited until
the later application phases.
