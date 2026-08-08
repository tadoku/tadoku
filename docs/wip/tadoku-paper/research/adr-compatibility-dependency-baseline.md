# ADR: Paper dependency and build baseline

Status: Accepted for Phase 1; Paper compiler boundary superseded by `adr-typescript-7-tooling-boundary.md`
Date: 2026-08-08

## Context

Tadoku's existing React applications use React 18.2 and TypeScript 4.9.3, while CI and containers use Node 20 and pnpm 10.10.0. Paper needs Base UI, a static Vite catalogue, a router, rendered component tests, ESM declarations, and standalone CSS without forcing upgrades through legacy applications.

Stable Base UI declarations use TypeScript 5 syntax. TypeScript 4.9 fails during parsing, including with `skipLibCheck`, for both Base UI 1.0.0 and 1.7.0.

## Decision

Create an isolated modern toolchain for `paper-ui` and `paper-styleguide` while keeping the repository runtime and React contract unchanged:

- Node 20.x and pnpm 10.10.0;
- React and React DOM 18.2.0;
- TypeScript 5.9.3 in the two new Paper workspaces only;
- `@base-ui/react` 1.7.0;
- Vite 6.4.3 with `@vitejs/plugin-react` 4.7.0;
- `react-router-dom` 7.18.2 in the catalogue application only;
- Vitest 4.1.10, jsdom 27.0.1, React Testing Library 16.3.2, DOM Testing Library 10.4.1, user-event 14.6.3, and jest-dom 6.9.1;
- tsup 8.5.1 for ESM and bundled declarations.

Pin exact versions in the first lockfile change. Preserve pnpm 10.10.0 explicitly in CI and images. Do not use unpinned Corepack.

`paper-ui` emits an ESM distribution and bundled declarations. It copies CSS and assets as independent public files. React, React DOM, and React Hook Form remain peer dependencies. Base UI remains a direct implementation dependency and is externalized by the wrapper build.

Public declarations must not contain Base UI types. A strict TypeScript 4.9 consumer smoke test is a package gate so legacy applications can consume built Paper types before their own compiler is upgraded.

## Consequences

### Benefits

- Paper can use supported stable Base UI without a repository-wide TypeScript migration.
- Vite 6 supports the full existing Node 20 contract; no minimum Node 20 patch must be introduced.
- Legacy applications retain their current compiler and Next builds during coexistence.
- One compact build step emits ESM and a deliberate public declaration surface.

### Costs and risks

- The monorepo temporarily contains TypeScript 4.9 and 5.9 compilers. Workspace scripts and CI filters must invoke the local compiler deterministically.
- TypeScript 4.9 consumer compatibility depends on keeping Base UI types out of Paper declarations; the smoke test is mandatory.
- Vite 6 is selected over newer majors to honor broad Node 20 compatibility. A later Node patch pin should trigger a Vite upgrade review.
- Static CSS/assets need an explicit copy and pack-list check because tsup is not their source of truth.

## Gate evidence

The disposable spike passed Paper TypeScript 5.9 typechecking, five Vitest/jsdom behaviors, tsup ESM/declarations, a Vite static build, React server rendering, strict TypeScript 4.9 declaration consumption, and `pnpm pack --json` file-list validation.
