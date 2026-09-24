---
title: admin
description: The Tadoku administration app for CMS pages, posts and announcements, users, languages and feature access.
sidebar_position: 4
---

# admin

Read this when you change the administration app at https://admin.tadoku.app.

Source: `frontend/apps/admin`. Development host: `admin.tadoku.dev.lab`.

## Scope

- The dashboard (`pages/index.tsx`) shows the latest official contest, leaderboards and the total user count.
- The CMS (`pages/pages/`, `pages/posts/`, `pages/announcements/`) manages pages, blog posts and announcements per namespace. It includes a CodeMirror editor, version history, and previews sanitised with DOMPurify and rehype-sanitize. CMS copy, such as the Contact and About pages, is changed here rather than in code.
- Users (`pages/users.tsx`) can be listed, banned or unbanned, and granted or denied per-user feature access.
- Languages (`pages/languages.tsx`) manages the immersion language list.

`pages/_app.tsx` requires a Kratos session and the `admin` role from Tadoku API. It sends visitors without a session to the auth login page and shows other users an access-denied screen.

## Stack

- Next.js 13 Pages Router, with feature code in `app/` imported as `@app/*`, Tailwind CSS and the legacy `ui` components.
- React Query 3 hooks call Tadoku API (`immersion`, `content`, `authz` and `profile`) with `fetch`, and Zod parses the responses. There is no generated client. See the [API reference](../api/index.md).
- The app also uses react-hook-form, Luxon, Jotai for the session, and `@ory/kratos-client`.
- Tests run on Vitest with jsdom and Testing Library.
- The `next.config.js` `NEXT_PUBLIC_*` variables work as they do in [webv2](webv2.md).

## Commands

From `frontend/`:

```sh
pnpm --filter admin dev      # port 3000
pnpm --filter admin lint
pnpm --filter admin exec tsc --noEmit
pnpm --filter admin test
pnpm --filter admin build
```
