# Tadoku Paper styleguide

This is the static, registry-driven catalogue for `paper-ui`. It deliberately
does not depend on Next.js, the legacy `ui` package, or Headless UI.

## Local development

From `frontend/`:

```sh
pnpm --filter paper-styleguide dev
```

For a temporary review URL inside T3 Code, let `t3-expose` select the host and
port that Vite uses:

```sh
t3-expose run --detach --idle-timeout 1h --max-lifetime 4h -- \
  pnpm --filter paper-styleguide dev
```

The returned HTTPS URL is private-lab/tailnet scoped and unauthenticated. Stop
it when review is complete with `t3-expose stop <session-id>`.

## Verification

```sh
pnpm --filter paper-styleguide lint
pnpm --filter paper-styleguide typecheck
pnpm --filter paper-styleguide test
pnpm --filter paper-styleguide build
pnpm check:paper-boundaries
```
