---
title: Frontend overview
description: The applications and shared packages in the frontend pnpm workspace, which design system each uses, and how to run them.
sidebar_position: 1
---

# Frontend overview

Read this when you start frontend work and need to find the right application, package or command.

`frontend/` is a single pnpm workspace: applications live in `frontend/apps/` and shared packages in `frontend/packages/`.

| Application | Purpose | Stack | Design system |
| --- | --- | --- | --- |
| [webv2](webv2.md) | Main site: logs, contests, leaderboards, profiles, blog and CMS pages | Next.js 13 | `ui` (legacy) |
| [auth](auth.md) | Account portal for Ory Kratos | Next.js 13 | `ui` (legacy) |
| [admin](admin.md) | Administration: CMS content, users, languages, feature access | Next.js 13 | `ui` (legacy) |
| [paper-styleguide](paper-styleguide.md) | Catalogue for `paper-ui` | Vite, React Router | `paper-ui` (Paper) |
| [paper-playground](paper-playground.md) | Unpublished Paper application study with sample data | Vite, React Router | `paper-ui` (Paper) |
| [styleguide](styleguide.md) | Reference for the `ui` components | Next.js 13 | `ui` (legacy) |

- `frontend/packages/ui` is the legacy design system: React components, Tailwind configuration and global CSS. The Next.js apps compile it from source through `transpilePackages`.
- `frontend/packages/paper-ui` is the Paper design system: framework-neutral controls, tokens, icons, a Tailwind preset and the catalogue registry. Consumers import the tsup build in `dist/`, which each Paper app's `predev` script rebuilds.

`frontend/paper-boundaries.json` marks each application as `legacy` or `paper`. `pnpm check:paper-boundaries` rejects `paper-ui` in legacy applications, and rejects `ui`, Next.js and Headless UI in Paper code.

## Commands

Always use pnpm, never npm. From `frontend/`:

```sh
pnpm install --frozen-lockfile
pnpm webv2                     # also auth, admin, styleguide, paper-styleguide, paper-playground
pnpm --filter <app> lint
pnpm --filter <app> test       # when the app defines it
pnpm build                     # runs build in every workspace package that defines it
pnpm check:paper-boundaries
```

The Next.js apps have no typecheck script, so use `pnpm --filter <app> exec tsc --noEmit`. The Paper apps have `pnpm --filter <app> typecheck`. Each application page lists its own commands.

To run webv2, auth or admin against the shared development backend, use DevCLI as described in [Local environment](../local-environment.md).

## Conventions

`AGENTS.md` is the binding source. In summary, legacy apps build on `ui` components and `btn` classes (`primary`, `secondary`, `danger`, `ghost`) instead of custom styles. Forms use `react-hook-form` with the `ui` form inputs. Paper apps use `paper-ui` and follow [Paper composition](paper-composition.md).
