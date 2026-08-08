# Tadoku Paper implementation plan

Status: ready for implementation

Last updated: 2026-08-08

Decision source: [decision-log.md](decision-log.md)

Design history: [README.md](README.md)

## Outcome

Deliver Tadoku Paper as a refined, tested, React-based design system in a new `paper-ui` package; build a new static React + Vite catalogue at `paper.tadoku.app`; migrate every product application without mixing old and new systems inside an application; move the completed catalogue to `ui.tadoku.app`; and finally retire legacy `ui`, the legacy styleguide, and Headless UI.

The finished work must provide:

- the approved sharp, editorial Bookplate visual language;
- a semantic token system with light and dark mappings;
- public CSS recipes and optional React components backed by the same implementation;
- a searchable, registry-driven `paper-styleguide` with one route per component;
- deterministic fixtures shared between documentation and component tests;
- a framework-independent package boundary with no Next.js dependency;
- an independent Paper catalogue followed by complete migrations of admin, auth, and webv2;
- compatibility, rollout, rollback, and eventual legacy-package removal gates.

This plan is the implementation companion to the decision log. If an early audit suggestion conflicts with a later recorded decision, the later decision wins. In particular, Tadoku Paper remains square rather than rounded, and buttons support both classes and React components rather than requiring a component everywhere.

## Overview

The work proceeds along two coordinated tracks. The first builds `paper-ui`, its catalogue schema, tests, and the new `paper-styleguide`. The second migrates admin, auth, and webv2 one complete application at a time after the Paper component surface is stable. The existing `ui` package and styleguide remain untouched enough to serve unmigrated applications and `ui.tadoku.app`; they never load into the new catalogue.

The new catalogue is a static Vite application. During development, agents expose it temporarily through the `t3-expose-development-server` workflow. Merged work deploys persistently to `paper.tadoku.app`. After admin, auth, and webv2 have migrated, `ui.tadoku.app` switches atomically to the Paper deployment and the legacy application is retired.

### Benefits

- No transitional CSS collision or misleading mixed catalogue.
- A Vite consumer proves `paper-ui` is not coupled to Next.js.
- Static production serving uses materially less idle CPU and memory than a persistent Next.js runtime.
- Base UI supplies difficult ARIA, keyboard, focus, portal, and positioning behavior while Paper retains its own API and visual identity.
- A typed registry makes navigation, search, lifecycle, documentation completeness, fixtures, and tests one coherent contract.
- Per-application image cutovers provide clear rollback boundaries.

### Risks and mitigations

| Risk | Mitigation |
| --- | --- |
| Two catalogues drift during the migration | Freeze the legacy styleguide except critical corrections; it documents only legacy `ui`, while Paper work goes only to `paper-styleguide`. |
| Base UI behavior leaks into Paper's public contract | Wrap primitives behind Paper components, avoid exporting Base UI types where possible, and add public API/boundary tests. |
| Copied shadcn conventions dilute the Paper identity | Borrow recipe, registry, composition, and documentation patterns only; do not generate shadcn components or adopt its theme. |
| Static SPA deep links return 404 | Configure the production server and ingress to fall back to `index.html`; test every canonical route directly. |
| Accessibility is assumed rather than verified | Reuse deterministic fixtures in rendered keyboard/semantic tests and complete issue #761 after the baseline. |
| Legacy global CSS hides migration work | Inventory selectors as an implicit API and require a zero-legacy-class audit per application. |
| Headless UI and Base UI conflict | Never combine them in a Paper component or application. Headless UI remains only behind legacy `ui` until cutover. |
| Parallel lanes conflict in shared files | Give one integrator ownership of tokens, public exports, catalogue aggregation, and migration markers. |
| Rarely visited styleguide still reserves excess cluster resources | Serve a static build from a minimal container, start with small measured requests/limits, and avoid SSR or scale-to-zero infrastructure unless data shows it is worthwhile. |

### Measurable success criteria

- `paper-ui` has zero `next` and zero `@headlessui/react` source imports or dependencies.
- `paper-styleguide` has zero `ui`, legacy stylesheet, Next.js, and Headless UI imports.
- Every public Stable component has exactly one canonical route, required instructional sections, deterministic fixtures, semantic/behavior tests, ownership, and lifecycle metadata.
- Catalogue navigation, search, routes, statuses, and test fixtures derive from one validated registry.
- Direct requests to every Paper catalogue route return the application successfully at `paper.tadoku.app` and, after cutover, `ui.tadoku.app`.
- Button classes, typed recipe output, and React component output use the same selectors for every documented variant and loading state.
- All required keyboard and focus tests pass for Base UI-backed components.
- Each migrated application has zero legacy `ui` imports, legacy global CSS, undocumented legacy-only classes, and unused Headless UI dependencies.
- Admin, auth, and webv2 pass their lint, typecheck, production build, image build, and critical-flow smoke gates.
- The static Paper catalogue operates within the resource requests/limits selected from observed idle and navigation usage.
- Repository-wide searches show zero legacy consumers before `ui`, the old styleguide, and Headless UI are removed.

## Scope and non-goals

### In scope

- All P0, P1, and P2 audit recommendations, including work already represented by follow-up issues.
- The complete current public surface of `ui`, including components that are poorly documented or available only through deep imports.
- The implicit styling contract currently supplied by `ui/styles/globals.css`.
- The new Paper styleguide's information architecture, examples, navigation, search, governance, contribution model, deployment, and final domain cutover.
- Component semantics, keyboard behavior, test fixtures, fast rendered tests, package-boundary checks, and application build gates.
- Additive repository-level coexistence followed by one complete cutover per application.

### Non-goals for the main migration

- Rewriting admin, auth, or webv2 away from Next.js. `paper-ui` must enable that future change, while `paper-styleguide` deliberately uses Vite as the first non-Next consumer.
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
- Dark default/emphasized actions use a deeper violet with light text.
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
- Button variants: `default`, `outline`, `ghost`, `link`, and `destructive`.
- `default` is the filled emphasized violet action; `outline` is the neutral bordered action. `loading` is a composable state.
- A class-only default button uses `btn`; additional variants compose as `btn outline`, `btn ghost`, `btn link`, or `btn destructive`. Loading composes as `btn loading` and must be paired with `aria-busy="true"` and normally disabled behavior.
- React components compose the same public recipes; they do not own private duplicate styles.
- `Button variant="link"` remains a real button action with link-like appearance. Navigation remains a real anchor and receives button appearance from `buttonClassName()` or an anchor-specific `LinkButton` adapter.
- Paper Button defaults to `type="button"`; form submissions declare `type="submit"`, and React Hook Form resets use `methods.reset()`.
- Applications import the Paper stylesheet once, after which class recipes work without a component import at each call site.
- Close is an action, not a variant: a footer Close action is usually outline; a top-right X is ghost and icon-only.
- Navigation components receive link rendering and current-route data from consumers rather than importing router APIs.

### Primitive policy

- Use Base UI as the sole headless primitive dependency for Paper.
- Wrap Base UI behind Paper-owned components and recipes; application code does not import Base UI for standard design-system components.
- Prefer native HTML when it already provides correct semantics and behavior.
- Use Base UI for complex Dialog, AlertDialog, Menu, Popover, Combobox, Autocomplete, custom Select, Tabs, Tooltip, and related interaction patterns.
- Do not layer Base UI on Headless UI or expose both implementations through one component.
- Headless UI remains only behind legacy `ui`; remove it from each application package when that application has no remaining direct or transitive use.
- Continue to use React Hook Form as the form-state contract. Use `useController()` for Base UI-backed custom controls rather than substituting Base UI Form for application form state.

## Delivery principles

1. **One source of truth.** Tokens generate styles and foundation documentation; the catalogue registry generates routes, navigation, search, and status metadata; named fixtures feed both docs and tests.
2. **Vertical slices before breadth.** Prove the package, fixture, component, documentation, and test contracts end to end with Button, Input, Modal, and ActionMenu before porting the full catalogue.
3. **Parallelize by ownership.** Independent lanes own separate package or category directories. One integrator owns shared registries and export boundaries to reduce merge conflicts.
4. **Promote by evidence.** A component becomes Stable only after its API, metadata, fixtures, semantics, behavior tests, migration notes, and ownership are present.
5. **Cut over whole applications.** Repository packages coexist; styles within an application do not.
6. **Keep changes reviewable.** Use atomic commits and small package/component PRs before the per-application deployment PRs.
7. **Preserve history.** The audit, visual studies, and decision log remain unchanged and are linked from the finished style guide.

## Autonomous execution protocol

The implementation should minimize synchronous user input. An execution request starts Phase 0 immediately; passing a phase gate starts the next phase automatically when the required actions remain within the authority granted at kickoff.

Agents follow these rules:

- Read this plan, the decision log, repository instructions, and relevant current source before acting.
- Treat recorded decisions as authoritative; do not reopen them because an implementation offers another reasonable preference.
- Resolve reversible technical choices through evidence, a small spike, and an architecture decision record rather than asking the user.
- Record non-blocking uncertainties and continue with unaffected work.
- Batch progress into phase/gate reports instead of requesting approval for routine commits or implementation details.
- Ask the user only when a choice changes product behavior or visual direction beyond this plan, required authority or credentials are missing, an irreversible external action is not pre-authorized, or safe alternatives have been exhausted.
- Keep deployments rollback-capable and report the exact image/commit before each production cutover.

### One-time kickoff authority

The user can eliminate repeated operational prompts by including the desired authority in the execution request. The useful scopes are:

- create branches, commits, and pull requests;
- merge PRs after required checks and phase gates pass;
- deploy the persistent `paper.tadoku.app` preview after its gate;
- deploy admin, auth, and webv2 cutovers after their individual gates;
- switch `ui.tadoku.app` and remove legacy source after the final gate.

Authority for one item does not imply the others. Without explicit deployment or merge authority, agents complete and verify the work, then stop only at that external-action boundary.

## Phase 0 research packet

Phase 0 produces durable evidence under `docs/wip/tadoku-paper/research/` so later phases and fresh agents do not repeat discovery.

| Research stream | Questions answered | Durable output |
| --- | --- | --- |
| Repository surface | Which exports, deep imports, selectors, classes, examples, routes, and Headless UI primitives exist and who consumes them? | Component/consumer matrix and selector inventory. |
| Dependency compatibility | Which Base UI, Vite, router, test, and package-build versions work with the repository's React, TypeScript, Node, and pnpm constraints? | Compatibility matrix, lockfile spike, and ADRs. |
| Primitive behavior | Which controls use native HTML and which use Base UI; do Dialog, Menu, Combobox, and Button meet the required semantics and composition model? | Disposable vertical spike plus keyboard/focus findings. |
| CSS and recipes | How will static public classes, typed variant recipes, optional Tailwind mappings, themes, and densities share one implementation? | Recipe API ADR and generated-selector contract. |
| Documentation/content | What legacy knowledge is worth carrying over, what becomes a Pattern or Experiment, and which old routes need redirects? | Page/route migration matrix and content gap report. |
| Testing | What deterministic fixtures and behavior assertions replace the source-text test; what CI jobs and path filters are missing? | Test inventory, initial fixture map, and CI proposal. |
| Fonts/brand/assets | Where do licensed font files and canonical logo geometry come from, and which production asset formats are required? | Asset provenance/license record and export list. |
| Deployment | Where do image, ingress, Kubernetes, DNS, Tilt, and workflow changes live; how are static routes, health, caching, and rollback handled? | Deployment topology, cross-repository touchpoint list, and rollback runbook. |
| Resource usage | What does the static container consume at startup, idle, and during representative navigation? | Measured recommendation for requests and limits. |
| Application risk | Which admin, auth, and webv2 flows are critical, implicit-CSS-dependent, responsive, router-aware, or difficult to roll back? | Per-application smoke and risk matrices. |

Research spikes are not production implementations. Keep only reusable tests or scaffolding that already meets the plan's package and quality contracts; otherwise preserve the findings and discard the spike through normal reviewable changes.

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

The package builds ESM, declarations, and CSS. React, React DOM, and React Hook Form are peer dependencies. Base UI and Heroicons are direct implementation dependencies. Headless UI is prohibited. Package output must be consumable without Tailwind and without Next transpilation.

### Paper styleguide application

```text
frontend/apps/paper-styleguide/
├── index.html
├── package.json
├── vite.config.ts
├── src/
│   ├── app/
│   │   ├── routes.tsx
│   │   ├── DocsShell.tsx
│   │   └── CatalogueSearch.tsx
│   ├── documentation/
│   │   ├── DocumentPage.tsx
│   │   ├── ExampleCanvas.tsx
│   │   ├── StateMatrix.tsx
│   │   └── TableOfContents.tsx
│   ├── styles/
│   └── main.tsx
└── tests/
```

- Plain React + Vite; no Next.js and no server rendering.
- Imports `paper-ui/styles.css` once and consumes `paper-ui/catalog`.
- Production output is static and served by a minimal container with SPA route fallback.
- Merged builds deploy to `paper.tadoku.app` throughout migration.
- Local live review uses `t3-expose run` with the Vite server honoring the supplied `HOST` and `PORT`; the returned URL is temporary and private-lab/tailnet scoped.
- The existing `frontend/apps/styleguide` continues to consume only legacy `ui` and remains at `ui.tadoku.app` until final cutover.

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

`paper-styleguide` uses React and Vite. Its routing is application-owned; `paper-ui`, fixtures, and catalogue metadata remain router- and bundler-neutral.

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

Build the new documentation shell with:

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

### Required instructional sequence

Every Stable component page presents information in this order:

1. What the component is and the problem it solves.
2. When to use it.
3. When not to use it.
4. How to choose between related components.
5. Labeled anatomy and required parts.
6. One recommended realistic Tadoku example with rationale.
7. Supported variants and why each exists.
8. Meaningful states, density, theme, viewport, long-copy, and overflow cases.
9. Pointer, keyboard, focus, dismissal, validation, and loading behavior.
10. Content guidance and common mistakes.
11. Accessibility requirements and known constraints.
12. Copyable implementation examples with complete imports and semantic attributes.
13. React props, public types, class recipes, defaults, and invalid combinations.
14. Related product patterns and composed primitives.
15. Legacy migration notes.
16. Lifecycle, ownership, source, version, last review, and changelog.

The schema distinguishes required sections from optional component-specific sections. Registry validation prevents a component from becoming Stable when a required section is missing.

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
paper-ui foundations + Vite catalogue shell
   ↓
four-component vertical slice
   ↓
complete component catalogue at paper.tadoku.app
   ↓
admin → auth → webv2 cutovers
   ↓
ui.tadoku.app domain cutover
   ↓
legacy ui, styleguide, Headless UI removal + deferred hardening
```

Work beside the spine runs concurrently. Only shared contracts, the end-to-end pilot, and application deployment order are deliberately serial.

### Phase 0 — Lock contracts and migration guardrails

Goal: give every parallel lane stable boundaries before implementation starts.

| Lane | Work |
| --- | --- |
| Architecture | Finalize export map, Base UI boundary, router/link contract, Button recipe/component parity, and built-output format. |
| Catalogue | Freeze document, fixture, lifecycle, route, redirect, and registry schemas. |
| Inventory | Map every legacy export, deep import, global selector, Headless UI-backed component, styleguide example, and application consumer to its Paper destination. |
| Delivery | Define `paper.tadoku.app` and final `ui.tadoku.app` routing, per-app smoke lists, rollback evidence, CI matrix, resource measurement, owners, and PR boundaries. |

Checklist:

- [ ] Sync with current `main` while preserving the planning history and record the starting commit.
- [ ] Create `docs/wip/tadoku-paper/research/` with an index linking every research output and ADR.
- [ ] Produce the complete legacy component/consumer, selector/class, example, route, and Headless UI inventory.
- [ ] Run an isolated compatibility spike for Base UI, Vite, the chosen router, Vitest, React Testing Library, and the package build against the repository's pinned React, TypeScript, Node, and pnpm versions.
- [ ] Prototype Button/link recipes plus Base UI Dialog, Menu, and Combobox; record DOM, keyboard, focus, form, styling, bundle, and typing findings.
- [ ] Decide native-versus-Base UI ownership for each component category and record it in an ADR.
- [ ] Decide the typed recipe helper, CSS output, optional Tailwind mapping, and selector contract through a reversible spike and ADR.
- [ ] Verify font licenses/provenance, canonical Cut Meter geometry, favicon needs, and distributable asset formats.
- [ ] Locate all repository and cross-repository deployment touchpoints for `paper.tadoku.app`, static serving, health, ingress, image publication, and final hostname cutover.
- [ ] Produce the route/content migration matrix, initial deterministic fixture map, CI gap report, and per-application risk/smoke matrices.
- [ ] Record the final Button vocabulary: default, outline, ghost, link, destructive, and loading state.
- [ ] Specify that `Button variant="link"` is an action and that anchors receive button appearance from `buttonClassName()` or an anchor adapter.
- [ ] Decide the exact Base UI import strategy and the Paper-owned wrapper boundary.
- [ ] Freeze package exports, token names, fixture/document schema, lifecycle values, route format, and old-route redirect map.
- [ ] Inventory every legacy `ui` export, deep import, raw class, and global selector.
- [ ] Inventory every Headless UI-backed legacy component and verify that applications do not need direct Headless UI APIs after Paper cutover.
- [ ] Define migration markers and repository checks for allowed package use per application.
- [ ] Define production image names, static-server fallback, `paper.tadoku.app` ingress, final domain switch, and immutable-image rollback procedure.
- [ ] Define measurable resource collection for the static styleguide container.
- [ ] Update coexistence contributor instructions without changing legacy application behavior.

Repository guards must reject `ui` in migrated applications, `paper-ui` in unmigrated applications, private `paper-ui/src/*` imports, Next or Headless UI imports in Paper, and duplicate Paper stylesheet imports.

**Phase gate:** the indexed research packet, compatibility/primitive spikes, ADRs, reviewed contract, inventories, route map, deployment/rollback design, smoke matrices, ownership map, and automated boundary checks all exist; no unresolved finding requires a product decision. No production component implementation begins before this gate passes.

### Phase 1 — Build the Paper foundation

Run four lanes in parallel.

| Lane | Deliverables |
| --- | --- |
| Package/build | `paper-ui` scaffold, ESM/declaration build, export map, Base UI dependency, immutable Tailwind preset, root scripts, lockfile, package boundary checks, CI and Tilt triggers. |
| Visual foundation | Semantic tokens, themes, density, typography/font assets, focus, motion, borders, accent rail, hard-offset elevation, chart palette, Cut Meter, wordmark, favicon sources. |
| Test/catalogue | Vitest setup, RTL helpers, fixture and metadata schemas, registry validation, test utilities, catalogue export. |
| Paper styleguide | New React + Vite application, route resolver, registry stubs, shell primitives, isolated preview-canvas prototype, static container, and `paper.tadoku.app` deployment workflow. |

One integrator owns root exports, registry aggregation, and shared token names. Lanes add modules without repeatedly editing those files.

Checklist:

- [ ] Scaffold `frontend/packages/paper-ui` with explicit public exports, side-effect CSS, ESM, declarations, and package scripts.
- [ ] Add Base UI and Heroicons; prohibit Next and Headless UI in source and dependency checks.
- [ ] Implement semantic token layers, Paper base styles, public class recipes, theme/density attributes, and self-hosted fonts.
- [ ] Add Cut Meter, wordmark, and derived favicon sources in monochrome, reversed, and accented forms.
- [ ] Add Vitest, jsdom, Testing Library, `user-event`, jest-dom, deterministic cleanup, and server-render consumer smoke tests.
- [ ] Define typed catalogue, document, fixture, lifecycle, and registry-validation contracts.
- [ ] Scaffold `frontend/apps/paper-styleguide` with React, Vite, routing, Paper-only imports, and SPA deep-link fallback.
- [ ] Build the responsive shell, registry navigation/search stubs, local outline, and isolated preview-canvas prototype.
- [ ] Add a static production container and CI workflow watching `paper-ui` and `paper-styleguide`.
- [ ] Add a `paper.tadoku.app` deployment target without changing `ui.tadoku.app`.
- [ ] Document the live-review command using `t3-expose run`; verify Vite honors the helper's `HOST` and `PORT`.
- [ ] Capture idle and basic-navigation CPU/memory observations from the static container and set initial requests/limits from evidence.

**Phase gate:** `paper-ui` builds, typechecks, tests, packages, and server-renders without Next or Headless UI; CSS works without Tailwind; both themes and densities render; fonts/assets resolve; the Vite shell builds to static files, serves direct routes, drives stub navigation/search, passes Paper-only boundary checks, and is reachable at `paper.tadoku.app`.

### Phase 2 — Prove an end-to-end vertical slice

Implement Button, Input, Modal, and ActionMenu completely. These four jointly prove class/component parity, form anatomy, interactive edge treatment, floating elevation, overlay focus behavior, fixtures, metadata, documentation rendering, and tests.

Parallel ownership:

| Lane | Slice |
| --- | --- |
| Actions | Button, anchor styling, and optional LinkButton: default/outline/ghost/link/destructive, loading, icon, full-width, both densities. |
| Forms | Input anatomy: label, hint, error, required, read-only, disabled, both densities and themes. |
| Overlays | Base UI-backed Modal and ActionMenu: focus, keyboard, close/return behavior, destructive/disabled items, floating treatment. |
| Documentation/tests | Final page anatomy, state matrix, Preview/Code/API controls, fixture rendering, registry and behavior tests. |

Show real hover and focus in browser documentation; do not introduce fake production state classes solely for jsdom.

Checklist:

- [ ] Implement Button with one public recipe shared by raw classes and the React component.
- [ ] Default Button to `type="button"`; document explicit submit behavior and React Hook Form reset behavior.
- [ ] Prove that `Button variant="link"` remains a button and that anchors can use every appropriate button recipe without button roles.
- [ ] Implement loading through class and component APIs with stable accessible names, `aria-busy`, and safe disabled behavior.
- [ ] Implement Input through native HTML and React Hook Form context with deterministic label, hint, and error associations.
- [ ] Implement Modal and ActionMenu on Base UI without exposing Base UI as the application API.
- [ ] Add named deterministic fixtures for variants, states, themes, densities, long content, overflow, and narrow viewports.
- [ ] Add rendered semantics, keyboard, pointer, focus containment/return, and class/component parity tests.
- [ ] Build full Button, Input, Modal, and ActionMenu pages using the required 16-part instructional sequence.
- [ ] Review the four pages through a temporary `t3-expose` URL before merging and verify the merged version at `paper.tadoku.app`.

**Phase gate:** all four pages satisfy the validated instructional schema; raw classes, recipes, and React APIs share styles; Button/link semantics are correct; Base UI focus/keyboard tests pass; desktop/mobile shell and preview controls work; the reviewed build is live at `paper.tadoku.app`.

### Phase 3 — Complete `paper-ui` and the style-guide catalogue

With the pilot contracts frozen, run parallel component lanes in separate directories.

| Lane | Components and docs |
| --- | --- |
| Actions/feedback | ButtonGroup, Flash, Loading, toasts, cards and elevation recipes. |
| Forms | TextArea, Select, Checkbox, RadioSelect, RadioGroup, AmountWithUnit, autocomplete, multi-autocomplete, TagsInput, public option types; all use React Hook Form. |
| Navigation/overlays | Base UI-backed or native Navbar, Sidebar, Breadcrumb, Tabbar, VerticalTabbar, Pagination, Modal, ActionMenu, and router-neutral link integration. |
| Data/brand/layout | Tables, HeatmapChart, chart defaults/palette, layout recipes, Cut Meter and wordmark lockups. |
| Docs foundations | Principles, color, typography, spacing/density, borders/shape, elevation, icons, motion, layout, brand. |
| Docs governance | Component indexes, Patterns, Experimental logging-v2, contribution guide, lifecycle, deprecation policy, changelog. |

Every lane ships its fixtures, metadata, tests, migration notes, and realistic Tadoku examples with the implementation. Explicitly document currently missed exports such as Loading, VerticalTabbar, and AmountWithUnit.

Checklist:

- [ ] Complete Actions and Feedback components with fixtures, metadata, tests, and instructional pages.
- [ ] Complete every React Hook Form control and Base UI-backed complex field with fixtures, metadata, tests, and instructional pages.
- [ ] Complete Navigation and Overlay components with router-neutral APIs and keyboard/focus tests.
- [ ] Complete data-display, table, chart, layout, brand, and icon APIs with non-color semantics prepared.
- [ ] Build all Foundation pages from canonical Paper tokens and assets rather than copied values.
- [ ] Build category indexes and searchable component routes from the registry.
- [ ] Rebuild realistic logging examples as Patterns and keep logging-v2 visibly Experimental.
- [ ] Add contribution, lifecycle, deprecation, ownership, review, migration, and changelog documentation.
- [ ] Add “choose between,” common-mistake, rationale, content, and accessibility guidance to every Stable component.
- [ ] Ensure copied code examples include complete imports, semantic attributes, and both class/component APIs where applicable.
- [ ] Add old legacy-styleguide route mappings to the final-cutover redirect manifest without changing the legacy application yet.

**Phase gate:** the public export surface is complete; every Stable component passes registry, fixture, behavior, accessibility-baseline, and documentation checks; every component is searchable; navigation/search/status derive from one registry; all Paper routes validate at `paper.tadoku.app`; foundation pages use canonical sources.

### Phase 4 — Stabilize the Paper catalogue at `paper.tadoku.app`

This phase makes the parallel catalogue authoritative for Paper while the original styleguide remains authoritative only for legacy `ui`.

| Lane | Work |
| --- | --- |
| Experience | Complete responsive shell, keyboard-first search, local outline, deep links, preview isolation, code copy, and theme/density/viewport controls. |
| Content quality | Review the 16-part page contract, cross-component choice guidance, realistic examples, Patterns, Experiments, and design history. |
| Static delivery | Harden direct-route fallback, caching, compression, font/assets, container health, resource limits, and `paper.tadoku.app` deployment. |
| Verification | Route/registry tests, keyboard/mobile shell tests, package tests, visual human review, production/image builds, and telemetry checks. |

Checklist:

- [ ] Verify every public Paper component has one canonical page and every legacy route has a planned final redirect.
- [ ] Verify every Stable page satisfies the complete instructional sequence and contains no placeholder or generic-only examples.
- [ ] Verify Preview, Code, API/Props, and Accessibility views work without layout jumps.
- [ ] Verify preview frames exercise real phone, tablet, and desktop media queries.
- [ ] Verify search, navigation, status, metadata, and source links all derive from the registry.
- [ ] Link the audit, all visual studies, decision log, and implementation plan under Design history.
- [ ] Verify direct HTTP requests to nested SPA routes load successfully.
- [ ] Measure static-container idle and navigation resources; tune requests and limits while retaining reliable startup.
- [ ] Run lint, typecheck, Paper tests, styleguide tests, production build, container build, and deployment smoke tests.
- [ ] Mark `paper.tadoku.app` as the authoritative Paper catalogue while keeping `ui.tadoku.app` unchanged.

**Phase gate:** `paper.tadoku.app` is a complete, stable, keyboard-operable Paper catalogue with no legacy `ui`, Next, or Headless UI dependencies; all content and delivery checks pass; the original styleguide still serves legacy documentation independently.

### Phase 5 — Admin big-bang cutover

Admin is the compact-density stress test. Work in parallel on one integration branch or coordinated PR stack, then switch the application once.

| Lane | Scope |
| --- | --- |
| Shell | Dashboard layout, Sidebar, branding, Loading, access-denied and redirect states. |
| Editors/forms | Content and announcement editors, namespace forms, language/user editing, CodeMirror adapter using `useController`. |
| Data/actions | Lists, tables, pagination, breadcrumbs, tabs, menus, dialogs, destructive actions, narrow layouts. |
| Verification | Import/class audit, compact screenshots for human review, CRUD and role smoke suite, build matrix. |

Review old `primary`, `secondary`, and `danger` combinations semantically; do not perform a blind class rename. The emphasized action becomes default, neutral becomes outline, and destructive intent becomes destructive. Compact controls must not make icon-only targets inaccessible.

Checklist:

- [ ] Migrate dashboard shell, Sidebar, branding, Loading, access-denied, and redirect states.
- [ ] Migrate content, announcement, namespace, language, and user forms while retaining React Hook Form.
- [ ] Migrate CodeMirror through `useController()` without making it a Paper primitive.
- [ ] Migrate tables, pagination, breadcrumbs, tabs, menus, dialogs, and destructive actions.
- [ ] Replace router behavior with admin-owned adapters.
- [ ] Remove all admin `ui` imports, legacy stylesheet use, legacy-only classes, and unused Headless UI dependency.
- [ ] Load Paper CSS once and apply compact density at the application root.
- [ ] Add every newly discovered shared state to Paper fixtures, tests, and documentation before marking it Stable.
- [ ] Run role, loading, access-denied, CRUD, preview, announcement, language/user, dialog, and narrow-layout smoke tests.
- [ ] Record the deployed image, deploy only admin, verify telemetry, and retain the previous image for rollback.

**Cutover gate:** admin contains only Paper, has no Headless UI dependency, passes its critical workflows at compact density, and passes lint, typecheck, production build, container build, deployment smoke, and rollback-readiness checks.

### Phase 6 — Auth big-bang cutover

Auth has few imports but high implicit-CSS and flow risk because Ory generates dynamic form nodes.

| Lane | Scope |
| --- | --- |
| Shell | Navigation, Cut Meter/wordmark, comfortable layout, feedback and loading. |
| Ory fields | Node adapters, labels, hints, errors, checkbox/text/select semantics, WebAuthn/script preservation. |
| Actions/flows | Explicit default/outline/ghost/link/destructive hierarchy, loading and double-submit prevention, removal of selector-order button hacks. |
| Verification | Fixtures and production-like login, registration, recovery, verification, settings, OIDC/passkey/WebAuthn checks. |

Checklist:

- [ ] Migrate the comfortable shell, navigation, brand, feedback, and loading states.
- [ ] Map every Ory node type to Paper fields with correct labels, hints, errors, names, and values.
- [ ] Preserve OIDC, passkey, WebAuthn, and required script behavior.
- [ ] Assign button variants by semantic job and remove selector-order hacks.
- [ ] Verify loading, `aria-busy`, disabled state, double-submit prevention, and keyboard submission.
- [ ] Remove all auth `ui` imports, legacy global-form dependencies, duplicate stylesheet imports, and unused Headless UI dependency.
- [ ] Test login, registration, recovery, verification, settings, configured identity methods, field errors, and global errors.
- [ ] Record the deployed image, deploy only auth, verify production-like Ory flows and telemetry, and retain rollback.

**Cutover gate:** auth contains only Paper, has no Headless UI dependency, preserves every configured identity flow and semantic error relationship, and passes lint, typecheck, production build, image build, focused Ory smoke, and rollback-readiness checks.

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

Checklist:

- [ ] Migrate global Navigation, Footer, announcement, signed-in/out/admin/banned, and mobile states.
- [ ] Migrate old/new/edit logging and contest create/register/submit flows without unrelated product consolidation.
- [ ] Migrate leaderboards, profiles, lists, tables, pagination, tabs, menus, responsive behavior, and overflow.
- [ ] Migrate blog, pages, manual, Page Counter exception, Chart.js palettes, heatmaps, and activity charts.
- [ ] Supply routing and current-page data through webv2-owned adapters.
- [ ] Remove all webv2 `ui` imports, legacy stylesheet use, legacy-only classes, and unused Headless UI dependency.
- [ ] Add discovered shared states to Paper fixtures/tests/docs before Stable promotion.
- [ ] Run signed-in/out/admin/banned, navigation, logging, contest, profile, content, pagination, tab, chart, and responsive smoke matrices.
- [ ] Record the deployed image, deploy only webv2, verify telemetry, and retain rollback.
- [ ] Run the complete frontend package/application and container-image matrix.

**Cutover gate:** webv2 contains only Paper, has no Headless UI dependency, passes every representative role and product workflow, and passes lint, typecheck, production build, image build, whole-frontend compatibility, deployment smoke, and rollback-readiness checks.

### Phase 8 — Move the catalogue and retire legacy code

Only begin when a repository-wide search shows zero consumers:

```sh
rg "from ['\"]ui|ui/styles|ui/components|\"ui\": \"workspace|@headlessui/react" frontend
```

Checklist:

- [ ] Confirm admin, auth, and webv2 production smoke windows have completed successfully.
- [ ] Build and record the exact Paper styleguide image that will receive `ui.tadoku.app`.
- [ ] Configure direct legacy route redirects and SPA fallback for the final hostname.
- [ ] Switch `ui.tadoku.app` atomically from the legacy styleguide image to the Paper styleguide image.
- [ ] Verify canonical routes, old bookmarks, search, assets, theme/density controls, and design-history links on `ui.tadoku.app`.
- [ ] Keep `paper.tadoku.app` as a temporary alias or redirect according to the deployment policy, then remove it when no longer needed.
- [ ] Retain the previous legacy styleguide image and verify the domain rollback procedure.
- [ ] Remove `frontend/apps/styleguide` and its build workflow after the final smoke window.
- [ ] Remove `frontend/packages/ui`, legacy dependencies, lockfile entries, transpilation entries, Tailwind inheritance, and old Tilt triggers.
- [ ] Remove Headless UI from the workspace after repository search shows zero uses.
- [ ] Update all CI paths, Tilt triggers, contributor instructions, documentation status, and styleguide changelog to Paper-only operation.
- [ ] Run Paper tests and the complete frontend/application-image matrix from the clean repository.

Previously deployed application images remain rollback-capable because their legacy dependencies are bundled.

**Phase gate:** `ui.tadoku.app` serves the verified static Paper catalogue; old bookmarks work; the repository has zero `ui` and Headless UI consumers; the legacy package and application are removed; all builds and smoke tests pass; immutable rollback images remain recorded.

### Phase 9 — Finish the deferred hardening issues

These workstreams may begin earlier once their prerequisites exist, but close after the migrations only when their full acceptance criteria are satisfied.

| Issue | Parallel work after prerequisites | Completion point |
| --- | --- | --- |
| #761 Accessibility contract | Component audits can start after Phase 2; product reflow checks start with each migrated app. | All documented controls, flows, contrast, keyboard, zoom/reflow, and exception tracking complete. |
| #762 Regression gates | Playwright fixture routes can start after the style-guide pilot; burn-in continues through app migrations. | Reviewed responsive baselines, failure artifacts, axe integration, and stable blocking CI complete. |
| #763 Themes/preferences | Dark/reduced-motion/forced-color fixtures can start after tokens; app persistence waits for product rollout decisions. | Preference behavior and representative app/theme matrix complete. |

Checklist:

- [ ] Complete #761's component and application WCAG 2.2 contract, contrast, keyboard, zoom/reflow, forced-color, and exception tracking.
- [ ] Complete #762's Playwright fixture routes, responsive screenshot baselines, real interaction states, axe scans, failure artifacts, and reviewed update workflow.
- [ ] Complete #763's theme persistence, operating-system preferences, reduced motion, forced colors, dark-mode application validation, and non-color status/chart cues.
- [ ] Run the resulting checks against `paper-ui`, `paper-styleguide`, admin, auth, and webv2.
- [ ] Update component lifecycle and known-constraint metadata from the final findings.

**Phase gate:** issues #761, #762, and #763 meet their full acceptance criteria; the resulting checks are documented, stable, and enforced at the agreed CI level.

## Parallel execution map

At maximum useful concurrency, staff the program as follows:

| Stream | Phases 0–2 | Phase 3 | Phase 4 | Phases 5–7 |
| --- | --- | --- | --- | --- |
| A | Package/build | Actions/feedback | Static delivery/resources | Active app shell |
| B | Visual foundations | Forms | Instructional content review | Active app forms/workflows |
| C | Test/catalogue | Navigation/overlays | Shell/search/canvas | Active app data/navigation |
| D | Vite styleguide | Data/brand/layout | Tests/governance | Verification plus next-app inventory |

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
- Convert buttons by semantic job rather than text substitution: emphasized to default, neutral to outline, quiet action to ghost or link, destructive intent to destructive.
- Load Paper styles once and set the application density.
- Replace Next-aware shared behavior with application adapters.
- Replace Headless UI-backed legacy components with their Paper/Base UI-backed equivalents without exposing Base UI to the application.
- Update package, Tailwind, framework, image, CI, Tilt, and asset references as applicable.

### 4. Prove zero mixing

For the target application, repository checks must find no:

- `ui` dependency or import;
- `ui/styles/globals.css` import;
- `ui/components/*` deep import;
- `require('ui/tailwind.config.js')` or equivalent;
- class that exists only through undocumented legacy globals;
- duplicate `paper-ui/styles.css` import;
- direct or unused `@headlessui/react` dependency.

### 5. Verify and deploy

- Run `paper-ui` typecheck, test, and build.
- Run target-app lint, `tsc --noEmit`, production build, and image build.
- Execute the app-specific smoke matrix on narrow and desktop layouts.
- Deploy only the target application.
- Re-run critical production smoke checks and observe frontend/error telemetry.

### 6. Roll back if necessary

Redeploy the recorded previous image. Do not attempt to fix a failed cutover by loading both stylesheets. Source rollback is a normal revert of the migration commits; no database or API rollback is involved.

## Delivery and migration matrix

| Order | Target | Default density | Primary risk | Required proof |
| ---: | --- | --- | --- | --- |
| 0 | `paper-styleguide` at `paper.tadoku.app` | switchable | The new system must document and test itself without legacy assistance. | Every route, search, canvas, fixture, registry, instructional, static-delivery, and resource contract works with zero legacy/Headless imports. |
| 1 | admin | compact | Dense CRUD, tables, editors, menus, and modal flows. | Core administrative workflows and narrow layouts work at 36px density with zero Headless UI. |
| 2 | auth | comfortable | Dynamic Ory nodes rely heavily on implicit global CSS. | Every configured identity flow, error state, keyboard path, and submit/loading state works. |
| 3 | webv2 | comfortable | Largest surface and most router/data/chart combinations. | Representative user roles and core logging, contest, content, profile, navigation, and chart flows work. |
| 4 | `ui.tadoku.app` and cleanup | switchable | Domain continuity, legacy bookmarks, and irreversible source cleanup. | Paper serves both canonical and redirected routes; rollback image is recorded; repository has zero `ui` or Headless UI consumers. |

## Validation commands

Exact package scripts will be added in Phase 1. The required command shape is:

```sh
cd frontend
pnpm --filter paper-ui typecheck
pnpm --filter paper-ui test
pnpm --filter paper-ui build
pnpm --filter paper-styleguide lint
pnpm --filter paper-styleguide test
pnpm --filter paper-styleguide exec tsc --noEmit
pnpm --filter paper-styleguide build
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
- `paper-ui` builds and is independently consumable by React without Next.js or Headless UI;
- class recipes and React components are documented and tested as one styling contract;
- Button variants use the accepted default/outline/ghost/link/destructive vocabulary with correct button-versus-link semantics;
- every public component has a canonical searchable page, deterministic fixtures, lifecycle metadata, ownership, tests, and migration guidance;
- foundations, components, product patterns, and experiments are visibly separate;
- `paper-styleguide` is a static React + Vite application that uses only Paper and satisfies the full instructional page contract;
- admin, auth, and webv2 each use only Paper and have passed their deployment smoke gates;
- `paper.tadoku.app` supports the migration and `ui.tadoku.app` serves the final Paper catalogue with legacy-route redirects;
- the original audit, all visual studies, and the decision log remain available;
- repository and CI configuration watches and validates `paper-ui`;
- legacy `ui`, the old styleguide, and Headless UI have zero consumers and are removed;
- issues #761–#763 meet their full accessibility, regression, and theme/preference acceptance criteria.

## Audit traceability

| Original recommendation | Delivery |
| --- | --- |
| P0.1 Semantic tokens | Phases 1 and 3 foundation work; token/package gates. |
| P0.2 Visual grammar | Confirmed contract; Phases 1–3 implementation and documentation. |
| P0.3 Complete docs shell | New Vite catalogue in Phases 1–4; no in-place legacy shell migration. |
| P0.4 Accessibility contract | Immediate semantic tests throughout; comprehensive completion in #761/Phase 9. |
| P0.5 Typed Button/LinkButton | Shadcn-inspired default/outline/ghost/link/destructive recipe, class/component parity, and semantic anchor styling proven in Phase 2. |
| P1.1 One route and search | Registry-driven docs in Phases 1, 3, and 4. |
| P1.2 Full state matrix | Fixture schema and every component page, beginning in Phase 2. |
| P1.3 Foundation pages | Phase 3. |
| P1.4 Co-located examples/metadata/tests | Package architecture and Phases 1–3. |
| P1.5 Automated quality gates | Fast tests and CI now; comprehensive visual/axe system in #762/Phase 9. |
| P2.1 Components/patterns/experiments | Information architecture and Phases 3–4. |
| P2.2 Lifecycle/governance | Registry schema, Stable gate, contribution and deprecation policy in Phases 1–4. |
| P2.3 Themes/preferences | Light/dark tokens and preview now; full preference behavior in #763/Phase 9. |
| P2.4 Package boundary | `paper-ui` architecture, Base UI wrapper boundary, Vite consumer proof, boundary tests, app adapters, and migration gates. |

The early quick wins—framed previews, constrained prose, refined violet, and explicit use/avoid guidance—are incorporated into the documentation shell and visual contract rather than scheduled as throwaway work.

## Follow-up opportunities unlocked by Paper

- Publish a Paper registry or codemod that can add approved patterns without making generated shadcn source the canonical implementation.
- Generate typed API tables and migration reports from TypeScript declarations and catalogue metadata.
- Generate design-token packages for non-React surfaces and design tooling from the same semantic sources.
- Move admin, auth, or webv2 away from Next.js independently now that their shared UI has no router/framework dependency.
- Evaluate object-storage or CDN delivery for the static styleguide if eliminating its Kubernetes pod becomes more valuable than deployment consistency.
- Add localized documentation and right-to-left specimens using the existing fixture and preview contracts.
- Add automated component dependency graphs and “used by” views to help assess breaking changes.
- Promote recurring product compositions into governed Patterns only after application migrations reveal real reuse.
