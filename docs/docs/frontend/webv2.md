---
title: webv2
description: The main Tadoku web application for logging, contests, leaderboards, profiles, the blog and CMS pages.
sidebar_position: 2
---

# webv2

Read this when you change the main Tadoku site at https://tadoku.app.

Source: `frontend/apps/webv2`. Development host: `tadoku.dev.lab`.

## Scope

webv2 covers logging, official and user-created contests (including registration and contest leaderboards), latest, yearly and all-time leaderboards, user profiles, the blog, and CMS pages. CMS copy is edited in [admin](admin.md), not in this app.

## Stack

- Next.js 13 Pages Router. Routes live in `pages/`, and feature code lives in `app/`, imported as `@app/*`. The `app/` directory is not the Next.js App Router.
- React Query 3 hooks in each feature's `api.ts` call Tadoku API with `fetch`, and Zod parses every response. There is no generated client.
- The API contract is `services/tadoku-api/spec/openapi.yaml` (see the [API reference](../api/index.md)). webv2 uses its `immersion`, `content` and `authz` operations.
- Sessions come from Ory Kratos through `@ory/kratos-client` and are held in Jotai.
- Forms use react-hook-form with `@hookform/resolvers` for Zod. The app also uses Luxon for dates, Chart.js for charts, Tailwind CSS, and the legacy `ui` components.
- Tests run on Vitest with jsdom and Testing Library.

## Configuration

`next.config.js` maps the `NEXT_PUBLIC_*` variables to `publicRuntimeConfig`. These cover the API endpoint, the Kratos endpoints, the auth, admin and home URLs, and cookie settings. `NEXT_SERVER_API_ENDPOINT` overrides the API endpoint for server-side requests. The defaults point at the public tadoku.app hosts, and the development stack sets them for `*.tadoku.dev.lab`.

## Commands

From `frontend/`:

```sh
pnpm --filter webv2 dev      # port 3000
pnpm --filter webv2 lint
pnpm --filter webv2 exec tsc --noEmit
pnpm --filter webv2 test
pnpm --filter webv2 build
```
