---
title: Tadoku API
description: What the Tadoku API service owns, its public API domains, how its code is layered, where its contract and persistence live, and which content is managed in the CMS instead of code.
sidebar_position: 1
---

# Tadoku API

Read this when you change backend behavior, the public HTTP contract or the
database schema, and need to know which page covers it.

Tadoku API (`services/tadoku-api/`) is the only backend. It serves every public
HTTP operation, which the gateway exposes under `/api`, and it runs the
leaderboard outbox worker that keeps the Valkey leaderboard caches current.

## Public API domains

The public contract is grouped into four domains. Each has a filtered view in
the [API reference](../api/index.md).

| Domain | Covers | Features |
| --- | --- | --- |
| Immersion | Logs, contests, registrations, leaderboards, scoring, languages and user profiles | `logs`, `contests`, `leaderboard`, `scoring`, `languages`, `profile` |
| Content | Blog posts, CMS pages and announcements | `posts`, `pages`, `announcements` |
| Profile | The administrator user list | `profile` |
| Authorization | The current user's role, role management and permission checks | `authz` |

## Code ownership

```text
transport/http -> app -> features/<feature> -> generated/sqlc/<feature>
app and features/<feature> -> domain/<concept>
cmd/tadoku-api constructs and owns the pgx/v5 pool, the raw Valkey, Kratos and Keto clients and the HTTP resources
```

All paths are relative to `services/tadoku-api/`.

| Package | Owns |
| --- | --- |
| `transport/http/` | The router (`router.go`) with standard method/path registrations, request deadlines and health checks; JWT authentication and the ban gate; mapping application errors to HTTP statuses. |
| `app/` | Application operations: actor authorization, cross-feature locks and transactions, and composition of feature results. |
| `features/<feature>/` | One feature: a service that owns its business decisions and a repository that queries and maps its rows. |
| `generated/` | Generated code: sqlc queries per feature (`generated/sqlc/<feature>/`) and HTTP bindings (`generated/openapi/`). |
| `storage/postgres/asyncoutbox/` | Shared PostgreSQL repository for typed background tasks, lease-fenced claims and terminal outcomes. |
| `domain/<concept>/` | Business values and pure rules shared by several features. |
| `internal/` | Technical support: typed background tasks (`asyncwork`), errors (`errx`), request identity (`identity`), actor permissions (`permissions`), business time (`timex`), callback authentication (`callbackauth`) and test fixtures (`test*`). |
| `infra/` | Infrastructure adapters: the PostgreSQL pool and transactions (`postgres`), the raw Valkey client (`valkey`), the Flipt management client (`fliptmanagement`) and scoring observability (`observability`). |
| `cmd/tadoku-api/` | The composition root: loads configuration, constructs and owns the pool, provider clients and HTTP resources, and wires them into the application. |

## Contract and persistence

- The canonical OpenAPI contract is `services/tadoku-api/spec/openapi.yaml`.
- PostgreSQL holds all application data. Migrations live in
  `services/tadoku-api/migrations/`. Queries live in
  `services/tadoku-api/sql/<feature>/`, and sqlc generates Go code into
  `services/tadoku-api/generated/sqlc/<feature>/`.
- Valkey holds leaderboard caches only; PostgreSQL remains the source of truth.
- `async_outbox` stores typed tasks with bounded claims and replay lineage.
  A producer enqueues through the shared repository inside its business transaction.
  Claims, renewals and outcomes use one SQL statement each; stale or expired
  claim tokens cannot acknowledge work.

## CMS-managed content

Public pages such as Contact and About store their HTML in the CMS (`pages` and
`pages_content`, namespace `tadoku`). That copy is edited in the admin CMS only;
see [Contributing workflow](../develop/contributing.md#cms-managed-content).

## Pages in this section

- [Conventions](./conventions.md): read before adding or changing an
  application operation, a feature service or repository, or a shared domain
  package.
- [Contract and OpenAPI](./contract.md): read before changing
  `spec/openapi.yaml`, a request or response shape, or generated HTTP code.
- [Database and migrations](./database.md): read before writing a migration,
  changing a SQL query or opening a transaction.
- [Testing](./testing.md): read before adding or changing any test; covers
  principles, test infrastructure and repository tests.
- [HTTP end-to-end tests](./http-e2e.md): read before adding or changing an
  operation's golden cases, fixture tokens, relationships or the Kratos fixture.
- [User journeys](./user-journeys.md): read before changing a chained user
  journey or its fixtures.
- [Import boundaries](./import-boundaries.md): read before adding an import, a
  package or a feature, or changing Bazel visibility.
- [Runtime configuration](./configuration.md): read before changing startup,
  environment variables, authentication wiring, provider clients or the
  leaderboard worker.

Access rules for operations are documented in
[Authorization](../architecture/authorization.md).
