---
title: auth
description: The Tadoku account portal, a Next.js frontend for Ory Kratos registration, login, recovery, verification and account settings.
sidebar_position: 3
---

# auth

Read this when you change registration, login, account recovery, verification or account settings at https://account.tadoku.app.

Source: `frontend/apps/auth`. Development host: `account.tadoku.dev.lab`.

## Scope

auth handles identity flows only: registration, login and logout, account recovery, email verification, and account settings for profile and password. Product features belong in [webv2](webv2.md) or [admin](admin.md).

Flow pages are `register.tsx`, `login.tsx`, `account-recovery.tsx`, `verification.tsx` and `post-registration.tsx` in `pages/`. Settings live at `pages/index.tsx`.

## Stack

- Next.js 13 Pages Router with Tailwind CSS. The legacy `ui` package supplies the navigation bar and toasts.
- Ory Kratos browser flows run through `@ory/kratos-client`. `src/ui/Flow.tsx` renders each flow's Kratos UI nodes as a react-hook-form form.
- Session state lives in Jotai (`src/session.ts`). The app does not call Tadoku API.
- The `next.config.js` `NEXT_PUBLIC_*` variables set the Kratos endpoints, home URL and cookie domain. They default to the public tadoku.app hosts.
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
