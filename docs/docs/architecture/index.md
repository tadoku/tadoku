---
sidebar_position: 1
title: System architecture
description: The components of the Tadoku development environment, how browser and service requests reach them, and where each one is defined in the repository.
---

# System architecture

Read this when you need a map of the running system before changing a service,
a route, a frontend or the development environment.

This page describes the development environment on `homelab-dev`, defined in
`k8s/dev/base/`. Production is deployed from a private repository and is not
documented here.

## Components

| Component | Role | Defined in |
| --- | --- | --- |
| Tadoku API | Serves every public HTTP operation and publishes typed jobs in the business transaction. | `services/tadoku-api/`, `k8s/dev/base/services/tadoku-api.yaml` |
| Tadoku worker | Executes registered jobs, composes feature operations and invalidates leaderboard caches. | `services/tadoku-api/cmd/tadoku-worker/`, `k8s/dev/base/services/tadoku-worker.yaml` |
| webv2 | Main site: logging, contests, leaderboards and content | `frontend/apps/webv2/`, `k8s/dev/base/frontend-webv2/` |
| auth | Account portal built on Kratos self-service flows | `frontend/apps/auth/`, `k8s/dev/base/frontend-auth/` |
| admin | Administration, moderation and CMS | `frontend/apps/admin/`, `k8s/dev/base/frontend-admin/` |
| Ory Kratos | Identities, sessions and login flows | `k8s/dev/base/kratos/` |
| Ory Keto | Role relationships such as administrators and bans | `k8s/dev/base/keto/` |
| Ory Oathkeeper | Access proxy for API and operator routes; issues user and service JWTs | `k8s/dev/base/oathkeeper/` |
| token-reflector | Publishes the Kubernetes service-account JWKS and returns exchanged service tokens | `services/token-reflector/` |
| Flipt | Feature flags | `k8s/dev/base/flipt/`, `feature-flags.contract.json` |
| PostgreSQL | Tadoku API, Kratos and Keto databases in one operator-managed server | `k8s/dev/base/data/postgres.yaml` |
| Valkey | Leaderboard cache | `k8s/dev/base/data/cache.yaml` |
| Mailhog | Captures outgoing email | `k8s/dev/base/data/mailhog.yaml` |

Each component runs in its own `tdk-dev-*` namespace. `tdk-dev-routing`
attaches the application routes to the platform Envoy Gateway.

## Request paths

```text
browser → ingress-nginx → Envoy → webv2 / auth / admin
                             → Oathkeeper → Envoy → Tadoku API
```

- **Browser API calls** go to `/api/internal/<domain>/…` on the main host.
  Oathkeeper accepts the Kratos session cookie or an anonymous request, replaces
  it with a signed JWT (`id_token` mutator) and forwards to Tadoku API.
  Tadoku API verifies that JWT against Oathkeeper's JWKS and rejects banned
  users before any operation runs. See
  [Authorization](./authorization.md).
- **Service calls** use short-lived service JWTs from the Oathkeeper token
  exchange. Tadoku API uses them to reach Flipt. See
  [Service-to-service authentication](./service-tokens.md).
- **Operator routes** such as the Flipt UI on `flags.tadoku.dev.lab` require a
  Kratos session, and Oathkeeper asks Tadoku API whether that user is an
  administrator.
- **Branch routing:** Envoy sends each request to a developer's branch overlay
  when one is selected and healthy, and to the base otherwise. See
  [Development environment](../develop/environment.md).

## Data

- Tadoku API owns the application schema. Its migrations live in
  `services/tadoku-api/migrations/` and run as Argo CD sync hooks before the API
  starts.
- PostgreSQL is the source of truth for leaderboards. Valkey holds sorted-set
  caches that the separate Tadoku worker keeps current; see
  [ADR 001](../adr/001-leaderboard.md).
- Kratos and Keto keep their own databases in the same PostgreSQL server.

## Where to go next

- [Module design](./module-design.md) for deep modules and information hiding
- [Tadoku API](../tadoku-api/index.md) for the backend
- [Frontend overview](../frontend/index.md) for the applications and design systems
- [Development environment](../develop/environment.md) to run and verify a branch
- [API reference](../api/index.md) for the public HTTP contract
