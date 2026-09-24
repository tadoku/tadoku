---
title: paper-styleguide
description: The static catalogue that documents paper-ui foundations, controls and product patterns.
sidebar_position: 5
---

# paper-styleguide

Read this when you document, review or preview `paper-ui` foundations, controls or product patterns.

Source: `frontend/apps/paper-styleguide`. Published at https://paper.tadoku.app.

## What it is

The Paper catalogue is a static Vite, React 18 and React Router app. It renders the registry exported by `paper-ui/catalog` from `frontend/packages/paper-ui/src/catalog/`. Catalogue entries, examples and pattern guidance live in `paper-ui`. The styleguide owns the documentation shell, search, routing and display preferences in `src/app/` and `src/documentation/`.

Together with `paper-ui`, it forms the Paper design system. It must not depend on Next.js, Headless UI or the legacy `ui` package, and `pnpm check:paper-boundaries` enforces this. [Paper composition](paper-composition.md) explains what belongs in a control, a product pattern or a screen.

## Development

dev-cli has no overlay for this app, so run it locally from `frontend/`:

```sh
pnpm paper-styleguide        # builds paper-ui, then serves Vite on port 5173
pnpm --filter paper-styleguide lint
pnpm --filter paper-styleguide typecheck
pnpm --filter paper-styleguide test
pnpm --filter paper-styleguide build
pnpm check:paper-boundaries
```

The app imports the built `paper-ui` from `dist/`. After you change `frontend/packages/paper-ui`, run `pnpm --filter paper-ui build` or restart the dev server. `paper-ui` has the same `lint`, `typecheck`, `test` and `build` scripts, plus `pack:check`, `examples:check` and `consumer:ts49`.
