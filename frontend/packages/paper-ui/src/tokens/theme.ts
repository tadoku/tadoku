import { layoutContainer, layoutGridTemplateColumns, layoutHeight, layoutSpacing, layoutText } from './layout'
import { palette as p } from './palette'

/** A colour that differs between the light and dark themes. */
export interface Themed {
  readonly light: string
  readonly dark: string
}

/** A size that differs between the comfortable and compact densities. */
export interface Dense {
  readonly comfortable: string
  readonly compact: string
}

/** A size that changes below breakpoints, desktop first: `{ base: '1.75rem', tablet: '1.25rem' }`. */
export type Responsive = { readonly base: string } & Partial<Record<keyof typeof breakpoint, string>>

export type TokenValue = string | Themed | Dense | Responsive

const themed = (light: string, dark: string): Themed => ({ light, dark })
const dense = (comfortable: string, compact: string): Dense => ({ comfortable, compact })
const mix = (color: string, percent: number, base = 'transparent') =>
  `color-mix(in srgb, var(--color-${color}) ${percent}%, ${base === 'transparent' ? base : `var(--color-${base})`})`

/**
 * Semantic colours. Utilities read them as `bg-paper`, `text-muted`, `border-rule` and so on; each resolves per
 * `data-theme`. Colours derived from other tokens are recomputed inside every theme scope.
 */
const color = {
  canvas: themed(p.graphite[100], p.graphite[900]),
  paper: themed(p.graphite[25], p.graphite[850]),
  raised: themed(p.white, p.graphite[750]),
  overlay: themed(p.graphite[25], p.graphite[750]),
  /** Text fields and pickers: a step lighter than the card, on a soft edge. */
  field: themed(p.white, p.graphite[700]),
  scrim: themed(p.shade.scrimLight, p.shade.scrimDark),

  ink: themed(p.graphite[850], p.graphite[100]),
  muted: themed(p.graphite[550], p.graphite[300]),
  inverse: p.white,
  link: themed(p.violet[600], p.violet[300]),
  'link-hover': themed(p.violet[750], p.violet[100]),

  'rule-subtle': themed(p.graphite[200], p.graphite[700]),
  rule: themed(p.graphite[300], p.graphite[600]),
  'rule-strong': themed(p.graphite[650], p.graphite[500]),
  /** The edge of a surface lifted by a hard offset shadow: the shadow carries the weight, so the line can be soft. */
  'surface-edge': themed(p.graphite[400], p.graphite[500]),
  /** Field edges identify the control, so they hold 3:1 against the field and the surface around it. */
  'field-edge': themed(p.graphite[380], p.graphite[450]),
  'field-edge-hover': themed(p.graphite[550], p.graphite[350]),
  'button-edge': themed(p.graphite[450], p.graphite[450]),
  'button-edge-hover': themed(p.graphite[650], p.graphite[400]),
  /** A filled button's edge: one translucent shade on every side, deeper at the bottom, so corners don't mitre. */
  'button-shade': themed('rgb(28 28 28 / 22%)', 'rgb(0 0 0 / 40%)'),

  action: themed(p.violet[600], p.violet[525]),
  'action-hover': themed(p.violet[700], p.violet[475]),
  'action-active': themed(p.violet[725], p.violet[550]),
  'action-soft': themed(p.violet[50], p.violet[900]),
  'neutral-hover': themed(p.graphite[175], p.graphite[700]),
  'raised-hover': themed(p.graphite[175], p.graphite[700]),
  'on-action': p.white,
  destructive: themed(p.red[500], p.red[450]),
  'destructive-hover': themed(p.red[600], p.red[400]),
  'on-destructive': p.white,

  information: themed(p.blue[600], p.blue[300]),
  success: themed(p.green[600], p.green[300]),
  warning: themed(p.amber[600], p.amber[300]),
  danger: themed(p.red[500], p.red[300]),

  focus: themed(p.violet[600], p.violet[300]),
  'focus-offset': 'var(--color-canvas)',
  'focus-halo': mix('focus', 12),
  'danger-halo': mix('danger', 22),
  shadow: themed(p.shade.ink, p.shade.violet),

  'chart-1': themed(p.violet[600], p.violet[350]),
  'chart-2': themed(p.green[600], p.green[300]),
  'chart-3': themed(p.amber[600], p.amber[300]),
  'chart-4': themed(p.blue[600], p.blue[300]),
  'chart-5': themed(p.red[500], p.red[300]),
  'chart-6': themed(p.violet[625], p.violet[250]),
  'chart-7': themed(p.teal[600], p.teal[300]),
  'chart-8': themed(p.rose[600], p.rose[300]),

  /** One hue per immersion activity, so lists and breakdowns can be scanned by activity. */
  /** Violet is left out: it is the link and action colour, and an activity must not read as a link. */
  'activity-reading': 'var(--color-chart-3)',
  'activity-listening': 'var(--color-chart-7)',
  'activity-writing': 'var(--color-chart-8)',
  'activity-speaking': 'var(--color-chart-2)',
  'activity-study': 'var(--color-chart-4)',
  'activity-other': 'var(--color-muted)',

  'heat-future': themed(p.graphite[150], p.graphite[800]),
  'heat-0': themed(p.graphite[200], p.graphite[650]),
  'heat-1': themed(p.violet[350], p.violet[550]),
  'heat-2': themed(p.violet[525], p.violet[400]),
  'heat-3': themed(p.violet[700], p.violet[200]),
  'heat-4': themed(p.violet[950], p.violet[25]),

  /** Component colours: tints that one control needs and the semantic set does not name. */
  'table-shade': mix('ink', 35),
  'canvas-table-shade': mix('ink', 22),
  'tab-hover': mix('ink', 4),
  'toggle-hover': mix('ink', 3, 'paper'),
  /** A selected value inside a field: a neutral step off the field, so it doesn't read as a link. */
  chip: themed(p.graphite[175], p.graphite[600]),
  'chip-remove-hover': themed(p.graphite[250], p.graphite[500]),
  'destructive-soft': mix('danger', 14),
  'current-row-hover': mix('ink', 5, 'action-soft'),
  'checkbox-disabled-edge': `color-mix(in srgb, var(--color-field-edge) 85%, var(--color-rule))`,

  /** The site footer sits on the paper surface and closes the page with a hairline. */
  footer: 'var(--color-canvas)',
  'footer-ink': 'var(--color-ink)',
  'footer-muted': 'var(--color-muted)',
  'footer-link': 'var(--color-ink)',
  'footer-label': 'var(--color-muted)',
  'footer-rule': 'var(--color-rule)',
  'footer-accent': 'var(--color-link)',
  'footer-hover': 'var(--color-link-hover)',
  'footer-tint': 'var(--color-tab-hover)',

  /** Leaderboard podium marks behind the top three ranks. */
  'podium-1': themed(p.gold[100], p.gold[800]),
  'on-podium-1': themed(p.gold[900], p.gold[300]),
  'podium-2': themed(p.silver[100], p.silver[800]),
  'on-podium-2': themed(p.silver[850], p.silver[200]),
  'podium-3': themed(p.bronze[100], p.bronze[800]),
  'on-podium-3': themed(p.bronze[850], p.bronze[300]),

  /** System colours for `forced-colors:` variants. */
  'forced-canvas': 'Canvas',
  'forced-ink': 'CanvasText',
  'forced-highlight': 'Highlight',
  'forced-on-highlight': 'HighlightText',
  'forced-muted': 'GrayText',
  'forced-mark': 'Mark',
  'forced-scrim': p.shade.forcedScrim,
} satisfies Record<string, TokenValue>

/** The 2px grid up to 16px, then 4px steps. Keys follow Tailwind's quarter-rem names: `p-3` is 12px. */
const space = {
  0: '0',
  px: '1px',
  0.5: '0.125rem',
  0.75: '0.1875rem',
  0.8: '0.2rem',
  1.2: '0.3rem',
  1: '0.25rem',
  1.25: '0.3125rem',
  1.4: '0.35rem',
  1.5: '0.375rem',
  1.6: '0.4rem',
  1.8: '0.45rem',
  2.2: '0.55rem',
  1.75: '0.4375rem',
  2: '0.5rem',
  2.5: '0.625rem',
  2.6: '0.65rem',
  3: '0.75rem',
  3.2: '0.8rem',
  3.4: '0.85rem',
  3.5: '0.875rem',
  3.6: '0.9rem',
  4: '1rem',
  4.5: '1.125rem',
  5: '1.25rem',
  5.5: '1.375rem',
  6: '1.5rem',
  7: '1.75rem',
  8: '2rem',
  9: '2.25rem',
  10: '2.5rem',
  11: '2.75rem',
  12: '3rem',
  14: '3.5rem',
  16: '4rem',
  18: '4.5rem',
  20: '5rem',
  22: '5.5rem',
  24: '6rem',
  25: '6.25rem',
  28: '7rem',
  32: '8rem',
  34: '8.5rem',
  48: '12rem',
  60: '15rem',
} as const

const spacing = {
  ...Object.fromEntries(Object.entries(space).map(([key, value]) => [key.replace('.', '_'), value])),
  /** Control sizes. */
  control: dense('2.75rem', '2.25rem'),
  target: '2.75rem',
  'icon-sm': '1rem',
  icon: '1.25rem',
  'icon-lg': '1.5rem',
  'icon-xl': '3rem',
  'check': '1.15rem',
  'chip-target': '1.5rem',
  /** Density-driven control insets. */
  'field-x': dense('0.75rem', '0.625rem'),
  'field-y': dense('0.625rem', '0.375rem'),
  inline: dense('0.625rem', '0.5rem'),
  'button-x': dense('0.9375rem', '0.6875rem'),
  'button-top': dense('0.5625rem', '0.375rem'),
  'button-bottom': dense('0.5rem', '0.3125rem'),
  'radio-min': dense('7rem', '6rem'),
  rail: '3px',
  'rail-outset': '1px',
  /** Derived control insets. */
  'select-end': 'calc(var(--spacing-field-x) * 2 + var(--spacing-icon-sm))',
  'combobox-end': 'calc(var(--spacing-target) + var(--spacing-1))',
  'chip-input-start': 'calc(var(--spacing-field-x) - var(--spacing-1))',
  'check-hint': 'calc(var(--spacing-check) + var(--spacing-inline))',
  'rail-inset': 'calc(var(--spacing-field-x) + var(--spacing-rail))',
  'flash-start': 'calc(var(--spacing-4) + var(--spacing-rail))',
  'page-title-trim': 'calc((var(--leading-page) - var(--spacing-control)) / 2)',
  'page-actions-trim': 'calc(var(--leading-page) - var(--spacing-control))',
  'compound-input': 'calc(var(--spacing-control) - 2px)',
  'compound-unit': 'calc(var(--spacing-control) - 1px)',
  ...layoutSpacing,
} satisfies Record<string, TokenValue>

/** Inline sizes: measures for text and widths for floating surfaces. */
const container = {
  'measure-label': '18ch',
  'measure-xs': '17ch',
  'measure-sm': '47ch',
  'measure-description': '58ch',
  'measure-note': '60ch',
  measure: '65ch',
  'measure-prose': '66ch',
  menu: '13rem',
  'unit-select': '6rem',
  'chip-input': '3rem',
  toggle: '28rem',
  heatmap: '44rem',
  'heatmap-legend': 'min(100%, calc(100vw - 5rem))',
  'nav-demo': '22rem',
  form: '32rem',
  'form-wide': '42rem',
  'panel-narrow': '36rem',
  panel: '40rem',
  'panel-wide': '44rem',
  drawer: 'min(28rem, calc(100vw - var(--spacing-control)))',
  modal: 'min(36rem, 100%)',
  ...layoutContainer,
  toast: 'min(24rem, calc(100vw - 2rem))',
} satisfies Record<string, TokenValue>

const height = {
  textarea: '5.5rem',
  popup: '18rem',
  modal: 'min(42rem, calc(100vh - 2rem))',
  ...layoutHeight,
} satisfies Record<string, TokenValue>

const font = {
  display:
    '"Merriweather", Georgia, "Hiragino Mincho ProN", "Yu Mincho", "Noto Serif JP", "Noto Serif CJK JP", "Times New Roman", serif',
  body:
    '"Open Sans", Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", "Hiragino Sans", "Yu Gothic UI", "Noto Sans JP", "Noto Sans CJK JP", sans-serif',
  mono: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace',
} as const

/** Font sizes. Type roles name Paper's hierarchy; the plain steps serve metadata and page copy. */
const text = {
  display: '3rem',
  page: '2rem',
  section: '1.5rem',
  component: '1.125rem',
  body: dense('1rem', '0.875rem'),
  label: '0.875rem',
  metadata: '0.875rem',
  button: dense('1em', '0.875rem'),
  em: '1em',
  code: '0.875em',
  /** Plain steps; names that Tailwind also uses keep Tailwind's sizes. */
  '2xs': '0.6875rem',
  xs: '0.75rem',
  'sm-tight': '0.8rem',
  fine: '0.8125rem',
  sm: '0.875rem',
  ui: '0.9375rem',
  base: '1rem',
  lead: '1.0625rem',
  lg: '1.125rem',
  xl: '1.25rem',
  '2xl': '1.5rem',
  '3xl': '2rem',
  'heat-axis': '9px',
  'heat-axis-phone': '12px',
  ...layoutText,
} satisfies Record<string, TokenValue>

const leading = {
  display: '3.25rem',
  page: '2.375rem',
  section: '1.875rem',
  component: '1.5rem',
  body: dense('1.625rem', '1.375rem'),
  label: '1.125rem',
  metadata: '1.25rem',
  100: '1',
  120: '1.2',
  125: '1.25',
  130: '1.3',
  135: '1.35',
  140: '1.4',
  145: '1.45',
  150: '1.5',
  155: '1.55',
  160: '1.6',
  165: '1.65',
  170: '1.7',
  175: '1.75',
  180: '1.8',
  185: '1.85',
} satisfies Record<string, TokenValue>

const fontWeight = {
  regular: '400',
  medium: '500',
  semibold: '600',
  bold: '700',
} as const

const tracking = {
  tight: '-0.04em',
  wide: '0.02em',
  hero: '-0.03em',
  caps: '0.08em',
  display: '-0.035em',
  page: '-0.025em',
  section: '-0.015em',
  label: '0.04em',
} as const

const borderWidth = {
  hairline: '1px',
  chevron: '1.5px',
  edge: '2px',
  rail: '3px',
  bar: '4px',
} as const

const outlineWidth = { ring: '2px', halo: '3px' } as const
const outlineOffset = { 0: '0', px: '1px', ring: '2px' } as const
const radius = { chip: '2px', pill: '999px' } as const
const textDecorationThickness = { hairline: '1px', hover: '2px' } as const
const textUnderlineOffset = { link: '0.2em' } as const

const shadow = {
  floating: '3px 3px 0 var(--color-shadow)',
  showcase: '5px 5px 0 var(--color-shadow)',
  'underline-link': 'inset 0 -1px 0 var(--color-link)',
  'current-bar': 'inset 0 -3px 0 var(--color-link)',
  'current-bar-start': 'inset 3px 0 0 var(--color-link)',
  'rule-bottom': 'inset 0 -1px 0 var(--color-rule)',
  'rule-start': 'inset 1px 0 0 var(--color-rule)',
  'rail-focus': 'inset 3px 0 0 var(--color-focus)',
} as const

const opacity = { disabled: '0.56', faded: '0.5', full: '1' } as const
const strokeWidth = { hairline: '1px', ring: '2px', check: '2.5' } as const
const scale = { enter: '0.98', mirror: '-1' } as const
const rotate = { eighth: '45deg', quarter: '90deg', flip: '180deg', 'chevron-up': '225deg' } as const
const translate = { offscreen: '180%' } as const
const transitionProperty = { enter: 'opacity, transform', inset: 'inset-inline-start' } as const
const ease = { out: 'ease-out' } as const
const gridTemplateRows = { docs: 'auto 1fr' } as const
const blur = { scrim: '0.375rem' } as const

const zIndex = {
  lift: '1',
  'lift-2': '2',
  'docs-navbar': '30',
  header: '40',
  backdrop: '80',
  dialog: '81',
  menu: '90',
  popover: '95',
  skip: '100',
  toast: '110',
} as const

const transitionDuration = {
  quick: '120ms',
  standard: '180ms',
  deliberate: '240ms',
} as const

const animate = {
  spin: 'paper-spin 700ms linear infinite',
  progress: 'paper-progress 1.2s linear infinite',
} as const

const keyframes = {
  'paper-spin': { to: 'transform: rotate(1turn);' },
  'paper-progress': { from: 'transform: translateX(-100%);', to: 'transform: translateX(350%);' },
} as const

const breakpoint = {
  '2xs': '23rem',
  xs: '25rem',
  sm: '30rem',
  md: '40rem',
  tablet: '48rem',
  lg: '64rem',
} as const

const gridTemplateColumns = {
  'content-end': 'minmax(0, 1fr) auto',
  'icon-content-end': 'auto minmax(0, 1fr) auto',
  'icon-content': 'auto minmax(0, 1fr)',
  'mark-label': 'auto 1fr',
  'label-check': '1fr auto',
  'item-indicator': '1.25rem 1fr',
  'radio-cards': 'repeat(auto-fit, minmax(12rem, 1fr))',
  stats: 'repeat(auto-fit, minmax(min(100%, 4.75rem), 1fr))',
  'facts-inline': 'minmax(6rem, 10rem) minmax(0, 1fr)',
  'tabs-vertical': 'max-content minmax(0, 1fr)',
  filters: 'repeat(auto-fit, minmax(11rem, 1fr))',
  single: 'minmax(0, 1fr)',
  'footer-links': 'repeat(3, minmax(0, 10rem))',
  'footer-links-narrow': 'repeat(auto-fill, minmax(9rem, 1fr))',
  ...layoutGridTemplateColumns,
} as const

/**
 * The Tailwind theme. Keys are Tailwind v4 theme namespaces; `--*: initial` removes Tailwind's defaults, so these
 * are the only values utilities can use.
 */
export const theme = {
  color,
  spacing,
  container,
  height,
  font,
  text,
  leading,
  'font-weight': fontWeight,
  tracking,
  'border-width': borderWidth,
  'outline-width': outlineWidth,
  'outline-offset': outlineOffset,
  radius,
  'text-decoration-thickness': textDecorationThickness,
  'text-underline-offset': textUnderlineOffset,
  shadow,
  opacity,
  'stroke-width': strokeWidth,
  scale,
  rotate,
  translate,
  blur,
  'z-index': zIndex,
  'transition-duration': transitionDuration,
  'transition-property': transitionProperty,
  ease,
  'grid-template-rows': gridTemplateRows,
  animate,
  breakpoint,
  'grid-template-columns': gridTemplateColumns,
} as const

export const themeKeyframes = keyframes

/** Tailwind defaults that its own utilities read, set from the tokens above. */
export const themeDefaults = {
  'default-font-family': 'var(--font-body)',
  'default-mono-font-family': 'var(--font-mono)',
  'default-border-width': borderWidth.hairline,
  'default-outline-width': outlineWidth.ring,
  'default-transition-duration': transitionDuration.quick,
  'default-transition-timing-function': 'ease',
} as const

export type Theme = typeof theme
export type ThemeNamespace = keyof Theme
