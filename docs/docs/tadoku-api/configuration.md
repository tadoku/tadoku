---
title: Runtime configuration
description: Tadoku API environment variables, startup and shutdown behavior, authentication wiring, permissions, metrics and production credentials.
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
  shared by the service-token exchange and Flipt clients.
  `API_IDLE_TIMEOUT` is also the listeners' idle-connection timeout.
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
- [Permissions](../architecture/permission-model.md) use this pool, with no
  endpoint or timeout of their own.

### Valkey, leaderboard caches and login throttling {#valkey-and-leaderboard-caches}

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

Identity also uses tenant-prefixed Valkey keys for login throttling. Its
15-minute fixed window permits ten attempts per normalized email and fifty per
client IP. Valkey failures allow login attempts and log a warning. See
[Identity](./identity.md) for session and credential behavior.

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

- `API_PRODUCTION` (default `true`) rejects `tadoku/prod` access tokens. Only
  development sets it to `false`; it does not alter JWT, cookie or administrator
  authorization. Test-tenant tokens remain available in both environments.
- `API_JWKS`, the gateway's public signing-key URL; same HTTP(S) URL rules as the other provider base URLs.
- `API_MAX_TOKEN_AGE` (default 24h), the maximum accepted age since `iat`.
- `API_JWT_ISSUER`, an optional exact issuer match. Empty leaves the issuer
  unchecked.
- `API_OATHKEEPER_AUTHZ_TOKEN`, the bearer credential required on trusted
  Oathkeeper authorization callbacks.

`API_IDENTITY_TRUSTED_PROXY_CIDRS` is an optional comma-separated list of
IP prefixes. Only a connection from those prefixes may supply `X-Real-IP` for
login throttling; configure only proxies whose edge overwrites that header.
The default trusts no forwarding header; clients behind an unconfigured proxy
share that proxy's peer-IP budget. Malformed prefixes fail startup.

`API_IDENTITY_COOKIE_DOMAIN` (default `tadoku.app`) is restricted to
`tadoku.app` or `tadoku.dev.lab`. The development base API and overlays set
the latter.

### Service tokens and feature flags

- `API_OATHKEEPER_URL` (default `http://oathkeeper-proxy.default:4455`), used
  for outgoing service-token exchange. `API_SERVICE_ACCOUNT_TOKEN_PATH`
  defaults to the projected credential at `/var/run/secrets/tokens/token`. A
  non-empty value must be an HTTP(S) base URL without credentials, query or
  fragment; trailing slashes are removed. Empty is allowed when Flipt is
  disabled.
- `API_FLIPT_ENABLED`, `API_FLIPT_URL`, `API_FLIPT_ENVIRONMENT`,
  `API_FLIPT_NAMESPACE`, `API_FLIPT_TEST_ENVIRONMENT`, `API_FLIPT_UPDATE_INTERVAL`,
  `API_FLIPT_REQUEST_TIMEOUT` and `API_FLIPT_STARTUP_TIMEOUT` configure evaluation.
  Evaluation exchanges credentials for `flipt-evaluation/tadoku-api`.
  `API_FLIPT_URL` is validated only when Flipt is enabled. Provider outages do
  not fail startup. Until a client has fetched its namespace snapshot,
  `Boolean` flag consumers use safe defaults; a later fetch recovers without
  a restart.
- `API_FLIPT_TEST_ENVIRONMENT` defaults to `test`. When Flipt is enabled, it
  must be nonempty and different from `API_FLIPT_ENVIRONMENT`; an invalid value
  fails startup. The canonical tenant `tadoku/prod` keeps the configured
  environment and namespace. Other tenants use the test environment and the
  tenant key with `/` mapped to `_`, because Flipt namespace keys cannot contain
  slashes. Evaluation and namespace provisioning use this same mapping. Test
  clients are lazy and bounded to 16. Missing namespaces, missing tenant
  contexts and disabled Flipt give `Boolean` consumers safe defaults. See
  [Feature flags](../architecture/feature-flags.md).

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
- An invalid permission model aborts startup.
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
  must be an account UUID; a token with any other subject is rejected with
  `401` like any other invalid token. Service tokens are never converted into
  human identities.
- The ban gate performs only the ban lookup, so no unrelated administrator
  lookup can discard a successful ban result. Request deadlines bound the ban
  query. A failed lookup is kept in the request context, so later authenticated
  and administrator checks return unavailable without another ban query.

## Identity reads

Profile and moderation receive `identity.NewReader(pool)` on the API's
PostgreSQL pool; it reads tenant-local accounts only. See
[Account reads](./identity.md#account-reads).

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
  connections, including service-token exchange and Flipt connections.
- A startup failure closes the same owned transport. Raw clients have no
  separate close operation.
- Valkey close follows the upstream client's per-connection close allowance
  rather than `API_VALKEY_TIMEOUT`.

## Production credentials

Production runtime credentials have only the grants the application requires,
separate from migration and administrator credentials. Confirm provider TLS,
pooling and connection limits when changing production configuration.
