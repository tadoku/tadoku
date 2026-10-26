---
title: Feature flags
description: Rules for declaring boolean flags, evaluating tenant-scoped decisions, evolving the public contract and managing provider state.
sidebar_position: 4
---

# Feature flags

Read this when you add a flag, change targeting or retire a flagged behavior.

## Declarations and defaults

Declare flag vocabulary centrally in `feature-flags.contract.json`,
`services/tadoku-api/internal/featureflags/registry.go` and
`frontend/apps/webv2/app/feature-flags/registry.ts`. Keep keys, order and safe
defaults consistent. Use typed definitions at call sites rather than supplying
provider keys or defaults independently.

Choose a behavior-preserving safe default. Evaluate through the application
layer; transport and product components must not assemble provider policy.
The contract checks beside the registries protect consistency across languages.

## Evaluation and tenancy

`services/tadoku-api/internal/featureflags/` accepts verified subjects. Guests
receive safe defaults without provider evaluation. Signed-in accounts use their
account IDs. Product code receives a boolean decision rather than a provider
error.

`services/tadoku-api/infra/flipt/` resolves the verified tenant before evaluating.
Canonical requests use the configured environment and namespace. Other tenants
use the separate test environment and the provider's tenant-to-namespace mapping.
A missing tenant must never fall back to canonical state. Preserve the parsed
tenant-key restrictions that make this mapping collision-free.

The provider polls evaluation snapshots and evaluates in process. Accept only
snapshots that identify the requested namespace and satisfy the SDK's required
fields. Install an SDK only after its own initial fetch and engine initialization
succeed. Failed initialization must not install an empty fallback snapshot; retry
through a fresh SDK. After successful initialization, a failed refresh retains the
last valid snapshot and marks decisions stale. Keep test-client caching bounded
and close clients on eviction or shutdown.

Configuration and lifecycle metrics belong to the provider. See
[Configuration](../tadoku-api/configuration.md) for its environment variables and
startup bounds. Application snapshot reads use the evaluation service audience
through Oathkeeper; see [Service-to-service auth](./service-tokens.md).
Provisioners use their separate direct management connection.

## Public decision contract

Expose browser decisions through the typed mapping in
`services/tadoku-api/internal/featureflags/public.go` and the OpenAPI fragments
under `services/tadoku-api/spec/openapi/`. Keep the response private and uncached.
The application uses the evaluator for public decisions and backend consumers.

webv2 fetches decisions during SSR and navigation, using registry defaults when
transport or validation fails. Generated clients reject missing and additional
keys. Adding or removing a public decision therefore changes a strict contract.

Before changing that shape, retire incompatible consumers, enforce a verified
refresh/cutover or retain a compatible endpoint while introducing a versioned
contract. Deploying unconditional product behavior does not retire already-open
JavaScript. Elapsed time alone is not proof that old consumers are gone.

Remove a release check and its obsolete product path only after the replacement
behavior is verified. Preserve behavior still shared by retained callers. Remove
managed API operations, generated clients and administration controls together;
retire older callers before removing their gateway permissions or provider state.

## Provider state and development ownership

`k8s/dev/base/flipt/features.yaml` supplies development declarations. Keep new
rollout policy disabled unless its release is explicitly authorized. Targeting
changes belong in the authorized Flipt operator interface; Tadoku API has no
management endpoint.

Branch provisioning publishes the branch's seed into its owned test namespace.
Provision resources through the management adapter's typed schema, preserving
existing resources and grants on repeated setup. Use the provider's namespace
mapping and validate resource references; do not copy another tenant's members
or revisions. Management requests do not write canonical Git-backed policy.

The test environment is memory-only. A provider restart discards its namespaces;
rerun the owned provisioning task to restore them. Teardown deletes only the
owned namespace. Do not repair a branch by changing canonical policy or restarting
shared services. See [Development environment](../develop/environment.md) for the
owner-scoped commands.

## Changing the contract

Update declarations, typed public mappings and schema fragments together. Run
the contract and generated-client checks, and regenerate the bundled OpenAPI,
SDK and API reference through their documented commands. See
[API contract](../tadoku-api/contract.md) and
[Frontend API client](../frontend/api-client.md).

Verify defaults, provider errors and tenant isolation through the actual API.
Provider-protocol fixtures are simulations, so record that boundary separately
from real-provider and browser verification. Retain meaningful adapter coverage
when a product flag is retired; generic provider tests may use test-only keys.
