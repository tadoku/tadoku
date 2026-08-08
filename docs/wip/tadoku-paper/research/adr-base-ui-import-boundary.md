# ADR: Base UI imports and Paper wrapper boundary

Status: Accepted for Phase 1  
Date: 2026-08-08

## Context

Paper uses Base UI for difficult interaction behavior but must retain a Paper-owned API and visual identity. Base UI was formerly published under `@base-ui-components/react`; current package documentation declares that name obsolete and directs consumers to `@base-ui/react`.

The package exposes both a root barrel and supported per-component subpaths. A bundle comparison of selected Dialog, Menu, and Combobox parts produced the same minified output from both paths, so this is an API-boundary decision rather than a code-size workaround.

## Decision

Paper implementation files import namespaces from supported component subpaths:

```ts
import { Combobox } from '@base-ui/react/combobox'
import { Dialog } from '@base-ui/react/dialog'
import { Menu } from '@base-ui/react/menu'
```

Do not use the obsolete package, internal Base UI paths, or `@base-ui/react`'s root barrel in Paper component implementations.

Wrap every Base UI primitive behind a Paper component. Paper owns:

- public component and option/value types;
- semantic props and defaults;
- React Hook Form adaptation;
- DOM labels/content requirements;
- variants, recipes, CSS classes, data-attribute selectors, tokens, and transitions;
- component composition and migration contract.

Base UI owns:

- ARIA wiring and composite-widget semantics;
- portal and positioning mechanics;
- modal isolation, focus containment, focus return, and dismissal;
- roving focus, typeahead, active descendant, and collection filtering behavior.

Do not re-export Base UI namespaces, components, event-detail types, state types, handles, or render-prop types. Paper callback types contain only product-relevant values. If a Paper escape hatch is later proven necessary, decide and test it separately rather than exposing arbitrary Base UI props.

Keep Base UI as a direct `paper-ui` dependency and externalize it from Paper's compiled JS. Public bundled declarations must contain no Base UI import or name. Application code imports only `paper-ui` and its declared subpaths.

## Consequences

- Base UI can be upgraded or replaced behind Paper wrappers without rewriting applications.
- Component source and reviews show exactly which primitive family is used.
- Tree-shaking remains effective; the spike measured identical selected code for root and subpath imports.
- Wrapper authors must map behavior intentionally instead of spreading all primitive props.
- Package gates must search built declarations/exports for Base UI leakage and source for obsolete/internal imports.
