---
title: Job worker configuration
description: The separate job worker's environment variables, queue and handler tenant scope, private listeners, cache behavior, concurrency and metrics.
sidebar_position: 9.1
---

# Job worker configuration

Read this when you change the job worker's settings, its tenant scope, its
listeners or its metrics. `services/tadoku-api/cmd/tadoku-worker` runs the
worker, and development values are in
`k8s/dev/base/services/tadoku-worker.yaml`.

The queue persists in the `jobs` table. See
[Successful-job retention](./jobs-operations.md#successful-job-retention) for
the automatic three-calendar-month policy. Failed records are retained
indefinitely.

`cmd/tadoku-worker` constructs `app/worker.Application` with the queue and
business features. Its immutable typed registration drives both claiming and
dispatch; [Jobs and worker](./jobs.md) describes publication, execution policies,
replay and consumer-first version migrations.

The worker uses `WORKER_POSTGRES_*` split connection configuration,
`WORKER_POSTGRES_MAX_CONNECTIONS` (default 4, range 1–32), `WORKER_VALKEY_URL`,
`WORKER_VALKEY_TIMEOUT` (default 1s),
`WORKER_DIAL_TIMEOUT` (default 3s), `WORKER_CONCURRENCY` (default 4), and
`WORKER_SHUTDOWN_TIMEOUT` (default 15s). Concurrency and shutdown timeout must
be positive; the command loads and validates both before application startup.

`WORKER_BRANCH` defaults to empty, selecting the base worker's all-tenants queue
scope, excluding tenants listed in `tenant_overrides` for `WORKER_COMPONENT`.
`WORKER_COMPONENT` defaults to `tadoku-worker`; its name must start with a
lowercase letter and contain only lowercase letters, digits or hyphens. An
explicit empty component is invalid.

A non-empty `WORKER_BRANCH` must be a valid `<name>/<id>` tenant key other than
`tadoku/prod`. Its scope claims only that tenant, including when an override is
present. Only base backlog metrics exclude overridden tenants. Base completed-job
cleanup stays global; branch cleanup stays tenant-local. Branch queue isolation
requires the `jobs` row-level security policy, so keep this variable unset when that policy is absent. Every handler
and lease transition still uses the tenant persisted on its claimed job.

Private health and metrics listeners default to `WORKER_PORT=8000` and
`WORKER_METRICS_PORT=9090`. It has no public route. The API and worker must use
the same database and tenant-derived cache keys. Development branches use
their route's full tenant key even when a migration profile selects a private
database.

The worker invalidates leaderboard caches through registered jobs. A cache miss
rebuilds from PostgreSQL only if its generation has not changed, and cached
reads recheck that generation before returning; unavailable Valkey falls back
to PostgreSQL. Worker Pod readiness reports whether the execution loop can
operate, independently of individual job success. Monitor queued and failed
jobs because cached results can remain stale while invalidation work is
outstanding.

Global and per-type concurrency are per process. Handler cancellation does not
release its slot until it returns. On shutdown the application stops claiming,
drains within its configured bound, then cancels remaining work and joins it
before provider resources close. A noncooperative handler can delay exit; see
[Execution and failure guarantees](./jobs.md#execution-and-failure-guarantees).

## Worker tenant scope

The worker separates queue scope from handler scope. Queue claims, backlog
metrics and completed-job cleanup receive the required `jobqueue.Scope`; each
handler and fenced transition receives only its persisted job tenant. Database
helpers never supply a default tenant. The replay command uses the explicit
canonical `tadoku/prod` context.

## Worker metrics

The separate worker serves its own metrics on `WORKER_METRICS_PORT`. Its
backlog gauges refresh at startup and every fifteen seconds. The `type` label
is a registered job type; unsupported-type gauges aggregate across unknown
types without an unbounded type label.

| Metric | Labels | Meaning |
| --- | --- | --- |
| `tadoku_worker_in_flight` | `type` | Jobs executing in this process. |
| `tadoku_worker_failed_attempts_total` | `type`, `code`, `tenant_kind` | Failed handler attempts for the persisted job tenant. |
| `tadoku_worker_handler_duration_seconds` | `type`, `tenant_kind` | Histogram of handler execution duration before the job transition. |
| `tadoku_worker_pending_jobs` | `type` | Pending jobs for each registered type. |
| `tadoku_worker_failed_jobs` | `type` | Terminally failed jobs for each registered type. |
| `tadoku_worker_oldest_due_age_seconds` | `type` | Age of the oldest due or expired job. |
| `tadoku_worker_unsupported_pending_jobs` | None | Pending jobs with an unsupported type. |
| `tadoku_worker_unsupported_running_jobs` | None | Running jobs with an unsupported type. |
| `tadoku_worker_unsupported_failed_jobs` | None | Failed jobs with an unsupported type. |
| `tadoku_worker_unsupported_oldest_due_age_seconds` | None | Age of the oldest due unsupported job. |
| `tadoku_worker_expired_leases_total` | `type` | Expired running leases reclaimed. |

## Email delivery

Sending defaults to disabled. `WORKER_EMAIL_ENABLED=true` constructs the SMTP
sender; disabled deployments need no SMTP settings and contact no mail server.
Even when enabled, only `tadoku/prod` rows can send. Configure:

- `WORKER_EMAIL_ADDRESS`: SMTP host and port supporting STARTTLS.
- `WORKER_EMAIL_USERNAME` and `WORKER_EMAIL_PASSWORD`: Secret references to the
  existing SES SMTP credential used by Kratos. Never log or commit these values.
- `WORKER_EMAIL_SENDER`: default `no-reply@tadoku.app`.
- `WORKER_EMAIL_SOURCE_ARN`, `WORKER_EMAIL_FROM_ARN` and
  `WORKER_EMAIL_RETURN_PATH_ARN`: the three matching `X-SES-*-ARN` headers from
  the production Kratos configuration.
- `WORKER_EMAIL_TIMEOUT`: positive per-send bound, default 10s, including
  connection, TLS, authentication and response reads.

Enabled deployments reject missing SMTP credentials, malformed addresses,
missing SES headers or non-positive timeout before opening runtime resources.
Production manifests and the namespace-local Secret prerequisite belong in
`tadoku-argocd`; the maintainer controls activation. The development base and
branch workers bind `WORKER_EMAIL_ENABLED=false` explicitly; see
[Development email](../operations/development-base.md#development-email).
[Retained email](./email.md) documents delivery and audit rules.
