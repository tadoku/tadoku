# Legacy `ui` API inventory

Inventory date: 2026-08-08  
Starting commit: `92083ac7fa3972648d154f78b4d7e398bb985502`

## Scope and resolution behavior

`frontend/packages/ui/package.json` has no `main`, `module`, `types`, `files`, or `exports` map. Workspace consumers resolve the root barrel through repository TypeScript/Next behavior and can address any file below the package by path. Therefore the migration surface is larger than `index.ts`: source modules, styles, Tailwind configuration, SVGs, Next configuration, and copied `public/` assets are all reachable deep paths.

The root package is React- and Next-coupled. `Breadcrumb`, `ButtonGroup`, `Flash`, `Navbar`, `Pagination`, `Sidebar`, `Tabbar`, and branding use `next/link`, `next/router`, or `next/image`; Paper equivalents must use the consumer-supplied link/router boundary in the implementation plan.

## Root barrel (`ui`)

Every symbol below is exported by `frontend/packages/ui/index.ts`.

| Root export | Kind | Legacy implementation | Paper destination |
| --- | --- | --- | --- |
| `ActionMenu` | component | Headless UI Menu | `components/actions/action-menu` |
| `Logo`, `LogoInverted` | components | Next Image + legacy SVGs | `foundations/brand` / Paper brand components |
| `Breadcrumb` | component | Next Link | `components/navigation/breadcrumb` |
| `ButtonGroup` | component | responsive group composed with ActionMenu and Next Link | `patterns/action-group` unless the final catalogue proves it primitive enough for Actions |
| `chartColors`, `chartDatasetDefaults` | values | raw Chart.js palette/defaults | `foundations/chart-palette` |
| `Flash` | component | optional Next Link; info/success/warning/error | `components/feedback/flash` (or final Alert naming with `Flash` migration alias only if approved) |
| `Input`, `TextArea`, `Checkbox`, `Select`, `RadioSelect` | RHF components | native controls + `useFormContext()` | `components/forms/*`; retain React Hook Form contract |
| `AutocompleteInput`, `AutocompleteMultiInput` | RHF components | Headless UI Combobox | Paper Base UI autocomplete/combobox wrappers |
| `RadioGroup` | RHF component | Headless UI RadioGroup | Paper Base UI RadioGroup wrapper |
| `TagsInput` | RHF component | Headless UI Combobox + custom tag state | Paper TagsInput, built on Paper/Base UI combobox behavior |
| `HeatmapChart` | component | SVG/DOM heatmap | `components/data-display/heatmap` |
| `Modal` | component | Headless UI Dialog/Transition | Paper Base UI Dialog wrapper; legacy name may be documented as migration terminology |
| `Navbar` | component | Headless UI Disclosure/Menu + Next router/link | Paper Navbar with injected links/current route |
| `Sidebar` | component | Next Link | `components/navigation/sidebar` with injected links |
| `Tabbar`, `VerticalTabbar` | components | Next Link | Paper Tabs/linked-tab navigation adapter; retain horizontal and vertical fixtures |
| `Pagination` | component | Next Link + router query | Paper Pagination with `getHref`/`onPageChange` |
| `ToastContainer` | component | react-toastify container | Paper toast viewport/provider decision; preserve notification behavior during app cutovers |
| `Loading` | component | animated loading surface | `components/feedback/loading` |

There are 26 root named exports. No props or other public types are re-exported from the root.

## Deep module exports

Because there is no export map, these named declarations are deep-importable even when absent from the root. “Root?” describes `ui/index.ts`, not whether an application currently uses the symbol.

| Deep path | Exported declarations | Root? | Observed direct consumer |
| --- | --- | ---: | --- |
| `ui/components/ActionMenu` | `ActionMenu` | yes | none (root used) |
| `ui/components/Breadcrumb` | `Breadcrumb` | yes | none |
| `ui/components/ButtonGroup` | `ButtonGroup` | yes | none |
| `ui/components/Flash` | `Flash` | yes | webv2 announcement banner |
| `ui/components/Form` | `AmountWithUnit`, `AutocompleteInput`, `AutocompleteMultiInput`, `Checkbox`, `Input`, `RadioGroup`, `RadioSelect`, `Select`, `TagsInput`, `TextArea`; types `Option`, `OptionGroup` | all except `AmountWithUnit`, `Option`, `OptionGroup` | styleguide and webv2 |
| `ui/components/Form/AmountWithUnit` | `AmountWithUnit` | no | through `ui/components/Form` |
| `ui/components/Form/AutocompleteInput` | `AutocompleteInput` | yes | none direct |
| `ui/components/Form/AutocompleteMultiInput` | `AutocompleteMultiInput` | yes | none direct |
| `ui/components/Form/Checkbox` | `Checkbox` | yes | none direct |
| `ui/components/Form/Input` | `Input` | yes | none direct |
| `ui/components/Form/RadioGroup` | `RadioGroup`; interface `RadioProps` | component only | webv2 imports `RadioProps` |
| `ui/components/Form/RadioSelect` | `RadioSelect` | yes | none direct |
| `ui/components/Form/Select` | `Select` | yes | through `ui/components/Form` |
| `ui/components/Form/TagsInput` | `TagsInput` | yes | none direct |
| `ui/components/Form/TextArea` | `TextArea` | yes | none direct |
| `ui/components/Form/types` | interfaces `FormElementProps`, `Option`, `OptionGroup` | `Option`, `OptionGroup` only through Form barrel | none at this exact path |
| `ui/components/HeatmapChart` | `HeatmapChart` | yes | none direct |
| `ui/components/Loading` | `Loading` | yes | none direct |
| `ui/components/Modal` | `Modal` | yes | none direct |
| `ui/components/Navbar` | `Navbar`; interfaces `NavigationDropDownProps`, `NavigationLinkProps` | component only | auth and webv2 import component and types |
| `ui/components/Pagination` | `Pagination` | yes | none direct |
| `ui/components/Sidebar` | `Sidebar` | yes | none direct |
| `ui/components/Tabbar` | `Tabbar`, `VerticalTabbar` | yes | none direct |
| `ui/components/branding` | `Logo`, `LogoInverted` | yes | none direct |
| `ui/components/charts` | `chartColors`, `chartDatasetDefaults` | yes | none direct |
| `ui/components/toasts` | `ToastContainer` | yes | admin and webv2 |

Most component prop interfaces are deliberately unexported. The exceptions are `RadioProps`, `NavigationDropDownProps`, `NavigationLinkProps`, `FormElementProps`, `Option`, and `OptionGroup`. Paper should export intentionally supported public types and block `paper-ui/src/*` imports rather than reproduce accidental filesystem visibility.

## Non-TypeScript deep surface

| Path | Current role | Migration disposition |
| --- | --- | --- |
| `ui/styles/globals.css` | shared unscoped base, component, and utility stylesheet | replace once per migrated app with `paper-ui/styles.css`; never co-load |
| `ui/tailwind.config.js` | mutable inherited Tailwind configuration | replace with immutable Paper preset or app-owned config |
| `ui/next.config.js` | package-local Next configuration | no Paper equivalent; Paper is framework-independent |
| `ui/postcss.config.js` | legacy CSS pipeline configuration | package-build concern only; not a Paper consumer export |
| `ui/components/logo.svg`, `logo-light.svg` | legacy wordmarks consumed through Next Image | replace with canonical Cut Meter/wordmark assets |
| `ui/public/static/favicon.png`, `favicon-dark.png` | source favicon copies | replace with Paper-derived favicons in each app/catalogue |
| `ui/tests/navbar-mobile.test.mjs` | source-text mobile Navbar assertions | replace with rendered behavior tests |

## Verification commands

```sh
sed -n '1,240p' frontend/packages/ui/index.ts
rg -n '^export' frontend/packages/ui/components frontend/packages/ui/index.ts
rg -n "from ['\"]ui(?:/[^'\"]*)?['\"]|import ['\"]ui/" frontend/apps
rg --files frontend/packages/ui | sort
```
