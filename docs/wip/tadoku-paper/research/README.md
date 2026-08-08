# Tadoku Paper research packet

Starting commit: `92083ac7fa3972648d154f78b4d7e398bb985502` (`origin/main`, 2026-08-08).

This packet freezes the evidence and reversible architecture decisions required by Phase 0 of the [implementation plan](../implementation-plan.md). Later phases should update a record only when new evidence invalidates it; product and visual decisions remain in the [decision log](../decision-log.md).

[Phase 0 gate report](phase-0-gate.md): passed 2026-08-08.

## Architecture decisions

- [Package, export, and router boundary](adr-001-package-boundary.md)
- [Tokens and shared CSS recipes](adr-002-tokens-and-recipes.md)
- [Catalogue, fixture, and route contracts](adr-003-catalogue-contract.md)
- [Dependency and build baseline](adr-compatibility-dependency-baseline.md)
- [Base UI import and wrapper boundary](adr-base-ui-import-boundary.md)
- [Native versus Base UI ownership](adr-native-base-ui-ownership.md)

## Evidence

- [Asset provenance and production exports](asset-provenance.md)
- [Compatibility matrix](compatibility-matrix.md) and [primitive spike findings](compatibility-primitive-spike.md)
- [Legacy public/deep API inventory](legacy-ui-api-inventory.md), [consumer matrix](legacy-consumer-matrix.md), [selector/class inventory](legacy-selector-and-class-inventory.md), and [Headless UI inventory](headless-ui-inventory.md)
- [Legacy route and example migration](legacy-styleguide-route-migration.md) and [application risk/smoke matrices](application-risk-and-smoke-matrices.md)
- [Migration marker/check design](migration-marker-and-check-proposal.md)
- [Deployment topology](deployment-topology.md), [CI and delivery boundaries](ci-delivery-boundaries.md), [resource measurement plan](resource-measurement-plan.md), and [rollback runbook](rollback-runbook.md)

## Frozen contracts

- Package name and source boundary: `paper-ui`; React-compatible and independent of Next.js.
- Catalogue application: `paper-styleguide`; React + Vite static application.
- Public component lifecycle: `Experimental`, `Stable`, or `Deprecated`.
- Component routes: `/components/:category/:slug`; foundation routes: `/foundations/:slug`; pattern and experiment routes use their matching top-level segment.
- Themes: `light` and `dark`; densities: `comfortable` and `compact`.
- Button variants: `default`, `outline`, `ghost`, `link`, and `destructive`; loading is a state.
- Primitive policy: native HTML first, Base UI for interaction patterns requiring managed focus, keyboard models, portals, or positioning.
- Coexistence: legacy applications may continue to use `ui`; no deployed application may load both `ui` and `paper-ui`.
