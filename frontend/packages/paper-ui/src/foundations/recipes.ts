import * as stylex from '@stylexjs/stylex'
import {
  border,
  breakpoints,
  color,
  decoration,
  density,
  font,
  fontWeight,
  lineHeight,
  measure,
  motion,
  shadow,
  size,
  space,
  type,
  zIndex,
} from '../tokens/tokens.stylex'

/** Paper's type roles. Headings reset their block margins; set spacing with the layout around them. */
export const text = stylex.create({
  display: {
    marginBlock: 0,
    fontFamily: font.display,
    fontSize: type.displaySize,
    fontWeight: fontWeight.bold,
    letterSpacing: type.displayTracking,
    lineHeight: type.displayLineHeight,
  },
  page: {
    marginBlock: 0,
    fontFamily: font.display,
    fontSize: type.pageSize,
    fontWeight: fontWeight.bold,
    letterSpacing: type.pageTracking,
    lineHeight: type.pageLineHeight,
  },
  section: {
    marginBlock: 0,
    fontFamily: font.display,
    fontSize: type.sectionSize,
    fontWeight: fontWeight.bold,
    letterSpacing: type.sectionTracking,
    lineHeight: type.sectionLineHeight,
  },
  component: {
    fontFamily: font.body,
    fontSize: type.componentSize,
    fontWeight: fontWeight.bold,
    lineHeight: type.componentLineHeight,
  },
  body: {
    fontFamily: font.body,
    fontSize: density.bodySize,
    fontWeight: fontWeight.regular,
    lineHeight: density.bodyLineHeight,
  },
  label: {
    fontFamily: font.body,
    fontSize: type.labelSize,
    fontWeight: fontWeight.semibold,
    letterSpacing: type.labelTracking,
    lineHeight: type.labelLineHeight,
  },
  metadata: {
    color: color.textMuted,
    fontFamily: font.body,
    fontSize: type.metadataSize,
    fontWeight: fontWeight.regular,
    lineHeight: type.metadataLineHeight,
  },
  muted: { color: color.textMuted },
  ink: { color: color.textInk },
  tabular: { fontVariantNumeric: 'tabular-nums' },
})

export const surface = stylex.create({
  paper: {
    borderColor: color.ruleDefault,
    borderStyle: 'solid',
    borderWidth: border.static,
    backgroundColor: color.surfacePaper,
    color: color.textInk,
  },
  raised: {
    borderColor: color.ruleDefault,
    borderStyle: 'solid',
    borderWidth: border.static,
    backgroundColor: color.surfaceRaised,
    color: color.textInk,
  },
  canvas: { backgroundColor: color.surfaceCanvas },
})

const rail = {
  backgroundColor: color.actionDefault,
  content: '""',
  insetInlineStart: 0,
  inlineSize: border.accentRail,
  pointerEvents: 'none',
  position: 'absolute',
  zIndex: zIndex.raised,
} as const

/** A 3px leading rail. Use `onBorder` when the element has Paper's one-pixel perimeter, so the rail covers it. */
export const accentRail = stylex.create({
  base: {
    position: 'relative',
    '::before': { ...rail, insetBlock: 0 },
  },
  onBorder: {
    position: 'relative',
    '::before': {
      ...rail,
      insetBlock: `calc(-1 * ${border.static})`,
      insetInlineStart: `calc(-1 * ${border.static})`,
    },
  },
  forcedColors: {
    '::before': {
      backgroundColor: { default: null, [breakpoints.forcedColors]: 'CanvasText' },
      forcedColorAdjust: { default: null, [breakpoints.forcedColors]: 'none' },
    },
  },
})

export const edge = stylex.create({
  field: {
    borderColor: color.ruleDefault,
    borderRadius: 0,
    borderStyle: 'solid',
    borderWidth: border.static,
    borderBottomColor: color.ruleFieldEdge,
    borderBottomWidth: border.fieldEdge,
  },
  action: {
    borderColor: color.actionDefault,
    borderRadius: 0,
    borderStyle: 'solid',
    borderWidth: border.static,
    borderBottomColor: color.ruleActionEdge,
    borderBottomWidth: border.actionEdge,
  },
})

export const elevation = stylex.create({
  flat: { boxShadow: 'none' },
  floating: { borderColor: color.ruleStrong, boxShadow: `${shadow.floating} ${color.shadow}` },
  showcase: { borderColor: color.ruleStrong, boxShadow: `${shadow.showcase} ${color.shadow}` },
})

export const icon = stylex.create({
  compact: { flex: 'none', blockSize: size[16], display: 'inline-block', inlineSize: size[16], verticalAlign: 'middle', },
  default: { flex: 'none', blockSize: size[20], display: 'inline-block', inlineSize: size[20], verticalAlign: 'middle', },
  prominent: { flex: 'none', blockSize: size[24], display: 'inline-block', inlineSize: size[24], verticalAlign: 'middle', },
  emptyState: {
    flex: 'none',
    blockSize: size[48],
    display: 'inline-block',
    inlineSize: size[48],
    verticalAlign: 'middle',
  },
  /** A fixed box that centres an icon of any intrinsic size. */
  slot: {
    flex: 'none',
    placeItems: 'center',
    blockSize: size[20],
    display: 'inline-grid',
    inlineSize: size[20],
  },
  /** An icon filling its slot. */
  fill: { blockSize: '100%', inlineSize: '100%' },
})

/** Composition stays fluid; applications own page grids and breakpoints. */
export const layout = stylex.create({
  stack: { gap: space[16], display: 'flex', flexDirection: 'column', minInlineSize: 0, },
  cluster: { gap: density.inlineGap, alignItems: 'center', display: 'flex', flexWrap: 'wrap', minInlineSize: 0, },
  measure: { maxInlineSize: measure.body, overflowWrap: 'anywhere' },
  /** Children of a stack or grid: no block margins and no minimum width, so they can shrink. */
  flow: { marginBlock: 0, minInlineSize: 0 },
  borderBox: { boxSizing: 'border-box' },
})

export const a11y = stylex.create({
  visuallyHidden: {
    margin: `calc(-1 * ${size[1]})`,
    padding: 0,
    borderWidth: 0,
    overflow: 'hidden',
    blockSize: size[1],
    clip: 'rect(0, 0, 0, 0)',
    clipPath: 'inset(50%)',
    inlineSize: size[1],
    position: 'absolute',
    whiteSpace: 'nowrap',
  },
  focusRing: {
    outlineColor: {
      default: null,
      ':focus-visible': color.focusRing,
    },
    outlineOffset: { default: null, ':focus-visible': border.focusOffset },
    outlineStyle: {
      default: null,
      ':focus-visible': "solid",
    },
    outlineWidth: {
      default: null,
      ':focus-visible': border.focusRing,
    },
  },
  /** Script-moved focus on headings and `main` announces the page without drawing a control ring. */
  programmaticFocus: { outlineStyle: { default: null, ':focus-visible': 'none' } },
})

/** Native controls take the surrounding type instead of the browser's control font. */
export const control = stylex.create({
  inherit: {
    color: 'inherit',
    fontFamily: 'inherit',
    fontSize: 'inherit',
    fontStyle: 'inherit',
    fontWeight: 'inherit',
    lineHeight: 'inherit',
  },
})

/** Every link answers the pointer with a deeper colour and a thicker underline, and shows the focus ring. */
export const link = stylex.create({
  base: {
    color: { default: color.textLink, ':hover': color.textLinkHover },
    outlineColor: {
      default: null,
      ':focus-visible': color.focusRing,
    },
    outlineOffset: { default: null, ':focus-visible': border.focusOffset },
    outlineStyle: {
      default: null,
      ':focus-visible': "solid",
    },
    outlineWidth: {
      default: null,
      ':focus-visible': border.focusRing,
    },
    textDecorationThickness: { default: null, ':hover': decoration.hover },
    textUnderlineOffset: decoration.offset,
  },
  /** Ink text with a muted underline that turns to the link colour on hover, for titles in lists and rows. */
  quiet: {
    color: { default: color.textInk, ":focus-visible": color.textLink, ":hover": color.textLink },
    textDecorationColor: { default: color.textMuted, ":focus-visible": "currentColor", ":hover": "currentColor" },
    textDecorationLine: "underline",
    textDecorationThickness: border.static,
  },
  /** A list entry's title: a quiet link in the label weight. */
  title: { fontWeight: fontWeight.semibold, lineHeight: lineHeight.normal, overflowWrap: "anywhere" },
  /** Links styled entirely by their component; only the focus ring is shared. */
  plain: {
    textDecoration: 'none',
    color: 'inherit',
    outlineColor: {
      default: null,
      ':focus-visible': color.focusRing,
    },
    outlineOffset: { default: null, ':focus-visible': border.focusOffset },
    outlineStyle: {
      default: null,
      ':focus-visible': "solid",
    },
    outlineWidth: {
      default: null,
      ':focus-visible': border.focusRing,
    },
  },
})

export const transition = stylex.create({
  quick: { transitionDuration: { default: motion.quick, [breakpoints.reducedMotion]: motion.instant } },
  standard: { transitionDuration: { default: motion.standard, [breakpoints.reducedMotion]: motion.instant } },
  deliberate: { transitionDuration: { default: motion.deliberate, [breakpoints.reducedMotion]: motion.instant } },
})
