# Legacy styleguide examples and route migration matrix

## Route policy

Paper canonical routes follow the implementation plan’s information architecture. During coexistence, `ui.tadoku.app` remains unchanged. The redirect manifest is activated only when the verified Paper catalogue replaces it. A legacy category page that covered several components redirects to the closest Paper category/index; individual canonical pages remain discoverable from that destination.

## Legacy route matrix

| Legacy route | Legacy content/examples | Canonical Paper destination | Content action |
| --- | --- | --- | --- |
| `/` | installation and root imports | `/` | rewrite for `paper-ui`, link principles/getting started; redirect not needed |
| `/color` | primary/secondary and Tailwind swatches | `/foundations/color` | rebuild from semantic tokens; do not preserve raw legacy palette as guidance |
| `/typography` | Title, Subtitle, Text Link | `/foundations/typography` | preserve real Tadoku prose; document named type roles and link recipe |
| `/branding` | light/dark legacy wordmarks | `/foundations/brand` | replace with canonical Cut Meter/wordmark/favicons; keep legacy note in migration section |
| `/templates` | vertical stack, horizontal stack, card | `/foundations/layout` | preserve stack/layout guidance; link Paper surface/card components rather than one mixed “Templates” page |
| `/forms` | log-pages RHF form; autocomplete; tags; blog form; misc controls; raw HTML elements | `/components/forms` | category redirect; split into canonical Input, TextArea, Checkbox, Select, RadioGroup, Autocomplete, TagsInput, AmountWithUnit pages; convert examples to deterministic fixtures |
| `/buttons` | variants, disabled, anchors, icons, ButtonGroup | `/components/actions/button` | canonical Button page; move ButtonGroup to action-group Pattern if retained; preserve class/component parity and semantic anchor guidance |
| `/navigation` | Navbar, Tabbar, Sidebar | `/components/navigation` | category redirect; split canonical Navbar, Tabs, Sidebar pages |
| `/toasts` | react-toastify triggers/setup | `/components/feedback/toast` | rebuild around Paper-owned public contract; do not carry third-party setup as primary guidance |
| `/flash` | standard/link/icon info/success/warning/error | `/components/feedback/flash` | preserve realistic variants, clarify action/link semantics |
| `/charts` | Heatmap, reading activity Chart.js, doughnut | `/components/data-display/charts` | category/pattern index; canonical Heatmap plus chart palette guidance and chart patterns |
| `/modals` | standard destructive modal | `/components/overlays/dialog` | migrate to Dialog; preserve destructive fixture and focus/dismissal contract |
| `/tables` | basic table and clickable rows with ActionMenu | `/components/data-display/table` | preserve overflow/clickable/action specimens; separate menu behavior fixture |
| `/breadcrumb` | one breadcrumb example | `/components/navigation/breadcrumb` | rebuild with injected link fixture and mobile collapse behavior |
| `/action-menu` | one menu example | `/components/actions/action-menu` | preserve orientation/danger/focus states; add keyboard matrix |
| `/pagination` | one pagination example | `/components/navigation/pagination` | rebuild with `getHref`/`onPageChange`, current/edge states |
| `/logging` | hand-built logs overview interaction | `/patterns/logging` | product Pattern; preserve content/data lessons, remove ad hoc local state from component docs |
| `/logging-v2` | new log form, submit-to-contest, log details | `/experiments/logging-v2` | remain explicitly Experimental; deterministic fixtures; do not consolidate product implementations |

Suggested redirect behavior is a `308` or static-server equivalent for exact legacy paths. Query and fragment should be retained where the serving layer supports it. The final redirect manifest must include all 18 routes above and direct-route fallback must serve every destination.

## Example inventory and fixture disposition

| Area | Files | Reuse decision |
| --- | ---: | --- |
| action menu | 1 | convert to named normal/danger/orientation fixtures |
| branding | 2 | replace visuals; retain only migration/history context |
| breadcrumb | 1 | convert content to injected-link fixture |
| buttons | 6 | high-value state matrix; rename legacy variants semantically and add loading |
| charts | 3 | retain deterministic data shapes; replace color values with semantic palette |
| flash | 3 | combine as stable status/link/icon fixtures |
| forms | 6 | retain RHF and Tadoku content; split by canonical control and eliminate Faker/network/current-time behavior |
| logging-v2 | 3 | keep under Experiments with deterministic domain fixtures |
| modals | 1 | retain destructive confirmation scenario |
| navigation | 3 | split into Navbar/Tabs/Sidebar fixtures; remove Next router assumption |
| pagination | 1 | retain page-state content with injected routing |
| tables | 2 | retain basic/action-menu/overflow cases |
| templates | 3 | migrate to layout/surface foundation examples |
| toasts | 1 | retain notification intent, replace third-party-specific setup |
| typography | 3 | retain hierarchy/link specimens with new type roles |

There are 39 example files. The `/logging` route is a large inline interaction and has no file under `examples/`; it must be inventoried separately as a Pattern source. `Loading` and `VerticalTabbar` are public components with no legacy styleguide example, while deep public types are undocumented.

## Content gaps revealed by migration

- Legacy pages have no lifecycle, owner, review date, source/version, accessibility contract, API table, migration notes, long-copy/overflow, density, theme, or viewport matrix.
- The source-view toggle has no stable fixture identity; examples are default exports plus `?raw` duplicates.
- Several pages are categories that mix primitives, recipes, patterns, and app flows.
- There is no search and no one-route-per-component guarantee.
- Faker is a styleguide dependency; every Paper fixture must be stable and must not use Faker randomness.

## Verification commands

```sh
rg --files frontend/apps/styleguide/pages frontend/apps/styleguide/examples | sort
rg -n '<Showcase|<h1|href:' frontend/apps/styleguide/pages
rg -n '^import .*@examples|\?raw' frontend/apps/styleguide/pages
```

