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
- **Permissions** own authorization facts: who is an administrator and who is
  banned in each tenant. Tadoku API stores them in its PostgreSQL database and
  reads them on every request that needs them. Roles are not stored in the
  token, and results are not cached.

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

Permissions are relationships stored in the tenant-owned table
`permission_relationships`:

| Column | Value |
| --- | --- |
| `tenant` | The tenant key. The `tenant_isolation` row-level security policy scopes reads and writes to the transaction's tenant. The foreign key to `tenants(key)` cascades, so deleting a tenant removes its relationships. |
| `object_type`, `object_id` | The object, such as `app` and `root` |
| `relation` | A relation the model defines for the object type |
| `subject_type`, `subject_id`, `subject_relation` | For a user: `user`, the lowercase account UUID from the JWT `sub` (never an email) and an empty subject relation |
| `created_at` | When the relationship was stored |

One primary key covers every identifying column, so each relationship is
stored once.

The permission engine in `services/tadoku-api/internal/permissions/` knows no
object type. Each feature defines the namespaces for the objects it owns with
`permissions.Define`, next to the code that uses them, and builds that
namespace's checks and relationships through it (`Namespace.Check` and
`Namespace.UserRelationship`). `features/authz` defines the only namespace
today, `authz.Permissions` for the `app` object type, with one object per
tenant, `app:root` (`services/tadoku-api/features/authz/authz_permissions.go`):

| Name | Kind | Passes for |
| --- | --- | --- |
| `admins` | Relation | Users stored in it |
| `banned` | Relation | Users stored in it |
| `admin` | Permit | `Related("admins")` |
| `is_banned` | Permit | `Related("banned")` |

Build permits from the engine's rules: `Related`, `Allowed`, `Via`, `PermitOn`,
`AnyOf`, `AllOf` and `Except`. `PermitOn` passes when a permit passes on one
fixed object, which may belong to another feature's namespace, such as
`PermitOn("app", "root", "admin")`. Features cannot import each other, so such
references are names. The composition roots (`cmd/tadoku-api`, the tenant
lifecycle application and the E2E harness) assemble the model with
`permissions.NewModel`, which validates the namespaces together, resolves those
names and rejects a duplicate object type, so `tadoku-api` refuses to start with
an invalid model. Evaluation fails closed: a storage error, an unknown permit, a
cancelled context or a walk deeper than eight levels returns an error, never an
allowed result.

The engine imports no database package. It reads and writes relationships
through the `permissions.Tuples` port it declares. `features/authz` owns
`permission_relationships` and implements the port with its `Repository`
(`services/tadoku-api/features/authz/authz_repository.go`) and the queries in
`services/tadoku-api/sql/authz/authz.sql`. The composition roots pass that
repository to `permissions.NewChecker` and `permissions.NewRelationships`.

Checks and relationships have unexported fields, so code builds them only
through a namespace. Bazel visibility keeps other features from importing a
feature's namespace, so only its owner and the application can build that
namespace's checks and relationships. To add a namespace, define it in its
feature and add it to `permissions.NewModel` in each composition root. The
engine builds user relationships today; add a namespace method for subject sets
or linked objects when a feature first stores one.

A feature declares its namespace as `var Permissions = permissions.Define(...)`
in `<feature>_permissions.go`, next to helpers that build its checks and
relationships. Name a check helper after the permit with a `Permission` suffix
and return a `permissions.Check`, such as
`contests.ModeratePermission(contestID)`. Name a relationship helper after the
relation with a `Relationship` suffix and return a `permissions.Relationship`,
such as `contests.OwnerRelationship(contestID, userID)`. Helpers take typed
IDs, so call sites never spell a namespace, permit or relation string. The
`authz` helpers follow the same names and stay unexported behind
`authz.Roles`.

Each tenant has its own `app:root`, and nothing is inherited between tenants.
Test and branch tenants start without administrators or bans: a production
administrator has no rights on a branch tenant, and a production ban does not
apply there. A regular production user holds neither `admins` nor `banned`.

Administrators can toggle only the `banned` relation through the API
(`PUT /authz/users/{id}/role`), and cannot change another administrator's role.
No API grants `admins`; see
[Granting an administrator](#granting-an-administrator).

Every role update that reaches the permission store is recorded in
`moderation_audit_log`: `ban_user` or `unban_user` when the relation changed,
and `ban_user_failed` or `unban_user_failed` with the error when it did not.
The entry is written after the relationship write, outside its transaction,
and never changes the response; see
[Auditing a provider change](../tadoku-api/conventions.md#auditing-a-provider-change).

### Granting an administrator

Grant or revoke an administrator in a deployed environment with operator SQL
as the runtime role. Set the tenant explicitly inside the transaction, as
[Tenancy](../tadoku-api/database.md#tenancy) requires, and use the user's
lowercase UUID. To grant a production administrator:

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

For a production branch tenant, replace both `tadoku/prod` values with that
tenant's key. In development, `tadoku-tenant development-accounts` grants Dev
Admin `admins` on each branch tenant; see
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
   (`services/tadoku-api/transport/http/banned_users.go`) calls the shared
   `Admit` from `authz.Roles`, which checks the `is_banned` permit on the
   verified tenant's `app:root` through the shared permission checker. A missing
   or zero tenant context fails with unavailable.

Every tenant admits requests by the same rules, so test tenants are open to
anonymous visitors and to signed-in users without any grant:

| Request | Result |
| --- | --- |
| Guest or empty subject | Continues without a permission query. |
| Signed-in user, not banned | One indexed query for the tenant's `is_banned` permit; continues. |
| Permission query error | Logged; continues with the ban state recorded as unknown. Read-only operations still run; strict checks return `503`. |
| Signed-in user with confirmed ban | Empty `403`, including for administrators. Only `GET /authz/current-user/role` continues and reports `banned`. |

A ban applies only to the tenant that stores it. The current-user-role
exception lets a banned user discover its ban.

Public identity entry rules use anonymous authentication with the signed tenant
and the `id_token` mutator. Their handlers resolve the identity cookie themselves,
so a banned account can inspect its session and log out without presenting a
banned user JWT. Protected routes retain cookie-session authentication and the
normal ban gate. Do not exempt identity paths from the backend ban middleware;
the gateway must use the anonymous-only identity rule.

## Checks in application operations

Application operations in `services/tadoku-api/app/` own actor authorization; no
HTTP middleware enforces administrator access. They call the generic
`*permissions.Checker` from `services/tadoku-api/internal/permissions/` and, for
administrator checks, `*authz.Roles` from `services/tadoku-api/features/authz/`:

| Method | Passes when | Otherwise |
| --- | --- | --- |
| `RequireAuthenticated` | An actor is present (a signed-in, non-guest user) and the ban lookup succeeded. | `401` for no user or `guest`; `503` if the ban state is unknown. |
| `RequireAuthenticatedAllowingUnknownBan` | An actor is present (a signed-in, non-guest user). | `401`. Read-only operations only; mutations must never use it. |
| `Roles.RequireAdmin` | `RequireAuthenticated` passes and the actor holds the `admin` permit on the request tenant's `app:root`. | As above, `403` for non-administrators, `503` when the check fails. |
| `Roles.RequireProductionAdmin` | `RequireAdmin` passes and the request tenant is `tadoku/prod`. | As `RequireAdmin`, and `403` for every other tenant. |
| `Roles.IsAdmin`, `Roles.IsAdminOrFalse` | Report administrator status to expand behavior inside an already-authorized operation. | `IsAdmin` returns unavailable on errors; `IsAdminOrFalse` returns `false`. |
| `Require`, `Allowed` | Evaluate any `permissions.Check` a namespace built, for the actor; `Allowed` reports the result. | `Require` returns the errors below. |

Checks return `401` when nobody is signed in, `403` when denied and `503` when
the check fails or the request's ban state is unknown. A signed-in subject that
is not a UUID has no roles.

Feature services may inspect permissions only to expand behavior inside an
operation the application has already authorized, and must not repeat the ban
gate. Facts about other users, such as whether a target user is an administrator,
come from `authz.Roles.UserRoles`, which reads the stored `admins` and `banned`
relationships for a batch of users with one query through
`permissions.Relationships.Stored`. `permissions.Relationships`
(`NewRelationships`) writes relationships with `Grant` and `Revoke`; it
validates each relationship against the model before any SQL and joins the
caller's transaction (`postgres.RunInTransaction`). Both use the request tenant.
Target facts are never actor authorization. The profile user list receives the
role lookup from the application, so profile does not depend on authz.

The administrator user list and role updates read the request tenant's stored
relations, so a tenant's list shows only its own grants, and administrators
granted in that tenant are protected from role changes. Unbanning a user
removes the ban in the request tenant only.

## Shared data

`languages` and `log_units` have no tenant column: every tenant, including
production, reads the same rows. The `tenants` and `tenant_overrides`
registry is also shared, but only the owner `tadoku-tenant` command writes
it. Any tenant may read shared data. Application operations that write it
call `RequireProductionAdmin`, so an administrator acting through a test
tenant receives `403` and cannot change data production uses. Row-level security
policies on the shared tables also reject writes from any other tenant.

## Oathkeeper administrator callback

Operator routes, such as the Flipt UI on `flags.tadoku.dev.lab`, require a native
session. Oathkeeper's `remote_json` authorizer then posts the session subject to
`POST /authz/internal/v1/proxy/admin-check`. This callback route bypasses the JWT
and ban pipeline and never creates a user identity. It accepts only the shared
bearer `API_OATHKEEPER_AUTHZ_TOKEN` (development Secret `dev-oathkeeper-authz`,
also mounted into Oathkeeper), checks only the subject's `admin` permit on
`app:root` in `tadoku/prod` through `IsProductionAdmin`, and returns `200` for
production administrators and `403` otherwise. `IsProductionAdmin` always
evaluates `tadoku/prod`, whatever tenant the context holds, so the callback
does not need a request tenant; a branch-only administrator cannot authorize
the operator UI.

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
- Waits for the `permission_relationships` table, then inserts the
  administrator's `admins` relationship on `app:root` in `tadoku/prod` with SQL
  through the database pod, keeping an existing one (`on conflict do nothing`).
  These relationships hold roles; native accounts have no role field, and the
  API never reads the `user_roles` row the SQL fixtures write. The reader gets
  no role.
- Loads the application seed data from `scripts/dev/seed/`.

See [Development data](../develop/development-data.md#seed-data) for fixture UUIDs, reruns and branch seeding, and
[Development environment](../develop/environment.md#clean-up) for cleanup.
