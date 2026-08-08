# ADR: TypeScript 7 compiler and TypeScript 6 tooling boundary

Status: Accepted
Date: 2026-08-08

## Context

Paper originally isolated TypeScript 5.9 from the legacy TypeScript 4.9
applications. TypeScript 7.0.2 is now the current compiler, but its native port
does not expose the programmatic API consumed by `typescript-eslint` and
`tsup`'s declaration builder. The TypeScript team recommends installing the
native compiler beside the `@typescript/typescript6` compatibility package
during this transition.

## Decision

Use the official side-by-side package layout in `paper-ui` and
`paper-styleguide`:

- `@typescript/native` aliases `typescript@7.0.2` and provides the `tsc`
  executable used by each package's `typecheck` script;
- `typescript` aliases `@typescript/typescript6@6.0.2`, allowing ESLint and
  declaration-generation tools to import the TypeScript 6 API;
- `paper-ui` retains its separate `typescript-4` alias and strict built-package
  consumer smoke;
- `paper-ui` suppresses TypeScript 6's `baseUrl` deprecation because tsup's
  declaration pipeline supplies that option internally. TypeScript 7 accepts
  the suppression and does not use `baseUrl` in Paper's authored config.

No legacy application compiler changes as part of this decision.

## Consequences

Paper gets the native TypeScript 7 compiler without breaking tools that still
need the TypeScript JavaScript API. The package aliases are intentionally
unusual: maintainers must not remove the TypeScript 6 alias until
`typescript-eslint` and declaration bundling support the TypeScript 7 API.

The dependency graph temporarily contains TypeScript 4.9, the TypeScript 6
compatibility API, and TypeScript 7. This cost is bounded to Paper and covered
by package-local scripts and CI filters.

## Gate evidence

Both packages report TypeScript 7.0.2 from `tsc` and pass lint, typecheck,
Vitest, and production builds. `paper-ui` also passes declaration generation,
the strict TypeScript 4.9 consumer, and package-content validation. Paper
boundaries and the complete frontend workspace build pass unchanged.

Reference: [Announcing TypeScript 7.0 — Running Side-by-Side with TypeScript 6.0](https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/#running-side-by-side-with-typescript-60)
