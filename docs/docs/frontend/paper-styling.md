---
title: Paper styling
description: How Paper and its applications are styled with Tailwind CSS from one typed token source, how themes, density and breakpoints switch, and which checks keep styles on tokens.
---

# Styling with Paper tokens

Read this when you style anything in `paper-ui`, webv3, authv2, paper-playground or paper-styleguide, or add a design token.

Paper applications have no handwritten CSS. Every value comes from the typed tokens in `frontend/packages/paper-ui/src/tokens`, which Paper's build turns into the Tailwind CSS v4 theme, and components and screens use the utilities that theme defines.

## Token tiers

| Tier | Where | What it holds |
| --- | --- | --- |
| Primitives | `frontend/packages/paper-ui/src/tokens/palette.ts` and the scales in `theme.ts` | Colour ramps and the spacing, type, line-height, radius, border, shadow, motion, z-index and breakpoint steps. Screens never read the palette directly |
| Semantic | `theme.ts` | Surface, text, rule, action, focus, status, chart and heat colours, each with a light and a dark value, and the type roles |
| Component | `theme.ts` and `effects.ts` | Knobs that one control needs and the semantic set does not name, such as the footer band colours or the table scroll shade |
| Layout | `layout.ts` | Page widths, gutters, grids and brand sizes that the applications use |

Each key is a Tailwind theme namespace, so `color.paper` becomes `bg-paper`, `spacing['field-x']` becomes `px-field-x` and `container.measure` becomes `max-w-measure`. Spacing steps follow Tailwind's quarter-rem names: `p-3` is 12px. A value can be a string, `{ light, dark }`, `{ comfortable, compact }` or a breakpoint map such as `{ base, tablet, sm }`, and may read other tokens with `var(--…)`.

`frontend/packages/paper-ui/scripts/build-assets.mjs` writes `dist/theme.css` from `themeCss()` in `frontend/packages/paper-ui/src/tokens/css.ts`. That file registers the tokens with `@theme reference` after `--*: initial`, so only Paper's utilities exist; declares every token on `:root`; re-declares what a `data-theme`, `data-density` or breakpoint changes, including tokens that read those, inside each scope; registers Paper's built components as a Tailwind source; declares the `dark:` and `compact:` variants; and loads the bundled fonts. The generator fails the build when a token reads an unknown token or a container or height key would be hidden by a spacing key of the same name.

## Set up an application

Each Paper application has exactly one stylesheet, `src/styles.css`, and adds `@tailwindcss/vite` to its Vite plugins:

```css
@import "paper-ui/theme.css";
@tailwind utilities;
@source "./";
```

`@tailwind utilities` keeps the utilities in the same file as the theme; a separate `@import` of Tailwind's utilities stylesheet is compiled on its own by the TanStack Start development server and loses the theme. Tailwind's Preflight reset is left out on purpose: Paper and the CMS content keep the browser's element defaults. Instead the root element takes `documentClassName` and the body `bodyClassName` from `paper-ui`; their zero-specificity utilities set the page surface, type, box sizing, link hover and focus ring, so any component utility overrides them. Set `data-theme` and `data-density` on `<html>`. An application that renders the catalogue adds `@source` for `paper-ui/dist/catalog.js`.

## Write styles

- Use utilities in `className`. Keep complete class names in source; Tailwind cannot see `` `p-${size}` ``.
- Component recipes are `tv()` calls from `paper-ui`'s `tv`, configured with Paper's token names so `cn()` and a component's `className` replace conflicting utilities instead of competing with them. A later `text-*` size also drops an earlier `leading-*`, as Tailwind's font-size utilities set line height.
- Components with several parts take `classNames` keyed by part, such as `PageHeader classNames={{ actions: 'gap-2' }}`.
- Shared compositions are exported as recipes and components: `stack()`, `cluster()`, `measure()`, `surface()`, `accentRail()`, `typography({ role })`, `appFrame()`, `prose()`, `entryList()`, `listRows()`, `entryLink()`, `metaLine()`, `listFooter()`, `formFooter()`, `IconTile`, `SkipLink`, `SiteFooter` and `Navbar contained`. Promote a pattern to Paper when a second application needs it.
- Arbitrary variants such as `[&_a]:underline` or `has-[>input:focus-visible]:outline-halo` are fine; they select elements. Inside one, write `.class` names without underscores, because Tailwind reads `_` as a space.
- Colours switch with `data-theme`, so `bg-paper` needs no `dark:` twin. Use `dark:` only for content that differs per theme, such as a reversed logo.
- Add a token when a value is needed and none fits. Name it for its role, put it in the right tier and reuse it rather than adding a near duplicate.
- Elements keep their `paper-*` part names, such as `paper-button--outline`, as stable hooks for tests and tooling. Never style them from CSS.

## Guard rails

| Check | Rejects |
| --- | --- |
| The `--*: initial` theme reset | Tailwind's default scale: `bg-slate-500` or `max-w-2xl` compile to nothing |
| `paper/no-arbitrary-values` in each Paper package's ESLint config | Arbitrary values such as `p-[13px]` or `bg-[#fff]`. The only arbitrary values allowed read a Paper token or a Base UI runtime property, such as `w-(--anchor-width)` or `[background:var(--effect-table-scroll)]`; `frontend/packages/paper-ui/src/foundations/document.ts` is exempt because it styles elements Paper does not render |
| `paper/no-literal-styles` | Literal values in a JSX `style` object; computed values such as a measured width stay allowed |
| `pnpm check:paper-boundaries` | Any `.css` file in `paper-ui` outside `dist`, a second stylesheet in a Paper application, or an entry stylesheet with anything other than `@import`, `@tailwind`, `@source`, `@layer`, `@custom-variant` and `@font-face`, or one that does not import `paper-ui/theme.css` |

Utilities with bare numeric values, such as `z-10` or `duration-150`, still compile without a token. Review catches those.
