---
title: Paper composition
description: How Tadoku Paper splits UI into paper-ui controls, documented product patterns and application-owned screens, when to extract or share a component, and how Paper code is styled with StyleX and design tokens.
---

# Composing Tadoku Paper

Read this when you build or review a screen in a Paper application or change `paper-ui`.

Paper uses three levels of composition. They describe ownership, not mandatory directory names.

| Level | Owns | Location |
| --- | --- | --- |
| Controls | Reusable appearance, semantics, and interaction without Tadoku business state | `frontend/packages/paper-ui/src/components/` |
| Product patterns | The content order, states, and actions of a recognizable Tadoku task section | Documented in the Paper catalogue; code starts beside its consuming screen |
| Screens | Routes, data fetching, permissions, form state, mutations, and arrangement of task sections | The consuming application |

Screens may import local pattern components and `paper-ui` controls. Pattern components may import `paper-ui`. `paper-ui` must not import an application, router, API client, or product service. An application passes data, links, and callbacks into any extracted pattern component; it still owns their effects.

## Form composition

Paper screens use `usePaperForm` from `paper-ui`, built with TanStack Form's
`createFormHookContexts` and `createFormHook`. Wrap the native form with
`form.AppForm`, use `form.Form` for validated submission and error focus, and
bind each control through `form.AppField` and its `field.Input`, `field.Select`
or other control. `form.Form` ignores a submission while the previous one is
still running, so a repeated submit can't send the operation twice. The screen supplies typed default values, validators and
`onSubmit`; controls own accessible labels, hints, errors and interaction.
Validation is configured on the field or form, including required values and
numeric bounds. The native `required`, `min` and `max` props describe the control.

`AmountWithUnit` binds the amount field and takes a separate `unitField` from a
nested `AppField`. Both names stay explicit in the screen's typed values.
Autocomplete values are selected objects or arrays, and `ToggleSelect` stores
an empty array or one selected value. Disabled controls retain their values;
submission handlers own which values belong in the operation.

## When to extract and share

Keep a section in its screen while it has one consumer and a clear local shape. Extract a named component when the section repeats, has meaningful independent states, or becomes hard to understand inside the screen. Name it for the user task or domain concept, such as `ReadingLogSummary`, rather than its layout.

A catalogue pattern describes a product convention; it does not require a React export. Publish a component from `paper-ui` only after the same stable appearance and interaction contract is needed by more than one consumer and its API does not encode application data, routing, permissions, or business rules. Keep product-specific shared code outside `paper-ui` until there is an actual cross-application consumer and an agreed owner.

Put a component in its own file when its API, examples, tests, or consumers warrant it. Small related helpers can share a file. Folder structure should follow responsibility and stay easy to navigate as the application grows.

## Page structure controls

`paper-ui` provides the parts that every Paper screen repeats, so screens compose them instead of restyling headings, facts and states:

| Need | Control |
| --- | --- |
| The page title, with an optional back link, lead, status and page actions | `PageHeader`, the page's only `h1`. Pass `renderLink` to render the back link with the application router |
| A region heading with an optional description, status and actions | `SectionHeading` with an explicit `level` from 2 to 6. Pass `id` and label the section with `aria-labelledby` |
| The current state of a contest, record or service | `StatusTag`, which requires visible text and maps its tone to the status colours |
| An empty, filtered, failed or not-found state | `EmptyState` with an explicit `headingLevel`; `surface` stands alone and `plain` sits inside a container |
| Headline figures | `StatGrid` |
| Facts about one subject | `DescriptionList`, `stacked` or `inline` |
| A summary of collapsed filters with one reset action | `AppliedFilters` |

These controls are Experimental; paper-playground is their only consumer. They own semantics, spacing and type roles, not page layout: screens keep page grids, columns, breakpoints, a main column beside an aside, and page-specific typography, and pass `xstyle` for them. The Filter bar catalogue pattern shows how a screen composes a Filters toggle, the filter panel, `AppliedFilters`, a result count and `EmptyState` while owning the filter state.

`PageHeader` and `SectionHeading` wrap according to their content at every width. Status tags stay beside the title while they fit. Actions share the row while the text, status and actions fit together; otherwise actions move below the text without squeezing the title into a narrow column. Long titles still wrap within the available width.

## Styling

Paper and its applications style with [StyleX](https://stylexjs.com) and hand-write no CSS. `paper-ui` ships its StyleX source uncompiled; each application's Vite config compiles it together with the application's own styles into one atomic stylesheet.

- **One token source.** `frontend/packages/paper-ui/src/tokens/source.ts` holds every design value as typed data in three tiers: palette ramps, semantic colour roles for the light and dark themes, and scales for space, size, type, borders, motion, layers and breakpoints, plus the two densities. `pnpm --filter paper-ui tokens` generates `src/tokens/tokens.stylex.ts` (`defineVars` and `defineConsts`) and `src/tokens/themes.ts` (`createTheme` for dark and compact). A paper-ui test fails when the generated files are stale. Add a value to the source and regenerate; never edit the generated files.
- **Values come from tokens.** Import tokens from `paper-ui/tokens.stylex`. `paperStyleConfig` from `paper-ui/eslint`, enabled in every Paper package, runs StyleX's own rules and two Paper rules: `paper/tokens-only` rejects literal lengths, colours and durations (`0`, `50%`, `100%` and viewport units excepted), and `paper/no-raw-styling` rejects `className` strings and `style` props. Values computed from data, such as a bar's width, use a dynamic StyleX style.
- **Themes and density.** `rootProps.html()` and `rootProps.body()` style the document root. `applyTheme` and `applyDensity` swap the generated theme classes on `<html>`; `themeBootScript(storageKey)` does the same in an inline `<head>` script before first paint for server-rendered apps. A nested boundary takes `themes.dark` or `densities.compact` through `stylex.props`. `ThemeVariant` renders content that differs by theme, such as a reversed logo.
- **Overrides.** Paper components take an `xstyle` prop, applied after their own styles. Every compiler uses `styleResolution: 'application-order'`, so the later style wins even when it is a shorthand and the earlier one a longhand. Shared recipes such as `typography`, `surfaces`, `layout`, `linkStyles` and `buttonStyles` cover roles applications repeat.
- **One stylesheet.** `paper-ui/fonts.css` holds only Paper's `@font-face` rules, which StyleX cannot declare, and is the asset the StyleX plugin appends compiled styles to. Import it once at the application entry; `pnpm check:paper-boundaries` rejects any other stylesheet.
- **Tests.** Locate elements by role, label or a `data-*` attribute that names their meaning, never by generated class names. Test behaviour, not the styles a component applies.

To add a component, put it under `frontend/packages/paper-ui/src/components/`, style it with `stylex.create` and tokens, accept `xstyle` on its root, export it from `src/index.ts`, and add a catalogue fixture with an example in `src/catalog/examples/`.

## Documenting a pattern

For each catalogue pattern, record:

- The task it helps with, when to use it, and when to avoid it.
- Its anatomy and content order, including what each action does.
- Relevant loading, empty, error, unavailable, permission-limited, and success states. Show only states that the task can actually reach.
- Responsive, keyboard, and assistive-technology behavior.
- Which data and effects belong to the consuming screen, and whether the example is guidance or an exported component.
- The controls it composes and at least one consuming screen or journey.

Use realistic, deterministic catalogue fixtures to illustrate the contract. Verify application behavior through the real screen; a fixture alone cannot prove data flow or persistence. When reviewing a new Paper page, ask whether each section is a control, a documented pattern, or screen-owned composition, and whether its current owner matches its data and reuse needs.

The logging pattern in `frontend/packages/paper-ui/src/catalog/phase-three-content.tsx` illustrates this boundary. Its reading-log summary uses `Surface` and `ButtonGroup` to show the work, amount, privacy, and contest state. A consuming logs screen owns the log data, navigation, and save or submission actions. The catalogue example is guidance; it is not a shared logging implementation.
