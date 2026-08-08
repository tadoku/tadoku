# ADR: Native HTML versus Base UI ownership

Status: Accepted for Phase 1  
Date: 2026-08-08

## Context

Paper should prefer native semantics while using Base UI where a correct interaction requires composite focus, modal isolation, portals, positioning, typeahead, or active-descendant behavior. Using Base UI everywhere would add indirection without improving standard HTML; avoiding it for complex widgets would recreate difficult accessibility behavior.

## Decision

Choose the lowest behavior layer that completely satisfies the component contract.

| Paper category | Owner | Boundary |
| --- | --- | --- |
| Button, icon button, submit/reset actions | Native HTML | Paper recipe plus thin native `<button>` adapter; default `type="button"`. |
| Link and link with button appearance | Native HTML / consumer router | Anchor or injected router link receives Paper recipe classes; never gains button semantics. |
| Text input, textarea, file input | Native HTML | Label, hint, error, and validation associations remain Paper/RHF concerns. |
| Checkbox and radio group | Native HTML first | Native inputs retain form and keyboard behavior; Paper may visually wrap them without replacing the control. |
| Simple select | Native HTML | Use when platform selection, styling, and option requirements are sufficient. |
| Fieldset, label, output, progress, meter | Native HTML | Preserve their built-in semantics; use Paper recipes. |
| Tables, lists, headings, navigation landmarks, breadcrumbs | Native HTML | Paper supplies structure guidance and styling; current state uses appropriate ARIA such as `aria-current`. |
| Dialog and AlertDialog | Base UI | Paper Modal/confirmation wrapper owns content/actions; Base UI owns modal isolation, portal, focus, and dismissal. |
| Menu, context menu, menubar | Base UI | Paper ActionMenu wrapper owns actions/destructive treatment; Base UI owns menu roles and composite focus. |
| Popover, preview card, tooltip | Base UI | Base UI owns floating positioning, dismissal, hover/focus coordination. |
| Combobox, autocomplete, multi-autocomplete, tags/chips autocomplete | Base UI | Paper/RHF owns values/options/errors; Base UI owns filtering, active descendant, selection, and popup. |
| Custom Select/listbox | Base UI | Use only when native Select cannot meet the documented product requirement. |
| Tabs, accordion, collapsible, toggle group, slider | Base UI | Composite keyboard/state behavior remains behind Paper wrappers. |
| Toast | Base UI candidate | Use its announcement/queue behavior behind the Paper notification API; validate with the notification tranche before marking Stable. |

Base UI Trigger parts may render native buttons inside a complex wrapper. That does not make Base UI the owner of the general Paper Button API; the trigger wrapper reuses Paper classes and lets Base UI compose the event behavior it needs.

React Hook Form remains the form-state contract. Native controls use registration through Paper form components. Base UI-backed custom controls use `useController()` and map only the necessary value, change, blur, name, disabled, and ref behavior. Do not introduce Base UI Form as an application state layer.

## Selection test for future components

Use native HTML unless at least one documented requirement needs behavior the native element cannot provide without building a composite widget. Base UI is warranted when the component needs one or more of:

- focus containment or reliable focus return;
- portal dismissal and collision-aware positioning;
- roving tabindex, typeahead, or directional composite navigation;
- `aria-activedescendant` with filtered/virtual collections;
- coordinated hover, focus, touch, and Escape behavior;
- selection behavior beyond a native control's supported presentation.

Visual customization alone is not sufficient reason to replace native semantics.

## Consequences

- Simple controls remain small, server-renderable, form-native, and framework-neutral.
- Complex controls inherit a tested interaction engine without exposing that engine as Paper's contract.
- Component tests must reflect actual primitive behavior. In particular, ARIA-disabled menu items remain focusable during arrow navigation but cannot activate.
- Product requirements for a custom Select must be explicit; Paper should not automatically replace every native select.
