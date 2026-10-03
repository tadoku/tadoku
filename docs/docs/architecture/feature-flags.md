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
| `services/tadoku-api/internal/featureflags/registry.go` | Typed Go `BooleanFlag` constants; call sites cannot supply keys or defaults |
| `frontend/apps/webv2/app/feature-flags/registry.ts` | webv2 keys, defaults and response schema |
| `k8s/dev/base/flipt/features.yaml` | Development Flipt seed: flags, rollouts and segments |

`services/tadoku-api/internal/featureflags/contract_test.go` and
`frontend/apps/webv2/app/feature-flags/registry.test.ts` compare the registries
with the contract. CI runs only the Go test, and only when Bazel inputs change,
so run both locally (`bazel test`, `pnpm --filter webv2 test`).

## Evaluation in Tadoku API

`Evaluator.Boolean` in `services/tadoku-api/internal/featureflags/` accepts a
verified subject and never returns an error to product code. Anonymous and guest
requests get the safe default without asking Flipt; signed-in users are evaluated
with their Kratos ID as the entity ID. The old common `UserIdentity` type is gone.

The provider in `services/tadoku-api/infra/flipt/` resolves the verified tenant
before evaluating a flag. `tadoku/prod` keeps the configured canonical target:
`production/default` in production and `local/default` in development. Other
tenants use `API_FLIPT_TEST_ENVIRONMENT` (default `test`) and their own namespace.
The single `Targets` mapping replaces the tenant's `/` with `_` for the Flipt
namespace key; `e2e/run-1` becomes `e2e_run-1`. Tenant segments cannot contain
underscores, so distinct tenant keys cannot collide. Keep the original
`<name>/<id>` as the namespace's display name. Do not create a literal slash
namespace: Flipt v2.11 accepts that management write but its feature schema
rejects the key, breaking evaluation snapshots for the environment.

The provider wraps the Flipt client SDK in polling mode: it fetches the resolved
namespace evaluation snapshot every
`API_FLIPT_UPDATE_INTERVAL` and evaluates in process. A failed fetch keeps the
last snapshot and marks results stale. Startup waits at most
`API_FLIPT_STARTUP_TIMEOUT` before serving safe defaults. The canonical client is
created at startup. Test clients are created on first evaluation, retained in a
16-client LRU cache and closed on eviction or shutdown. A missing tenant never
falls back to canonical flags. Test clients have no provider-lifecycle observer,
so `tadoku_feature_flag_config_age_seconds` continues to describe canonical
configuration freshness. See [Configuration](../tadoku-api/configuration.md) for the `API_FLIPT_*` settings.

Flipt calls pass through Oathkeeper with service JWTs that
`services/common/client/s2s/` exchanges for the `flipt-evaluation` or
`flipt-management` audience ([Service-to-service auth](./service-tokens.md)).
Their routes in `k8s/dev/base/oathkeeper/rules.yaml` allow only the snapshot
read and managed segment calls. Metrics are named `tadoku_feature_flag_*`.

## Decisions in the browser

`GET /immersion/feature-flags` (`ImmersionFeatureFlagDecisions`) returns the
decisions for the current user or guest with `Cache-Control: private, no-store`. Only flags in
`PublicDecisions` (`services/tadoku-api/internal/featureflags/public.go`) are exposed.
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
(`services/tadoku-api/infra/fliptmanagement/`) at the same tenant target used for
evaluation. A test grant or revoke does not change the canonical allowlist. The
client rejects a provider response from another namespace, or a segment that
differs from the Go definition, and audits each change.

## Development Flipt

Flipt runs in `tdk-dev-flipt` from `k8s/dev/base/flipt/` and loads
`features.yaml` into in-memory storage when the pod starts; recreating the pod
discards changes. The separate `test` environment uses in-memory storage with
no Git remote; its namespaces also disappear when the pod restarts. Production
uses the same separate memory-only test environment, so test flags are never
committed or pushed to the canonical feature-flags repository. The operator UI at `https://flags.tadoku.dev.lab` requires a
Kratos session of an administrator, checked by Oathkeeper through Tadoku API.
Flipt is shared: do not change flag policy for other developers. HTTP E2Es use `services/tadoku-api/internal/testflipt/` instead.

Every dev-cli branch, including frontend-only work, provisions its own
`test/tadoku_<route>` namespace with display name `tadoku/<route>`. The `tenant`
hook publishes the branch's own `features.yaml` in its tenant image, converts
the seed to Flipt v2 resources and creates missing segments and flags. Repeating
the task preserves existing resources and named-user grants. Evaluation and
admin grant changes use that namespace when the main hostname is selected;
canonical development requests still use `local/default`.

`dev down` and expired-route cleanup delete the owned test namespace. After a
Flipt restart, rerun `dev task --owner <owner> tenant` to restore missing
resources. Do not edit canonical flags to repair a branch or copy another
tenant's members.


## Provisioning a test namespace

Provisioners use Flipt's management API directly with their own authorized
operator or provisioning credentials. Tadoku API's Oathkeeper service-token rules
permit evaluation snapshots and the managed segment GET/PUT operations; they do
not permit namespace creation or deletion.

1. Parse the tenant key and use `Targets` to obtain the test environment and
   mapped namespace key. The test environment must differ from the canonical
   environment. For `e2e/run-1`, create the namespace with
   `POST /api/v2/environments/test/namespaces` and body
   `{"key":"e2e_run-1","name":"e2e/run-1"}`. Creation returns `409` when it
   already exists.
2. Create each segment and flag from the branch's
   `k8s/dev/base/flipt/features.yaml` using
   `POST /api/v2/environments/test/namespaces/e2e_run-1/resources`. Send the
   resource key and typed `payload`, such as
   `{"key":"release-log-entry-v2-access","payload":{"@type":"flipt.core.Segment",...}}`.
   Translate the seed into Flipt's v2 resource schema: a segment rollout needs
   `type: "SEGMENT_ROLLOUT_TYPE"`, `segment.segments` and
   `segment.segmentOperator`. Raw seed YAML field names are not the v2 JSON
   contract. `PUT` updates existing resources; creating a resource under a
   missing namespace returns `500`.
3. Verify
   `GET /internal/v1/evaluation/snapshot/namespace/e2e_run-1` with header
   `x-flipt-environment: test`. An unprovisioned namespace returns `404` and
   evaluation uses safe defaults. Test namespaces never share canonical members
   or revisions.
4. Delete only the owned namespace with
   `DELETE /api/v2/environments/test/namespaces/e2e_run-1` during teardown.
   Provision it again after a Flipt restart.

These calls do not write the production flags repository or modify canonical
flags. Test environment namespaces are visible to administrators in the shared
operator UI.

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
