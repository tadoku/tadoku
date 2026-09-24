---
sidebar_position: 3
title: Authorization (Keto)
---

# Authorization (Ory Keto)

This document describes how Tadoku uses Ory Keto for authorization, specifically for user roles (admin/banned).

The public HTTP contract is available in the
[Authorization API reference](../api/authorization/authz-api), with the
[OpenAPI source](https://github.com/tadoku/tadoku/blob/main/services/tadoku-api/spec/openapi.yaml)
kept in the repository.

## Overview

- **Authentication (who are you?)** is handled by **Ory Kratos**. User JWTs include a stable subject id in the `sub` claim.
- **Authorization (what can you do?)** is handled by **Ory Keto**. We store role membership as Keto relation tuples and evaluate them per request.

In the backend request pipeline, services:

1. Verify the JWT and attach an identity to the request context.
2. Enrich the request context with role claims from Keto (admin/banned).
3. Block banned users.
4. Domain code uses role claims (not a DB/config role field) to authorize actions.

## Keto Data Model Used for Roles

We model global, application-scoped roles under a single object:

- **namespace**: `app`
- **object**: `tadoku`
- **relations**:
  - `admins`
  - `banned`

The OPL (namespace config) lives at:

- `k8s/dev/base/keto/namespaces.keto.ts` for the development GitOps base
- `infra/dev/ory/namespaces.keto.ts` for isolated backend test fixtures

Example tuples:

- Admin: `app:tadoku#admins@<kratos_subject_id>`
- Banned: `app:tadoku#banned@<kratos_subject_id>`

Important detail: the **Keto subject id** we use is the **Kratos identity id** from the JWT `sub` claim (not an email).

## Backend Integration

### Keto client wrapper

The shared Keto client lives at:

- `services/common/client/keto/`

Key points:

- Services that only read roles use a read-only client (`NewReadClient`).
- Services that manage roles use a combined read+write client (`NewClient`) via the `AuthorizationClient` interface.
- Direct subjects are sent using Keto's `subject_id` field (not `subject_set.*`).

### Roles service (claims)

Role evaluation is implemented in:

- `services/common/authz/roles/`

The middleware stores a `roles.Claims` struct on the request context:

- `Authenticated` (derived from identity presence)
- `Admin`
- `Banned`
- `Err` (set when authz evaluation failed, e.g. Keto unavailable)

The primary helpers used by domain code are:

- `roles.RequireAuthenticated(ctx)`:
  - returns `ErrUnauthorized` if not logged in
  - returns `ErrAuthzUnavailable` if we could not evaluate claims
- `roles.RequireAdmin(ctx)`:
  - returns `ErrUnauthorized` if not logged in
  - returns `ErrAuthzUnavailable` if we could not evaluate claims
  - returns `ErrForbidden` if non-admin or banned

Service-specific domain packages typically wrap these (for example `requireAdmin(ctx)` in the domain package).

### Middleware flow

Services wire middleware in this order (see `main.go` in each service):

1. `VerifyJWT(...)`
2. `Identity()` (attaches `domain.UserIdentity` or `domain.ServiceIdentity`)
3. `RolesFromKeto(rolesSvc)` (attaches `roles.Claims` for authenticated users)
4. `RequireServiceAudience(serviceName)` (for service tokens)
5. `RejectBannedUsers()` (blocks banned users with `403`)

Notes:

- `RolesFromKeto` only enriches **user** requests (guests and service identities are skipped).
- `RejectBannedUsers` is **fail-open** if a user is authenticated but role evaluation failed (`claims.Err != nil`): it logs and allows the request to proceed. Admin-only endpoints are still protected by `roles.RequireAdmin`, which will return `ErrAuthzUnavailable`.

## HTTP Error Mapping

Backend domain code returns shared sentinel errors from:

- `services/common/domain/errors.go`

REST handlers map these to status codes via:

- `services/common/http/httperr/httperr.go`

Relevant mappings:

- `ErrUnauthorized` -> `401`
- `ErrForbidden` -> `403`
- `ErrAuthzUnavailable` -> `503`

## Development Environment: Seeding an Admin

Run `make dev-seed` (`scripts/dev/seed-db.sh`) against the development GitOps base.
The shared Kratos and Keto providers must be ready first. Seeding is explicit,
not an automatic startup task.

Behavior:

- Creates or refreshes the marked synthetic Kratos identity for `TADOKU_DEV_ADMIN_EMAIL` (default `dev@tadoku.app`); unmarked identities are never taken over.
- Seeds `app:tadoku#admins@<kratos_subject_id>` into Keto using the write admin API.
- Re-running the seed is idempotent.

DevCLI branch seed tasks reuse these shared identities; they do not provision
per-branch auth providers. See [Development environment](../local-environment.md)
for branch migration/seeding and scoped cleanup. Database reset is not implicit.
