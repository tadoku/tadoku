/**
 * The single source of Paper's design tokens. `scripts/tokens.ts` generates `tokens.stylex.ts` and `themes.ts`
 * from this file; edit values here and run `pnpm --filter paper-ui tokens`.
 *
 * Tiers:
 * - `palette` and the scales are primitives. Components never read the palette directly.
 * - `colorRoles` maps every semantic colour to a palette entry for each theme.
 * - `density` holds the control metrics that change between comfortable and compact.
 */

export const palette = {
  warm: {
    0: '#ffffff',
    25: '#fffefa',
    50: '#f2efe8',
    75: '#ebe6dd',
    100: '#e9e5dc',
    150: '#e2ddd4',
    200: '#d3ccc0',
    300: '#bcb4a8',
    500: '#857d72',
    700: '#5e5659',
    800: '#393235',
    900: '#211a1d',
  },
  /** The footer band on ink: a step of muted warm greys between the paper and ink ramps. */
  band: {
    100: '#ddd6d9',
    300: '#b4abaf',
    400: '#9d9498',
    700: '#4a4245',
    950: '#1b1618',
  },
  graphite: {
    100: '#e8e8e8',
    300: '#aeaeae',
    500: '#808080',
    600: '#5f5f5f',
    700: '#444444',
    800: '#2c2c2c',
    850: '#252525',
    875: '#232323',
    900: '#1c1c1c',
    950: '#141414',
  },
  violet: {
    50: '#ece7ff',
    75: '#e8e4fa',
    100: '#cbc2f8',
    110: '#c4b9ff',
    125: '#c0b4f5',
    150: '#a99cf0',
    200: '#9a8bea',
    210: '#9a8ae6',
    225: '#8f80e8',
    250: '#8a7ae0',
    300: '#7a6ad2',
    325: '#7262dd',
    350: '#6a58cf',
    375: '#6656d4',
    400: '#5b4bc8',
    450: '#5747b8',
    500: '#4d3daa',
    550: '#49399f',
    600: '#3f3196',
    625: '#40318c',
    650: '#3e3290',
    700: '#35277f',
    750: '#33277c',
    950: '#383646',
  },
  red: {
    300: '#d77171',
    400: '#c05050',
    450: '#b84a4a',
    500: '#a33333',
    600: '#892828',
    700: '#7a2b2b',
    800: '#651b1b',
  },
  green: { 300: '#7fb488', 600: '#3f7048' },
  amber: { 300: '#ddb66f', 500: '#9b6a25', 600: '#8a5b16' },
  blue: { 300: '#70acd0', 600: '#2f6f8f' },
  plum: { 300: '#b69ad0', 600: '#6c568c' },
  sage: { 300: '#83aaa3', 600: '#55716d' },
  rose: { 300: '#c58e9b', 600: '#83565d' },
  gold: { 100: '#f3e2a6', 200: '#f2d77c', 800: '#4d3f12', 900: '#5c4300' },
  silver: { 100: '#e3e3ea', 150: '#dcdce6', 800: '#3b3b46', 900: '#3d3d4a' },
  bronze: { 100: '#f0d3bd', 200: '#eebc98', 800: '#4a2c1b', 900: '#6b3412' },
  /** Translucent colours: scrims, shadows and washes over a known base. */
  alpha: {
    inkScrim: 'rgb(33 26 29 / 28%)',
    inkShadow: 'rgb(33 26 29 / 16%)',
    blackScrim: 'rgb(0 0 0 / 38%)',
    blackForcedScrim: 'rgb(0 0 0 / 50%)',
    violetShadow: 'rgb(124 104 224 / 34%)',
    violetWash: 'rgb(143 128 232 / 14%)',
  },
} as const

type Palette = typeof palette
export type PaletteRef = {
  [Ramp in keyof Palette]: `${Ramp}.${Extract<keyof Palette[Ramp], string | number>}`
}[keyof Palette]

const colorRoleNames = [
  'surfaceCanvas',
  'surfacePaper',
  'surfaceRaised',
  'surfaceOverlay',
  'surfaceScrim',
  /** The scrim kept in forced-colours mode, where the theme's own scrim is replaced. */
  'forcedScrim',
  'textInk',
  'textMuted',
  'textInverse',
  'textLink',
  'textLinkHover',
  'ruleSubtle',
  'ruleDefault',
  'ruleStrong',
  'ruleFieldEdge',
  'ruleActionEdge',
  'ruleDestructiveActionEdge',
  'actionDefault',
  'actionHover',
  'actionActive',
  'actionSoft',
  'actionNeutralHover',
  'actionText',
  'actionDestructive',
  'actionDestructiveHover',
  'actionDestructiveText',
  'statusInformation',
  'statusSuccess',
  'statusWarning',
  'statusDanger',
  'focusRing',
  'focusOffset',
  'shadow',
  'chart1',
  'chart2',
  'chart3',
  'chart4',
  'chart5',
  'chart6',
  'chart7',
  'chart8',
  'heatFuture',
  'heat0',
  'heat1',
  'heat2',
  'heat3',
  'heat4',
  /** The footer band: ink in light mode, the paper surface in dark mode. */
  'bandSurface',
  'bandText',
  'bandMuted',
  'bandLink',
  'bandLabel',
  'bandRule',
  'bandHover',
  'bandHoverWash',
  'bandAccent',
  'podiumGoldFill',
  'podiumGoldInk',
  'podiumSilverFill',
  'podiumSilverInk',
  'podiumBronzeFill',
  'podiumBronzeInk',
  /** Translucent washes and halos derived from the roles above. */
  'focusHalo',
  'dangerHalo',
  'inkWash',
  'linkWash',
  'scrollShade',
  'proseScrollShade',
  'disabledCheckEdge',
  'toggleHover',
  'currentRowHover',
] as const

export type ColorRole = (typeof colorRoleNames)[number]
export type Theme = 'light' | 'dark'

/** A palette entry, or `amount` percent of one mixed with another entry or with transparency. */
export type ColorValue = PaletteRef | { readonly mix: PaletteRef; readonly amount: number; readonly with?: PaletteRef }

const derived = (theme: Record<string, PaletteRef>) =>
  ({
    focusHalo: { mix: theme.focusRing, amount: 22 },
    dangerHalo: { mix: theme.statusDanger, amount: 22 },
    inkWash: { mix: theme.textInk, amount: 4 },
    linkWash: { mix: theme.textLink, amount: 16 },
    scrollShade: { mix: theme.textInk, amount: 35 },
    proseScrollShade: { mix: theme.textInk, amount: 22 },
    disabledCheckEdge: { mix: theme.ruleFieldEdge, amount: 85, with: theme.ruleDefault },
    toggleHover: { mix: theme.textInk, amount: 3, with: theme.surfacePaper },
    currentRowHover: { mix: theme.textInk, amount: 5, with: theme.actionSoft },
  }) as const

type BaseRole = Exclude<ColorRole, keyof ReturnType<typeof derived>>

const lightRoles = {
  surfaceCanvas: 'warm.50',
  surfacePaper: 'warm.25',
  surfaceRaised: 'warm.0',
  surfaceOverlay: 'warm.25',
  surfaceScrim: 'alpha.inkScrim',
  forcedScrim: 'alpha.blackForcedScrim',
  textInk: 'warm.900',
  textMuted: 'warm.700',
  textInverse: 'warm.25',
  textLink: 'violet.450',
  textLinkHover: 'violet.600',
  ruleSubtle: 'warm.150',
  ruleDefault: 'warm.300',
  ruleStrong: 'warm.800',
  ruleFieldEdge: 'warm.500',
  ruleActionEdge: 'violet.700',
  ruleDestructiveActionEdge: 'red.800',
  actionDefault: 'violet.450',
  actionHover: 'violet.550',
  actionActive: 'violet.625',
  actionSoft: 'violet.75',
  actionNeutralHover: 'warm.100',
  actionText: 'warm.25',
  actionDestructive: 'red.500',
  actionDestructiveHover: 'red.600',
  actionDestructiveText: 'warm.25',
  statusInformation: 'blue.600',
  statusSuccess: 'green.600',
  statusWarning: 'amber.600',
  statusDanger: 'red.500',
  focusRing: 'violet.450',
  focusOffset: 'warm.50',
  shadow: 'alpha.inkShadow',
  chart1: 'violet.450',
  chart2: 'green.600',
  chart3: 'amber.500',
  chart4: 'blue.600',
  chart5: 'red.500',
  chart6: 'plum.600',
  chart7: 'sage.600',
  chart8: 'rose.600',
  heatFuture: 'warm.75',
  heat0: 'warm.200',
  heat1: 'violet.250',
  heat2: 'violet.350',
  heat3: 'violet.500',
  heat4: 'violet.750',
  bandSurface: 'band.950',
  bandText: 'warm.50',
  bandMuted: 'band.300',
  bandLink: 'band.100',
  bandLabel: 'band.400',
  bandRule: 'band.700',
  bandHover: 'violet.110',
  bandHoverWash: 'alpha.violetWash',
  bandAccent: 'violet.225',
  podiumGoldFill: 'gold.100',
  podiumGoldInk: 'gold.900',
  podiumSilverFill: 'silver.100',
  podiumSilverInk: 'silver.900',
  podiumBronzeFill: 'bronze.100',
  podiumBronzeInk: 'bronze.900',
} satisfies Record<BaseRole, PaletteRef>

const darkRoles = {
  surfaceCanvas: 'graphite.950',
  surfacePaper: 'graphite.900',
  surfaceRaised: 'graphite.850',
  surfaceOverlay: 'graphite.850',
  surfaceScrim: 'alpha.blackScrim',
  forcedScrim: 'alpha.blackForcedScrim',
  textInk: 'graphite.100',
  textMuted: 'graphite.300',
  textInverse: 'warm.25',
  textLink: 'violet.150',
  textLinkHover: 'violet.100',
  ruleSubtle: 'graphite.800',
  ruleDefault: 'graphite.700',
  ruleStrong: 'graphite.600',
  ruleFieldEdge: 'graphite.500',
  ruleActionEdge: 'violet.650',
  ruleDestructiveActionEdge: 'red.700',
  actionDefault: 'violet.375',
  actionHover: 'violet.325',
  actionActive: 'violet.400',
  actionSoft: 'violet.950',
  actionNeutralHover: 'graphite.875',
  actionText: 'warm.25',
  actionDestructive: 'red.450',
  actionDestructiveHover: 'red.400',
  actionDestructiveText: 'warm.25',
  statusInformation: 'blue.300',
  statusSuccess: 'green.300',
  statusWarning: 'amber.300',
  statusDanger: 'red.300',
  focusRing: 'violet.150',
  focusOffset: 'graphite.950',
  shadow: 'alpha.violetShadow',
  chart1: 'violet.200',
  chart2: 'green.300',
  chart3: 'amber.300',
  chart4: 'blue.300',
  chart5: 'red.300',
  chart6: 'plum.300',
  chart7: 'sage.300',
  chart8: 'rose.300',
  heatFuture: 'graphite.875',
  heat0: 'graphite.700',
  heat1: 'violet.300',
  heat2: 'violet.210',
  heat3: 'violet.125',
  heat4: 'violet.50',
  bandSurface: 'graphite.900',
  bandText: 'graphite.100',
  bandMuted: 'graphite.300',
  bandLink: 'graphite.100',
  bandLabel: 'graphite.300',
  bandRule: 'graphite.700',
  bandHover: 'violet.110',
  bandHoverWash: 'alpha.violetWash',
  bandAccent: 'violet.225',
  podiumGoldFill: 'gold.800',
  podiumGoldInk: 'gold.200',
  podiumSilverFill: 'silver.800',
  podiumSilverInk: 'silver.150',
  podiumBronzeFill: 'bronze.800',
  podiumBronzeInk: 'bronze.200',
} satisfies Record<BaseRole, PaletteRef>

export const colorRoles: Record<Theme, Record<ColorRole, ColorValue>> = {
  light: { ...lightRoles, ...derived(lightRoles) },
  dark: { ...darkRoles, ...derived(darkRoles) },
}

/** Lengths keyed by their size in CSS pixels at the default 16px root, emitted in rem. */
const px = <const Steps extends readonly number[]>(steps: Steps) =>
  Object.fromEntries(steps.map(step => [step, step === 0 ? '0' : `${step / 16}rem`])) as Record<Steps[number], string>

export const space = px([
  0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 14, 15, 16, 18, 20, 22, 24, 28, 32, 36, 40, 44, 48, 56, 64, 72, 80, 88, 96,
])

/** Block and inline sizes: icons, targets, columns, sheets and page widths. */
export const size = px([
  1, 2, 3, 4, 8, 12, 14, 16, 18, 20, 24, 28, 32, 35, 36, 40, 44, 48, 56, 61, 64, 72, 80, 88, 96, 100, 106, 112, 120,
  124, 128, 136, 140, 142, 144, 160, 176, 180, 192, 200, 208, 224, 230, 240, 256, 260, 272, 288, 300, 320, 352, 360,
  384, 400, 416, 420, 440, 448, 480, 512, 520, 544, 560, 576, 580, 608, 640, 672, 680, 700, 704, 720, 736, 760, 768, 800, 900, 1000,
  1040, 1050, 1152, 1160, 1216,
])

/** Viewport-relative lengths for fluid gutters and type, keyed by their percentage of the viewport width. */
export const fluid = { 2: '2vw', 3.5: '3.5vw', 4: '4vw', 5: '5vw', 6: '6vw', 6.75: '6.75vw', 3.25: '3.25vw', vh7: '7vh', vh55: '55vh' } as const

/** Reading measures in characters of the current font. */
export const measure = {
  caption: '10ch',
  headline: '17ch',
  title: '18ch',
  card: '25ch',
  narrow: '34ch',
  short: '42ch',
  intro: '47ch',
  compact: '48ch',
  aside: '53ch',
  summary: '54ch',
  note: '56ch',
  lead: '58ch',
  text: '60ch',
  copy: '62ch',
  answer: '64ch',
  body: '65ch',
  prose: '66ch',
  column: '68ch',
  wide: '76ch',
} as const

export const font = {
  display:
    '"Merriweather", Georgia, "Hiragino Mincho ProN", "Yu Mincho", "Noto Serif JP", "Noto Serif CJK JP", ' +
    '"Times New Roman", serif',
  body:
    '"Open Sans", Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", ' +
    '"Hiragino Sans", "Yu Gothic UI", "Noto Sans JP", "Noto Sans CJK JP", sans-serif',
  monospace: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace',
} as const

export const fontSize = {
  ...px([9, 11, 12, 13, 14, 15, 16, 17, 18, 20, 22, 24, 26, 28, 32, 36, 40, 44, 48, 56]),
  /** Relative sizes for runs inside text. */
  em75: '0.75em',
  em875: '0.875em',
  inherit: 'inherit',
} as const

export const lineHeight = {
  none: '1',
  tight: '1.15',
  hero: '1.2',
  snug: '1.25',
  statValue: '1.3',
  heading: '1.35',
  compact: '1.4',
  statement: '1.45',
  normal: '1.5',
  description: '1.55',
  relaxed: '1.6',
  loose: '1.65',
  airy: '1.7',
  open: '1.75',
  prose: '1.8',
  extraLoose: '1.85',
  ...px([16, 18, 20, 22, 24, 26, 28, 30, 32, 36, 38, 44, 52, 56, 60]),
} as const

export const fontWeight = { regular: '400', medium: '500', semibold: '600', bold: '700' } as const

export const tracking = {
  tightest: '-0.045em',
  tighter: '-0.04em',
  tight: '-0.035em',
  snug: '-0.025em',
  slight: '-0.015em',
  normal: '0',
  wide: '0.02em',
  wider: '0.04em',
  widest: '0.08em',
} as const

export const type = {
  displaySize: fontSize[48],
  displayLineHeight: lineHeight[52],
  displayTracking: tracking.tight,
  pageSize: fontSize[32],
  pageLineHeight: lineHeight[38],
  pageTracking: tracking.snug,
  sectionSize: fontSize[24],
  sectionLineHeight: lineHeight[30],
  sectionTracking: tracking.slight,
  componentSize: fontSize[18],
  componentLineHeight: lineHeight[24],
  labelSize: fontSize[13],
  labelLineHeight: lineHeight[18],
  labelTracking: tracking.wider,
  metadataSize: fontSize[14],
  metadataLineHeight: lineHeight[20],
} as const

/** Underlines: the offset below text and the thicker stroke a hovered link draws. */
export const decoration = { offset: '0.2em', hover: '2px' } as const

export const radius = { none: '0', hairline: '1px', small: '2px', round: '50%', pill: '999px' } as const

export const border = {
  none: '0',
  static: '1px',
  fieldEdge: '2px',
  actionEdge: '2px',
  accentRail: '3px',
  indicator: '3px',
  spinner: '2px',
  spinnerLarge: '3px',
  focusRing: '2px',
  focusOffset: '2px',
  /** The stroke of a disclosure chevron drawn with borders. */
  chevron: '1.5px',
  /** Fields show focus on their own edge plus a soft halo, so the ring never floats away from the field. */
  focusHalo: '3px',
  forcedFocus: '3px',
} as const

export const motion = {
  instant: '0s',
  quick: '120ms',
  standard: '180ms',
  deliberate: '240ms',
  spin: '700ms',
  progress: '1.2s',
  easeOut: 'ease-out',
  linear: 'linear',
} as const

/** Glyph-relative sizes for marks drawn beside text, such as a check or disclosure chevron. */
export const glyph = { size: '0.7em', gap: '0.35em' } as const

/** Background blurs behind dialogs. */
export const blur = { scrim: '0.375rem' } as const

/** Proportions of a parent, such as the moving segment of an indeterminate progress bar. */
export const ratio = { progress: '40%', scrollCover: '40%', offscreen: '-180%' } as const

export const opacity = { hidden: '0', inactive: '0.5', disabled: '0.56', muted: '0.7', visible: '1' } as const

/** Stacking order, lowest first. */
export const zIndex = {
  base: '0',
  raised: '1',
  focused: '2',
  sticky: '30',
  navbar: '40',
  scrim: '80',
  dialog: '81',
  menu: '90',
  popover: '95',
  modal: '100',
  skipLink: '100',
  toast: '110',
} as const

/**
 * Media query conditions; widths below the named breakpoint match. Widths are in px because StyleX orders
 * overlapping width queries, so the narrower one wins, only when they use px.
 */
export const breakpoints = {
  xxs: '@media (max-width: 360px)',
  tiny: '@media (max-width: 368px)',
  xs: '@media (max-width: 400px)',
  compactPhone: '@media (max-width: 420px)',
  phone: '@media (max-width: 480px)',
  sm: '@media (max-width: 580px)',
  narrow: '@media (max-width: 720px)',
  md: '@media (max-width: 768px)',
  tablet: '@media (max-width: 800px)',
  lg: '@media (max-width: 900px)',
  xl: '@media (max-width: 1000px)',
  belowWide: '@media (max-width: 1024px)',
  wide: '@media (min-width: 1024px)',
  aboveMd: '@media (min-width: 768px)',
  motionOk: '@media (prefers-reduced-motion: no-preference)',
  /** Between the phone navigation and a full desktop bar. */
  mdToLg: '@media (min-width: 768px) and (max-width: 899px)',
  /** A container narrower than a two-up row of choices. */
  containerNarrow: '@container (max-width: 28rem)',
  reducedMotion: '@media (prefers-reduced-motion: reduce)',
  forcedColors: '@media (forced-colors: active)',
} as const

/** A value that changes at breakpoints; keys are `breakpoints` names, applied in order. */
export type Responsive = { readonly default: string } & { readonly [Name in keyof typeof breakpoints]?: string }

/**
 * Page frame metrics shared by an application's header, content and footer. They respond to the viewport, so every
 * band that uses them stays aligned.
 */
export const page: Record<'contentWidth' | 'gutter' | 'brandWidth', string | Responsive> = {
  contentWidth: '72.5rem',
  gutter: { default: '1.75rem', tablet: '1.25rem', phone: '1rem' },
  brandWidth: { default: '8.875rem', tablet: '7.5rem', mdToLg: '5.5rem', phone: '6.625rem', xs: '5.5rem', xxs: '4.5rem' },
}

export type Density = 'comfortable' | 'compact'

export const density: Record<Density, Record<string, string>> = {
  comfortable: {
    controlHeight: '2.75rem',
    buttonPaddingBlockStart: '0.5625rem',
    buttonPaddingBlockEnd: '0.5rem',
    buttonPaddingInline: '0.9375rem',
    buttonFontSize: '1em',
    radioMinWidth: '7rem',
    inlineGap: '0.625rem',
    fieldPaddingBlock: '0.625rem',
    fieldPaddingInline: '0.75rem',
    bodySize: '1rem',
    bodyLineHeight: '1.625rem',
    sidebarGap: '1.5rem',
  },
  compact: {
    controlHeight: '2.25rem',
    buttonPaddingBlockStart: '0.375rem',
    buttonPaddingBlockEnd: '0.3125rem',
    buttonPaddingInline: '0.6875rem',
    buttonFontSize: '0.875rem',
    radioMinWidth: '6rem',
    inlineGap: '0.5rem',
    fieldPaddingBlock: '0.375rem',
    fieldPaddingInline: '0.625rem',
    bodySize: '0.875rem',
    bodyLineHeight: '1.375rem',
    sidebarGap: '1rem',
  },
}

/** Offsets of Paper's hard shadows; their colour is the theme's `shadow` role. */
export const shadow = { floating: '3px 3px 0', showcase: '5px 5px 0' } as const

export const tokenSource = {
  palette,
  colorRoles,
  space,
  size,
  fluid,
  measure,
  font,
  fontSize,
  lineHeight,
  fontWeight,
  tracking,
  type,
  decoration,
  radius,
  border,
  motion,
  glyph,
  blur,
  ratio,
  opacity,
  zIndex,
  breakpoints,
  density,
  page,
  shadow,
} as const

export { colorRoleNames }
