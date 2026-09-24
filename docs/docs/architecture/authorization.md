---
sidebar_position: 3
title: Authentication and authorization
description: How Tadoku API authenticates Kratos users from gateway JWTs and authorizes them with Keto administrator and ban relations.
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
  `session` claim; `type` is `user`.
- **Ory Keto** owns authorization facts: who is an administrator and who is
  banned. Tadoku API reads them from Keto on every request that needs them. Roles
  are not stored in PostgreSQL or in the token, and results are not cached.

## Keto data model

| Field | Value |
| --- | --- |
| namespace | `app` |
| object | `tadoku` |
| relations | `admins`, `banned` |
| subject | direct `subject_id`: the Kratos identity ID from the JWT `sub`, never an email |

Example tuples are `app:tadoku#admins@<identity-id>` and
`app:tadoku#banned@<identity-id>`. A regular user holds neither relation.

The namespace configuration (OPL) is `k8s/dev/base/keto/namespaces.keto.ts` for
the development environment and `infra/dev/ory/namespaces.keto.ts` for backend
test fixtures. Tadoku API calls Keto through `services/common/client/keto/`:
`NewReadClient` for checks and `NewClient` for read/write access. Keto's `403`
answer to a check means "denied", not an error.

Administrators can toggle only the `banned` relation through the API
(`PUT /authz/users/{id}/role`), and cannot change another administrator's role. No API
grants `admins`; that tuple is written directly through the Keto write API, as
the development seed does.

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
   (`services/tadoku-api/internal/identity/`). This step checks no roles.
2. **Ban gate** (`services/tadoku-api/transport/http/banned_users.go`) looks up
   `app:tadoku#banned` once for every subject except an empty one or `guest`.

| Ban lookup | Result |
| --- | --- |
| Not banned | The request continues. |
| Banned | Empty `403`, including for administrators. Only `GET /authz/current-user/role` continues, and reports the role `banned`. |
| Keto error | Logged; the request continues with the ban state recorded as unknown. Strict checks then return `503`. |

## Checks in application operations

Application operations in `services/tadoku-api/app/` own caller authorization; no
HTTP middleware enforces administrator access. They call the
`*permissions.Checker` from `services/tadoku-api/internal/permissions/`:

| Method | Passes when | Otherwise |
| --- | --- | --- |
| `RequireAuthenticated` | The caller is a non-guest user and the ban lookup succeeded. | `401` for no user or `guest`; `503` if the ban state is unknown. |
| `RequireAuthenticatedAllowingUnknownBan` | The caller is a non-guest user. | `401`. Read-only operations only; mutations must never use it. |
| `RequireAdmin` | `RequireAuthenticated` passes and the caller holds `admins`. | As above, `403` for non-administrators, `503` on Keto errors. |
| `IsAdmin`, `IsAdminOrFalse` | Report administrator status to expand behavior inside an already-authorized operation. | `IsAdmin` returns unavailable on errors; `IsAdminOrFalse` returns `false`. |

Feature services may inspect permissions only to expand behavior inside an
operation the application has already authorized, and must not repeat the ban
gate. Facts about other users, such as whether a target user is an administrator,
come from `services/common/authz/roles/` (`KetoService` for reads, `KetoManager`
for writes). Target facts are never caller authorization.

## Oathkeeper administrator callback

Operator routes, such as the Flipt UI on `flags.tadoku.dev.lab`, require a Kratos
session. Oathkeeper's `remote_json` authorizer then posts the session subject to
`POST /authz/internal/v1/proxy/admin-check`. This callback route bypasses the JWT
and ban pipeline and never creates a user identity. It accepts only the shared
bearer `API_OATHKEEPER_AUTHZ_TOKEN` (development Secret `dev-oathkeeper-authz`,
also mounted into Oathkeeper), checks only the subject's `admins` relation, and
returns `200` for administrators and `403` otherwise.

## Error to HTTP status mapping

Application errors are `services/tadoku-api/internal/errx/` kinds, mapped in
`services/tadoku-api/transport/http/errors.go`.

| Cause | Status |
| --- | --- |
| Missing or malformed `Authorization: Bearer` header | `400`, JSON `missing or malformed jwt` |
| Invalid, expired, too old or service JWT | `401`, JSON `invalid or expired jwt` |
| Confirmed ban | `403`, empty body |
| Invalid or missing callback credential | `401`, empty body |
| `errx.Unauthorized` (no user or `guest`) | `401` |
| `errx.Forbidden` (not an administrator) | `403` |
| `errx.Unavailable` (Keto error, unknown ban state) | `503` |
| `errx.InvalidInput` | `400` |
| Request deadline exceeded | `504` |

## Seeding an administrator in development

Run `make dev-seed` (`scripts/dev/seed-db.sh`) once Kratos and Keto are ready. It
only runs against the `homelab-dev` Kubernetes context and is safe to re-run.

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
