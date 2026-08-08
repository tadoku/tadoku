# ADR 001: Paper package, exports, and router boundary

Status: accepted

## Context

The legacy `ui` package exposes React components, global Tailwind CSS, Next-specific branding and navigation, and Headless UI behavior as one implicit contract. Paper must be a Vite-consumable React package, allow raw CSS-class use, and coexist in the repository without coexisting inside an application.

## Decision

`frontend/packages/paper-ui` is an ESM package with declarations and these explicit entry points:

| Entry point | Contents |
| --- | --- |
| `paper-ui` | React components, recipe helpers, and public types |
| `paper-ui/styles.css` | Fonts, semantic tokens, documented base rules, component recipes, and utilities |
| `paper-ui/tokens.css` | Semantic theme and density variables without component recipes |
| `paper-ui/fonts.css` | Self-hosted `@font-face` declarations |
| `paper-ui/icons` | Curated Heroicons and icon sizing helpers |
| `paper-ui/catalog` | Document metadata and deterministic fixtures; opt-in only |
| `paper-ui/tailwind-preset` | Read-only semantic mappings for Tailwind consumers |

React, React DOM, and React Hook Form are peer dependencies. Base UI and Heroicons are implementation dependencies. Neither `next` nor `@headlessui/react` may be a dependency, import, mock, exported type, or router assumption.

The package owns no router. Navigation receives a consumer-supplied anchor/link renderer and current-route data. Pagination accepts `getHref` or `onPageChange`; breadcrumbs, tab bars, sidebars, and grouped links use the same injected link contract. Ordinary images and inline SVGs replace `next/image`; packaged CSS replaces `next/font`.

Private `paper-ui/src/*` imports are prohibited. Applications load `paper-ui/styles.css` exactly once at their root. During coexistence, migration markers allow `paper-ui` only in `paper-styleguide`; later phases flip an entire application marker only as part of that application's coordinated cutover.

## Build and verification contract

- Output contains ESM JavaScript, `.d.ts`, CSS, fonts, brand SVGs/favicons, and declared entry points.
- A minimal React consumer can import the package and render public components with `react-dom/server`.
- `pnpm pack --dry-run` contains only declared production files.
- Boundary checks reject Next/Headless imports, cross-system mixing, private imports, and duplicate stylesheet entry imports.
- The Tailwind preset consumes public semantic variables and is never the source of component CSS.

## Consequences

The opt-in catalogue entry may contain fixture render functions and metadata without entering application bundles. Router adapters remain application-owned. Repository-level additive rollout is safe, but each deployed application still has a single-system cutover boundary.

