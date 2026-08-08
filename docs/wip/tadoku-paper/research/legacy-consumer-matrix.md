# Legacy `ui` consumer matrix

Inventory date: 2026-08-08  
Starting commit: `92083ac7fa3972648d154f78b4d7e398bb985502`

## Application integration matrix

| Consumer | Root symbols observed | Deep paths observed | Styles/configuration | Headless UI manifest state |
| --- | --- | --- | --- | --- |
| admin | `ActionMenu`, `AutocompleteInput`, `Breadcrumb`, `Input`, `Loading`, `Logo`, `Modal`, `Pagination`, `Select`, `Sidebar`, `Tabbar`, `TextArea` | `ui/components/toasts` (`ToastContainer`) | one globals import in `pages/_app.tsx`; inherits/mutates `ui/tailwind.config.js`; `transpilePackages: ['ui']` | direct dependency declared; no direct source import |
| auth | `ToastContainer` | `ui/components/Navbar` (`Navbar`, `NavigationDropDownProps`, `NavigationLinkProps`) | globals imported in both `pages/_app.tsx` and `src/Navigation.tsx`; inherits/mutates legacy Tailwind config; transpiles `ui` | direct dependency declared; no direct source import |
| styleguide | `ActionMenu`, `AutocompleteInput`, `AutocompleteMultiInput`, `Breadcrumb`, `ButtonGroup`, chart values, `Checkbox`, `Flash`, `HeatmapChart`, `Input`, both logos, `Modal`, `Navbar`, `Pagination`, `RadioSelect`, `Select`, `Sidebar`, `Tabbar`, `TagsInput`, `TextArea`, `ToastContainer` | `ui/components/Form` (`AmountWithUnit`, `RadioGroup`) | globals imported in `_app.tsx` and also displayed as literal code in `index.tsx`; inherited legacy Tailwind config adds examples glob; transpiles `ui` | no direct dependency; receives it through `ui` |
| webv2 | `ActionMenu`, `AutocompleteMultiInput`, `Breadcrumb`, `ButtonGroup`, chart values, `Checkbox`, `Flash`, `HeatmapChart`, `Input`, `Loading`, `LogoInverted`, `Modal`, `Pagination`, `RadioGroup`, `Tabbar`, `TagsInput`, `TextArea`, `VerticalTabbar` | `ui/components/Flash`; `ui/components/Form` (`AmountWithUnit`, `Option`, `OptionGroup`, `Select`); `ui/components/Form/RadioGroup` (`RadioProps`); `ui/components/Navbar` (component + types); `ui/components/toasts` | one globals import in `pages/_app.tsx`; inherits/mutates legacy Tailwind config and dynamic activity safelist; transpiles `ui` | direct dependency declared; no direct source import |

Counts should be regenerated at cutover because the app source is active. This inventory intentionally records symbol presence rather than treating repeated imports as distinct API requirements.

## Consumer-specific coupling

### admin

- Dense CRUD/editor pages consume nearly every navigation, overlay, list, and form family needed for the compact-density gate.
- `CodeEditor.tsx` is app-owned CodeMirror and is not a legacy `ui` export. It must remain app-owned and integrate through React Hook Form `useController()`.
- Content HTML uses the global `.auto-format` recipe; data screens use raw `.default` table cells/headers and `.table-container`.
- Buttons include default, primary, secondary, danger, and ghost jobs. The class names cannot be mechanically mapped because legacy default is neutral while Paper default is emphasized.

### auth

- Few component imports conceal a large implicit CSS dependency: Ory nodes render native inputs/buttons and class strings in `src/ui/*`.
- `Flow.tsx` owns `.kratos-form`; `NodeInputSubmit.tsx` emits `btn primary`; `NodeInputButton.tsx` emits `btn`.
- The duplicate stylesheet import violates the future “load Paper once” contract and is a migration guard test case.
- Navbar’s exported deep types are used to construct app-owned route data.

### styleguide

- It is the broadest legacy package showcase, but not the full public surface: `Loading`, `VerticalTabbar`, and several deep public types are absent from examples.
- Raw recipes (buttons, templates, tables, typography) are first-class content even when no React component exists.
- Examples are source-loaded via Webpack `?raw`; the new Vite catalogue must replace the mechanism while keeping deterministic, copyable fixtures.

### webv2

- It is the widest product consumer and the only user of `HeatmapChart`, chart palette exports, `VerticalTabbar`, `RadioProps`, `OptionGroup`, and most deep form exports.
- It contains two logging form implementations plus the logging-v2 details/contest flow. The migration must preserve both and avoid product consolidation.
- Navigation, breadcrumbs, tabs, and pagination are router-aware through legacy components; Paper needs webv2-owned adapters.
- `tailwind.config.js` mutates the shared config and adds runtime-derived activity-color safelist entries. The cutover must preserve those product mappings until semantic chart/activity tokens replace them.

## Shared package/build coupling

All four apps depend on the mutable legacy Tailwind object and/or Next transpilation. Product-image workflows watch `frontend/packages/ui/**`; styleguide has its own image workflow. Current production-style image names are:

- `ghcr.io/tadoku/tadoku/frontend-admin`
- `ghcr.io/tadoku/tadoku/frontend-auth`
- `ghcr.io/tadoku/tadoku/frontend-webv2`
- `ghcr.io/tadoku/tadoku/frontend-styleguide`

The new package workflow must be added without removing these watches until each consumer is migrated.

## Verification commands

```sh
rg -n --glob '*.{ts,tsx,js,jsx,mjs,cjs}' "from ['\"]ui(?:/[^'\"]*)?['\"]|import ['\"]ui(?:/[^'\"]*)?['\"]|require\(['\"]ui" frontend/apps
rg -n --glob package.json '"ui"|"@headlessui/react"' frontend
rg -n 'transpilePackages|ui/tailwind.config|ui/styles/globals.css' frontend/apps
rg -n 'frontend/packages/ui|PROJECT_NAME|IMAGE_NAME' .github/workflows/build-frontend-*.yaml
```

