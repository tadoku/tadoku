# admin

The Tadoku administration app at https://admin.tadoku.app. It manages CMS pages, posts and announcements, users, languages and feature access. It is built with Next.js 13, uses the legacy `ui` package, and requires the `admin` role.

From `frontend/`:

```sh
pnpm --filter admin dev      # port 3000
pnpm --filter admin lint
pnpm --filter admin exec tsc --noEmit
pnpm --filter admin test
pnpm --filter admin build
```

See [the admin docs](../../../docs/docs/frontend/admin.md).
