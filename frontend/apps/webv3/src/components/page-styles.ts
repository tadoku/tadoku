import * as stylex from '@stylexjs/stylex'
import { linkStyles } from 'paper-ui'
import { border, breakpoints, color, font, fontSize, fontWeight, lineHeight, size, space, type } from 'paper-ui/tokens.stylex'

/** Layout shared by Webv3's record and form pages. */
export const pageStyles = stylex.create({
  /** A record or form reads as one column of this width. */
  column: { maxInlineSize: size[768] },
  section: { marginBlockStart: space[32] },
  muted: {
    marginInline: 0,
    color: color.textMuted,
    fontSize: type.metadataSize,
    lineHeight: lineHeight.relaxed,
    marginBlockEnd: 0,
    marginBlockStart: space[8],
  },
  form: { minInlineSize: 0 },
  /** Two fields side by side, stacked on a phone. */
  formRow: {
    gap: space[20],
    alignItems: 'start',
    display: 'grid',
    gridTemplateColumns: { default: 'minmax(0, 1fr) minmax(0, 1fr)', [breakpoints.phone]: 'minmax(0, 1fr)' },
  },
  /** A titled group of fields under a rule, inside a longer form. */
  formGroup: {    borderWidth: 0,
    gap: space[16],
    marginInline: 0,
    paddingInline: 0,
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
 boxSizing: "border-box",
    display: 'grid',
    marginBlockEnd: 0,
    marginBlockStart: space[8],
    minInlineSize: 0,
    paddingBlockEnd: 0,
    paddingBlockStart: space[24],
  },
  formGroupLegend: {
    padding: 0,
    color: color.textInk,
    float: 'left',
    fontFamily: font.display,
    fontSize: fontSize[22],
    fontWeight: fontWeight.bold,
    lineHeight: lineHeight[30],
    width: '100%',
  },
  flush: { margin: 0 },
  actions: { alignItems: 'center', columnGap: space[20], display: 'flex', flexWrap: 'wrap', marginBlockStart: space[4], rowGap: space[12] },
})

/** The standings pages' frame: a main column beside a narrow aside, a bordered sheet and its parts. */
export const standingsLayout = stylex.create({
  header: { marginBlockEnd: space[20] },
  notice: { marginBlockStart: space[20] },
  workspace: {
    gap: { default: space[32], [breakpoints.xl]: space[24] },
    alignItems: 'start',
    display: 'grid',
    gridTemplateColumns: {
      default: `minmax(0, 1fr) ${size[256]}`,
      [breakpoints.xl]: `minmax(0, 1fr) ${size[224]}`,
      [breakpoints.tablet]: 'minmax(0, 1fr)',
    },
    marginBlockStart: space[20],
  },
  column: { minInlineSize: 0 },
  sheet: {    borderColor: color.ruleDefault,
    borderStyle: 'solid',
    borderWidth: border.static,
    backgroundColor: color.surfacePaper,
 boxSizing: "border-box",
    minInlineSize: 0,
  },
  /** Inset the tabs, not the tab bar, so its rule runs edge to edge across the sheet. */
  sheetTabs: { paddingInline: { default: space[20], [breakpoints.sm]: space[16] } },
  context: {
    marginBlock: 0,
    paddingInline: { default: space[20], [breakpoints.sm]: space[16] },
    paddingBlockEnd: { default: space[20], [breakpoints.sm]: space[16] },
    paddingBlockStart: space[20],
    rowGap: 0,
  },
  contextTitle: { fontSize: { default: fontSize[22], [breakpoints.sm]: fontSize[18] } },
  empty: {
    margin: 0,
    paddingBlock: space[24],
    paddingInline: { default: space[20], [breakpoints.sm]: space[16] },
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
  },
  cell: { paddingInline: { default: space[20], [breakpoints.sm]: space[16] } },
  pagination: {
    paddingBlock: space[12],
    paddingInline: { default: space[20], [breakpoints.sm]: space[16] },
    alignItems: 'center',
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
    columnGap: space[16],
    display: 'flex',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    rowGap: space[8],
  },
  paginationList: { flexWrap: { default: null, [breakpoints.sm]: 'wrap' } },
  paginationSummary: { margin: 0, color: color.textMuted, fontSize: type.metadataSize, },
  tieNote: {
    margin: 0,
    paddingBlock: space[12],
    paddingInline: space[20],
    borderBlockStartColor: color.ruleSubtle,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
    color: color.textMuted,
    fontSize: type.metadataSize,
  },
  plainAbbr: { textDecorationLine: 'none' },
  aside: { fontSize: fontSize[14], lineHeight: lineHeight.airy, overflowWrap: 'anywhere', paddingBlockStart: space[4] },
  asideSection: {
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: { default: border.static, ':first-child': 0 },
    marginBlockStart: { default: space[28], ':first-child': 0 },
    paddingBlockStart: { default: space[24], ':first-child': 0 },
  },
  asideHeading: {
    marginInline: 0,
    fontFamily: font.display,
    fontSize: type.componentSize,
    fontWeight: fontWeight.bold,
    lineHeight: type.componentLineHeight,
    marginBlockEnd: space[12],
    marginBlockStart: 0,
  },
  asideParagraph: { marginInline: 0, marginBlockEnd: space[16], marginBlockStart: 0, },
  asideNote: { color: color.textMuted },
  description: { whiteSpace: 'pre-line' },
  /** The aside's closing link stands apart from the last fact rather than reading as part of its value. */
  asideLink: { display: 'inline-block', marginBlockStart: space[16] },
})

/** Titles and metadata lines shared by the profile's lists of logs and contests. */
export const entryStyles = stylex.create({
  meta: { marginInline: 0, color: color.textMuted, fontSize: type.metadataSize, marginBlockEnd: 0, marginBlockStart: space[6], },
  contestList: { margin: 0, padding: 0, listStyle: 'none', },
  contestItem: {
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: 'solid',
    borderBlockEndWidth: { default: border.static, ':last-child': 0 },
    display: 'grid',
    overflowWrap: 'anywhere',
    paddingBlockEnd: { default: space[12], ':last-child': 0 },
    paddingBlockStart: { default: space[12], ':first-child': 0 },
  },
})

/** A list entry's title link: ink with a muted underline that takes the link colour on hover. */
export const entryTitle = [linkStyles.base, linkStyles.quiet, linkStyles.title] as const

/** Manual, CMS pages and blog, following the playground's guide article. */
export const contentStyles = stylex.create({
  article: { maxInlineSize: size[720], minInlineSize: 0 },
})
