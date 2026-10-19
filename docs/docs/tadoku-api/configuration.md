---
title: Runtime configuration
description: Tadoku API environment variables, startup and shutdown behavior, authentication wiring, the raw Keto and Kratos clients, metrics and production credentials.
sidebar_position: 9
---

# Runtime configuration

Read this when you change startup, an environment variable, authentication
wiring or a provider client.
[Job worker configuration](./worker-configuration.md) covers the leaderboard
worker, and [Tenant lifecycle command](./tenant-lifecycle.md) covers
`tadoku-tenant`.

`services/tadoku-api/cmd/tadoku-api/config.go` loads the configuration from
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
  defaults to the projected credential at `/var/run/secrets/tokens/token`. A
  non-empty value must be an HTTP(S) base URL without credentials, query or
  fragment; trailing slashes are removed. Empty is allowed when Flipt is
  disabled.
- `API_FLIPT_ENABLED`, `API_FLIPT_URL`, `API_FLIPT_ENVIRONMENT`,
  `API_FLIPT_NAMESPACE`, `API_FLIPT_TEST_ENVIRONMENT`, `API_FLIPT_UPDATE_INTERVAL`,
  `API_FLIPT_REQUEST_TIMEOUT`, `API_FLIPT_STARTUP_TIMEOUT` and
  `API_FLIPT_MANAGEMENT_URL` configure the feature-flag evaluation provider and
  the management client. Evaluation and management exchange separate
  credentials for `flipt-evaluation/tadoku-api` and
  `flipt-management/tadoku-api`. A non-empty `API_FLIPT_MANAGEMENT_URL` must be
  an HTTP(S) base URL without credentials, query or fragment even when Flipt is
  disabled; empty is allowed when Flipt is disabled. `API_FLIPT_URL` is
  validated only when Flipt is enabled. Provider outages keep safe defaults
  active and do not fail startup; later polling recovers without a restart.
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

## Identity reads

- The composition root keeps a concrete `identity.LegacyReader` on
  `application.kratosRead`. The identity reader uses this same read client
  for v1 lookups and cursor pagination. Profile and moderation receive
  `identity.NewReader(pool, legacy, evaluator)`, which selects the tenant's v1
  adapter or local account reads.
- `identity.NewLegacyReader(baseURL, identity.WithHTTPClient(httpClient))` uses the same client
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
- Reads return the SDK identity models. Identity ownership maps traits and account creation time into its provider-free
  `Account` type. `ListIdentities` fetches
  one requested cursor page and returns its next page token.
- The total client timeout and any earlier caller deadline bound requests;
  caller cancellation also interrupts response-body reads.

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
