# Headless UI legacy inventory

## Source imports

Only `frontend/packages/ui` imports `@headlessui/react` in repository TypeScript source. Admin, auth, and webv2 nevertheless declare it as a direct dependency, so their manifests contain removable dependencies after each app no longer consumes legacy `ui`.

| Legacy component | Headless UI primitives | Important behavior to preserve/test | Paper ownership |
| --- | --- | --- | --- |
| `ActionMenu` | `Menu`, `MenuButton`, `MenuItems`, `MenuItem` | button/menu semantics, Arrow/Escape navigation, focus, outside dismissal, non-modal positioning, left/right anchor, dangerous item styling | Base UI Menu behind Paper `ActionMenu` |
| `AutocompleteInput` | `Combobox`, `ComboboxInput`, `ComboboxButton`, `ComboboxOptions`, `ComboboxOption` | controlled RHF value, matching/limit, accessible label, keyboard selection, close/reset, empty state | Base UI Autocomplete/Combobox + `useController()` |
| `AutocompleteMultiInput` | same Combobox family | multi-value identity comparator, display formatting, limit/filtering, keyboard/focus | Base UI Combobox + `useController()` |
| `TagsInput` | `Combobox`, `ComboboxInput`, `ComboboxOptions`, `ComboboxOption` | async/debounced suggestions, free entry, deduplication/normalization, max tags, remove buttons, loading/empty text | Paper TagsInput composed over Paper/Base UI combobox |
| `RadioGroup` | aliased `RadioGroup`, `Radio`, `Label` | RHF controlled value, required rule, disabled options, label/name, keyboard selection | Base UI RadioGroup wrapper + `useController()` |
| `Modal` | `Dialog`, `DialogPanel`, `DialogTitle`, `Transition`, `TransitionChild` | modal semantics, focus entry/containment/return, Escape/backdrop close, title, scroll, transition | Base UI Dialog behind Paper Modal/Dialog API |
| `Navbar` | `Disclosure`, `DisclosureButton`, `DisclosurePanel`, `Menu`, `MenuButton`, `MenuItems`, `MenuItem` | mobile disclosure, desktop account menus, current-route state, focus/Escape, body scroll lock/restore, mobile-primary/bottom grouping | Paper navigation primitives; app supplies routes/link renderer |

No legacy component imports both Headless UI and another headless primitive library. Paper must not expose Headless types or combine Headless UI with Base UI.

## Manifest and transitive state

| Package | Declares `@headlessui/react` | Direct source import | Required removal point |
| --- | ---: | ---: | --- |
| `ui` | yes | yes (seven component files) | final legacy removal |
| admin | yes | no | admin cutover |
| auth | yes | no | auth cutover |
| webv2 | yes | no | webv2 cutover |
| styleguide | no | no | retired with legacy styleguide; dependency is transitive through `ui` |

The pinned range is `^2.2.0`; the lockfile currently resolves `2.2.9` for React 18. This is inventory evidence only, not the Base UI compatibility decision.

## Guard implications

- Check `dependencies`, `devDependencies`, and source imports separately.
- Allow Headless UI only under `frontend/packages/ui/**` and only while legacy consumers exist.
- Reject it from `paper-ui` and `paper-styleguide` from their first commit.
- For a migrated app, reject direct imports, manifest entries, and lockfile importer entries; do not require global lockfile absence until all apps and `ui` are removed.
- Behavioral replacement tests must be rendered tests. The current lone Navbar test inspects source text and does not prove disclosure/menu keyboard or focus behavior.

## Verification commands

```sh
rg -n '@headlessui/react' frontend --glob '*.{ts,tsx,js,jsx,json}'
rg -n --glob package.json '"@headlessui/react"' frontend
rg -n 'packages/ui:|@headlessui/react@' frontend/pnpm-lock.yaml
```
