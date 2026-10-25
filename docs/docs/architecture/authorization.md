---
sidebar_position: 3
title: Authentication and authorization
description: How Tadoku API authenticates gateway JWTs, admits requests on every tenant, checks tenant-scoped administrator and ban permits stored in PostgreSQL and restricts shared-data writes to production.
---

# Authentication and authorization

Read this when you add or change an operation's access rules, work with administrator or ban roles, or debug a 400, 401, 403 or 503 from Tadoku API.

The public contract is in the [Authorization API reference](../api/authorization/authz-api);
its source is `services/tadoku-api/spec/openapi.yaml`.

## Authentication and authorization

- **Identity** keeps tenant-owned accounts and sessions in Tadoku API. The
  [identity entry point](../tadoku-api/identity.md) provides the authenticated
  session callback. Oathkeeper's cookie-session authenticator forwards the
  complete Cookie and trusted X-Dev-Branch headers through the branch-aware API
  hop. Oathkeeper's `id_token` mutator produces an RS256 JWT whose `sub` is the
  account UUID, or `guest` for anonymous requests. Session traits travel in the
  `session` claim; `type` is `user`. API tokens carry a signed `tenant` claim
  for both anonymous and signed-in requests: `tadoku/prod`, or `tadoku/<route>`
  when a trusted branch selection reaches Oathkeeper. The edge removes client
  branch headers, so only branch hosts and in-cluster callers forwarding the
  selection choose a test tenant; production Oathkeeper signs an empty, rejected
  tenant for a malformed route. Service-token exchanges use their own claims.
- **Permissions** record who is an administrator and who is banned in each
  tenant. They are rows in the Tadoku API database, read on every request that
  needs them, never cached and never put in the token.

## Email access tokens

Only `GET /api/internal/emails` and its message-detail route accept tenant access
tokens. Their Oathkeeper rule tries `cookie_session` and then `noop`, signs the
normal cookie-session JWT, and restores the original Authorization header only
when it contains a `tdk_at_` bearer. Other routes keep their existing authentication.
The backend verifies tokens itself; proxy admission never grants email access.
Anonymous requests and forged JWTs are rejected.

The authz token feature parses the tenant from the token, checks `API_BRANCH`,
and looks up its SHA-256 hash under that tenant's row-level security scope.
The hash must be unrevoked, unexpired and scoped exactly to `email:read`.
Successful verification updates `last_used_at` and creates a private application
email-reader context, without creating a user identity or bypassing other routes.
Token authentication rejects `tadoku/prod` when `API_PRODUCTION` is true (the
default). Development explicitly sets it false; production administrators still
use cookies, administrator permits and transactional, fail-closed read auditing.
Canonical development token reads record the token UUID as actor and include
`access_token_id` in the same fail-closed audit transaction.

See [Tenant lifecycle command](../tadoku-api/tenant-lifecycle.md#email-access-tokens)
for attached issuance, limits, listing and revocation.

## Identity reads

Profile and moderation use `identity.Reader` for public profile details, account
age, admin listing and ban-target existence. Targets must be active accounts in
the request tenant. Account deletion has no API; see
[Account deletion](../operations/account-deletion.md). Read access does not
authorize shared provider mutations.

## Permission model

Features define permissions as [Permission model](./permission-model.md)
describes.

`authz.Permissions` (`services/tadoku-api/features/authz/authz_permissions.go`)
defines the `app` namespace, with one object per tenant, `app:root`: the
relations `admins` and `banned`, and the permits `admin`, `Related("admins")`,
and `is_banned`, `Related("banned")`.

Nothing is inherited between tenants: a test or branch tenant starts with no
administrators and no bans.

Administrators can change only the `banned` relation through the API
(`PUT /authz/users/{id}/role`), and never for another administrator. No API
grants `admins`; see [Granting an administrator](#granting-an-administrator).
A role update writes the relationship and its `moderation_audit_log` entry,
`ban_user` or `unban_user`, in one transaction.

### Granting an administrator

Grant or revoke an administrator in a deployed environment with operator SQL
as the runtime role, setting the tenant inside the transaction as
[Tenancy](../tadoku-api/database.md#tenancy) requires. To grant a production
administrator:

```sql
begin;
select set_config('tadoku.tenant', 'tadoku/prod', true);
insert into permission_relationships (tenant, object_type, object_id, relation, subject_type, subject_id)
values ('tadoku/prod', 'app', 'root', 'admins', 'user', '<lowercase user UUID>')
on conflict do nothing;
commit;
```

To revoke one:

```sql
begin;
select set_config('tadoku.tenant', 'tadoku/prod', true);
delete from permission_relationships
where tenant = 'tadoku/prod'
  and object_type = 'app'
  and object_id = 'root'
  and relation = 'admins'
  and subject_type = 'user'
  and subject_id = '<lowercase user UUID>';
commit;
```

For a production branch tenant, use its key instead of `tadoku/prod`. In
development, `tadoku-tenant development-accounts` grants Dev Admin `admins` on
each branch tenant; see
[Development data](../develop/development-data.md#branch-tenants).

## Request pipeline

Public `/api/internal/identity/*` entry rules use only Oathkeeper's anonymous
authenticator and `id_token` mutator. The signed guest JWT supplies the tenant;
the original Cookie reaches the identity API, which owns session authentication
and CSRF. These rules skip the global cookie-session callback so a banned user
can read their session or log out. This does not exempt unrelated business routes
from the ban pipeline below. The identity API's admission rules are owned by
the identity feature; changing gateway rules does not implement those handlers.

Every business route registered through `services/tadoku-api/transport/http/router.go`
passes two shared middlewares before its handler. Health probes (`/livez`,
`/readyz`) are outside this pipeline.

1. **JWT authentication** (`services/tadoku-api/transport/http/authentication.go`) verifies the RS256
   signature against `API_JWKS`, Oathkeeper's public key set. Startup fetches it
   and fails if it cannot. Tokens must carry `exp` and `iat`, be younger than
   `API_MAX_TOKEN_AGE` (default 24h) and, when `API_JWT_ISSUER` is set, match that
   issuer. Tokens with `type: service` are rejected. The verified subject, email
   and display name are placed on the request context as `identity.User`
   (`services/tadoku-api/internal/identity/`). The signed `tenant` must match
   `^[a-z0-9][a-z0-9-]{0,55}/[a-z0-9][a-z0-9-]{0,55}$`. Authentication parses it
   into `internal/tenant.Key` and puts it on the context beside the identity.
   Missing or malformed tenant claims are invalid credentials. A scoped
   `API_BRANCH` deployment rejects a different valid tenant with `421`; the
   unscoped base accepts every parsed tenant. This step checks no roles.
2. **Tenant admission and ban gate**
   (`services/tadoku-api/transport/http/banned_users.go`) calls
   `authz.Roles.Admit`, which checks the `is_banned` permit on the tenant's
   `app:root`. A missing tenant fails with unavailable.

Every tenant admits requests by the same rules, so test tenants are open to
anonymous visitors and to signed-in users without any grant:

| Request | Result |
| --- | --- |
| Guest or empty subject | Continues without a permission query. |
| Signed-in user, not banned | One query; continues. |
| Permission query error | Logged; continues with the ban state recorded as unknown. Read-only operations still run; strict checks return `503`. |
| Signed-in user with confirmed ban | Empty `403`, including for administrators. Only `GET /authz/current-user/role` continues and reports `banned`. |

The current-user-role exception lets a banned user discover its ban.

Public identity entry rules use anonymous authentication with the signed tenant
and the `id_token` mutator. Their handlers resolve the identity cookie themselves,
so a banned account can inspect its session and log out without presenting a
banned user JWT. Protected routes retain cookie-session authentication and the
normal ban gate. Do not exempt identity paths from the backend ban middleware;
the gateway must use the anonymous-only identity rule.

## Checks in application operations

Application operations in `services/tadoku-api/app/` own actor authorization; no
HTTP middleware enforces administrator access. They call
`*permissions.Checker` (`services/tadoku-api/internal/permissions/`) and, for
administrator checks, `*authz.Roles` (`services/tadoku-api/features/authz/`):

| Method | Passes when | Otherwise |
| --- | --- | --- |
| `RequireAuthenticated` | An actor is present (a signed-in, non-guest user) and the ban lookup succeeded. | `401` for no user or `guest`; `503` if the ban state is unknown. |
| `RequireAuthenticatedAllowingUnknownBan` | An actor is present (a signed-in, non-guest user). | `401`. Read-only operations only; mutations must never use it. |
| `Roles.RequireAdmin` | `RequireAuthenticated` passes and the actor holds the `admin` permit on the request tenant's `app:root`. | As above, `403` for non-administrators, `503` when the check fails. |
| `Roles.RequireProductionAdmin` | `RequireAdmin` passes and the request tenant is `tadoku/prod`. | As `RequireAdmin`, and `403` for every other tenant. |
| `Roles.IsAdmin`, `Roles.IsAdminOrFalse` | Report administrator status to expand behavior inside an already-authorized operation. | `IsAdmin` returns unavailable on errors; `IsAdminOrFalse` returns `false`. |
| `Require`, `Allowed` | Evaluate any `permissions.Check` a namespace built, for the actor; `Allowed` reports the result. | `Require` returns the errors below. |
| `AllowedEach` | Reports many checks for the actor, each as `Allowed` would. | `false` for every check for a guest or a banned user; `503` when the ban state is unknown or any check fails. |

Checks return `401` when nobody is signed in, `403` when denied and `503` when
the check fails or the request's ban state is unknown. A signed-in subject that
is not a UUID has no roles.

Feature services may inspect permissions only to expand behavior inside an
operation the application has already authorized, and never repeat the ban
gate. Roles of other users come from `authz.Roles.UserRoles` and never
authorize the actor. Relationship writes and reads use the
[building blocks](./permission-model.md#building-blocks).

## Shared data

`languages` and `log_units` have no tenant column: every tenant, including
production, reads the same rows. The `tenants` and `tenant_overrides`
registry is also shared, but only the owner `tadoku-tenant` command writes
it. Any tenant may read shared data. Application operations that write it
call `RequireProductionAdmin`, so an administrator acting through a test
tenant receives `403`. Row-level security policies on the shared tables also
reject writes from any other tenant.

## Oathkeeper administrator callback

Operator routes, such as the Flipt UI on `flags.tadoku.dev.lab`, require a native
session. Oathkeeper's `remote_json` authorizer then posts the session subject to
`POST /authz/internal/v1/proxy/admin-check`. This callback route bypasses the JWT
and ban pipeline and never creates a user identity. It accepts only the shared
bearer `API_OATHKEEPER_AUTHZ_TOKEN` (development Secret `dev-oathkeeper-authz`,
also mounted into Oathkeeper). It returns `200` when `IsProductionAdmin` finds
the subject's `admin` permit in `tadoku/prod`, whatever tenant the context
holds, and `403` otherwise.

## Error to HTTP status mapping

Application errors are `services/tadoku-api/internal/errx/` kinds, mapped in
`services/tadoku-api/transport/http/errors.go`.

| Cause | Status |
| --- | --- |
| Missing or malformed `Authorization: Bearer` header | `400`, JSON `missing or malformed jwt` |
| Invalid, expired, too old, service JWT or invalid tenant claim | `401`, JSON `invalid or expired jwt` |
| Valid tenant not served by this deployment | `421`, JSON `tenant not served by this deployment` |
| Confirmed ban | `403`, empty body |
| Invalid or missing callback credential | `401`, empty body |
| `errx.Unauthorized` (no user or `guest`) | `401` |
| `errx.Forbidden` (not an administrator) | `403` |
| `errx.Unavailable` (failed permission check, missing tenant, unknown ban state, [PostgreSQL outage](../tadoku-api/database.md#postgresql-errors)) | `503` |
| `errx.InvalidInput` | `400` |
| Request deadline exceeded | `504` |

## Seeding an administrator in development

Run `make dev-seed` (`scripts/dev/seed-db.sh`) once the base API is ready. It
only runs against the `homelab-talos-dev` Kubernetes context and is safe to
re-run.

- Creates a native administrator account (`TADOKU_DEV_ADMIN_EMAIL`, default
  `dev@tadoku.app`) and reader account (`TADOKU_DEV_READER_EMAIL`, default
  `reader@tadoku.app`) in the canonical tenant. Passwords come from
  `TADOKU_DEV_ADMIN_PASSWORD` and `TADOKU_DEV_READER_PASSWORD`. An account with
  the same email that the seed does not own is never modified.
- Grants the administrator `admins` on `app:root` in `tadoku/prod` with SQL,
  keeping an existing grant. The reader gets no role. The API never reads the
  `user_roles` rows the SQL fixtures write.
- Loads the application seed data from `scripts/dev/seed/`.

See [Development data](../develop/development-data.md#seed-data) for fixture UUIDs, reruns and branch seeding, and
[Development environment](../develop/environment.md#clean-up) for cleanup.
