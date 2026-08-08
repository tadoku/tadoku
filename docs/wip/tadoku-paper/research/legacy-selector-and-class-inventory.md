# Legacy global selector and raw-class inventory

Source of truth: `frontend/packages/ui/styles/globals.css` (461 lines at the starting commit).

## External and token-like side effects

- Runtime Google Fonts import: Merriweather 700 and Open Sans 400/400i/500/600.
- Tailwind base, components, and utilities are emitted together.
- `:root` mutates react-toastify color variables.
- `::selection` sets global foreground/background.
- The shared Tailwind preset adds `primary` (`#6969FF`), `secondary` (`#2a282c`), font families, a `fill-48` grid, loading animation/keyframes, and a fixed safelist.

These are implicit application APIs even when no matching class string appears in app source.

## Broad/global selectors

| Selector family | Behavior to replace explicitly |
| --- | --- |
| `html`, `body`, and separate `html`/`body` rules | full-height, horizontal overflow/margin compensation, neutral canvas, Open Sans, reset margin/padding |
| `a:not(.reset):not(.flash):not(.btn)` and hover/active/focus variants | global link transition/color behavior with three escape classes |
| `p`, `p a:not(...)`, `p > code` | paragraph margins, underlined prose links, inline-code styling |
| `h1`–`h7` | Merriweather assignment (includes nonstandard `h7`) |
| `button` | Open Sans assignment |
| `a[href]`, enabled submit/image inputs, labels, selects, buttons | global pointer cursor |
| `table`, `table thead tr`, `table.zebra tbody tr:nth-child(2n+1)` | surface/shadow, header rule, zebra rows |
| text-like `input[type=…]` list | `.input-frame`, 44px sizing, full width, padding |
| range/file inputs | appearance, dimensions, focus, file-button styling |
| `textarea`, `select` | global frame/dimensions |
| checkbox/radio selectors and checked adjacent `span` | control appearance, focus, checked color/font coupling, DOM-order behavior |
| `fieldset:disabled` | opacity and pointer-event suppression |

Paper must not recreate broad form/table/heading/link selectors. Each legacy dependency needs a named recipe/component or deliberately documented small base rule.

## Declared class recipes and selector combinations

| Family | Complete legacy selectors/classes |
| --- | --- |
| typography | `.text-link`, `.title`, `.subtitle` |
| stacks | `.v-stack`, `.v-stack > *`, `.h-stack`, `.h-stack > *`, `.h-stack > .btn`, `.v-stack.spaced`, `.h-stack.spaced`, `.h-stack.fill > *` |
| form labels/errors | `.label`, `.label-text`, `.label .label-inline`, `.label .label-inline input[type=checkbox]`, `.label-inline input[type=radio]`, `.label .label-inline .label-text`, `.label-hint`, `.label.error input`, `.label.error textarea`, `.label .error`, `.label.error .error` |
| buttons/links | `.btn`, `.btn > svg`, `.btn.small > svg`, `a.btn`, `.btn:disabled`, `.btn.small`, `.btn.primary`, `.btn.primary:disabled`, `.btn.secondary`, `.btn.secondary:disabled`, `.btn.danger`, `.btn.danger:disabled`, `.btn.ghost`, `.btn.ghost:disabled`, `.btn.ghost.danger`, `.btn.disabled` |
| Auth special case | `.kratos-form .btn.primary + .btn.primary` (changes the second adjacent primary into a quiet absolute-positioned action) |
| surfaces/overlays | `.card`, `.card:not(.narrow):not(.p-0)`, `.card.narrow`, `.modal-body`, `.modal-actions` |
| content/data | `ul.list`, `.table-container`, `.tag`, `.auto-format table`, `.auto-format .table-container`, `.auto-format ul`, `.auto-format a`, `.auto-format h2`, `.auto-format h3`, `.input-frame` |
| tables | `table.default`, `table th.default`, `.auto-format table th`, `table td.default`, `.auto-format table td`, `table tfoot`, `table td.disabled`, `table tr.link`, `table td.link`, `table td.link a` |
| flashes | `.flash`, `a.flash`, `.flash.info`, `a.flash.info`, `.flash.success`, `a.flash.success`, `.flash.warning`, `a.flash.warning`, `.flash.error`, `a.flash.error` |

Escape/state classes with styling significance include `.reset`, `.zebra`, `.small`, `.primary`, `.secondary`, `.danger`, `.ghost`, `.disabled`, `.narrow`, `.p-0`, `.spaced`, `.fill`, `.default`, `.link`, `.info`, `.success`, `.warning`, and `.error`. Several collide with generic/product vocabulary; a raw-token search needs selector context and cannot safely auto-rewrite them.

## Verified raw-class consumers

This table is based on static/template `className` strings. It excludes classes created inside legacy components and treats conditional strings conservatively.

| App | Direct legacy recipes/states observed | Migration hot spots |
| --- | --- | --- |
| admin | `auto-format`, `btn` + primary/secondary/danger/ghost, `card`, `default`, `fill`, `modal-actions`, `modal-body`, `spaced`, `subtitle`, `table-container`, `tag`, `title`, `v-stack` | editors/preview, users/languages modals, default tables, dashboard layout |
| auth | `btn` + primary/ghost/small, `card`, `h-stack`, `kratos-form`, `spaced`, `subtitle`, `title`, `v-stack` | dynamic Ory node output and adjacent-primary selector-order hack |
| legacy styleguide | all documented button states, `card`, `h-stack`, `input-frame`, form label recipes, `list`, modal recipes, `reset`, stacks, table recipe, typography recipes | raw-class documentation is part of the content migration, not dead demo code |
| webv2 | `auto-format`, `btn` + primary/danger/ghost, `card`/`narrow`, `default`, `h-stack`/`v-stack`/`spaced`, `input-frame`, modal recipes, `reset`, subtitles/titles, table/container, tags | product prose, manual, both logging generations, contest/profile tables, page-counter size exception |

### Button semantics requiring review

Legacy `.btn` is neutral, `.btn.primary` is emphasized, `.btn.secondary` is dark, and `.btn.danger` is destructive. Paper `.btn` is the emphasized default. Every call site therefore needs a semantic decision:

- legacy primary → Paper default in the ordinary emphasized case;
- legacy neutral/default or secondary → usually Paper outline, but inspect job and surrounding hierarchy;
- legacy ghost → Paper ghost or link based on action semantics;
- legacy danger → Paper destructive;
- anchors retain anchor semantics and use a recipe/adapter.

## Audit command proposal

Keep the exact selector list above as a versioned machine-readable allow/deny set when migration checks are implemented. A coarse discovery command is:

```sh
rg -n --glob '*.{ts,tsx,js,jsx}' \
  'class(Name)?=.*(btn|card|table-container|auto-format|kratos-form|text-link|title|subtitle|v-stack|h-stack|modal-body|modal-actions|label-hint|input-frame)' \
  frontend/apps
```

It is evidence discovery, not a zero-result gate, because words such as `title`, `error`, `default`, and `link` are too generic. The production guard should parse JSX/class composition and compare tokens against the versioned legacy set.

