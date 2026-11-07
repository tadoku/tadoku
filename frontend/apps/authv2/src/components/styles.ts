import * as stylex from '@stylexjs/stylex'
import { linkStyles } from 'paper-ui'
import {
  border,
  breakpoints,
  color,
  density,
  fluid,
  font,
  fontSize,
  fontWeight,
  lineHeight,
  page,
  ratio,
  size,
  space,
  type,
  zIndex,
} from 'paper-ui/tokens.stylex'

const gutter = `clamp(${space[16]}, ${fluid[4]}, ${space[32]})`

/** Every plain link in the account pages: Paper's link colour and underline. */
export const textLink = linkStyles.base

/** Authv2's shell: the header matches Webv3's height and content column, so moving between hosts keeps the brand in place. */
export const shellStyles = stylex.create({
  shell: { display: 'flex', flexDirection: 'column', minHeight: '100vh' },
  skip: {
    insetBlockStart: space[12],
    insetInlineStart: space[12],
    position: 'fixed',
    transform: { default: `translateY(${ratio.offscreen})`, ':focus': 'none' },
    zIndex: zIndex.modal,
  },
  header: {
    paddingBlock: space[8],
    paddingInline: `max(${gutter}, calc((100% - ${page.contentWidth}) / 2))`,
    alignItems: 'center',
    backgroundColor: color.surfacePaper,
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: 'solid',
    borderBlockEndWidth: border.static,
    boxSizing: 'border-box',
    columnGap: { default: space[24], [breakpoints.phone]: space[12] },
    display: 'flex',
    flexWrap: { default: 'wrap', [breakpoints.phone]: 'nowrap' },
    justifyContent: 'space-between',
    rowGap: space[12],
    minHeight: size[61],
  },
  brand: { gap: space[14], alignItems: 'center', color: color.textMuted, display: 'inline-flex', textDecorationLine: 'none', },
  lockup: { gap: space[7], alignItems: 'center', display: 'flex', },
  mark: { blockSize: 'auto', display: 'block', inlineSize: { default: size[35], [breakpoints.phone]: size[28] } },
  wordmark: { blockSize: 'auto', display: 'block', inlineSize: { default: size[124], [breakpoints.phone]: size[100] } },
  brandLabel: {
    borderInlineStartColor: color.ruleDefault,
    borderInlineStartStyle: 'solid',
    borderInlineStartWidth: border.static,
    display: { default: null, [breakpoints.phone]: 'none' },
    fontSize: fontSize[14],
    fontWeight: fontWeight.semibold,
    lineHeight: lineHeight[20],
    paddingInlineStart: space[14],
  },
  session: { margin: 0, gap: { default: space[16], [breakpoints.phone]: space[12] }, alignItems: 'center', display: 'flex', minInlineSize: 0, },
  sessionName: {
    overflow: 'hidden',
    color: color.textInk,
    fontSize: fontSize[14],
    maxInlineSize: { default: size[192], [breakpoints.phone]: 'none' },
    minInlineSize: 0,
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  /** The session row reads as a name and a plain action of equal size and weight, aligned to the content edge. */
  logout: {
    borderWidth: 0,
    paddingInline: 0,
    textDecoration: 'underline',
    borderBlockEndWidth: 0,
    flexShrink: { default: null, [breakpoints.phone]: 0 },
    fontSize: fontSize[14],
    fontWeight: fontWeight.regular,
    minBlockSize: size[24],
    paddingBlockEnd: 0,
    paddingBlockStart: 0,
    whiteSpace: { default: null, [breakpoints.phone]: 'nowrap' },
  },
  main: {
    flex: '1',
    paddingInline: gutter,
    boxSizing: 'border-box',
    outlineStyle: { default: null, ':focus': 'none' },
    paddingBlockEnd: space[72],
    paddingBlockStart: `clamp(${space[32]}, ${fluid.vh7}, ${space[72]})`,
    width: '100%',
  },
})

/** Single-task cards: log in, create an account, recovery and status pages. */
export const cardStyles = stylex.create({
  page: { marginInline: 'auto', width: `min(100%, ${size[440]})` },
  card: {
    paddingBlock: { default: `clamp(${space[24]}, ${fluid[6]}, ${space[36]})`, [breakpoints.phone]: space[24] },
    paddingInline: { default: `clamp(${space[24]}, ${fluid[6]}, ${space[36]})`, [breakpoints.phone]: space[20] },
  },
  header: { marginBlockEnd: space[28] },
  footer: { marginInline: 0, color: color.textMuted, fontSize: fontSize[15], marginBlockEnd: 0, marginBlockStart: space[24], textAlign: 'center', },
  confirmation: { gap: space[16], display: 'grid', },
  spaced: { marginBlockEnd: space[28] },
  confirmationText: { margin: 0, lineHeight: lineHeight.airy, },
})

export const formStyles = stylex.create({
  form: { gap: space[20], display: 'grid', minInlineSize: 0, },
  compact: { gap: space[16] },
  aside: { fontSize: fontSize[14], justifySelf: 'center' },
  actions: { alignItems: 'center', columnGap: space[20], display: 'flex', flexWrap: 'wrap', rowGap: space[12] },
  action: { flexShrink: 0 },
  noticeLink: { whiteSpace: 'nowrap' },
})

/** Settings: headed groups of rows; the row the URL opens shows its editor on the raised surface. */
export const settingsStyles = stylex.create({
  page: { marginInline: 'auto', width: `min(100%, ${size[760]})` },
  header: { marginBlockEnd: space[32] },
  backLink: { gap: space[8] },
  notice: { marginBlockEnd: space[20] },
  group: { marginBlockStart: { default: space[32], ':first-of-type': 0 } },
  groupTitle: {
    marginInline: 0,
    fontFamily: font.display,
    fontSize: type.componentSize,
    lineHeight: type.componentLineHeight,
    marginBlockEnd: space[12],
    marginBlockStart: 0,
  },
  row: {
    paddingBlock: space[16],
    paddingInline: space[20],
    alignItems: 'center',
    borderBlockStartColor: color.ruleSubtle,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: { default: border.static, ':first-child': 0 },
    columnGap: space[24],
    display: 'grid',
    gridTemplateColumns: { default: `${size[160]} minmax(0, 1fr) auto`, [breakpoints.narrow]: 'minmax(0, 1fr) auto' },
    rowGap: space[4],
  },
  open: {
    alignItems: 'start',
    backgroundColor: color.surfaceRaised,
    boxShadow: `inset ${border.accentRail} 0 0 ${color.focusRing}`,
    gridTemplateColumns: { default: `${size[160]} minmax(0, 1fr)`, [breakpoints.narrow]: 'minmax(0, 1fr) auto' },
  },
  label: {
    margin: 0,
    gridColumn: { default: null, [breakpoints.narrow]: '1 / -1' },
    color: color.textMuted,
    fontFamily: font.body,
    fontSize: type.metadataSize,
    fontWeight: fontWeight.regular,
    lineHeight: type.metadataLineHeight,
  },
  summary: { minInlineSize: 0 },
  value: {
    margin: 0,
    color: color.textInk,
    fontSize: density.bodySize,
    lineHeight: density.bodyLineHeight,
    overflowWrap: 'anywhere',
    whiteSpace: 'break-spaces',
  },
  hint: { margin: 0, color: color.textMuted, fontSize: type.metadataSize, lineHeight: type.metadataLineHeight, },
  editor: {
    gridColumn: { default: null, [breakpoints.narrow]: '1 / -1' },
    maxInlineSize: { default: size[416], [breakpoints.narrow]: 'none' },
    minInlineSize: 0,
    overflowWrap: 'anywhere',
  },
  /** Deletion stays off the rows: one quiet link below the groups. */
  delete: { marginInline: 0, marginBlockEnd: 0, marginBlockStart: space[40], },
  deleteLink: { color: color.statusDanger },
  saved: {
    margin: 0,
    color: color.statusSuccess,
    fontSize: type.metadataSize,
    fontWeight: fontWeight.regular,
    lineHeight: type.metadataLineHeight,
  },
  rows: { padding: 0 },
  /** The email hint holds the stored address, whose raw spaces would otherwise hang past the row. */
  storedHint: { whiteSpace: 'break-spaces' },
})
