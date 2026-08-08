# Tadoku Paper primitive spike findings

Status: Phase 0 evidence complete  
Primitive dependency: `@base-ui/react` 1.7.0  
Runtime: React 18.2.0, Node 20.20.2

## Spike surface

The disposable vertical spike implemented one native Paper Button and anchor recipe plus Paper-owned wrappers around Base UI Dialog, Menu, and Combobox. Five Vitest/React Testing Library tests covered semantic roles, accessible names, default button type, loading state, portals, focus entry/return, Escape, arrow navigation, disabled menu behavior, filtering, active-descendant selection, and native form data.

All five tests, TypeScript 5.9 typechecking, tsup package output, Vite production output, React server rendering, TypeScript 4.9 declaration consumption, and package-list validation passed. No spike scaffolding remains in the repository.

## Button and anchor recipe

One `buttonClassName({ variant, loading, className })` helper generated the same public classes used by the React adapter: `btn`, `btn--<variant>`, and `is-loading`. The exact selector spelling remains owned by the recipe/CSS ADR; this spike validated the composition model.

| Case | Result |
| --- | --- |
| `<PaperButton>` default | Native `<button type="button">`; no accidental form submission. |
| loading | `disabled`, `aria-busy="true"`, and loading class appear together. |
| explicit submit | Native `type="submit"` remains available to the caller. |
| anchor with button appearance | Native `<a href>` retains role `link`; recipe classes do not add button semantics. |
| `variant="link"` Button | Remains a button action; visual vocabulary does not change semantics. |
| server rendering | Produces ordinary deterministic HTML without a browser or Next.js. |

Do not use Base UI Button for the standard Paper Button. Native HTML already supplies the required semantics, form behavior, disabled behavior, and keyboard activation.

## Dialog

The wrapper rendered Base UI through the supported `@base-ui/react/dialog` subpath.

| Area | Observed behavior |
| --- | --- |
| trigger DOM | Native button with `aria-haspopup="dialog"`, `aria-expanded`, `aria-controls`, and `data-popup-open`. |
| popup DOM | Portalled `div[role="dialog"]`, generated `aria-labelledby` and `aria-describedby`, `tabindex="-1"`, and `data-open`. |
| backdrop/viewport | Presentation elements; backdrop is hidden from assistive technology. |
| initial focus | First tabbable form field received focus. |
| containment | Tab advanced only through popup controls; Base UI installed focus guards and made outside roots inert while modal. |
| dismissal | Escape unmounted the popup. `Dialog.Close` rendered a native `type="button"`. |
| return focus | Focus returned to the original trigger after Escape. |
| document effects | Modal behavior temporarily locks scrolling and applies inert/ARIA isolation outside the portal. |
| styling | Paper classes passed directly to parts; state is exposed with `data-open`, `data-starting-style`, and `data-ending-style`. |

Paper Modal must always include a visible `Dialog.Close` inside the popup, as Base UI requires for touch-screen-reader escape. Paper should own labels, layout, action order, transitions, and classes; Base UI should own modal isolation, focus, portal, and dismissal mechanics.

## Menu

The wrapper rendered Base UI through `@base-ui/react/menu`.

| Area | Observed behavior |
| --- | --- |
| trigger DOM | Native button with `aria-haspopup="menu"` and `aria-expanded`. |
| popup/items | Portalled `role="menu"`; actions use `role="menuitem"`. |
| open/focus | Arrow Down from the focused trigger opens the menu and focuses the first item. |
| navigation | Arrow keys move the roving focus; Home returns to the first item. |
| disabled item | Uses `aria-disabled="true"` and `data-disabled`; it remains focusable during arrow navigation, consistent with ARIA menu practice, but Enter does not activate it. Paper guidance must not promise that arrows skip disabled items. |
| activation | Enter activates an enabled item, closes the popup, and returns focus to the trigger. |
| styling/positioning | Paper classes attach to Trigger, Positioner, Popup, and Item; `data-highlighted`, `data-disabled`, and positioning CSS variables are available. |

jsdom 26 does not provide `PointerEvent`; Base UI 1.7 synthesizes one during keyboard activation, so the event path crashes in that environment. jsdom 27.0.1 fixed the test without a custom polyfill.

The keyboard trigger and complete item flow passed. A `userEvent.click()` trigger attempt did not open this Menu in jsdom 27 even though the primitive's runtime trigger is wired to mouse-down; treat this as a harness limitation to resolve with the Phase 1 real-browser pointer smoke, not as evidence that browser pointer behavior passed.

## Combobox

The wrapper rendered Base UI through `@base-ui/react/combobox` and used the collection render function so Base UI controlled the filtered item set.

| Area | Observed behavior |
| --- | --- |
| input DOM | Native `input[role="combobox"]` with label association, expanded state, controls, and active descendant. |
| list DOM | Portalled `role="listbox"`; results use `role="option"` and selected/highlighted state attributes. |
| filtering | Typing `spa` reduced Japanese/German/Spanish to Spanish when the `Combobox.List` render function consumed the filtered collection. Rendering a fixed manual `.map()` does not remove filtered items and is therefore the wrong composition. |
| keyboard/focus | Arrow Down set `aria-activedescendant`; Enter selected Spanish while DOM focus remained in the input. |
| form behavior | `name="language"` created an internal form value; `FormData` returned `Spanish` after keyboard selection. |
| styling | Input, Positioner, Popup, List, Empty, and Item accept Paper classes; state attributes include open, empty, highlighted, selected, and disabled, plus anchor/available-size CSS variables. |
| typing | Root item/value generics infer controlled values and callbacks, but low-level Item exposes a broad value type. Paper must publish its own `Option<Value>` and callback types and keep Base UI event-detail types private. |

For React Hook Form consumers, the Paper wrapper should use `useController()` and map `field.value`, `field.onChange`, `field.onBlur`, `field.name`, and `field.ref` deliberately. Do not replace React Hook Form state with Base UI Form.

## Test-harness requirements

- Use jsdom 27.0.1 or newer only after rechecking its Node engine; do not add a local PointerEvent shim while the selected compatible version supplies it.
- Call React Testing Library `cleanup()` from an explicit Vitest `afterEach`. Relying on auto-cleanup while Vitest globals are disabled allowed a dialog portal, inert markers, and body scroll styles to leak into the next test during the spike.
- Query portalled parts through `screen`, not only the render container.
- Test state through roles, names, focus, values, ARIA, and documented Base UI data attributes; do not assert generated IDs.
- Keep a real-browser smoke for pointer positioning and focus wrap because jsdom has no layout and its focus-guard behavior is not a complete browser simulation.

## Bundle and boundary findings

The externalized Paper wrapper output was 3.5 kB before CSS. A disposable Vite single chunk with React, React Router, and all three Base UI primitives was 360.20 kB raw / 119.92 kB gzip; the React/router baseline was 181.24 kB / 59.70 kB gzip. The styleguide should lazy-load component routes, and production applications should import only Paper wrappers they use.

Bundling the same selected parts from Base UI's root barrel and from its component subpaths produced identical 208,308-byte minified code with React external. Subpaths are still preferred because they make the dependency boundary explicit, avoid traversing unrelated public surface, and align Paper source with Base UI's per-component documentation.

## Blockers

There is no unresolved product decision in this lane. The only implementation prerequisite is the recorded, contained TypeScript 5 compiler boundary for the two new Paper workspaces.
