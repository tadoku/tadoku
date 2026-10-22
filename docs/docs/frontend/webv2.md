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
- React Query 3 hooks in each feature's `api.ts` call Tadoku API through the [generated client](api-client.md), which validates every response.
- The API contract is `services/tadoku-api/spec/openapi.yaml` (see the [API reference](../api/index.md)). webv2 uses its `immersion`, `content` and `authz` operations.
- Sessions come from `GET /identity/session` and are held in Jotai. Login, registration and settings links open [authv2](authv2.md); logout posts to `POST /identity/logout`.
- Forms use react-hook-form with `@hookform/resolvers` for Zod. The app also uses Luxon for dates, Chart.js for charts, Tailwind CSS, and the legacy `ui` components.
- Tests run on Vitest with jsdom and Testing Library.

## Configuration

`next.config.js` maps the `NEXT_PUBLIC_*` variables to `publicRuntimeConfig`. These cover the API endpoint and the admin and home URLs; the authv2 URL is the `authv2.` sibling of the home URL. `NEXT_SERVER_API_ENDPOINT` overrides the API endpoint for server-side requests. The defaults point at the public tadoku.app hosts, and the development stack sets them for `*.tadoku.dev.lab`.

`frontend/packages/ui/app-urls.tsx` resolves sibling URLs and the public API
endpoint per host. SSR supplies those URLs through its provider; browser API
calls resolve them at call time. The private server API endpoint stays on its configured host. Session lookup
forwards the incoming Cookie and X-Dev-Branch headers, including for guests.
Session responses are private and never cached; frontend code does not clear cookies
on an API 401. The API owns session revocation and cookie clearing.

## Commands

From `frontend/`:

```sh
pnpm --filter webv2 dev      # port 3000
pnpm --filter webv2 lint
pnpm --filter webv2 exec tsc --noEmit
pnpm --filter webv2 test
pnpm --filter webv2 build
```
