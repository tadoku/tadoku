---
title: auth
description: The legacy Tadoku account portal, a Next.js frontend for Ory Kratos registration, login, recovery, verification and account settings for tenants that select v1.
sidebar_position: 3
---

# auth

Read this when you change the legacy Kratos account flows for tenants that select v1.

Source: `frontend/apps/auth`. Development host: `account.tadoku.dev.lab`. Production does not run this app: `account.tadoku.app` redirects to [authv2](authv2.md), and production Kratos is offline.

## Scope

auth handles identity flows only: registration, login and logout, account recovery, email verification, and account settings for profile and password. Product features belong in [webv2](webv2.md) or [admin](admin.md).

Flow pages are `register.tsx`, `login.tsx`, `account-recovery.tsx`, `verification.tsx` and `post-registration.tsx` in `pages/`. Settings live at `pages/index.tsx`.

## Stack

- Next.js 13 Pages Router with Tailwind CSS. The legacy `ui` package supplies the navigation bar and toasts.
- Ory Kratos browser flows run through `@ory/kratos-client`. `src/ui/Flow.tsx` renders each flow's Kratos UI nodes as a react-hook-form form.
- Session state lives in Jotai (`src/session.ts`). Before starting a Kratos flow, the app asks `GET /identity/session` for the tenant's `account_app`. V2 redirects to the sibling authv2 host, preserving the path and query; Settings maps to `/settings` and account recovery to `/recovery`. V1 keeps its Kratos flows. If session selection is unavailable, a retry screen blocks Kratos initialization rather than guessing a mode. SSR forwards Cookie and X-Dev-Branch through the private API endpoint and never caches the response.
- The `next.config.js` `NEXT_PUBLIC_*` variables set the Kratos endpoints, sibling application URLs, API endpoint and cookie domain. They default to the public tadoku.app hosts. `frontend/packages/ui/app-urls.tsx` resolves home/account/admin URLs from the request or browser host; Kratos remains on the base account host.
- Playwright journeys in `e2e/` run against the development stack. See `frontend/apps/auth/e2e/README.md` for the required hosts and environment.

## Commands

From `frontend/`:

```sh
pnpm --filter auth dev       # Next.js default port 3000
pnpm --filter auth lint
pnpm --filter auth exec tsc --noEmit
pnpm --filter auth build
pnpm --filter auth test:e2e  # needs the environment from e2e/README.md
```
