# Tadoku Paper implementation plan

Status: ready for implementation  
Last updated: 2026-08-07  
Decision source: [decision-log.md](decision-log.md)  
Design history: [README.md](README.md)

## Outcome

Deliver Tadoku Paper as a refined, tested, React-based design system in a new `paper-ui` package, make the style guide its authoritative catalogue, migrate every frontend application without mixing old and new systems inside an application, and finally retire legacy `ui`.

The finished work must provide:

- the approved sharp, editorial Bookplate visual language;
- a semantic token system with light and dark mappings;
- public CSS recipes and optional React components backed by the same implementation;
- a searchable, registry-driven style guide with one route per component;
- deterministic fixtures shared between documentation and component tests;
- a framework-independent package boundary with no Next.js dependency;
- complete migrations of styleguide, admin, auth, and webv2;
- compatibility, rollout, rollback, and eventual legacy-package removal gates.

This plan is the implementation companion to the decision log. If an early audit suggestion conflicts with a later recorded decision, the later decision wins. In particular, Tadoku Paper remains square rather than rounded, and buttons support both classes and React components rather than requiring a component everywhere.

## Scope and non-goals

### In scope

- All P0, P1, and P2 audit recommendations, including work already represented by follow-up issues.
- The complete current public surface of `ui`, including components that are poorly documented or available only through deep imports.
- The implicit styling contract currently supplied by `ui/styles/globals.css`.
- The style-guide information architecture, examples, navigation, search, governance, and contribution model.
- Component semantics, keyboard behavior, test fixtures, fast rendered tests, package-boundary checks, and application build gates.
- Additive repository-level coexistence followed by one complete cutover per application.

### Non-goals for the main migration

- Rewriting any application away from Next.js. `paper-ui` must enable that future change, but application framework migration is separate work.
- Consolidating or redesigning product behavior such as the two webv2 logging implementations.
- Requiring Storybook before the custom catalogue has been proven insufficient.
- Blocking the initial rollout on the complete Playwright screenshot system, exhaustive WCAG audit, or full user-preference matrix. Those remain tracked in issues #761–#763 and build on the foundation delivered here.
- Mixing `ui` and `paper-ui` in one deployed application, even temporarily.

## Confirmed design contract

### Visual grammar

- Sharp component geometry with no border radius.
- Warm canvas and paper surfaces, dark ink structure, and restrained ink violet.
- Light structural accent `#5747B8`; lifted dark structural accent `#9A8BEA`.
- Static surfaces use quiet 1px neutral rules.
- A colored left accent is a straight positioned rail that replaces the neutral left edge.
- Pale form fields use a subtle neutral 2px lower edge.
- Filled actions use a stronger 2px lower edge and a distinct hover state.
- Ordinary surfaces stay flat; floating overlays use a 3px hard-offset shadow; rare showcase surfaces may use 5px.
- Dark primary actions use a deeper violet with light text.
- Comfortable density is 44px; compact density is 36px.
- Focus uses a consistent 2px `:focus-visible` ring with a 2px offset.
- Motion roles are 120ms, 180ms, and 240ms.

### Type, icons, and brand

- Merriweather: display, page, and section hierarchy.
- Open Sans: component titles, labels, body copy, and dense UI, with real 400/600/700 font files.
- Fonts are self-hosted and have no Google runtime import or `next/font` dependency.
- Heroicons outline: actions and navigation; solid: status and confirmation.
- Icon roles: 16px compact, 20px default, 24px prominent or icon-only, 48px only for empty-state illustration.
- The compact brand mark is the monochrome-first square-cut Cut Meter with three rising bars.
- Monochrome and reversed marks must work independently of the optional ink-violet accent.

### Public API

- Package name: `paper-ui`.
- React is assumed; Next.js is not.
- CSS class recipes and React components are equal public contracts.
- Button hierarchy: `default`, `primary`, `ghost`.
- `danger` is a composable tone and `loading` is a composable state.
- A class-only loading button remains valid, for example `btn primary loading`, when paired with `aria-busy="true"` and normally `disabled`.
- React components compose the same public recipes; they do not own private duplicate styles.
- `LinkButton` is a router-neutral React adapter over a native anchor or caller-supplied link element; router-specific code can always use the same public class recipe directly.
- Applications import the Paper stylesheet once, after which class recipes work without a component import at each call site.
- Close is an action, not a variant: a footer Close action is usually default; a top-right X is ghost and icon-only.
- Navigation components receive link rendering and current-route data from consumers rather than importing router APIs.

## Delivery principles

1. **One source of truth.** Tokens generate styles and foundation documentation; the catalogue registry generates routes, navigation, search, and status metadata; named fixtures feed both docs and tests.
2. **Vertical slices before breadth.** Prove the package, fixture, component, documentation, and test contracts end to end with Button, Input, Modal, and ActionMenu before porting the full catalogue.
3. **Parallelize by ownership.** Independent lanes own separate package or category directories. One integrator owns shared registries and export boundaries to reduce merge conflicts.
4. **Promote by evidence.** A component becomes Stable only after its API, metadata, fixtures, semantics, behavior tests, migration notes, and ownership are present.
5. **Cut over whole applications.** Repository packages coexist; styles within an application do not.
6. **Keep changes reviewable.** Use atomic commits and small package/component PRs before the per-application deployment PRs.
7. **Preserve history.** The audit, visual studies, and decision log remain unchanged and are linked from the finished style guide.

## Target architecture

### Package layout

```text
frontend/packages/paper-ui/
├── package.json
├── tsconfig.json
├── vitest.config.ts
├── src/
│   ├── assets/
│   │   ├── brand/
│   │   └── fonts/
│   ├── foundations/
│   │   ├── tokens.css
│   │   ├── fonts.css
│   │   ├── base.css
│   │   └── chart-palette.ts
│   ├── components/
│   │   ├── actions/
│   │   ├── forms/
│   │   ├── navigation/
│   │   ├── feedback/
│   │   ├── overlays/
│   │   └── data-display/
│   ├── catalog/
│   │   ├── schema.ts
│   │   └── registry.ts
│   ├── icons/
│   └── index.ts
├── styles/
│   ├── recipes.css
│   ├── utilities.css
│   └── index.css
├── tailwind-preset.cjs
└── tests/
    └── setup.ts
```

Each component directory owns implementation, styles, fixtures, metadata, and tests:

```text
Button/
├── Button.tsx
├── button.css
├── Button.fixtures.tsx
├── Button.meta.ts
└── Button.test.tsx
```

### Export contract

| Export | Responsibility |
| --- | --- |
| `paper-ui` | React components, recipe helpers, and public types |
| `paper-ui/styles.css` | Application-level Paper stylesheet |
| `paper-ui/tokens.css` | Semantic light/dark variables when tokens are needed independently |
| `paper-ui/fonts.css` | Self-hosted font faces |
| `paper-ui/icons` | Curated icon exports and size helpers |
| `paper-ui/tailwind-preset` | Immutable semantic Tailwind mappings |
| `paper-ui/catalog` | Documentation metadata and named fixtures; excluded from production bundles unless explicitly imported |

The package builds ESM, declarations, and CSS. React, React DOM, and React Hook Form are peer dependencies. Implementation libraries such as Headless UI and Heroicons can be direct dependencies. Package output must be consumable without Tailwind and without Next transpilation.

### Token layers

Keep primitives private and expose semantic roles:

- surfaces: canvas, paper, raised, overlay;
- text: ink, muted, inverse, link;
- rules: subtle, default, strong, interactive lower edge;
- actions: base, hover, active, text, danger;
- status: information, success, warning, danger;
- focus: ring and offset;
- density: comfortable and compact;
- motion: quick, standard, deliberate;
- elevation: flat, floating, showcase;
- typography: named display, page, section, component, body, label, and metadata roles;
- charts: a separate categorical sequence with non-color cues.

Light aliases live on `:root` and `[data-theme="light"]`; dark aliases live on `[data-theme="dark"]`. Density is set with `[data-density="comfortable|compact"]`. Theme previewing is part of the style guide now; operating-system preference and persistence remain in #763.

### CSS boundary

- `styles.css` includes tokens, fonts, a small documented base layer, and static class recipes.
- Avoid broad form, table, heading, and link selectors that turn global CSS into an undocumented API.
- Forms and tables receive Paper styling through components or named recipes.
- Raw colors are allowed only in primitive token definitions, documented specimens, and an explicit chart-palette allow-list.
- Keep convenient unprefixed recipes such as `.btn` because only one system is loaded per application.

### Router boundary

- No `next`, `next/link`, `next/router`, `next/image`, or `next/font` import or dependency in `paper-ui` or its catalogue.
- Use inline SVG or ordinary image rendering for brand assets.
- Expose a recipe helper for styling any router's link.
- Navbar receives current state and a consumer-supplied link renderer.
- Pagination receives page state plus `getHref` or `onPageChange`.
- Breadcrumbs, tabs, sidebars, and grouped links use the same injected link contract.

## Documentation model

### Information architecture

```text
/
├── foundations/
│   ├── principles
│   ├── color
│   ├── typography
│   ├── spacing-and-density
│   ├── shape-and-borders
│   ├── elevation
│   ├── iconography
│   ├── motion
│   ├── layout
│   └── brand
├── components/
│   ├── actions/
│   ├── forms/
│   ├── navigation/
│   ├── feedback/
│   ├── overlays/
│   └── data-display/
├── patterns/
│   └── logging/
├── experiments/
│   └── logging-v2
├── contributing
└── changelog
```

The style guide may remain a Next.js application during this work. Framework routing stays in a thin application adapter; `paper-ui`, its fixtures, and its catalogue remain framework-neutral.

### Required page anatomy

Every component page contains:

- name, summary, lifecycle status, owner, source, package version, and last meaningful review;
- when to use and when to avoid;
- content guidance;
- accessibility and keyboard contract;
- API and CSS-class contract;
- realistic Tadoku examples;
- meaningful variants, states, density, theme, viewport, long-copy, and overflow coverage;
- migration guidance and changelog notes;
- replacement and planned removal information when Deprecated.

Use exactly `Experimental`, `Stable`, and `Deprecated`. Category pages summarize their children; every actual component gets one canonical searchable route. Old bookmarks redirect to the new route map.

### Documentation shell

Replace `Showcase` with:

- a responsive docs shell and mobile navigation;
- keyboard-first search with `/` and Command/Ctrl-K;
- registry-driven navigation and component index;
- a local table of contents with stable deep links;
- a framed example canvas;
- Preview, Code, API/Props, and Accessibility tabs as applicable;
- code copy and fixture deep links;
- theme, density, and real viewport controls;
- status, metadata, usage, and accessibility panels.

Viewport previewing should create a real media-query environment, preferably an isolated iframe document, not only a narrow wrapper. Reading prose should stay around 65–75 characters or 720–780px while examples and data specimens may use a wider column.

### Catalogue and fixture schema

Each document entry requires a stable ID, route, name, category, aliases, summary, keywords, lifecycle, owner, review date, source path, package version, guidance, accessibility notes, API documentation, fixture IDs, dependencies, and migration details.

Fixtures must be deterministic: no Faker randomness, `Math.random`, current dates, network calls, or application-router assumptions. A fixture contains stable metadata and a render function; assertions stay in tests.

Registry validation fails on:

- duplicate IDs or slugs;
- invalid category or lifecycle values;
- missing Stable metadata, fixtures, behavior tests, or accessibility notes;
- documentation references to nonexistent fixture IDs;
- Deprecated entries without replacement and removal guidance.

## Test and quality strategy

### Immediate baseline

Add Vitest with jsdom, React Testing Library, `user-event`, and jest-dom. Replace source-text assertions with rendered semantic and behavioral tests. Do not use an arbitrary coverage-percentage gate initially; gate the documented component contract instead.

Initial tests cover:

- roles and accessible names;
- label, hint, and error relationships;
- `aria-invalid`, `aria-describedby`, `aria-current`, and `aria-busy` behavior;
- native disabled behavior and prevention of duplicate activation;
- keyboard operation, focus entry/containment/return, and Escape behavior;
- raw-class and React-component parity;
- comfortable and compact fixture coverage;
- deterministic fixture and registry validation;
- public exports, CSS selectors, built assets, and server-render import smoke tests.

The first required behavioral tranche is Button/LinkButton recipes, Input, Select, Checkbox, Modal, ActionMenu, Navbar, Breadcrumb, Tabs, Pagination, Loading, and form error plumbing. A component receives equivalent tests when promoted to Stable.

### Package gates

- No Next import, dependency, mock, or router assumption.
- Public exports import and `renderToString` in a minimal React consumer.
- Built output contains JS, declarations, CSS, tokens, fonts, icons, and brand assets.
- `pnpm pack --dry-run` or equivalent validates the published file set.
- Built CSS contains every documented raw class recipe and state selector.
- Token lint rejects raw palette values outside approved definition files.
- Font checks verify packaged Merriweather and Open Sans 400/600/700 assets.

### CI

Add a Paper quality workflow on pull requests and pushes to main for package, catalogue, style-guide, lockfile, and workflow changes. Its fast required job runs:

```sh
cd frontend
pnpm install --frozen-lockfile
pnpm --filter paper-ui typecheck
pnpm --filter paper-ui test
pnpm --filter paper-ui build
```

As each application migrates, its workflow must:

- watch `frontend/packages/paper-ui/**`;
- run application lint and typecheck before its production build;
- retain production and container-image builds as compatibility checks;
- join the Paper compatibility matrix.

At completion, the repository gate is the complete frontend build plus all four application images.

### Deferred comprehensive gates

- #761: full WCAG 2.2 component contract, contrast, zoom/reflow, forced colors, and exception register.
- #762: Playwright fixture routes, phone/desktop screenshots, real hover/focus, axe scans, traces/diffs, baseline approval workflow, and stability burn-in.
- #763: preference detection and persistence, reduced motion, forced colors, exhaustive dark-theme validation, and non-color chart/status cues.

The main plan establishes the catalogue, deterministic fixtures, semantics, dark token aliases, and CI hooks needed by these issues; the issues stay open until their complete acceptance criteria are met.

## Phased execution

The dependency spine is:

```text
contracts
   ↓
package foundations + catalogue/test schema
   ↓
four-component vertical slice
   ↓
full component catalogue + docs platform
   ↓
styleguide cutover
   ↓
admin → auth → webv2 cutovers
   ↓
legacy ui removal and deferred hardening
```

Work beside the spine runs concurrently. Only shared contracts, the end-to-end pilot, and application deployment order are deliberately serial.

### Phase 0 — Lock contracts and migration guardrails

Goal: give every parallel lane stable boundaries before implementation starts.

| Lane | Work |
| --- | --- |
| Architecture | Finalize export map, dependency policy, router/link contract, recipe/component parity, and built-output format. |
| Catalogue | Freeze document, fixture, lifecycle, route, redirect, and registry schemas. |
| Inventory | Map every legacy export, deep import, global selector, style-guide example, and consumer to its Paper destination. |
| Delivery | Define per-app smoke lists, rollback evidence, CI matrix, owners, and PR boundaries. |

Add repository guards that reject:

- `ui` imports in an application marked migrated;
- `paper-ui` imports in an application not yet marked migrated;
- `paper-ui/src/*` private imports;
- Next imports inside `paper-ui`;
- duplicate application-level Paper stylesheet imports.

Update contributor instructions during coexistence to explain which package each application may use. Keep the existing design archive untouched.

**Exit gate:** approved contract document, complete migration inventory, route/redirect map, application smoke matrix, and enforcement scripts.

### Phase 1 — Build the Paper foundation

Run four lanes in parallel.

| Lane | Deliverables |
| --- | --- |
| Package/build | Package scaffold, ESM/declaration build, export map, immutable Tailwind preset, root scripts, lockfile, package boundary checks, CI and Tilt triggers. |
| Visual foundation | Semantic tokens, themes, density, typography/font assets, focus, motion, borders, accent rail, hard-offset elevation, chart palette, Cut Meter, wordmark, favicon sources. |
| Test/catalogue | Vitest setup, RTL helpers, fixture and metadata schemas, registry validation, test utilities, catalogue export. |
| Docs platform | Style-guide route resolver, registry stubs, shell primitives, isolated preview-canvas prototype, old-route redirect table. |

One integrator owns root exports, registry aggregation, and shared token names. Lanes add modules without repeatedly editing those files.

**Exit gate:** `paper-ui` builds, typechecks, tests, and server-renders outside Next; CSS works without Tailwind; both themes and densities render in a minimal harness; fonts and brand assets resolve; the stub catalogue drives navigation and search.

### Phase 2 — Prove an end-to-end vertical slice

Implement Button, Input, Modal, and ActionMenu completely. These four jointly prove class/component parity, form anatomy, interactive edge treatment, floating elevation, overlay focus behavior, fixtures, metadata, documentation rendering, and tests.

Parallel ownership:

| Lane | Slice |
| --- | --- |
| Actions | Button and LinkButton recipes/components: default/primary/ghost, danger, loading, icon, full-width, both densities. |
| Forms | Input anatomy: label, hint, error, required, read-only, disabled, both densities and themes. |
| Overlays | Modal and ActionMenu: focus, keyboard, close/return behavior, danger/disabled items, floating treatment. |
| Documentation/tests | Final page anatomy, state matrix, Preview/Code/API controls, fixture rendering, registry and behavior tests. |

Show real hover and focus in browser documentation; do not introduce fake production state classes solely for jsdom.

**Exit gate:** all four pages satisfy the final schema; raw classes and React APIs share styles; tests pass; desktop/mobile shell and all preview controls work; the team reviews visual coherence before scaling.

### Phase 3 — Complete `paper-ui` and the style-guide catalogue

With the pilot contracts frozen, run parallel component lanes in separate directories.

| Lane | Components and docs |
| --- | --- |
| Actions/feedback | ButtonGroup, Flash, Loading, toasts, cards and elevation recipes. |
| Forms | TextArea, Select, Checkbox, RadioSelect, RadioGroup, AmountWithUnit, autocomplete, multi-autocomplete, TagsInput, public option types; all use React Hook Form. |
| Navigation/overlays | Navbar, Sidebar, Breadcrumb, Tabbar, VerticalTabbar, Pagination, Modal, ActionMenu, router-neutral link integration. |
| Data/brand/layout | Tables, HeatmapChart, chart defaults/palette, layout recipes, Cut Meter and wordmark lockups. |
| Docs foundations | Principles, color, typography, spacing/density, borders/shape, elevation, icons, motion, layout, brand. |
| Docs governance | Component indexes, Patterns, Experimental logging-v2, contribution guide, lifecycle, deprecation policy, changelog. |

Every lane ships its fixtures, metadata, tests, migration notes, and realistic Tadoku examples with the implementation. Explicitly document currently missed exports such as Loading, VerticalTabbar, and AmountWithUnit.

**Exit gate:** the public export surface is complete; every Stable component has a searchable route and contract evidence; navigation/search/status derive from one registry; all routes and old-route redirects validate; foundation pages read canonical tokens instead of copied values.

### Phase 4 — Styleguide big-bang cutover

Prepare independent changes concurrently, then integrate them into one deployable application cutover.

| Lane | Work |
| --- | --- |
| Shell | Replace `_app.tsx`, hard-coded sidebar, mobile navigation, `Showcase`, and old example helpers with the registry-driven docs shell. |
| Content | Replace repeated category pages, split Forms and Navigation into canonical component routes, reclassify logging as Patterns and logging-v2 as Experimental. |
| Build/assets | Switch package, CSS, Tailwind, Next config, document font preloads, brand/favicon assets, and remove obsolete raw-source loader support. |
| Verification | Search/route tests, registry checks, keyboard/mobile docs-shell tests, link/redirect checks, production and container builds. |

Import `paper-ui/styles.css` exactly once. Preserve the audit, studies, and decision log and link them under Design history or Contributing.

**Cutover gate:** no legacy `ui` import, dependency, stylesheet, Tailwind inheritance, or config reference remains in styleguide; every shipped component is documented; search and deep links work keyboard-first; both themes, densities, and real narrow/desktop preview environments work; lint, typecheck, tests, application build, and image build pass.

### Phase 5 — Admin big-bang cutover

Admin is the compact-density stress test. Work in parallel on one integration branch or coordinated PR stack, then switch the application once.

| Lane | Scope |
| --- | --- |
| Shell | Dashboard layout, Sidebar, branding, Loading, access-denied and redirect states. |
| Editors/forms | Content and announcement editors, namespace forms, language/user editing, CodeMirror adapter using `useController`. |
| Data/actions | Lists, tables, pagination, breadcrumbs, tabs, menus, dialogs, destructive actions, narrow layouts. |
| Verification | Import/class audit, compact screenshots for human review, CRUD and role smoke suite, build matrix. |

Review old `secondary` and `danger` combinations semantically; do not perform a blind class rename. Compact controls must not make icon-only targets inaccessible.

**Cutover gate:** no legacy package or implicit legacy-only class remains; Paper CSS is loaded once and compact density is applied at the app root; dashboard/navigation, role/loading/access-denied, content CRUD and preview, announcements, languages, users, dialogs, and narrow layouts pass; lint, typecheck, production build, and image build pass.

### Phase 6 — Auth big-bang cutover

Auth has few imports but high implicit-CSS and flow risk because Ory generates dynamic form nodes.

| Lane | Scope |
| --- | --- |
| Shell | Navigation, Cut Meter/wordmark, comfortable layout, feedback and loading. |
| Ory fields | Node adapters, labels, hints, errors, checkbox/text/select semantics, WebAuthn/script preservation. |
| Actions/flows | Explicit default/primary/ghost hierarchy, loading and double-submit prevention, removal of selector-order button hacks. |
| Verification | Fixtures and production-like login, registration, recovery, verification, settings, OIDC/passkey/WebAuthn checks. |

**Cutover gate:** no legacy import or broad legacy-form dependency; Paper CSS loaded once and comfortable density applied; server errors associate to controls; keyboard submission, `aria-busy`, disabled state, and dynamic Ory behavior work; lint, typecheck, production build, image build, and focused Ory smoke tests pass.

### Phase 7 — webv2 big-bang cutover

webv2 is last because it has the largest and broadest surface. Use four parallel workstreams while preserving the application as one cutover.

| Lane | Scope |
| --- | --- |
| Global shell | Navigation, Footer, announcement banner, signed-in/out/admin/banned states, mobile disclosure. |
| Logging/contest forms | New/old/edit log flows, contest create/register/submit flows, form validation and overlays. |
| Browse/data | Leaderboards, profiles, lists, tables, pagination, tabs, menus, responsive and overflow behavior. |
| Content/charts | Blog, pages, manual, Chart.js palettes, heatmaps, activity charts, page counter exception. |
| Verification | Import/class audit, route and role smoke matrix, Paper tests, full frontend and container builds. |

Do not combine product refactors with the visual migration. App adapters provide routing data to Navbar, Breadcrumb, Tabs, and Pagination. Charts use semantic categorical tokens and prepare non-color cues for #763.

**Cutover gate:** no legacy import or undocumented CSS dependency; Paper CSS loaded once; signed-in/out navigation, mobile navigation, logging create/edit/details/delete, contest creation/registration/leaderboards, profiles/statistics, blog/content, pagination/tabs, charts, and page counter pass; lint, typecheck, production build, image build, and the complete frontend build matrix pass.

### Phase 8 — Retire legacy `ui`

Only begin when a repository-wide search shows zero consumers:

```sh
rg "from ['\"]ui|ui/styles|ui/components|\"ui\": \"workspace" frontend
```

Then:

- remove `frontend/packages/ui`;
- remove old dependencies, lockfile entries, transpilation entries, and Tailwind inheritance;
- update all CI paths and `frontend/Tiltfile` to Paper only;
- simplify contributor instructions from coexistence rules to Paper-only guidance;
- run package tests and the complete frontend/application-image matrix;
- record package retirement and final migration status in this directory and the style-guide changelog.

Previously deployed application images remain rollback-capable because their legacy dependencies are bundled.

### Phase 9 — Finish the deferred hardening issues

These workstreams may begin earlier once their prerequisites exist, but close after the migrations only when their full acceptance criteria are satisfied.

| Issue | Parallel work after prerequisites | Completion point |
| --- | --- | --- |
| #761 Accessibility contract | Component audits can start after Phase 2; product reflow checks start with each migrated app. | All documented controls, flows, contrast, keyboard, zoom/reflow, and exception tracking complete. |
| #762 Regression gates | Playwright fixture routes can start after the style-guide pilot; burn-in continues through app migrations. | Reviewed responsive baselines, failure artifacts, axe integration, and stable blocking CI complete. |
| #763 Themes/preferences | Dark/reduced-motion/forced-color fixtures can start after tokens; app persistence waits for product rollout decisions. | Preference behavior and representative app/theme matrix complete. |

## Parallel execution map

At maximum useful concurrency, staff the program as follows:

| Stream | Phases 0–2 | Phase 3 | Phase 4 | Phases 5–7 |
| --- | --- | --- | --- | --- |
| A | Package/build | Actions/feedback | Build/assets | Active app shell |
| B | Visual foundations | Forms | Foundations/content | Active app forms/workflows |
| C | Test/catalogue | Navigation/overlays | Shell/search/canvas | Active app data/navigation |
| D | Docs platform | Data/brand/layout | Tests/governance | Verification plus next-app inventory |

While one application is in final integration and deployment, other streams may prepare the next application's inventory, fixtures, and smoke checklist or continue #761–#763. They must not introduce `paper-ui` into the next application until its coordinated cutover branch begins.

Recommended integration discipline:

- one owner for token names, public exports, and the catalogue registry;
- directory ownership for parallel component lanes;
- generated registry aggregation where possible;
- small component/foundation PRs with their tests and docs;
- one deployable migration PR per application, composed of atomic commits;
- short-lived integration branches and daily main rebases during large cutovers;
- no application deployment before the preceding application has passed its production smoke window.

## Per-application migration protocol

Each application follows the same protocol.

### 1. Inventory

- Direct imports and deep imports.
- Legacy global selectors and raw classes.
- Tailwind mutation, package, transpile, and stylesheet configuration.
- Router-dependent behavior.
- Brand/favicons and font loading.
- Critical routes, roles, responsive states, form errors, and overlays.

### 2. Prepare

- Ensure every required Paper component is Stable or explicitly approved Experimental.
- Add missing product states as deterministic fixtures and tests before migration.
- Create application-owned router adapters and product patterns.
- Record the currently deployed immutable image SHA/digest and verify the rollback command.

### 3. Convert

- Replace component imports and deep imports.
- Replace implicit global styling with components or named recipes.
- Convert buttons by semantic hierarchy, tone, and state rather than text substitution.
- Load Paper styles once and set the application density.
- Replace Next-aware shared behavior with application adapters.
- Update package, Tailwind, framework, image, CI, Tilt, and asset references as applicable.

### 4. Prove zero mixing

For the target application, repository checks must find no:

- `ui` dependency or import;
- `ui/styles/globals.css` import;
- `ui/components/*` deep import;
- `require('ui/tailwind.config.js')` or equivalent;
- class that exists only through undocumented legacy globals;
- duplicate `paper-ui/styles.css` import.

### 5. Verify and deploy

- Run `paper-ui` typecheck, test, and build.
- Run target-app lint, `tsc --noEmit`, production build, and image build.
- Execute the app-specific smoke matrix on narrow and desktop layouts.
- Deploy only the target application.
- Re-run critical production smoke checks and observe frontend/error telemetry.

### 6. Roll back if necessary

Redeploy the recorded previous image. Do not attempt to fix a failed cutover by loading both stylesheets. Source rollback is a normal revert of the migration commits; no database or API rollback is involved.

## Application migration matrix

| Order | Application | Default density | Primary risk | Required proof |
| ---: | --- | --- | --- | --- |
| 1 | styleguide | switchable | The new system documents and tests itself. | Every route, search, canvas, fixture, redirect, and registry contract works with zero legacy imports. |
| 2 | admin | compact | Dense CRUD, tables, editors, menus, and modal flows. | Core administrative workflows and narrow layouts work at 36px density. |
| 3 | auth | comfortable | Dynamic Ory nodes rely heavily on implicit global CSS. | Every configured identity flow, error state, keyboard path, and submit/loading state works. |
| 4 | webv2 | comfortable | Largest surface and most router/data/chart combinations. | Representative user roles and core logging, contest, content, profile, navigation, and chart flows work. |

## Validation commands

Exact package scripts will be added in Phase 1. The required command shape is:

```sh
cd frontend
pnpm --filter paper-ui typecheck
pnpm --filter paper-ui test
pnpm --filter paper-ui build
pnpm --filter styleguide lint
pnpm --filter styleguide exec tsc --noEmit
pnpm --filter styleguide build
pnpm --filter admin lint
pnpm --filter admin exec tsc --noEmit
pnpm --filter admin build
pnpm --filter auth lint
pnpm --filter auth exec tsc --noEmit
pnpm --filter auth build
pnpm --filter webv2 lint
pnpm --filter webv2 exec tsc --noEmit
pnpm --filter webv2 build
pnpm build
```

Application and container-image workflow commands remain the final production compatibility proof. Any new form implementation continues to use React Hook Form and Paper form components.

## Definition of done

Tadoku Paper is complete when:

- the approved visual grammar and brand assets are implemented only through semantic roles and documented exceptions;
- `paper-ui` builds and is independently consumable by React without Next.js;
- class recipes and React components are documented and tested as one styling contract;
- every public component has a canonical searchable page, deterministic fixtures, lifecycle metadata, ownership, tests, and migration guidance;
- foundations, components, product patterns, and experiments are visibly separate;
- styleguide, admin, auth, and webv2 each use only Paper and have passed their deployment smoke gates;
- the original audit, all visual studies, and the decision log remain available;
- repository and CI configuration watches and validates `paper-ui`;
- legacy `ui` has zero consumers and is removed;
- issues #761–#763 meet their full accessibility, regression, and theme/preference acceptance criteria.

## Audit traceability

| Original recommendation | Delivery |
| --- | --- |
| P0.1 Semantic tokens | Phases 1 and 3 foundation work; token/package gates. |
| P0.2 Visual grammar | Confirmed contract; Phases 1–3 implementation and documentation. |
| P0.3 Complete docs shell | Phases 1–4. |
| P0.4 Accessibility contract | Immediate semantic tests throughout; comprehensive completion in #761/Phase 9. |
| P0.5 Typed Button/LinkButton | Hybrid class/component contract proven in Phase 2; router-neutral link recipes. |
| P1.1 One route and search | Registry-driven docs in Phases 1, 3, and 4. |
| P1.2 Full state matrix | Fixture schema and every component page, beginning in Phase 2. |
| P1.3 Foundation pages | Phase 3. |
| P1.4 Co-located examples/metadata/tests | Package architecture and Phases 1–3. |
| P1.5 Automated quality gates | Fast tests and CI now; comprehensive visual/axe system in #762/Phase 9. |
| P2.1 Components/patterns/experiments | Information architecture and Phases 3–4. |
| P2.2 Lifecycle/governance | Registry schema, Stable gate, contribution and deprecation policy in Phases 1–4. |
| P2.3 Themes/preferences | Light/dark tokens and preview now; full preference behavior in #763/Phase 9. |
| P2.4 Package boundary | `paper-ui` architecture, boundary tests, app adapters, and migration gates. |

The early quick wins—framed previews, constrained prose, refined violet, and explicit use/avoid guidance—are incorporated into the documentation shell and visual contract rather than scheduled as throwaway work.
