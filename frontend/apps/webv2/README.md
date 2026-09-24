# webv2

The main Tadoku web application at https://tadoku.app. It covers logging, contests, leaderboards, profiles, the blog and CMS pages. It is built with Next.js 13, uses the legacy `ui` package, and calls Tadoku API.

From `frontend/`:

```sh
pnpm --filter webv2 dev      # port 3000
pnpm --filter webv2 lint
pnpm --filter webv2 exec tsc --noEmit
pnpm --filter webv2 test
pnpm --filter webv2 build
```

See [the webv2 docs](../../../docs/docs/frontend/webv2.md).
