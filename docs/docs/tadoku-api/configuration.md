---
title: Runtime configuration
description: Tadoku API environment variables, startup and shutdown behavior, authentication wiring, the raw Keto and Kratos clients, the leaderboard worker and metrics.
sidebar_position: 9
---

# Runtime configuration

Read this when you change startup, an environment variable, authentication
wiring, a provider client or the leaderboard worker.

`services/tadoku-api/cmd/tadoku-api/main.go` loads the configuration from
environment variables. Development values are in
`k8s/dev/base/services/tadoku-api.yaml`.

## Environment variables

### Server and outgoing HTTP transport

- `API_PORT` (default 8000) and `API_METRICS_PORT` (default 9090).
- `API_REQUEST_TIMEOUT` (default 30s), the shared request deadline.
- `API_DIAL_TIMEOUT` (default 3s), `API_RESPONSE_HEADER_TIMEOUT` (default 10s)
  and `API_IDLE_TIMEOUT` (default 30s) bound the owned outgoing HTTP transport
  shared by the Flipt, Kratos and Keto clients. `API_IDLE_TIMEOUT` is also the
  listeners' idle-connection timeout.
- `API_SHUTDOWN_TIMEOUT` (default 10s).

### PostgreSQL

- `API_POSTGRES_HOST`, `API_POSTGRES_PORT` (default 5432),
  `API_POSTGRES_DATABASE`, `API_POSTGRES_USER`, `API_POSTGRES_PASSWORD` and
  `API_POSTGRES_SSLMODE`. A single `API_POSTGRES_URL` is rejected.
- `services/common/postgresconfig` holds the password and the connection URL
  built from these fields as `Secret` values. Formatting and `log/slog` print
  `[REDACTED]`, and JSON encoding omits them. Call `Reveal` only where the
  value is passed to the PostgreSQL pool or the migration tool.
- `API_POSTGRES_MAX_CONNECTIONS` (default 4, validated range 1–32).

### Valkey and leaderboard caches

- `API_VALKEY_URL`, one standalone TCP URL accepted by `valkey-go`.
- `API_VALKEY_TIMEOUT` (default 1s), the positive bound for each connection and
  handshake attempt and the established-connection keepalive and I/O interval.
Leaderboard keys come from the parsed request or job tenant:
`tenant:<name>/<id>:leaderboard:…`. The API and worker use the same key format;
a missing context tenant fails before any cache operation.

`API_LEADERBOARD_CACHE_PREFIX` and `WORKER_LEADERBOARD_CACHE_PREFIX` default to
empty and remain only for development overlays with private databases whose
requests and jobs still carry `tadoku/prod`. Those overlays set the same
`dev:${DEV_ROUTE}:` prefix, producing
`dev:${DEV_ROUTE}:tenant:tadoku/prod:leaderboard:…`. Keep it unique per database.
Production, the development base and real test tenants leave these settings
unset. Remove both compatibility settings when every private-database overlay
has a distinct signed request tenant and persists that tenant on its jobs.

`services/tadoku-api/infra/valkey/README.md` documents which URL options are
accepted and how commands, timeouts, cancellation and close behave.

### Tenant deployment

`API_BRANCH` is optional and defaults to empty. The base deployment serves every
valid signed tenant. A configured deployment serves only its exact parsed
`<name>/<id>` key and rejects another tenant with `421`. Each segment starts
with a lowercase letter or digit and contains at most 56 lowercase letters,
digits or hyphens. Bare names, malformed keys and `tadoku/prod` are invalid
branch settings and fail startup. The key is not a DNS slug.

`services/tadoku-api/internal/tenant` owns parsing, context propagation and
deployment scoping. Obtain keys through `Parse` or `Production()`; the canonical
production key is `tadoku/prod`, also used by the development base. The Go zero
value is not a tenant and context lookup rejects it. Development and production
base deployments leave `API_BRANCH` unset.

### Authentication

- `API_JWKS`, the gateway's public signing-key URL.
- `API_MAX_TOKEN_AGE` (default 24h), the maximum accepted age since `iat`.
- `API_JWT_ISSUER`, an optional exact issuer match. Empty leaves the issuer
  unchecked.
- `API_OATHKEEPER_AUTHZ_TOKEN`, the bearer credential required on trusted
  Oathkeeper authorization callbacks.

### Keto and Kratos

- `API_KETO_READ_URL`, the Keto read API URL. Ban and administrator checks use
  a separate read-only client with a 2s total request timeout.
- `API_KETO_WRITE_URL`, an absolute HTTP(S) base URL for the Keto write
  service.
- `API_KETO_WRITE_TIMEOUT` (default 2s), a positive total request timeout for
  the raw read/write client, including response-body reads.
- `API_KRATOS_ADMIN_URL`, an absolute HTTP(S) base URL for the Kratos admin
  service.
- `API_KRATOS_TIMEOUT` (default 2s), a positive total request timeout, including
  reading the response body.

For both base URLs, credentials, query strings and fragments are rejected; path
prefixes are supported and trailing slashes are removed.

### Service tokens and feature flags

- `API_OATHKEEPER_URL` (default `http://oathkeeper-proxy.default:4455`), used
  for outgoing service-token exchange. `API_SERVICE_ACCOUNT_TOKEN_PATH`
  defaults to the projected credential at `/var/run/secrets/tokens/token`.
- `API_FLIPT_ENABLED`, `API_FLIPT_URL`, `API_FLIPT_ENVIRONMENT`,
  `API_FLIPT_NAMESPACE`, `API_FLIPT_UPDATE_INTERVAL`,
  `API_FLIPT_REQUEST_TIMEOUT`, `API_FLIPT_STARTUP_TIMEOUT` and
  `API_FLIPT_MANAGEMENT_URL` configure the feature-flag evaluation provider and
  the management client. Evaluation and management exchange separate
  credentials for `flipt-evaluation/tadoku-api` and
  `flipt-management/tadoku-api`. Provider outages keep safe defaults active and
  do not fail startup; later polling recovers without a restart.

### Scoring engine

`API_SCORING_ENGINE_ENABLED` is a required boolean; it is not derived from
Flipt. Missing or malformed values fail startup. Configure it deliberately for
each environment; the development manifests set it to `false`.

## Startup

- Business route registration requires authentication middleware. Startup
  always constructs it from `API_JWKS`; there is no opt-out or feature flag.
- Startup fetches the JWKS within `API_DIAL_TIMEOUT` and pings PostgreSQL
  before opening listeners. Either failure aborts startup.
- After startup, the middleware refreshes the JWKS in the background every hour
  and when a token names an unknown key ID, at most once a minute. Refresh
  failures are logged as `jwks refresh failed`.
- Startup always constructs and owns the raw Valkey client. Invalid Valkey
  configuration or cancelled setup aborts startup. An unavailable server logs a
  warning and starts in degraded mode; the client reconnects on a later command.
- Startup constructs and retains the raw Kratos SDK client and the shared Keto
  read and read/write clients. Constructing them makes no provider request, and
  provider-backed caches stay cold until a request needs them.
- Provider availability and cache refresh completion are not startup or
  health-check gates. Startup and health checks never mutate provider state.
- `/readyz` checks PostgreSQL. `/livez` is independent of dependency health.
  Valkey is deliberately neither a readiness nor a liveness gate.

## Authentication

[Authorization](../architecture/authorization.md) documents JWT verification, the
ban gate, the checker methods, the administrator callback and their error
responses. In addition:

- Authentication also verifies `nbf` when present. The audience is not
  restricted, and JWT parsing applies no role, ban, permission or
  service-audience policy.
- The identity's `CreatedAt` is the token's issue time, not the account's
  creation time.
- Anonymous gateway traffic carries a signed token with subject `guest`, which
  is distinct from a direct request without credentials. Every other subject
  must be a UUID, the Kratos identity ID; a token with any other subject is
  rejected with `401` like any other invalid token. Service tokens are never
  converted into human identities.
- The ban gate performs only the ban lookup, so no unrelated administrator
  lookup can discard a successful ban result. Request deadlines bound the Keto
  call. A failed lookup is kept in the request context, so later authenticated
  and administrator checks return unavailable without another ban query.

## Raw Keto relationship primitive

- The composition root retains a concrete `*ketoclient.Client` from
  `services/tadoku-api/infra/keto` on `application.keto`. It feeds the concrete
  relationship mutation services.
- A bounded `NewReadClient` instance, with no write API, feeds shared role
  facts, request ban checks and actor administrator checks.
- Pass concrete clients and shared application services explicitly into
  consumers.
- `keto.NewClient(readURL, writeURL, keto.WithHTTPClient(httpClient))` applies
  options to both APIs; the caller owns the HTTP client and transport. Without
  options, the constructor keeps the SDK defaults. Tadoku API supplies its
  bounded HTTP client.

Use the existing primitives with the operation's caller context and a complete
tuple target. Set exactly one of `Subject.ID` or `Subject.Set`:

```go
subject := keto.Subject{ID: subjectID}
err := client.AddRelation(ctx, namespace, object, relation, subject)
// Delete only this namespace/object/relation/subject tuple.
err = client.DeleteRelation(ctx, namespace, object, relation, subject)

group := keto.Subject{Set: &keto.SubjectSet{
    Namespace: groupNamespace,
    Object:    groupID,
    Relation:  membershipRelation,
}}
err = client.AddRelation(ctx, namespace, object, relation, group)
err = client.DeleteRelation(ctx, namespace, object, relation, group)
```

- Direct subjects use `subject_id`; subject sets keep their namespace, object
  and relation. Delete sends every target component, including every subject-set
  field.
- Add treats HTTP 409 as success and delete treats HTTP 404 as success.
- Other errors keep their wrapped provider error: `errors.As` can inspect
  `*ketoapi.GenericOpenAPIError` and its `Body()` and `Model()`, and `errors.Is`
  identifies `context.Canceled` and `context.DeadlineExceeded`. The primitives
  return only an error, without separate HTTP response metadata.
- The total client timeout, any earlier caller deadline and caller cancellation
  bound requests, including response-body reads. There is no mutation retry
  loop; a timeout or cancellation does not establish whether Keto committed the
  write.
- Application operations remain responsible for actor authorization and for
  ordering feature work around these primitives. Audit services own recording
  details such as business timestamps and persistence.

## Raw Kratos primitive

- The composition root keeps a concrete `*kratosapi.APIClient` on
  `application.kratos` for startup integration checks. Identity features use
  the `services/tadoku-api/infra/kratos` client for lookups and cursor pagination.
- `services/tadoku-api/infra/kratos.NewAPIClient(baseURL, kratos.WithHTTPClient(httpClient))`
  constructs the pinned `github.com/ory/kratos-client-go` v0.11.1 SDK. It
  neither validates deployment configuration nor owns the supplied HTTP client.
  Tadoku API validates the configuration and supplies a client with a total
  timeout on its owned transport.
- `NewClient(baseURL, kratos.WithHTTPClient(httpClient))` uses the same client
  for SDK operations and cursor pagination. Without the option, both
  constructors use the SDK's default HTTP client.

Call the SDK directly with the operation's caller context:

```go
identity, response, err := kratos.IdentityApi.GetIdentity(ctx, identityID).Execute()
```

- `response` can be nil for transport and cancellation failures. When present,
  its status and headers remain available even with an error.
- Use `errors.As` to inspect `*kratosapi.GenericOpenAPIError`, including its
  `Body()` and `Model()`, and `errors.Is` for `context.Canceled` or
  `context.DeadlineExceeded`.
- The raw client returns the SDK's models, response metadata and errors
  unchanged. It applies no domain mapping, trait policy, account-age rules, or
  not-found and idempotent-delete translations; consuming features own those
  decisions and any cache lifecycle.
- Use the pinned SDK's request builders and pagination options directly.
- The total client timeout and any earlier caller deadline bound requests;
  caller cancellation also interrupts response-body reads.

## Separate job worker

The queue persists in the `jobs` table. See
[Successful-job retention](./jobs.md#successful-job-retention) for the automatic
three-calendar-month policy. Failed records are retained indefinitely.

`cmd/tadoku-worker` constructs `app/worker.Application` with the queue and
business features. Its immutable typed registration drives both claiming and
dispatch; [Jobs and worker](./jobs.md) describes publication, execution policies,
replay and consumer-first version migrations.

The worker uses `WORKER_POSTGRES_*` split connection configuration,
`WORKER_POSTGRES_MAX_CONNECTIONS` (default 4, range 1–32), `WORKER_VALKEY_URL`,
`WORKER_VALKEY_TIMEOUT` (default 1s), `WORKER_LEADERBOARD_CACHE_PREFIX`,
`WORKER_DIAL_TIMEOUT` (default 3s), `WORKER_CONCURRENCY` (default 4), and
`WORKER_SHUTDOWN_TIMEOUT` (default 15s). Concurrency and shutdown timeout must
be positive; the command loads and validates both before application startup.

`WORKER_BRANCH` defaults to empty, selecting the base worker's all-tenants queue
scope. A non-empty value must be a valid `<name>/<id>` tenant key other than
`tadoku/prod`. Branch queue isolation requires the `jobs` row-level security
policy, so keep this variable unset when that policy is absent. Every handler
and lease transition still uses the tenant persisted on its claimed job.

Private health and metrics listeners default to `WORKER_PORT=8000` and
`WORKER_METRICS_PORT=9090`. It has no public route. The API and worker must use
the same database and tenant-derived cache keys. Private development databases
that still use `tadoku/prod` also require the matching compatibility prefix
[described above](#valkey-and-leaderboard-caches).

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

### Worker tenant scope

The current worker binary supplies the canonical `tadoku/prod` context for its
run and replay composition boundaries. Detached handler and completion contexts
retain that explicit scope while preserving graceful shutdown. Database helpers
never supply a default tenant. This is the intermediate worker behavior until
persisted job tenants and separate cross-tenant queue scopes are supported.

## Metrics

- Request and Go process metrics are served on the metrics listener
  (`API_METRICS_PORT`).
- The request duration metric keeps its dashboard-compatible name,
  `tadoku_api_proxy_request_duration_seconds`, and its `route`, `upstream`,
  `mode` and `status` labels. Every route reports mode `native` and an empty
  upstream.
- The API feature-flag metrics report bounded provider initialization,
  refresh, error and evaluation labels, without user identities.

### Worker metrics

The separate worker serves its own metrics on `WORKER_METRICS_PORT`. Its
backlog gauges refresh at startup and every fifteen seconds. The `type` label
is a registered job type; unsupported-type gauges aggregate across unknown
types without an unbounded type label.

| Metric | Labels | Meaning |
| --- | --- | --- |
| `tadoku_worker_in_flight` | `type` | Jobs executing in this process. |
| `tadoku_worker_failed_attempts_total` | `type`, `code` | Failed handler attempts. |
| `tadoku_worker_handler_duration_seconds` | `type` | Histogram of handler execution duration before the job transition. |
| `tadoku_worker_pending_jobs` | `type` | Pending jobs for each registered type. |
| `tadoku_worker_failed_jobs` | `type` | Terminally failed jobs for each registered type. |
| `tadoku_worker_oldest_due_age_seconds` | `type` | Age of the oldest due or expired job. |
| `tadoku_worker_unsupported_pending_jobs` | None | Pending jobs with an unsupported type. |
| `tadoku_worker_unsupported_running_jobs` | None | Running jobs with an unsupported type. |
| `tadoku_worker_unsupported_failed_jobs` | None | Failed jobs with an unsupported type. |
| `tadoku_worker_unsupported_oldest_due_age_seconds` | None | Age of the oldest due unsupported job. |
| `tadoku_worker_expired_leases_total` | `type` | Expired running leases reclaimed. |

## Shutdown

- Shutdown closes the request and metrics listeners. After request handling
  stops, it closes the database and provider transport dependencies: the Flipt
  polling provider, the PostgreSQL pool, the Valkey client and idle HTTP
  connections, including Flipt, Kratos and Keto connections.
- A startup failure closes the same owned transport. Raw clients have no
  separate close operation.
- Valkey close follows the upstream client's per-connection close allowance
  rather than `API_VALKEY_TIMEOUT`.

## Production credentials

Production runtime credentials have only the grants the application requires,
separate from migration and administrator credentials. Confirm provider TLS,
pooling and connection limits when changing production configuration.
