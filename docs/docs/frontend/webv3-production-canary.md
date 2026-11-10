---
title: Webv3 production canary
description: The signed-in journey that checks tadoku.app on a schedule, the account and secrets it needs, how it avoids writes, and where its green or red result is read.
sidebar_position: 3
---

# Webv3 production canary

Read this when you change the production canary, create its account or read its
result.

`.github/workflows/webv3-production-canary.yaml` signs in on
`https://tadoku.app` through authv2 and runs
`frontend/apps/webv3/e2e/production-canary.spec.ts`. The journey loads the
home page, an official contest, the account's profile and `/logs/new`.

## When it runs

The workflow runs hourly, the same interval as the production Image Updater that
rolls out the `prod` image, so a run follows each sync. GitHub starts this
repository's schedules late and can skip them. A `workflow_dispatch` runs it at
once, for example after a rollout, and only from `main`.

It does nothing until the repository variable `WEBV3_CANARY_ENABLED` is `true`.

## Account and secrets

| Item | Where | Value |
| --- | --- | --- |
| Account | Production identity, registered on `account.tadoku.app` | A dedicated account that joins no contest and holds no real activity |
| `WEBV3_CANARY_EMAIL` | Secret in the `production-canary` GitHub environment | The account's email |
| `WEBV3_CANARY_PASSWORD` | Secret in the same environment | The account's password |
| `production-canary` | GitHub environment, deployment branches limited to `main` | Holds the secrets above |
| `WEBV3_CANARY_ENABLED` | Repository variable | `true` once the secrets exist |

Once enabled, missing secrets fail the run rather than skipping the journey.

## No writes

From its first request, the journey refuses every request that is not a read,
except the login POST. A refused request fails the run, so the account's
activity and every standing stay unchanged. Signing in creates one session per
run; sessions expire on their own.

Only failure screenshots are uploaded, for seven days. The HTML report is kept
out of the artifact because it records typed values.

## Reading the result

| Reader | Source |
| --- | --- |
| Dashboards | `conclusion` of the newest completed run at `GET https://api.github.com/repos/tadoku/tadoku/actions/workflows/webv3-production-canary.yaml/runs?branch=main&status=completed&per_page=1`; `success` is green and anything else is red |
| Alerts | Notify topic `webv3-production-canary`: severity `info` for green and `critical` for red, each linking to the run |
| Auto-revert alert | Gauges `tadoku_webv3_canary_success` and `tadoku_webv3_canary_last_run_timestamp_seconds`, sent to the homelab's Alloy gateway after every run that reached a result |

A red result delivers through Notify's critical channels. Only the journey's own
result reaches the gauges: a run that fails before the journey, or is cancelled,
reports nothing there. A red gauge fires `TadokuWebv3CanaryFailed`, which reverts
Webv3; setting `WEBV3_CANARY_ENABLED` to anything else stops the gauges, so
silence `TadokuWebv3CanaryStale` first.

## Running it against a branch

From `frontend/`, against a branch tenant you own, with an account that exists
in that tenant:

```sh
WEBV3_BASE_URL=https://<route>.next.tadoku.dev.lab \
TADOKU_API_URL=https://<route>.tadoku.dev.lab/api/internal \
E2E_READER_EMAIL=<email> E2E_READER_PASSWORD=<password> \
NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/lab-root-ca.crt \
pnpm --filter webv3 exec playwright test e2e/production-canary.spec.ts
```
