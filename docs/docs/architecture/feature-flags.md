---
title: Feature flags
description: How Tadoku declares boolean feature flags, evaluates them in Tadoku API against Flipt, exposes decisions to webv2 and admins, and what adding a flag requires.
sidebar_position: 4
---

# Feature flags

Read this when you gate behavior behind a flag, change who sees a flagged feature or add a new flag.

This page describes the development environment. Flags are boolean; without a
decision, every consumer uses the flag's behavior-preserving safe default.

## Where flags are declared

| File | Role |
| --- | --- |
| `feature-flags.contract.json` | Cross-stack list of flag keys and safe defaults |
| `services/common/featureflags/registry.go` | Typed Go `BooleanFlag` constants; call sites cannot supply keys or defaults |
| `frontend/apps/webv2/app/feature-flags/registry.ts` | webv2 keys, defaults and response schema |
| `k8s/dev/base/flipt/features.yaml` | Development Flipt seed: flags, rollouts and segments |

`services/common/featureflags/contract_test.go` and
`frontend/apps/webv2/app/feature-flags/registry.test.ts` compare the registries
with the contract. CI runs only the Go test, and only when Bazel inputs change,
so run both locally (`bazel test`, `pnpm --filter webv2 test`).

## Evaluation in Tadoku API

`Evaluator.Boolean` in `services/common/featureflags/` never returns an error to
product code. Anonymous and guest requests get the safe default without asking
Flipt; signed-in users are evaluated with their Kratos ID as the entity ID.

The provider in `services/common/client/flipt/` wraps the Flipt client SDK in
polling mode: it fetches the namespace evaluation snapshot every
`API_FLIPT_UPDATE_INTERVAL` and evaluates in process. A failed fetch keeps the
last snapshot and marks results stale. Startup waits at most
`API_FLIPT_STARTUP_TIMEOUT` before serving safe defaults. See
[Configuration](../tadoku-api/configuration.md) for the `API_FLIPT_*` settings.

Flipt calls pass through Oathkeeper with service JWTs that
`services/common/client/s2s/` exchanges for the `flipt-evaluation` or
`flipt-management` audience ([Service-to-service auth](./service-tokens.md)).
Their routes in `k8s/dev/base/oathkeeper/rules.yaml` allow only the snapshot
read and managed segment calls. Metrics are named `tadoku_feature_flag_*`.

## Decisions in the browser

`GET /immersion/feature-flags` (`ImmersionFeatureFlagDecisions`) returns the
decisions for the current user or guest with `Cache-Control: private, no-store`. Only flags in
`PublicDecisions` (`services/common/featureflags/public.go`) are exposed.
[webv2](../frontend/webv2.md) fetches them server-side in `frontend/apps/webv2/pages/_app.tsx` and
after each client-side route change, falling back to defaults on any failure.
Components use `useFeatureFlag`, or `useLatchedFeatureFlag` to keep the first
value for the component's lifetime.

## Named-user access

Administrators grant or revoke a flag for one user from the admin users page
(`frontend/apps/admin/app/feature-access/`) through
`/immersion/admin/feature-flags/{flagKey}/users/{userId}`. Tadoku API requires
an administrator, accepts only the managed flags in
`services/tadoku-api/features/featureflags/domain.go`, updates the flag's
allowlist segment through the Flipt management API
(`services/tadoku-api/infra/fliptmanagement/`), which rejects a segment that
differs from the Go definition, and audits each change.

## Development Flipt

Flipt runs in `tdk-dev-flipt` from `k8s/dev/base/flipt/` and loads
`features.yaml` into in-memory storage when the pod starts; recreating the pod
discards changes. The operator UI at `https://flags.tadoku.dev.lab` requires a
Kratos session of an administrator, checked by Oathkeeper through Tadoku API.
Flipt is shared: do not change flag policy for other developers. HTTP E2Es use `services/tadoku-api/internal/testflipt/` instead.

## Adding a flag

1. Add the key and `safeDefault` to `feature-flags.contract.json`.
2. Add a `BooleanFlag` constant and definition in `registry.go`; extend `contract_test.go`.
3. Add the key to `featureFlagRegistry` and `featureFlagDecisionsSchema` in webv2 `registry.ts`, in contract order.
4. Add the flag, disabled, to `k8s/dev/base/flipt/features.yaml`.
5. To gate backend behavior, evaluate it in the application layer via the `featureflags` feature service.
6. To expose it to the browser, add it to both `PublicDecisions` types and the
   service mapping, the `ImmersionFeatureFlagDecisions` schema in
   `services/tadoku-api/spec/openapi.yaml` and `services/tadoku-api/transport/http/feature_flags.go`.
   Run `./scripts/generate-openapi.sh` and `pnpm api:generate`, and update the
   `ImmersionFeatureFlagDecisions` goldens.
7. For named-user access, add a segment and rollout to `features.yaml`, the key
   and segment to `domain.go`, the `ImmersionManagedFeatureFlagKey` enum, the
   `frontend/apps/admin/app/feature-access/contracts.ts` schema, and an Oathkeeper rule for reading the segment.
