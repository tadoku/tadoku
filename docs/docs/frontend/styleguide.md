---
title: styleguide
description: The reference site for the legacy ui component library used by webv2, auth and admin.
sidebar_position: 7
---

# styleguide

Read this when you use or change a component from the legacy `ui` package.

Source: `frontend/apps/styleguide`. Published at https://ui.tadoku.app.

## What it is

`ui` (`frontend/packages/ui`) and this styleguide make up the legacy design system that webv2, auth and admin still use. New Paper work uses `paper-ui` and [paper-styleguide](paper-styleguide.md) instead.

The styleguide is a Next.js 13 site with one page for each area of `ui`: branding, colour, typography, buttons, forms, modals, flash messages, toasts, navigation, breadcrumbs, pagination, action menus, tables, charts, templates and logging. Each page renders live examples from `examples/` next to their source code. The examples use react-hook-form with Zod resolvers, Luxon, Chart.js, TanStack Table and faker fixtures.

## Using `ui`

1. Add `"ui": "workspace:*"` to the app's dependencies and list `ui` in `transpilePackages` in `next.config.js`.
2. Import `ui/styles/globals.css` once in `_app.tsx`.
3. Import components from `ui`, and form inputs from `ui/components/Form`.

Buttons are classes, not components: `className="btn primary"`, `secondary`, `danger` or `ghost`. A plain `btn` gives the default style.

## Commands

From `frontend/`:

```sh
pnpm styleguide              # port 3002
pnpm --filter styleguide lint
pnpm --filter styleguide exec tsc --noEmit
pnpm --filter styleguide build
pnpm --filter ui test        # tests for the ui package
```
