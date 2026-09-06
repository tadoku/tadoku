# Paper playground

An unpublished Tadoku application study built from [redesign master version 3](https://draft.apps.lab/documents/tadoku-redesign-master.html?version=3). It uses the public `paper-ui` package and its Tailwind preset. Data and actions are local to the browser; there is no API connection, deployment image, or publishing workflow.

From `frontend`:

```sh
pnpm install --frozen-lockfile
pnpm paper-playground
```

The dev script builds Paper first. `HOST` and `PORT` are respected; the default port is 5174.

The footer’s **Playground controls** opens the page gallery, source scenarios, sample account, theme, density, and three Guide navigation alternatives. **Reset sample changes** clears local edits and restores the initial fixtures. The fixed sample date is 5 September 2026; lifecycle scenarios describe other dates explicitly. Scores illustrate separate personal and contest contexts and are not production scoring policy.

```sh
pnpm --filter paper-playground lint
pnpm --filter paper-playground typecheck
pnpm --filter paper-playground test
pnpm --filter paper-playground build
```

The page and scenario inventory, implementation plan, final report, and browser evidence live in `docs/wip/paper-playground` at the repository root. Do not connect these sample sign-in or administrative flows to live services.
