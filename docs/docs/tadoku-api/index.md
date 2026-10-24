---
title: Tadoku API
description: What the Tadoku API service owns, its public API domains, how its code is layered, where its contract and persistence live, and which content is managed in the CMS instead of code.
sidebar_position: 1
---

# Tadoku API

Read this when you change backend behavior, the public HTTP contract or the
database schema, and need to know which page covers it.

Tadoku API (`services/tadoku-api/`) is the only backend. It serves every public
HTTP operation, which the gateway exposes under `/api/internal`. A separate
`tadoku-worker` process executes its typed background jobs, beginning with
leaderboard cache invalidation.

## Public API domains

The public contract is grouped by domain. Each has a filtered view in
the [API reference](../api/index.md).

| Domain | Covers | Features |
| --- | --- | --- |
| Immersion | Logs, contests, registrations, leaderboards, scoring, languages and user profiles | `logs`, `contests`, `leaderboard`, `scoring`, `languages`, `profile` |
| Content | Blog posts, CMS pages and announcements | `posts`, `pages`, `announcements` |
| Profile | The administrator user list | `profile` |
| Authorization | The current user's role, role management and the administrator callback | `authz` |
| Identity | Registration, sessions, recovery and profile/password settings | `identity` |
| Email | Transactional templates, delivery, audited reads and retention | `email` |

## Code ownership

```text
transport/http -> app -> features/<feature> -> generated/sqlc/<feature>
app and features/<feature> -> domain/<concept>
cmd/tadoku-api constructs and owns the pgx/v5 pool, the Valkey client and the HTTP resources
```

All paths are relative to `services/tadoku-api/`.

| Package | Owns |
| --- | --- |
| `transport/http/` | The router (`router.go`) with standard method/path registrations, request deadlines and health checks; JWT authentication and the ban gate; mapping application errors to HTTP statuses. |
| `app/` | Application operations: actor authorization, cross-feature locks and transactions, and composition of feature results. |
| `features/<feature>/` | One feature: a service that owns its business decisions, a repository that queries and maps its rows and, where needed, a cache or store for non-authoritative data. |
| `generated/` | Generated code: sqlc queries per feature (`generated/sqlc/<feature>/`) and HTTP bindings (`generated/openapi/`). |
| `app/worker/` | Typed handlers, registration, bounded execution and feature composition. |
| `cmd/tadoku-worker/` | Worker composition root and private health/metrics lifecycle. |
| `domain/<concept>/` | Business values and pure rules shared by several features. |
| `internal/` | Technical support: errors (`errx`), request identity (`identity`), the permission engine that evaluates namespaces features define (`permissions`), feature-flag evaluation (`featureflags`), business time (`timex`), callback authentication (`callbackauth`) and test fixtures (`test*`). |
| `infra/` | Infrastructure adapters: PostgreSQL (`postgres`), permission relationship storage (`relationships`), Valkey (`valkey`), Flipt evaluation and management (`flipt`, `fliptmanagement`), and scoring observability (`observability`). |
| `cmd/tadoku-api/` | The composition root: loads configuration, constructs and owns the pool, provider clients and HTTP resources, and wires them into the application. |

The service-token exchange client and PostgreSQL migration commands remain in
`services/common/` so another backend can use the same infrastructure contract.

## Contract and persistence

- The canonical OpenAPI contract is `services/tadoku-api/spec/openapi.yaml`.
- PostgreSQL holds all application data. Migrations live in
  `services/tadoku-api/migrations/`. Queries live in
  `services/tadoku-api/sql/<feature>/`, and sqlc generates Go code into
  `services/tadoku-api/generated/sqlc/<feature>/`.
- Valkey holds leaderboard caches only; PostgreSQL remains the source of truth.

## CMS-managed content

Public pages such as Contact and About store their HTML in the CMS (`pages` and
`pages_content`, namespace `tadoku`). That copy is edited in the admin CMS only;
see [Contributing workflow](../develop/contributing.md#cms-managed-content).

## Pages in this section

- [Conventions](./conventions.md): read before adding or changing an
  application operation, a feature service or repository, or a shared domain
  package.
- [Go style](./go-style.md): read before writing or changing any handwritten
  Go, including tests; covers wrapping, blank-line grouping and SQL in Go.
- [Contract and OpenAPI](./contract.md): read before changing
  the OpenAPI contract, a request or response shape, or generated HTTP code.
- [Contest event feeds](./contest-events.md): poll submission changes and read
  current contest log contents for integrations.
- [Database and migrations](./database.md): read before writing a migration,
  changing a SQL query or opening a transaction.
- [Jobs and worker](./jobs.md): read before adding background work or changing
  how a job is defined, published or handled.
- [Operate jobs and the worker](./jobs-operations.md): read before checking
  worker health, running jobs on a branch or replaying failed work.
- [Migrate a job contract from v1 to v2](./jobs-versioning.md): read before
  changing a persisted message version.
- [Testing](./testing.md): read before adding or changing any test; covers
  principles, test infrastructure and repository tests.
- [HTTP end-to-end tests](./http-e2e.md): read before adding or changing an
  operation's golden cases, fixture tokens, relationships or local identity fixtures.
- [User journeys](./user-journeys.md): read before changing a chained user
  journey or its fixtures.
- [Import boundaries](./import-boundaries.md): read before adding an import, a
  package or a feature, or changing Bazel visibility.
- [Runtime configuration](./configuration.md): read before changing startup,
  environment variables, authentication wiring or provider clients.
- [Job worker configuration](./worker-configuration.md): read before changing
  the leaderboard worker's settings, tenant scope or metrics.
- [Tenant lifecycle command](./tenant-lifecycle.md): read before changing test
  tenant provisioning, teardown or worker overrides.

Access rules for operations are documented in
[Authorization](../architecture/authorization.md).
