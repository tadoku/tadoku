# Tadoku Paper design-system decision log

This is the permanent record of the Tadoku Paper planning discussion. The original audit remains preserved at [artifacts/design-system-refinement-audit.html](artifacts/design-system-refinement-audit.html) and [Presenter](https://presentr.lab/html/tadoku-design-system-refinement-audit).

Last updated: 2026-08-07

## Final decision snapshot

### Identity and visual grammar

- Human-facing design-system name: **Tadoku Paper**.
- Compact brand mark: the original monochrome-first Cut Meter with three rising bars and a square diagonal negative-space cut.
- Preserve the editorial book/library character with sharp geometry and no component border radius.
- Use warm paper/canvas neutrals, dark ink structure, and restrained ink violet as the brand/action accent.
- Use straight accent rails; a colored left rail replaces the neutral left hairline.
- Static surfaces use quiet 1px neutral borders. Pale form controls use a subtle 2px lower edge. Filled actions retain a stronger 2px lower edge.
- Ordinary surfaces remain flat. Floating overlays use a 3px hard-offset shadow; rare showcase surfaces may use 5px.
- Dark mode uses a lifted structural violet and deeper filled-action violet with light action text.
- Support comfortable 44px and compact 36px density modes.

### Typography and iconography

- Merriweather is reserved for display, page, and section hierarchy.
- Open Sans is used for component titles, labels, body copy, and dense interfaces, with real 400/600/700 weights.
- Font delivery must be self-hostable and framework-independent.
- Heroicons remain the initial icon family: outline for navigation/actions and solid for status/confirmation.
- Icon sizes are 16px compact, 20px default, 24px prominent/icon-only, and 48px only for empty-state illustration.

### Package and APIs

- Package name: **`paper-ui`**.
- `paper-ui` assumes React but must not depend on Next.js or Next-specific link, image, routing, or font APIs.
- Both class recipes and optional React components are first-class public APIs and share one CSS implementation.
- Standard button hierarchy is default, primary, and ghost. Danger is a composable tone; loading is a composable state.
- Every standard button state remains usable through classes without importing a component.
- Loading visuals must be paired with `aria-busy="true"` and normally native disabled behavior.

### Documentation, tests, and migration

- Replace the current style-guide `Showcase` with a complete documentation shell.
- Give each component a searchable route, complete state matrix, realistic Tadoku examples, co-located metadata/examples/tests, and lifecycle status.
- Separate primitives/components, product patterns, and experiments.
- Reuse named style-guide fixtures in Vitest, React Testing Library, `user-event`, and `jest-dom` behavior/accessibility tests.
- Introduce `paper-ui` alongside legacy `ui` at the repository level, but never mix them inside one application.
- Perform a complete migration per application in this order: style guide, admin, auth, webv2.
- Remove legacy `ui` only after the final migration and a repository-wide consumer search.

## Agreed scope

### Implement in the upcoming plan

- P0.1 — Introduce semantic design tokens.
- P0.3 — Replace `Showcase` with a complete documentation shell.
- P1.1 — Use one route per component, with search.
- P1.2 — Show the full component state matrix.
- P1.3 — Rebuild foundation pages.
- P1.4 — Co-locate examples, metadata, and tests.
- P2.1 — Separate components, patterns, and experiments.
- P2.2 — Add lifecycle and governance metadata.

### Discussed and ready for planning

1. P0.2 — Define Tadoku's visual grammar. **Resolved.**
2. P0.5 — Define the Button/class API. **Resolved enough for planning.**
3. P2.4 — Tighten the package boundary. **Resolved.**

### Deferred GitHub issues

- P0.4 — Accessibility contract: https://github.com/tadoku/tadoku/issues/761
- P1.5 — Automated component regression gates: https://github.com/tadoku/tadoku/issues/762
- P2.3 — Themes and user preferences: https://github.com/tadoku/tadoku/issues/763

## Cross-cutting constraints

- The migration must cover `webv2`, `admin`, and `auth`, not only the style guide.
- All four applications currently import the same unscoped `ui/styles/globals.css`.
- Shared UI changes build and publish four application images independently.
- Introduce `paper-ui` additively at the repository level while unmigrated applications continue using legacy `ui`.
- Do not mix legacy `ui` and `paper-ui` inside an application; each application receives one complete migration.
- Migrate in this order: style guide, admin, auth, webv2.
- Keep the legacy package until repository searches show no remaining application consumers.
- Preserve the existing audit report; do not overwrite its file or Presentr key.
- Parallel sub-agents may implement independent work after the decisions and plan are locked.

## Immediate testing baseline for the implementation plan

- Add rendered component tests using Vitest, React Testing Library, `user-event`, and `jest-dom`.
- Test new/refactored component contracts, roles, names, disabled/loading behavior, and keyboard interaction.
- Reuse named examples between the documentation renderer and tests.
- Continue using all four application builds as the compatibility matrix.
- Keep full Playwright visual regression, axe scanning, and CI baseline management in issue #762.

---

## Discussion 1 — P0.2: Tadoku's visual grammar

### Existing character to preserve

- Violet as the main brand/action color.
- Merriweather headings and Tadoku's editorial, reading-oriented identity.
- Product-specific examples instead of generic component-library demos.

### Visual preview

- Bookplate and flat-archival comparison: https://presentr.lab/html/tadoku-bookplate-direction-preview
- Archived source: [artifacts/design-system-bookplate-preview.html](artifacts/design-system-bookplate-preview.html)

### Preview feedback

- Bookplate is preferred over flat archival.
- The first Bookplate preview uses too many unrelated border colors and weights.
- Border treatment must be simplified into a consistent, deliberate grammar.
- The current bright purple needs to be evaluated against the book/library direction.
- The logo and its relationship to the refined visual grammar need exploration.

### Refinement under review

- One neutral border family rather than component-specific border colors.
- Border roles limited to structural, component, interactive weight, and floating depth.
- Status color appears as a small marker, never as a competing full outline.
- Purple is treated like printed annotation ink and used sparingly.
- Preserve the existing geometric wordmark unless a lockup study shows a clear improvement.

### Refinement preview v2

- Refined borders, accent comparison, and logo lockups: https://presentr.lab/html/tadoku-bookplate-refinement-v2
- Archived source: [artifacts/design-system-bookplate-refinement-v2.html](artifacts/design-system-bookplate-refinement-v2.html)
- Interactive accent choices: current violet, ink violet, and library plum.
- Working recommendation in the preview: ink violet and the existing unboxed wordmark, with a catalog lockup as a secondary composition.

### Refinement v2 feedback and decisions

- **Decided:** use ink violet for the light theme; it retains Tadoku's purple brand while fitting the printed Bookplate direction.
- When a component has a colored left accent, that accent replaces the neutral left hairline. Do not draw both.
- Soften the interactive lower edge: keep a 1px border on every side and use only a darker 1px bottom color, not extra bottom thickness.
- The v2 logo preview is invalid: the reusable SVG lost the original even-odd fill behavior and the accent fill did not reliably carry through the SVG `<use>` instance.
- Correct subsequent logo studies before making any logo decision.
- Dark mode requires a tonal mapping of the same ink-violet hue, not reuse of the light-theme hex.

### Light/dark refinement preview v3

- Corrected light/dark study: https://presentr.lab/html/tadoku-bookplate-dark-preview-v3
- Archived source: [artifacts/design-system-bookplate-dark-preview-v3.html](artifacts/design-system-bookplate-dark-preview-v3.html)
- Light accent: ink violet `#5747B8`.
- Proposed dark counterpart: lifted ink violet `#9A8BEA` with dark foreground on filled actions.
- Accent-edge rule and softened 1px interactive lower color are demonstrated in the product specimen.
- Logo preview restores `fill-rule="evenodd"` and applies the accent wedge independently.

### Refinement v3 feedback

- The thick accent border creates triangular/mitered joins where it meets the 1px top and bottom borders.
- **Direction:** replace the border join with a positioned accent rail that overlays the left ends of the neutral borders, producing one uninterrupted straight edge without extra wrapper nesting.
- Try a 2px lower border on form controls and interactive components; the 1px color-only treatment is too subtle.
- Explore five distinct compact logo-mark concepts that represent Tadoku rather than extracting the K wedge from the existing wordmark.

### Refinement v4 preview

- Presenter: `https://presentr.lab/html/tadoku-bookplate-border-logo-v4`
- Archived source: [artifacts/design-system-bookplate-border-logo-v4.html](artifacts/design-system-bookplate-border-logo-v4.html)
- The accent is now a positioned rectangular rail that overlays the top and bottom hairline endpoints; it has no mitered joins and requires no additional wrapper.
- Interactive controls use 1px top/side rules with a 2px lower rule. Static containers retain a consistent 1px hairline.
- Five schematic logo territories are included: Open Book T, Bookmark T, Library Shelf, Reading Loop, and Catalog Card.
- Initial shortlist for discussion: Bookmark T for compact clarity; Reading Loop for distinctiveness.

### Refinement v4 feedback

- Keep the 2px lower edge, but make it lighter and more subtle on pale input fields.
- Filled submit buttons can retain their stronger lower edge.
- Add a visible hover state to primary submit buttons.
- Reject all letter-driven logo concepts; the forced T does not feel natural and none of the first set reads correctly.
- Explore abstract marks without hidden initials or elaborate letter construction.

### Refinement v5 preview

- Presenter: `https://presentr.lab/html/tadoku-bookplate-controls-logo-v5`
- Archived source: [artifacts/design-system-bookplate-controls-logo-v5.html](artifacts/design-system-bookplate-controls-logo-v5.html)
- Pale fields now use a separate soft-neutral 2px lower-edge token; filled primary actions retain the deeper ink-violet edge.
- Primary actions transition to a lighter ink-violet fill on hover while retaining a stable lower edge.
- The logo study has been reset with five non-letter territories: Page Turn, Open Passage, Accumulation, Reading Current, and Marginal Rhythm.
- Logo guardrails: no hidden initials, one meaningful violet gesture, and a legible two-color silhouette at 16px.

### Refinement v5 feedback

- The quieter 2px lower edge on inputs is approved.
- In dark mode, try light text on primary buttons.
- Reject the second logo set; the abstract forms still read as generic or unusable interface icons.
- New identity territories should engage more directly with listening, language learning, contests, friendly competition, and media consumption.

### Refinement v6 preview

- Presenter: `https://presentr.lab/html/tadoku-bookplate-brand-v6`
- Archived source: [artifacts/design-system-bookplate-brand-v6.html](artifacts/design-system-bookplate-brand-v6.html)
- Dark primary actions now use light text on a deeper action-specific violet; default and hover contrast are approximately 6.15:1 and 5.19:1.
- Bright dark-mode ink violet remains reserved for rails, focus rings, and small structural accents.
- Five new territories: Immersion Field, Living Signal, Language Weave, Friendly Pace, and Media Feed.
- Every mark is shown with a neutral Tadoku wordmark and at 28px/16px.
- Brand recommendation: keep the master mark centered on immersion/language; treat competition as social energy and potentially give contests a separate badge rather than making the master identity trophy-like.

### Refinement v6 feedback

- Add the hard-offset floating treatment from the Bookplate v2 study to the design language/system.
- Language Weave is visually strongest in the third set, but still too busy and not viable as-is.
- Use the Bookmeter logo as conceptual inspiration: a very simple growing bar chart with immediate recognition.
- The Tadoku compact mark must stand out with few shapes and remain recognizable with no color.
- Color is an optional accent only; it must not create or rescue the mark.

### Refinement v7 preview

- Presenter: `https://presentr.lab/html/tadoku-bookplate-meter-v7`
- Archived source: [artifacts/design-system-bookplate-meter-v7.html](artifacts/design-system-bookplate-meter-v7.html)
- **Approved direction:** add hard-edged elevation as named roles: `floating` uses a strong 1px border plus a 3px hard offset; rare `showcase` surfaces may use a 5px hard offset. Ordinary cards remain flat.
- Logo exploration is narrowed to one meter family rather than five unrelated metaphors.
- Five monochrome-first variations: Three Beat, Page Meter, Cut Meter, Signal Rise, and Stacked Measure.
- Cut Meter reduces the directional movement of Language Weave to one negative-space diagonal across growing bars.
- Every mark is shown in solid monochrome, reversed, optionally accented, and beside a wordmark.
- Non-negotiable order: black first, reverse second, accent third.
- Reference principle from Bookmeter: few shapes, one baseline, immediate growth rhythm. Do not copy its exact three-bar proportions or green treatment.

### Refinement v7 feedback

- Cut Meter is the first compact-mark direction approved as a solid foundation.
- Explore variations without changing the three-bar silhouette.
- Vary the negative-space cut shape, including a slightly rounded line.
- Begin surfacing the remaining design-system questions while the mark is refined.

### Refinement v8 preview

- Presenter: `https://presentr.lab/html/tadoku-cut-meter-variations-v8`
- Archived source: [artifacts/design-system-cut-meter-variations-v8.html](artifacts/design-system-cut-meter-variations-v8.html)
- Six Cut Meter variations: square diagonal, soft diagonal, tapered passage, gentle wave, stepped cut, and separate notches.
- Suggested refinement for initial review: soft diagonal. Only the cut endpoints are rounded; the bars and interface geometry remain square.
- All variations use identical bar proportions and are tested at 28px, 16px, reversed, and in a wordmark lockup.
- The study now surfaces the remaining question queue: density, typography/icon grammar, typed Button APIs, package boundary, application rollout order, and the immediate regression-test baseline.
- Suggested discussion sequence: finish the Cut Meter choice, decide density to close P0.2, then discuss P0.5 Button APIs and P2.4 package boundaries before creating the implementation plan.

### Refinement v8 feedback

- The soft diagonal does not appear materially softer than the original.
- Explore a few genuinely rounded versions of the negative-space cut.

### Refinement v9 preview

- Presenter: `https://presentr.lab/html/tadoku-cut-meter-soft-v9`
- Archived source: [artifacts/design-system-cut-meter-soft-v9.html](artifacts/design-system-cut-meter-soft-v9.html)
- Construction correction: the previous line's rounded endpoints sat outside the bars and were clipped away, so the visible intersections remained sharp.
- Each new visible cut segment has its own rounded geometry inside the bar.
- Four softness levels: restrained capsules, full capsules, bowed capsules, and circular passage.
- All four retain the fixed three-bar silhouette and are tested at 28px, 16px, reversed, and in a wordmark lockup.

### Decisions after v9

- **Density approved:** use two named modes—comfortable 44px controls and compact 36px controls.
- **Compact mark selected for now:** use the original Cut Meter with the square diagonal cut; do not continue with the rounded-cut variants.
- Discuss typography and icon grammar next.
- A mandatory imported Button component is not assumed. The existing class/Tailwind-style convenience is important and needs to remain available.
- UI v2 must move away from Next.js-specific dependencies because the applications may move away from Next.js.
- Create UI v2 additively so legacy `ui` remains usable throughout migration.
- Confirmed application rollout order: style guide, admin, auth, then webv2.
- Explain and discuss the immediate regression baseline before locking the implementation plan.

### Typography, icon, API, and regression preview v10

- Presenter: `https://presentr.lab/html/tadoku-type-icon-baseline-v10`
- Archived source: [artifacts/design-system-type-icon-baseline-v10.html](artifacts/design-system-type-icon-baseline-v10.html)
- Typography proposal: Merriweather for display/page/section hierarchy; Open Sans for component titles, labels, body copy, and dense interfaces.
- Proposed text roles: display 48/52, page 32/38, section 24/30, component 18/24, comfortable body 16/26, compact body 14/22, label 12/16, metadata 12/18.
- Fix the current synthetic-bold risk: Open Sans 700 is used by classes but is not currently loaded.
- Framework-independent font direction: optional self-hosted WOFF2/CSS assets; no `next/font` and no runtime Google Fonts import in the implementation.
- Icon proposal: keep Heroicons initially; outline icons for navigation/actions, solid icons for status/confirmation; 16px compact, 20px default, 24px prominent/icon-only, 48px empty-state exception.
- Rounded icon stroke caps are allowed for recognition even though component containers stay square.
- Button API recommendation: class recipes are the cross-framework primitive (`btn primary`); an optional thin React adapter adds behavior but is not required for styling.
- Regression baseline clarified: named fixtures are shared by style-guide documentation and Vitest/Testing Library behavior/accessibility tests. It does not freeze old visuals or require screenshots. The later visual-regression issue can reuse the same fixtures.
- Proposed first fixture groups: buttons/links, form controls, overlays, and navigation.

### Decisions after v10

- **Typography approved:** Merriweather for display/page/section hierarchy; Open Sans for component titles, labels, body text, and dense interfaces. Use the proposed named scale and load real 400/600/700 sans weights.
- **Icon grammar approved:** outline for navigation/actions, solid for status/confirmation; 16px compact, 20px default, 24px prominent/icon-only, and 48px only for empty-state illustration.
- Rounded icon stroke endings are allowed for legibility while component geometry remains sharp.
- **Button direction:** keep class recipes as the primary cross-framework API. Loading may be expressed with `btn primary loading` and does not require a React component.
- Loading styling must be paired with semantic/behavioral state: `aria-busy="true"` and normally `disabled` on a native button. Prefer the semantic attribute as the CSS source of truth, with `.loading` available as a convenience alias if desired.
- Optional framework adapters remain permissible but are not required to obtain any standard visual state.
- **Regression baseline approved:** reuse named style-guide fixtures in behavior/accessibility tests first; add screenshot comparisons through the deferred visual-regression issue.
- **Button architecture approved:** UI v2 supports both the class API and optional Button components.
- The class API remains a first-class, documented, and tested public contract—not a legacy fallback.
- Button components must compose the same public classes and semantic attributes; they must not own a separate private styling implementation.
- Every standard state and variant, including loading, must remain usable without importing a framework component.

### Agreed direction

- No border radius. Sharp edges are an intentional Tadoku design principle.
- The square geometry supports the book, library, and slightly old-school character of the product.
- Refinement should come from proportion, typography, color, borders, spacing, and interaction states—not rounded corners.

### Remaining recommended direction

- Border-led surfaces: subtle 1px borders as the default separator.
- No shadow on ordinary inline surfaces; restrained elevation for raised cards and popovers; stronger elevation only for overlays.
- Test a restrained 2px lower rule on interactive controls only; static surfaces remain 1px.
- Two named densities:
  - Comfortable: 44px controls for public and authentication experiences.
  - Compact: 36px controls for admin and dense data interfaces.
- Icon scale: 16px, 20px, and 24px.
- One consistent 2px `:focus-visible` ring with a 2px offset.
- Motion durations: 120ms, 180ms, and 240ms, with reduced-motion support deferred to issue #763.

### Safe migration model

1. Create the new React-based, Next-independent `paper-ui` package with Tadoku Paper tokens, CSS recipes, components, assets, and tests.
2. Rebuild and completely migrate the style guide so it becomes the authoritative Paper catalogue and fixture host.
3. Completely migrate admin, using compact density as the primary stress test.
4. Completely migrate auth, using comfortable density and form flows as the primary stress test.
5. Completely migrate webv2 after Paper has been proven in the other applications.
6. Keep legacy `ui` available only for applications that have not yet reached their cutover; remove it after webv2 migration and a repository-wide consumer search.

### Decision required

- None blocking the implementation plan. Treat the original square-diagonal Cut Meter as the selected compact mark.

### Decision

- **Decided:** Tadoku components and surfaces retain sharp edges with no border radius.
- **Rationale:** square geometry fits Tadoku's book/library identity and slightly old-school visual character.
- **Decided:** ordinary surfaces are border-led and flat; floating/showcase roles use named hard-offset elevation.
- **Decided:** comfortable 44px and compact 36px density modes.
- **Decided:** the proposed typography hierarchy and icon grammar.

---

## Discussion 2 — P0.5: Typed Button and LinkButton APIs

Status: resolved enough for implementation planning.

Questions to resolve later:

- **Decided:** imported React components are not mandatory.
- **Decided:** preserve a framework-independent class recipe as a first-class API so buttons and links can share styling without imports.
- **Decided:** loading is a supported class/attribute state and does not require a component; pair its visual state with `aria-busy` and native disabled behavior.
- **Decided:** provide both class-based buttons and optional Button components.
- **Decided:** optional components compose the same public classes and semantic attributes so the two APIs cannot visually drift.
- **Decided:** name the unmodified hierarchy `default`.
- **Working default:** model destructive styling as a composable danger tone across default, primary, and ghost hierarchy. This is not plan-blocking.
- If adapters remain, prefer separate `Button` and link adapters over a Next-specific polymorphic API.

### Button and package preview v11

- Presenter: `https://presentr.lab/html/tadoku-paper-button-grammar-v11`
- Archived source: [artifacts/tadoku-paper-button-grammar-v11.html](artifacts/tadoku-paper-button-grammar-v11.html)
- User selected `default` as the base hierarchy name.
- Proposal: ordinary hierarchy is limited to default, primary, and ghost; omit a separate secondary name because default already fills that role.
- “Close” is an action, not a visual variant: footer Close/Cancel is normally default; a top-right × is normally a ghost icon-only button.
- Proposal: danger is a composable tone, allowing default + danger, primary + danger, and ghost + danger.
- Loading composes with any hierarchy through both classes and semantic attributes.
- Class and React examples are shown side-by-side and share one CSS implementation.

---

## Discussion 3 — P2.4: Package boundary

Status: resolved.

Questions to resolve later:

- **Direction supplied:** UI v2 must be framework-independent and must not depend on Next.js.
- Decide whether the additive package is named `ui-v2`, split into core/adapters, or exposed through versioned entry points.
- Establish canonical styles, tokens, Tailwind-preset, assets, and optional framework-adapter entry points.
- Retain the compatibility global stylesheet until all application migrations are complete.

### Supplied architecture decisions

- Human-facing design-system name: **Tadoku Paper**.
- New package name: **`paper-ui`**.
- `paper-ui` may assume React, but must not depend on Next.js or Next-specific components and font systems.
- One React package is sufficient; a framework-core/React package split is unnecessary.
- Old `ui` and new `paper-ui` may coexist in the repository during rollout, but an individual application does not mix them.
- Each application receives one complete/big-bang migration in this order: style guide, admin, auth, webv2.
- Therefore selector-level v1/v2 coexistence inside the same application is not a design requirement.

---

## Plan status

Ready to create the parallel implementation plan. No blocking design or architecture discussion remains.
