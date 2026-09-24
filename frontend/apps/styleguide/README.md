# styleguide

The reference site at https://ui.tadoku.app for the legacy `ui` component library in `frontend/packages/ui`, which webv2, auth and admin use. It is built with Next.js 13. For Paper components, use `paper-styleguide` instead.

From `frontend/`:

```sh
pnpm --filter styleguide dev   # port 3002
pnpm --filter styleguide lint
pnpm --filter styleguide exec tsc --noEmit
pnpm --filter styleguide build
```

See [the styleguide docs](../../../docs/docs/frontend/styleguide.md).
