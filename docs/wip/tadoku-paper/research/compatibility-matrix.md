# Tadoku Paper compatibility matrix

Status: Phase 0 evidence complete  
Measured: 2026-08-08  
Starting commit: `92083ac7fa3972648d154f78b4d7e398bb985502`

## Repository baseline

The frontend lockfile is lockfile format 9.0 and globally overrides React and React DOM to 18.2.0. Existing applications compile with TypeScript 4.9.3 and React 18 types. CI and both frontend Dockerfiles use Node 20 and explicitly install pnpm 10.10.0.

An unpinned `corepack pnpm` resolved pnpm 11.20.0 in the spike environment and failed on Node 20.20.2 because pnpm 11 requires Node 22.13 or newer. Paper must preserve the repository's explicit pnpm 10.10.0 pin rather than rely on Corepack's current default.

## Selected versions

| Concern | Selected version | Compatibility evidence and reason |
| --- | --- | --- |
| Node | 20.x | Matches CI and `node:20-alpine`; tested on 20.20.2. |
| pnpm | 10.10.0 | Matches CI and Dockerfiles; reads and writes lockfile 9.0. |
| React / React DOM | 18.2.0 | Matches root overrides; all rendered, server-render, and Vite tests passed. |
| TypeScript for Paper workspaces | 5.9.3 | Required to parse Base UI declarations; does not change legacy application compilers. |
| Base UI | `@base-ui/react` 1.7.0 | Supports React 17–19 and Node 14+; Dialog, Menu, and Combobox spike passed on React 18.2. |
| Vite | 6.4.3 | Supports every Node 20 release (`^20.0.0`), unlike Vite 7/8's Node 20.19 floor; static React build passed. |
| Vite React plugin | `@vitejs/plugin-react` 4.7.0 | Peer range includes Vite 6; Vite build and TypeScript 5.9 configuration passed. |
| Catalogue router | `react-router-dom` 7.18.2 | React 18 and Node 20 peer/runtime ranges pass; `BrowserRouter`, `Routes`, `Route`, and `Link` built in the static Vite consumer. |
| Vitest | 4.1.10 | Supports Node 20 and Vite 6; five rendered behavior tests passed. |
| DOM environment | jsdom 27.0.1 | Supports Node 20 and supplies `PointerEvent`, which Base UI 1.7 uses for keyboard activation dispatch. |
| React Testing Library | 16.3.2 | Peer range includes React 18 and DOM Testing Library 10; render/query/cleanup passed. |
| DOM Testing Library | 10.4.1 | Satisfies React Testing Library and user-event peers. |
| user-event | 14.6.3 | Dialog/Button pointer, typing, tab, and keyboard flows passed with jsdom 27; Menu pointer opening still needs the required real-browser smoke. |
| jest-dom | 6.9.1 | Node 14+ and DOM Testing Library 10 compatible; v7 was rejected because it requires Node 22. |
| Package builder | tsup 8.5.1 | Node 18+; emitted ESM and bundled declarations while leaving React, React DOM, and Base UI external. |
| Node types | `@types/node` 20.19.25 | Matches the Paper runtime/toolchain rather than the legacy apps' Node 18 type snapshots. |
| React types | `@types/react` 18.0.26 / `@types/react-dom` 18.0.9 | Matches the current repository family and React 18 runtime. |

Use exact versions for the initial Phase 1 lockfile change. Renovation can happen after the first Paper gate with this matrix rerun.

## TypeScript compatibility boundary

Both `@base-ui/react` 1.7.0 and its earliest stable release, 1.0.0, failed under TypeScript 4.9.3 before semantic checking. Their `@base-ui/utils` declarations use TypeScript 5 `const` type parameters, which `skipLibCheck` cannot hide because TypeScript 4.9 cannot parse the syntax.

The contained solution is:

- compile `paper-ui` and `paper-styleguide` with TypeScript 5.9.3;
- keep React at 18.2.0 and keep legacy application compilers unchanged;
- bundle Paper declarations so no Base UI type or import appears in the public `.d.ts` files;
- smoke-test the built Paper declarations from TypeScript 4.9.3.

The spike's 767-byte bundled declaration file contained only Paper and React types. A strict TypeScript 4.9.3 consumer imported Button, Dialog, Menu, Combobox, and the button recipe successfully with `skipLibCheck: false` after the normal React 18 scheduler type was present.

## Package and catalogue build evidence

The disposable package used tsup with ESM output, declaration bundling, target ES2020, and React/React DOM externals. Base UI stayed as direct external subpath imports. Results:

| Artifact | Result |
| --- | --- |
| Paper wrapper ESM | 3,502 bytes |
| Paper bundled declaration | 767 bytes; no Base UI, Next, or Headless UI reference |
| `renderToString(<PaperButton>Save</PaperButton>)` | `<button type="button" class="btn btn--default">Save</button>` |
| pnpm package contents | package JSON plus the ESM and declaration artifacts only |
| Vite React/router baseline | 181.24 kB raw, 59.70 kB gzip |
| Vite app with Dialog, Menu, and Combobox | 360.20 kB raw, 119.92 kB gzip |

The primitive tranche added about 178.96 kB raw / 60.22 kB gzip to the one-chunk disposable app. This is a diagnostic upper bound, not a catalogue performance budget: route-level lazy chunks can defer complex primitives in the real catalogue.

tsup and Vite must use distinct output directories. Vite cleans its output directory by default; a sequential spike that pointed both tools at `dist/` correctly exposed that Vite would erase the package artifacts.

The initial package build should therefore:

- emit JS and declarations with tsup into the package distribution directory;
- copy already-authored CSS, fonts, icons, and brand assets into that directory as separate public files;
- keep React, React DOM, React Hook Form, and Base UI external to wrapper compilation according to their peer/direct dependency roles;
- declare CSS files as side effects and expose them explicitly in `exports`;
- validate the pack list with `pnpm pack --json` (pnpm 10.10.0 has no `pack --dry-run` option).

## Alternatives rejected for the first lock

| Alternative | Finding |
| --- | --- |
| Vite 7.3.6 or Vite 8.2.1 | Both require Node 20.19+ rather than the repository's broader Node 20 contract. Reconsider after pinning a minimum Node patch. |
| `@vitejs/plugin-react` 5.2.0 with TypeScript 4.9 | Its declarations use syntax TypeScript 4.9 cannot parse. Paper already needs TypeScript 5, but plugin 4.7 is the matching stable Vite 6 line. |
| Vitest 3.2.4 | Compatible, but Vitest 4.1.10 passed on the selected Node/Vite pair and avoids beginning a greenfield suite on the previous major. |
| jest-dom 7.0.0 | Requires Node 22. |
| jsdom 26.1.0 | Base UI Menu keyboard activation crashed because `window.PointerEvent` is absent; a custom polyfill is avoidable with jsdom 27.0.1. |
| `@base-ui-components/react` | Obsolete package name. The installed Base UI documentation explicitly says to use `@base-ui/react`. |
| Base UI 1.0.0 to retain TypeScript 4.9 | Failed: the earliest stable line also contains TypeScript 5 declarations. |
| Vite library mode for `paper-ui` | It can emit ESM/CSS but does not replace a declaration-bundling step. tsup produced the required small, externalized JS and stable public declaration in one command. |

## Reproduction commands

The disposable spike was removed after recording results. Its gates were equivalent to:

```sh
pnpm install --frozen-lockfile=false
pnpm typecheck
pnpm test
pnpm build
node node_modules/typescript-4/bin/tsc --noEmit -p consumer-4.9/tsconfig.json
pnpm pack --json --pack-destination /tmp
```

All final gates passed on Node 20.20.2 and pnpm 10.10.0.
