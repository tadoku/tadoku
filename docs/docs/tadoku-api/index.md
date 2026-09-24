---
title: Tadoku API
description: What the Tadoku API service owns, its public API domains, where its contract, persistence and features live, and which content is managed in the CMS instead of code.
---

# Tadoku API

Read this when you change backend behavior, the public HTTP contract or the
database schema.

Tadoku API (`services/tadoku-api/`) is the only backend. It serves every public
HTTP operation, and it runs the leaderboard outbox worker that keeps the Valkey
leaderboard caches current.

## Public API domains

The public contract is grouped into four domains. Each has a filtered view in
the [API reference](../api/index.md).

| Domain | Covers | Features |
| --- | --- | --- |
| Immersion | Logs, contests, registrations, leaderboards, scoring, languages and user profiles | `logs`, `contests`, `leaderboard`, `scoring`, `languages`, `profile` |
| Content | Blog posts, CMS pages and announcements | `posts`, `pages`, `announcements` |
| Profile | The administrator user list | `profile` |
| Authorization | The current user's role, role management and permission checks | `authz` |

Features live under `services/tadoku-api/features/<feature>/`. Shared business
concepts live under `services/tadoku-api/domain/`, and application operations
that compose features live under `services/tadoku-api/app/`.

## Contract and persistence

- The canonical OpenAPI contract is `services/tadoku-api/spec/openapi.yaml`.
  Server code is generated from it with `./scripts/generate-openapi.sh`.
- PostgreSQL holds all application data. Migrations live in
  `services/tadoku-api/migrations/`. Queries live in
  `services/tadoku-api/sql/<feature>/`, and sqlc generates Go code into
  `services/tadoku-api/generated/sqlc/<feature>/`.
- Valkey holds leaderboard caches only; PostgreSQL remains the source of truth.

## CMS-managed content

Public pages such as Contact and About store their HTML in the CMS (`pages` and
`pages_content`, namespace `tadoku`). That copy is edited in the admin CMS only;
see [Contributing workflow](../develop/contributing.md#cms-managed-content).

## More detail

The architecture, conventions, runtime configuration and testing approach are
documented in `services/tadoku-api/README.md`.
