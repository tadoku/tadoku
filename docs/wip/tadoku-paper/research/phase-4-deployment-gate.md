# Tadoku Paper Phase 4 deployment gate

Status: **passed on 2026-08-08**

The complete report is available as [the repository HTML artifact](../artifacts/tadoku-paper-phase-4-gate.html) and [the Presentr copy](https://presentr.lab/html/tadoku-paper-phase-4-gate). Raw counter snapshots and request summaries are preserved in [phase-4-resource-observations.md](phase-4-resource-observations.md); the repeatable driver is [scripts/measure-paper-resources.sh](scripts/measure-paper-resources.sh).

## Production identity

- Tadoku commit: `f297aa7899d54dba2302d6d26f377fe9fe9d570b`
- Image: `ghcr.io/tadoku/tadoku/frontend-paper-styleguide:prod@sha256:4f15ab54ef5d6844df798ad3b08530f1a162ee0c6424a77977a362786a07b714`
- Base image: `nginxinc/nginx-unprivileged:1.27-alpine@sha256:65e3e85dbaed8ba248841d9d58a899b6197106c23cb0ff1a132b7bfe0547e4c0`
- Static output: 10 files, 3,696,801 uncompressed bytes
- Registry image: linux/amd64 manifest `sha256:43f3a5e40be55ac899c8bf45c7c4862617baecd4d2a29c9db46b574b5349b19c`, 22,325,290 compressed layer bytes
- Deployment: `lke20948-ctx/tdk-prod-paper-styleguide/tadoku-paper-styleguide`, rolling update with `maxUnavailable: 0`

## Gate result

- Source gates: 73 `paper-ui` tests, 31 styleguide tests, both lint/typecheck/build suites, TS 4.9 declaration consumer, pack smoke, Paper boundary scan, seven-project frontend build, production image build and image smoke passed.
- Production content: all 40 canonical routes, static assets, SPA fallback, nested routes, health, caching, gzip, immutable assets, dot-path rejection, and missing-asset behavior passed.
- Production visual/interaction: desktop plus 320×700 mobile, no horizontal overflow, exact 360px fixture width, drawer/search focus behavior, keyboard search navigation, persistent workbench panels, copy feedback, dark wordmark, and direct hash positioning passed.
- Three production-shaped fresh-pod runs completed 49,456 requests with zero failed responses, restarts, OOM/pressure events, probe warnings, or readiness loss. Three observer-neutral navigation controls added 60 successes and zero throttled periods.
- Container readiness was 2s, 2s, and 1s. Worst sustained CPU was `0.341m`; worst burst throttling was `0.17%`; worst memory peak was 11.48 MB; worst recovery delta was +4.2%.

## Resource decision

| Resource | Formula | Selected |
| --- | ---: | ---: |
| CPU request | below 5m; 10m floor | `10m` |
| CPU limit | 110m from valid partial p99 | `250m` |
| Memory request | 12Mi; 32Mi floor | `32Mi` |
| Memory limit | 24Mi; 64Mi floor | `64Mi` |

The existing values stay unchanged. The cluster has no Metrics API and reset the continuous exec sampler after 27 valid one-second samples; its p99 was `52.626m`, and complete boundary counters show burst averages of `37.9m`, `42.7m`, and `46.4m`. The safer `250m` limit remains until complete telemetry justifies lowering it. No GitOps resource PR is needed.

## Stop boundary

`paper.tadoku.app` is the authoritative Paper catalogue. `ui.tadoku.app` remains the independent legacy catalogue. Admin, Auth, webv2, and production application migration work have not started. Phase 5 requires a new explicit authorization.
