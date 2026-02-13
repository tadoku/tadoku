---
sidebar_position: 3
title: Authorization (Keto)
---

# Authorization (Ory Keto)

This document describes how Tadoku uses Ory Keto for authorization, specifically for user roles (admin/banned).

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

- `/Users/io/xdev/tadoku/infra/dev/ory/namespaces.keto.ts`

Example tuples:

- Admin: `app:tadoku#admins@<kratos_subject_id>`
- Banned: `app:tadoku#banned@<kratos_subject_id>`

Important detail: the **Keto subject id** we use is the **Kratos identity id** from the JWT `sub` claim (not an email).

## Backend Integration

### Keto client wrapper

The shared Keto client lives at:

- `/Users/io/xdev/tadoku/services/common/client/keto/`

Key points:

- Services that only read roles use a read-only client (`NewReadClient`).
- Services that manage roles use a combined read+write client (`NewClient`) via the `AuthorizationClient` interface.
- Direct subjects are sent using Keto's `subject_id` field (not `subject_set.*`).

### Roles service (claims)

Role evaluation is implemented in:

- `/Users/io/xdev/tadoku/services/common/authz/roles/`

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

- `/Users/io/xdev/tadoku/services/common/domain/errors.go`

REST handlers map these to status codes via:

- `/Users/io/xdev/tadoku/services/common/http/httperr/httperr.go`

Relevant mappings:

- `ErrUnauthorized` -> `401`
- `ErrForbidden` -> `403`
- `ErrAuthzUnavailable` -> `503`

## Local Dev: Seeding an Admin

Tilt runs an idempotent Kubernetes Job at startup to ensure a local dev admin exists:

- `/Users/io/xdev/tadoku/infra/dev/ory/keto_seed_admin_job.yaml`

Behavior:

- Looks up the Kratos identity id for `SEED_ADMIN_EMAIL` (default `dev@tadoku.app`) using the Kratos admin API.
- Seeds `app:tadoku#admins@<kratos_subject_id>` into Keto using the write admin API.
- If the tuple already exists, Keto returns `409` and the job treats that as success.
