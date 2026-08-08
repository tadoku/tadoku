# ADR 002: Semantic tokens and shared CSS recipes

Status: accepted

## Context

Raw class recipes and React components are equal public APIs. They must not drift, require Tailwind, or expose primitive palette values as the application contract.

## Decision

Paper emits static, unprefixed recipe classes from one CSS implementation. Typed helpers only compose those public classes; React components call the same helpers and add semantic attributes or behavior.

`buttonClassName({ variant, loading, fullWidth, iconOnly, className })` returns `btn` plus the public variant/state classes. `Button` calls this helper, defaults to `type="button"`, and maps loading to `aria-busy="true"` plus disabled behavior. Anchors may call the helper but remain anchors; `Button variant="link"` remains a button.

CSS layers are ordered as `paper-tokens`, `paper-base`, `paper-recipes`, and `paper-utilities`. `styles.css` imports all four. `tokens.css` contains only primitives and semantic aliases. Theme aliases are selected by `:root`, `[data-theme="light"]`, and `[data-theme="dark"]`; density aliases by `[data-density="comfortable"]` and `[data-density="compact"]`.

Public semantic roles are frozen by family:

- surface: canvas, paper, raised, overlay;
- text: ink, muted, inverse, link;
- rule: subtle, default, strong, field-edge, action-edge;
- action: default, hover, active, text, destructive;
- status: information, success, warning, danger;
- focus: ring, offset;
- density: control-height, inline-gap, field-padding;
- motion: quick (120ms), standard (180ms), deliberate (240ms);
- elevation: flat, floating (3px hard offset), showcase (5px hard offset);
- typography: display, page, section, component, body, label, metadata;
- charts: categorical sequence plus pattern/dash metadata.

Raw colors are accepted only in primitive token definitions, brand SVG sources, documented color specimens, and the chart allow-list. A repository check enforces the exception list.

The optional Tailwind preset maps semantic utilities to CSS variables and never mutates or extends recipe selectors. CSS works when Tailwind is absent.

## Consequences

Class examples, typed helpers, and React output can be compared mechanically. Visual changes have one CSS source, while semantic behavior remains explicit in markup and tests.

