# auth

The Tadoku account portal at https://account.tadoku.app. It is a Next.js 13 frontend for Ory Kratos registration, login, recovery, verification and account settings, and it uses the legacy `ui` package.

From `frontend/`:

```sh
pnpm --filter auth dev       # port 3000
pnpm --filter auth lint
pnpm --filter auth exec tsc --noEmit
pnpm --filter auth build
pnpm --filter auth test:e2e  # Playwright; see e2e/README.md
```

See [the auth docs](../../../docs/docs/frontend/auth.md).
