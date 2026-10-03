---
sidebar_position: 3
title: Authentication and authorization
description: How Tadoku API authenticates gateway JWTs and checks tenant-scoped Keto administrator, ban and access permits.
---

# Authentication and authorization

Read this when you add or change an operation's access rules, work with administrator or ban roles, or debug a 400, 401, 403 or 503 from Tadoku API.

The public contract is in the [Authorization API reference](../api/authorization/authz-api);
its source is `services/tadoku-api/spec/openapi.yaml`.

## Authentication and authorization

- **Ory Kratos** owns identities, sessions and login. Browser API calls reach
  Oathkeeper with a Kratos session cookie or no credentials. Oathkeeper's
  `id_token` mutator replaces them with an RS256 JWT whose `sub` is the Kratos
  identity ID, or `guest` for anonymous requests. Session traits travel in the
  `session` claim; `type` is `user`. API tokens carry the signed constant
  `tenant` claim `tadoku/prod` for both anonymous and signed-in requests in
  the development base and production. Client branch headers do not select the
  tenant. Service-token exchanges use their own claims.
- **Ory Keto** owns authorization facts: who is an administrator and who is
  banned. Tadoku API reads them from Keto on every request that needs them. Roles
  are not stored in PostgreSQL or in the token, and results are not cached.

## Kratos identity writes

Kratos identities and sessions are shared across tenants. Test tenants never
modify them. `Writer` in `services/tadoku-api/infra/kratosidentity/` is the only
application path for deactivating an identity, deleting its sessions or deleting
the identity itself. It requires a tenant context: `tadoku/prod` applies the
operation, every other parsed tenant returns `SkippedForTestTenant` and logs
the tenant and identity ID, and a missing tenant returns an error before making
a provider request. Repeating an operation against a missing identity succeeds.

The identity read client in `services/tadoku-api/infra/kratos` exposes only
`FetchIdentity`, `UserExists` and `ListIdentities`. The writer has no HTTP entry
point and is not wired into account deletion. When that flow is added, it must
use the writer and handle the skipped outcome after deleting the tenant's own
application data.

## Keto data model

| Field | Value |
| --- | --- |
| namespace | `app` |
| canonical tenant | `tadoku/prod`, using the existing object `tadoku` |
| test object | The full tenant key, such as `e2e/alpha-0000000a` |
| relations | `admins`, `banned`, `parents`, `testers` |
| permits | `admin`, `is_banned`, `access` |
| subject | direct `subject_id`: the Kratos identity ID from the JWT `sub`, never an email |

The canonical tenant keeps `app:tadoku` and its existing administrator and ban
relations. Tenant keys use `name/id`; the provider object for this existing
canonical tenant remains `tadoku`, without copying or rewriting its tuples.
A regular production user holds neither `admins` nor `banned`.

A test object inherits administrators and bans from its `parents` objects.
Its own administrator or ban tuples apply only to that object. The `access`
permit allows administrators or direct members of `testers`; a tester on one
object has no access to another object without its own grant.

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

## Request pipeline

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

| Tenant and lookup | Result |
| --- | --- |
| Canonical guest or empty subject | Continues without a Keto call. |
| Canonical signed-in user, not banned | One `is_banned` permit lookup; continues. No `access` lookup is added. |
| Canonical Keto error | Logged; continues with the ban state recorded as unknown. Strict checks then return `503`. |
| Test guest or empty subject | Empty `403`, without a Keto call. |
| Test signed-in user | `is_banned` and `access` are checked together through `CheckPermissions`. Both must complete successfully. |
| Test user without `access` | Empty `403`, including on the current-user-role route. |
| Test Keto error | Empty `503`; no fail-open path. |
| Admitted user with confirmed ban | Empty `403`, including for administrators. Only `GET /authz/current-user/role` continues and reports `banned`. |

An administrator or tester on a test object passes its access gate. The
current-user-role exception allows an admitted banned user to discover its
ban; it never grants access to an otherwise inaccessible test tenant.

## Checks in application operations

Application operations in `services/tadoku-api/app/` own actor authorization; no
HTTP middleware enforces administrator access. They call the
`*permissions.Checker` from `services/tadoku-api/internal/permissions/`:

| Method | Passes when | Otherwise |
| --- | --- | --- |
| `RequireAuthenticated` | An actor is present (a signed-in, non-guest user) and the ban lookup succeeded. | `401` for no user or `guest`; `503` if the ban state is unknown. |
| `RequireAuthenticatedAllowingUnknownBan` | An actor is present (a signed-in, non-guest user). | `401`. Read-only operations only; mutations must never use it. |
| `RequireAdmin` | `RequireAuthenticated` passes and the actor holds the tenant object's `admin` permit. | As above, `403` for non-administrators, `503` on Keto errors. |
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
while its access and role checks still inherit the parent permits. Unbanning
a user on a test tenant removes only that object's direct ban; an inherited
production ban continues to apply.

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
| Confirmed ban or missing test-tenant access | `403`, empty body |
| Invalid or missing callback credential | `401`, empty body |
| `errx.Unauthorized` (no user or `guest`) | `401` |
| `errx.Forbidden` (not an administrator) | `403` |
| `errx.Unavailable` (Keto error, missing tenant, unknown ban state, [PostgreSQL outage](../tadoku-api/database.md#postgresql-errors)) | `503` |
| `errx.InvalidInput` | `400` |
| Request deadline exceeded | `504` |

## Seeding an administrator in development

Run `make dev-seed` (`scripts/dev/seed-db.sh`) once Kratos and Keto are ready. It
only runs against the `homelab-talos-dev` Kubernetes context and is safe to re-run.

- Creates or refreshes two Kratos identities marked with
  `metadata_admin.seeded_by=tadoku-dev-seed`: an administrator
  (`TADOKU_DEV_ADMIN_EMAIL`, default `dev@tadoku.app`) and a reader
  (`TADOKU_DEV_READER_EMAIL`, default `reader@tadoku.app`). Passwords come from
  `TADOKU_DEV_ADMIN_PASSWORD` and `TADOKU_DEV_READER_PASSWORD`. An existing
  identity with the same email but no marker is never modified.
- Writes `app:tadoku#admins@<administrator identity ID>` through the Keto write
  API (`http://keto-write.tdk-dev-keto:4467`).
- Loads the application seed data from `scripts/dev/seed/`.

See [Development environment](../develop/environment.md) for branch seeding and
cleanup.
