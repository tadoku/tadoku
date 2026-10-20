---
sidebar_position: 3
title: Authentication and authorization
description: How Tadoku API authenticates gateway JWTs, admits requests on every tenant, checks tenant-scoped Keto administrator and ban permits and restricts shared-data writes to production.
---

# Authentication and authorization

Read this when you add or change an operation's access rules, work with administrator or ban roles, or debug a 400, 401, 403 or 503 from Tadoku API.

The public contract is in the [Authorization API reference](../api/authorization/authz-api);
its source is `services/tadoku-api/spec/openapi.yaml`.

## Authentication and authorization

- **Identity** owns account/session selection per tenant. V1 uses Ory Kratos;
  v2 uses tenant-owned accounts and sessions in Tadoku API. The
  [identity entry point](../tadoku-api/identity.md) selects with the tenant-level
  `identity-v2` flag and provides the authenticated session callback. Its cookie-session authenticator forwards the
  complete Cookie and trusted X-Dev-Branch headers through the branch-aware API
  hop. Oathkeeper's
  `id_token` mutator produces an RS256 JWT whose `sub` is the
  account UUID, or `guest` for anonymous requests. Session traits travel in the
  `session` claim; `type` is `user`. API tokens carry a signed `tenant`
  claim for both anonymous and signed-in requests: `tadoku/prod`, or
  `tadoku/<route>` when a trusted branch selection reaches Oathkeeper. The
  edge removes client branch headers, so only branch hosts and in-cluster
  callers forwarding the selection choose a test tenant; production Oathkeeper
  signs an empty, rejected tenant for a malformed route. Service-token
  exchanges use their own claims.
- **Ory Keto** owns authorization facts: who is an administrator and who is
  banned. Tadoku API reads them from Keto on every request that needs them. Roles
  are not stored in PostgreSQL or in the token, and results are not cached.

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
use cookies, Keto authorization and transactional, fail-closed read auditing.
Canonical development token reads record the token UUID as actor and include
`access_token_id` in the same fail-closed audit transaction.

See [Tenant lifecycle command](../tadoku-api/tenant-lifecycle.md#email-access-tokens)
for attached issuance, limits, listing and revocation.

## Identity reads

Profile and moderation use `identity.Reader` for public profile details, account
age, admin listing and ban-target existence. The identity feature chooses the
implementation using the request tenant's flag. V2 targets must be active
accounts in that tenant; v1 reads continue through Kratos. Account deletion is
not available, and read access does not authorize shared provider mutations.

The identity session adapter in `services/tadoku-api/features/identity/kratos`
checks the selected Kratos cookie and revokes only that session for the canonical
tenant. Test-tenant logout clears the browser cookie and leaves shared provider
sessions active; a missing tenant cannot revoke a session.

## Keto data model

| Field | Value |
| --- | --- |
| namespace | `app` |
| canonical tenant | `tadoku/prod`, using the existing object `tadoku` |
| test object | The full tenant key, such as `e2e/alpha-0000000a` |
| relations | `admins`, `banned`, `parents`, `testers` |
| permits | `admin`, `is_banned`, `access` |
| subject | direct `subject_id`: the account UUID from the JWT `sub`, never an email |

The canonical tenant keeps `app:tadoku` and its existing administrator and ban
relations. Tenant keys use `name/id`; the provider object for this existing
canonical tenant remains `tadoku`, without copying or rewriting its tuples.
A regular production user holds neither `admins` nor `banned`.

A test object inherits administrators and bans from its `parents` objects.
Its own administrator or ban tuples apply only to that object. The `access`
permit allows administrators or direct members of `testers`; a tester on one
object has no access to another object without its own grant. Request
admission does not check `access`: anyone may use a test tenant.

Provision a test object's parent with this JSON tuple. Keto v25.4.0 accepts
the empty parent relation and traverses the parent's permits:

```json
{
  "namespace": "app",
  "object": "e2e/alpha-0000000a",
  "relation": "parents",
  "subject_set": {
    "namespace": "app",
    "object": "tadoku",
    "relation": ""
  }
}
```

Add a direct tester tuple as
`app:e2e/alpha-0000000a#testers@<identity-id>`. Raw `admins` and `banned`
relations do not include inherited memberships: an administrator inherited
from `app:tadoku` has the child's `admin` and `access` permits but no local
`admins` tuple. Removing a branch ban does not remove an inherited production
ban.

The namespace configuration (OPL) is `k8s/dev/base/keto/namespaces.keto.ts` for
the development environment and `infra/dev/ory/namespaces.keto.ts` for backend
test fixtures. These files stay byte-identical. Development runs Keto v26.2.0;
production and the real Bazel fixture run v25.4.0. Tadoku API calls Keto through
`services/tadoku-api/infra/keto/`: `NewReadClient` for checks and `NewClient`
for read/write access. Keto's `403` answer to a check means "denied", not an
error.

Administrators can toggle only the `banned` relation through the API
(`PUT /authz/users/{id}/role`), and cannot change another administrator's role.
No API grants `admins`; that tuple is written directly through the Keto write
API, as the development seed does.

Every role update that reaches Keto is recorded in `moderation_audit_log`:
`ban_user` or `unban_user` when the relation changed, and `ban_user_failed` or
`unban_user_failed` with the provider error when it did not. The entry is
written after the Keto write and never changes the response; see
[Auditing a provider change](../tadoku-api/conventions.md#auditing-a-provider-change).

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
   (`services/tadoku-api/transport/http/banned_users.go`) calls the shared
   permission checker's `Admit`. The checker resolves the verified tenant to
   its Keto object; missing or zero tenant contexts fail with unavailable.
   Canonical requests keep `app:tadoku`, including private development
   databases using `tadoku/prod`. Every other tenant uses its full `name/id`
   key as the object.

Every tenant admits requests by the same rules, so test tenants are open to
anonymous visitors and to signed-in users without any grant on their object:

| Request | Result |
| --- | --- |
| Guest or empty subject | Continues without a Keto call. |
| Signed-in user, not banned | One `is_banned` permit lookup on the tenant's object; continues. |
| Keto error | Logged; continues with the ban state recorded as unknown. Strict checks then return `503`. |
| Signed-in user with confirmed ban | Empty `403`, including for administrators. Only `GET /authz/current-user/role` continues and reports `banned`. |

A test object's `is_banned` permit includes bans inherited from its parents,
while a ban on the test object applies only to that tenant. The
current-user-role exception lets a banned user discover its ban.

Public identity entry rules use anonymous authentication with the signed tenant
and the `id_token` mutator. Their handlers resolve the identity cookie themselves,
so a banned account can inspect its session and log out without presenting a
banned user JWT. Protected routes retain cookie-session authentication and the
normal ban gate. Do not exempt identity paths from the backend ban middleware;
the gateway must use the anonymous-only identity rule.

## Checks in application operations

Application operations in `services/tadoku-api/app/` own actor authorization; no
HTTP middleware enforces administrator access. They call the
`*permissions.Checker` from `services/tadoku-api/internal/permissions/`:

| Method | Passes when | Otherwise |
| --- | --- | --- |
| `RequireAuthenticated` | An actor is present (a signed-in, non-guest user) and the ban lookup succeeded. | `401` for no user or `guest`; `503` if the ban state is unknown. |
| `RequireAuthenticatedAllowingUnknownBan` | An actor is present (a signed-in, non-guest user). | `401`. Read-only operations only; mutations must never use it. |
| `RequireAdmin` | `RequireAuthenticated` passes and the actor holds the tenant object's `admin` permit. | As above, `403` for non-administrators, `503` on Keto errors. |
| `RequireProductionAdmin` | `RequireAdmin` passes and the request tenant is `tadoku/prod`. | As `RequireAdmin`, and `403` for every other tenant. |
| `IsAdmin`, `IsAdminOrFalse` | Report administrator status to expand behavior inside an already-authorized operation. | `IsAdmin` returns unavailable on errors; `IsAdminOrFalse` returns `false`. |

Feature services may inspect permissions only to expand behavior inside an
operation the application has already authorized, and must not repeat the ban
gate. Facts about other users, such as whether a target user is an administrator,
come from `services/tadoku-api/internal/permissions/` (`KetoService` reads
`TargetRoles`; `KetoManager` writes bans). Target facts are never actor
authorization. Both resolve the tenant object for each operation. Individual
facts check the `admin` and `is_banned` permits, so inherited administrators
remain protected from role changes.

The administrator user list reads raw `admins` and `banned` relations on the
resolved object. A test tenant's list therefore shows only its own grants,
while its administrator and ban checks still inherit the parent permits.
Unbanning a user on a test tenant removes only that object's direct ban; an
inherited production ban continues to apply.

## Shared data

`languages` and `log_units` have no tenant column: every tenant, including
production, reads the same rows. The `tenants` and `tenant_overrides`
registry is also shared, but only the owner `tadoku-tenant` command writes
it. Any tenant may read shared data. Application operations that write it
call `RequireProductionAdmin`, so an administrator acting through a test
tenant, whether granted on the branch object or inherited from production,
receives `403` and cannot change data production uses. Row-level security
policies on the shared tables also reject writes from any other tenant.

## Oathkeeper administrator callback

Operator routes, such as the Flipt UI on `flags.tadoku.dev.lab`, require a Kratos
session. Oathkeeper's `remote_json` authorizer then posts the session subject to
`POST /authz/internal/v1/proxy/admin-check`. This callback route bypasses the JWT
and ban pipeline and never creates a user identity. It accepts only the shared
bearer `API_OATHKEEPER_AUTHZ_TOKEN` (development Secret `dev-oathkeeper-authz`,
also mounted into Oathkeeper), checks only the subject's `admin` permit on `app:tadoku` through
`IsProductionAdmin`, and returns `200` for production administrators and
`403` otherwise. The callback does not need a request tenant; a branch-only
administrator cannot authorize the operator UI.

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
| `errx.Unavailable` (Keto error, missing tenant, unknown ban state, [PostgreSQL outage](../tadoku-api/database.md#postgresql-errors)) | `503` |
| `errx.InvalidInput` | `400` |
| Request deadline exceeded | `504` |

## Seeding an administrator in development

Run `make dev-seed` (`scripts/dev/seed-db.sh`) once the base API and Keto are
ready, and Kratos too while the canonical tenant uses v1. It only runs against
the `homelab-talos-dev` Kubernetes context and is safe to re-run.

- Creates an administrator (`TADOKU_DEV_ADMIN_EMAIL`, default `dev@tadoku.app`)
  and a reader (`TADOKU_DEV_READER_EMAIL`, default `reader@tadoku.app`) through
  the canonical tenant's identity provider: Kratos identities marked with
  `metadata_admin.seeded_by=tadoku-dev-seed` on v1, native accounts on v2.
  Passwords come from `TADOKU_DEV_ADMIN_PASSWORD` and
  `TADOKU_DEV_READER_PASSWORD`. An identity or account with the same email that
  the seed does not own is never modified.
- Writes `app:tadoku#admins@<administrator UUID>` through the Keto write API
  (`http://keto-write.tdk-dev-keto:4467`) unless the read API already lists it.
  Keto holds roles for both providers; native accounts have no role field, and
  the API never reads the `user_roles` row the SQL fixtures write. The reader
  gets no role.
- Loads the application seed data from `scripts/dev/seed/`.

See [Development data](../develop/development-data.md#seed-data) for provider
selection, fixture UUIDs, reruns and branch seeding, and
[Development environment](../develop/environment.md#clean-up) for cleanup.
