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

Development branches receive their own signed tenant and persist it on every
job. The cache key is identical for a tenant served by the base or a branch
process, including branches with an isolated migration database. No separate
cache prefix is configured.

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

- `API_KETO_READ_URL`, an absolute HTTP(S) base URL for the Keto read
  service. Ban and administrator checks use a separate read-only client with a
  2s total request timeout.
- `API_KETO_WRITE_URL`, an absolute HTTP(S) base URL for the Keto write
  service.
- `API_KETO_WRITE_TIMEOUT` (default 2s), a positive total request timeout for
  the raw read/write client, including response-body reads.
- `API_KRATOS_ADMIN_URL`, an absolute HTTP(S) base URL for the Kratos admin
  service.
- `API_KRATOS_TIMEOUT` (default 2s), a positive total request timeout, including
  reading the response body.

For the Keto and Kratos base URLs, credentials, query strings and fragments are
rejected; path prefixes are supported and trailing slashes are removed.

### Service tokens and feature flags

- `API_OATHKEEPER_URL` (default `http://oathkeeper-proxy.default:4455`), used
  for outgoing service-token exchange. `API_SERVICE_ACCOUNT_TOKEN_PATH`
  defaults to the projected credential at `/var/run/secrets/tokens/token`.
- `API_FLIPT_ENABLED`, `API_FLIPT_URL`, `API_FLIPT_ENVIRONMENT`,
  `API_FLIPT_NAMESPACE`, `API_FLIPT_TEST_ENVIRONMENT`, `API_FLIPT_UPDATE_INTERVAL`,
  `API_FLIPT_REQUEST_TIMEOUT`, `API_FLIPT_STARTUP_TIMEOUT` and
  `API_FLIPT_MANAGEMENT_URL` configure the feature-flag evaluation provider and
  the management client. Evaluation and management exchange separate
  credentials for `flipt-evaluation/tadoku-api` and
  `flipt-management/tadoku-api`. Provider outages keep safe defaults active and
  do not fail startup; later polling recovers without a restart.
- `API_FLIPT_TEST_ENVIRONMENT` defaults to `test`. When Flipt is enabled, it
  must be nonempty and different from `API_FLIPT_ENVIRONMENT`; an invalid value
  fails startup. The canonical tenant `tadoku/prod` keeps the configured
  environment and namespace. Other tenants use the test environment and the
  tenant key with `/` mapped to `_`, because Flipt namespace keys cannot contain
  slashes. Evaluation and named-user management use this same mapping. Test
  clients are lazy and bounded to 16; missing namespaces or missing tenant
  contexts use safe evaluation defaults. Disabled Flipt serves safe defaults
  for every tenant. See [Feature flags](../architecture/feature-flags.md).

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
- Startup constructs and retains the read-only Kratos client and the shared Keto
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

## Kratos reads and guarded identity writes

- The composition root keeps a concrete `*kratos.Client` on
  `application.kratosRead`. Identity features use this same read client
  for lookups and cursor pagination.
- `NewClient(baseURL, kratos.WithHTTPClient(httpClient))` uses the same client
  for SDK reads and cursor pagination. It constructs the pinned
  `github.com/ory/kratos-client-go` v0.11.1 SDK internally and exposes no raw
  SDK constructor. Tadoku API validates the configuration and supplies a
  client with a total timeout on its owned transport; the read client does not
  own that HTTP client. Without the option, it uses the SDK's default client.

Call the read client with the operation's caller context:

```go
identity, err := kratosIdentities.FetchIdentity(ctx, identityID)
```

- `FetchIdentity` maps provider HTTP 404 to `kratos.ErrNotFound` and returns
  other provider failures as wrapped errors. `UserExists` translates that
  sentinel to `false`.
- Use `errors.As` to inspect `*kratosapi.GenericOpenAPIError`, including its
  `Body()` and `Model()`, and `errors.Is` for `context.Canceled` or
  `context.DeadlineExceeded`.
- Reads return the SDK identity models. Features own domain mapping, trait
  policy, account-age rules and any cache lifecycle. `ListIdentities` fetches
  one requested cursor page and returns its next page token.
- The total client timeout and any earlier caller deadline bound requests;
  caller cancellation also interrupts response-body reads.

`Writer` in `services/tadoku-api/infra/kratosidentity/` owns application identity
writes. `NewWriter(adminURL, httpClient, logger)` constructs its SDK internally
from the admin URL and a caller-owned HTTP client with a positive total timeout.
It checks the tenant before every mutation and remains unwired until an
account-deletion flow exists. See
[Kratos identity writes](../architecture/authorization.md#kratos-identity-writes)
for its applied, skipped and missing-tenant outcomes.

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

### Worker tenant scope

The worker separates queue scope from handler scope. Queue claims, backlog
metrics and completed-job cleanup receive the required `jobqueue.Scope`; each
handler and fenced transition receives only its persisted job tenant. Database
helpers never supply a default tenant. The replay command uses the explicit
canonical `tadoku/prod` context.

## Metrics

- Request and Go process metrics are served on the metrics listener
  (`API_METRICS_PORT`).
- The request duration metric keeps its dashboard-compatible name,
  `tadoku_api_proxy_request_duration_seconds`, and its `route`, `upstream`,
  `mode` and `status` labels, with the additional `tenant_kind` label. Every
  route reports mode `native` and an empty upstream.
- `tadoku_scoring_shadow_comparisons_total` also carries `tenant_kind`, alongside
  its existing scoring outcome, operation, mode, activity and source labels.
- The API feature-flag metrics report bounded provider initialization,
  refresh, error and evaluation labels, without user identities.

`tenant_kind` has only three values: `production` for `tadoku/prod`, `test` for
every other validated tenant key, and `unknown` when the context has no tenant.
Request observations use the tenant from the verified JWT, including the outer
completion log after authentication returns. Rejections before a token is
accepted and routes without tenant authentication report `unknown`.

The API and worker's context-aware structured logs carry the full `tenant` key.
Use the caller or job context when logging tenant work; context-free startup
logs have no tenant attribute. Request completion logs use `tenant="unknown"`
when authentication has not established a tenant. Full tenant keys are never
metric labels, so creating or deleting test tenants cannot create an unbounded
set of series.

### Worker metrics

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

## Shutdown

- Shutdown closes the request and metrics listeners. After request handling
  stops, it closes the database and provider transport dependencies: the Flipt
  polling provider, the PostgreSQL pool, the Valkey client and idle HTTP
  connections, including Flipt, Kratos and Keto connections.
- A startup failure closes the same owned transport. Raw clients have no
  separate close operation.
- Valkey close follows the upstream client's per-connection close allowance
  rather than `API_VALKEY_TIMEOUT`.

## Tenant lifecycle command

`services/tadoku-api/cmd/tadoku-tenant` provides provisioning, teardown and
worker overrides. Its tenant argument is a full parsed key, such
as `tadoku/branch-0123abcd`; the id must end in a hyphen and eight lowercase
hexadecimal digits. Bare route ids, `tadoku/prod` and production registry rows
are refused before provider writes.

```sh
bazel run //services/tadoku-api/cmd/tadoku-tenant -- provision \
  --tenant tadoku/branch-0123abcd --flipt-features k8s/dev/base/flipt/features.yaml \
  --tester 11111111-1111-4111-8111-111111111111
bazel run //services/tadoku-api/cmd/tadoku-tenant -- override set \
  --tenant tadoku/branch-0123abcd --component tadoku-worker
bazel run //services/tadoku-api/cmd/tadoku-tenant -- override clear \
  --tenant tadoku/branch-0123abcd --component tadoku-worker
bazel run //services/tadoku-api/cmd/tadoku-tenant -- teardown \
  --tenant tadoku/branch-0123abcd
```

All commands refuse credentials that neither own the tenant registry nor
belong to the `tadoku_tenant_lifecycle` role. That role needs only schema usage
and `select`, `insert` and `delete` on `tenants` and `tenant_overrides`;
registry policies limit its writes to test rows, and deleting a row cascades
through the owner's foreign keys. Production branch Jobs use it, and
development uses the owner. Credentials come from `TENANT_POSTGRES_HOST`,
`TENANT_POSTGRES_PORT` (default 5432), `TENANT_POSTGRES_DATABASE`,
`TENANT_POSTGRES_USER`, `TENANT_POSTGRES_PASSWORD` and `TENANT_POSTGRES_SSLMODE`.
`TENANT_POSTGRES_URL` is rejected. Provision and teardown additionally require
`TENANT_KETO_READ_URL`, `TENANT_KETO_WRITE_URL`, `TENANT_VALKEY_URL` and
`TENANT_FLIPT_MANAGEMENT_URL`; `TENANT_FLIPT_ENVIRONMENT` defaults to `test` and
must differ from `production`. Use the direct provisioning-authorized Flipt
endpoint, rather than the API's restricted service-token route. Override
commands need only PostgreSQL. The only registered override component is
`jobqueue.WorkerComponent`, currently `tadoku-worker`.

`//services/tadoku-api/cmd/tadoku-tenant:image` packages the command for
branch provisioning outside development: a distroless base with the CA bundle
`verify-full` connections need, `/tadoku-tenant` as its entrypoint and the
repository's flag definitions at `/flipt/features.yaml`. Its `branch_push`
target sets no repository or tags; the caller supplies both. The development
`//.dev:tenant_push` image has no CA bundle.

Provision commits an idempotent test registry row under a transaction-local
advisory lock, grants the canonical Keto parent and repeated `--tester` UUIDs,
then creates only missing Flipt resources from the strict boolean seed format.
Keto grants are checked by their complete tuple before creating them because
the provider accepts duplicate relationship writes.
Reruns preserve existing grants and flag values. Teardown deletes the tenant's
Keto relationships, cache keys and Flipt namespace before deleting the registry
row and its cascading data. Failure leaves the row for a retry; cleanup still
runs when a prior attempt already removed that row. Commands sharing a tenant
must be serialized by the caller's route lease. Provider HTTP calls have a
30-second timeout; the command has a five-minute deadline and observes shutdown
signals. Secrets are supplied through environment variables and never printed.

## Production credentials

Production runtime credentials have only the grants the application requires,
separate from migration and administrator credentials. Confirm provider TLS,
pooling and connection limits when changing production configuration.
