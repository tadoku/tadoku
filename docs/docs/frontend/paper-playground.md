---
title: paper-playground
description: An unpublished Paper application study with browser-local sample data and no API connection.
sidebar_position: 6
---

# paper-playground

Read this when you explore or change the Paper application study in `frontend/apps/paper-playground`.

## What it is

paper-playground is a Vite and React Router app that composes `paper-ui` into full Tadoku screens: home, guide, leaderboards, contests, logs and the log editor, profiles, settings, sign-in and admin. Forms use react-hook-form.

All data is sample fixtures (`src/data.ts`), and every action changes only browser-local state (`src/state.tsx`, persisted in `localStorage`). There is no API connection, deployment image or publishing workflow. Do not connect its sample sign-in or administrative flows to live services.

The screens illustrate composition; they are not a shared implementation. A production screen still owns its data, routing and permissions, as described in [Paper composition](paper-composition.md).

The footer's **Playground controls** switch between pages, sample scenarios, the sample account, theme and density. **Reset sample changes** restores the fixtures. The sample date is fixed at 5 September 2026.

## Development

From `frontend/`:

```sh
pnpm paper-playground        # builds paper-ui, then serves Vite on port 5174 (respects HOST and PORT)
pnpm --filter paper-playground lint
pnpm --filter paper-playground typecheck
pnpm --filter paper-playground test
pnpm --filter paper-playground build
```
