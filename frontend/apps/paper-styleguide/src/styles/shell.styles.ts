import * as stylex from '@stylexjs/stylex'
import { border, breakpoints, color, density, fluid, font, fontSize, fontWeight, lineHeight, motion, ratio, size, space, tracking, type, zIndex } from 'paper-ui/tokens.stylex'

/** The sticky docs navbar's height: one control row plus its lower rule. */
const navbarBlockSize = `calc(${density.controlHeight} + ${border.static})`

export const shellStyles = stylex.create({
  skipLink: {
    paddingBlock: density.fieldPaddingBlock,
    paddingInline: density.fieldPaddingInline,
    transition: { default: null, [breakpoints.motionOk]: `transform ${motion.quick}` },
    backgroundColor: color.actionDefault,
    color: color.textInverse,
    insetBlockStart: density.inlineGap,
    insetInlineStart: density.inlineGap,
    position: "fixed",
    transform: { default: `translateY(${ratio.offscreen})`, ":focus": "translateY(0)" },
    zIndex: zIndex.skipLink,
  },
  docsShell: {
    display: { default: "grid", [breakpoints.belowWide]: "block" },
    gridTemplateColumns: `${size[288]} minmax(0, 1fr)`,
    gridTemplateRows: "auto 1fr",
    minBlockSize: "100vh",
  },
  docsNavbar: {
    gridColumnEnd: '-1',
    gridColumnStart: '1',
    insetBlockStart: 0,
    position: "sticky",
    zIndex: zIndex.sticky,
  },
  docsWordmark: {
    gap: space[7],
    alignItems: "center",
    display: "flex",
  },
  docsWordmark__mark: {
    blockSize: size[36],
    display: "block",
    inlineSize: size[36],
  },
  docsWordmark__image: {
    blockSize: size[24],
    display: "block",
    inlineSize: size[128],
  },
  docsWordmark__mark_reversed: {
    display: "none",
  },
  docsWordmark__image_reversed: {
    display: "none",
  },
  mobileNavTrigger: {
    paddingInline: { default: null, [breakpoints.tiny]: 0 },
    display: { default: "none", [breakpoints.belowWide]: "inline-flex" },
    inlineSize: { default: null, [breakpoints.tiny]: density.controlHeight },
    justifyContent: { default: null, [breakpoints.tiny]: "center" },
  },
  docsSidebar: {
    paddingInline: space[16],
    alignSelf: "start",
    backgroundColor: color.surfacePaper,
    blockSize: `calc(100vh - ${navbarBlockSize})`,
    borderInlineEndColor: color.ruleSubtle,
    borderInlineEndStyle: "solid",
    borderInlineEndWidth: border.static,
    display: { default: null, [breakpoints.belowWide]: "none" },
    insetBlockStart: navbarBlockSize,
    position: "sticky",
    overflowY: "auto",
    paddingBottom: space[48],
    paddingTop: space[20],
  },
  searchResults: {
    margin: 0,
    padding: 0,
    listStyle: "none",
  },
  docsMain: {
    paddingInline: `clamp(${space[16]}, ${fluid[5]}, ${space[80]})`,
    minInlineSize: 0,
    paddingBlockEnd: size[112],
    paddingBlockStart: { default: `clamp(${space[32]}, ${fluid[5]}, ${space[80]})`, [breakpoints.md]: space[32] },
  },
  catalogueIndex: {
    marginInline: "auto",
    maxInlineSize: size[1216],
  },
  notFound: {
    marginInline: "auto",
    marginBlockEnd: space[64],
    maxInlineSize: size[1216],
  },
  catalogueIndex__hero: {
    marginBlockEnd: { default: space[64], [breakpoints.md]: space[40] },
    maxInlineSize: size[768],
  },
  designHistory__links: {
    padding: 0,
    gap: space[16],
    listStyle: "none",
    marginInline: 0,
    display: "grid",
    gridTemplateColumns: `repeat(auto-fit, minmax(min(100%, ${size[256]}), 1fr))`,
    marginBottom: 0,
    marginTop: space[20],
  },
  designHistory__link: {
    textDecoration: "none",
    blockSize: "100%",
    color: color.textInk,
    display: "block",
  },
  documentSummary: {
    color: color.textMuted,
    fontSize: type.componentSize,
    lineHeight: lineHeight.loose,
  },
  eyebrow: {
    margin: 0,
    color: color.textLink,
    fontSize: type.metadataSize,
    fontWeight: fontWeight.semibold,
    letterSpacing: tracking.widest,
    textTransform: "uppercase",
  },
  documentCardGrid: {
    gap: { default: space[16], [breakpoints.md]: space[12] },
    display: "grid",
    gridTemplateColumns: `repeat(auto-fit, minmax(min(100%, ${size[256]}), 1fr))`,
  },
  documentCard: {
    padding: { default: null, [breakpoints.md]: space[14] },
    textDecoration: "none",
    color: color.textInk,
    minBlockSize: { default: size[176], [breakpoints.md]: 0 },
  },
  documentLayout: {
    gap: `clamp(${space[32]}, ${fluid[5]}, ${space[80]})`,
    display: "grid",
    gridTemplateColumns: { default: `minmax(0, ${size[768]}) ${size[192]}`, [breakpoints.belowWide]: `minmax(0, ${size[768]})` },
    justifyContent: "center",
  },
  documentHero: {
    marginBlockEnd: space[64],
  },
  documentSection: {
    maxInlineSize: size[720],
    scrollMarginBlockStart: space[96],
  },
  documentSection_wide: {
    maxInlineSize: "none",
  },
  metadataStrip: {
    gap: space[16],
    borderBlockStartColor: color.ruleSubtle,
    borderBlockStartStyle: "solid",
    borderBlockStartWidth: border.static,
    display: "grid",
    gridTemplateColumns: { default: "repeat(3, minmax(0, 1fr))", [breakpoints.md]: "1fr 1fr" },
    marginBlockStart: space[32],
    paddingBlockStart: space[16],
  },
  documentListGrid: {
    gap: space[28],
    display: "grid",
    gridTemplateColumns: "minmax(0, 1fr)",
  },
  foundationSpecimen: {
    overflow: "hidden",
    marginBlockStart: space[16],
  },
  foundationSpecimen__lead: {
    marginBlockEnd: space[20],
    marginBlockStart: 0,
    maxInlineSize: size[672],
  },
  foundationPrinciples__grid: {
    gap: space[16],
    display: "grid",
    gridTemplateColumns: `repeat(auto-fit, minmax(min(100%, ${size[176]}), 1fr))`,
  },
  foundationDensityGrid: {
    gap: space[16],
    display: "grid",
    gridTemplateColumns: `repeat(auto-fit, minmax(min(100%, ${size[176]}), 1fr))`,
  },
  foundationElevations: {
    gap: space[16],
    display: "grid",
    gridTemplateColumns: `repeat(auto-fit, minmax(min(100%, ${size[176]}), 1fr))`,
  },
  foundationSwatches: {
    margin: 0,
    padding: 0,
    gap: space[16],
    listStyle: "none",
    display: "grid",
    gridTemplateColumns: `repeat(auto-fit, minmax(min(100%, ${size[144]}), 1fr))`,
  },
  foundationSwatch: {
    borderColor: color.ruleDefault,
    borderStyle: "solid",
    borderWidth: border.static,
    blockSize: size[72],
    display: "block",
  },
  foundationSwatch_1: {
    backgroundColor: color.surfaceCanvas,
  },
  foundationSwatch_2: {
    backgroundColor: color.surfacePaper,
  },
  foundationSwatch_3: {
    backgroundColor: color.textInk,
  },
  foundationSwatch_4: {
    backgroundColor: color.actionDefault,
  },
  foundationSwatch_5: {
    backgroundColor: color.statusSuccess,
  },
  foundationSwatch_6: {
    backgroundColor: color.chart1,
  },
  foundationTypeScale: {
    display: "grid",
  },
  foundationControlHeight: {
    borderColor: color.ruleFieldEdge,
    borderStyle: "solid",
    borderWidth: border.fieldEdge,
    placeItems: "center",
    blockSize: density.controlHeight,
    display: "grid",
    minInlineSize: size[128],
  },
  foundationLayout__flow: {
    gap: space[16],
    display: "grid",
    gridTemplateColumns: `minmax(${size[192]}, 2fr) minmax(${size[160]}, 1fr)`,
  },
  foundationLayout__cluster: {
    gap: density.inlineGap,
    alignContent: "start",
    display: "flex",
    flexWrap: "wrap",
  },
  foundationLayout__overflow: {
    padding: space[12],
    borderBlockEndColor: color.ruleActionEdge,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.actionEdge,
    marginBlockStart: space[16],
    whiteSpace: "nowrap",
    overflowX: "auto",
  },
  foundationIcons: {
    gap: space[32],
    display: "flex",
    flexWrap: "wrap",
  },
  foundationMotion: {
    gap: space[16],
    display: "grid",
  },
  foundationMotion__track: {
    padding: space[4],
  },
  foundationBrand: {
    gap: space[32],
    alignItems: "end",
    display: "grid",
    gridTemplateColumns: `minmax(${size[128]}, 0.7fr) minmax(${size[224]}, 2fr)`,
  },
  localOutline: {
    alignSelf: "start",
    display: { default: null, [breakpoints.belowWide]: "none" },
    insetBlockStart: space[96],
    position: "sticky",
  },
  exampleCanvas: {
    containerType: "inline-size",
    minInlineSize: 0,
  },
  componentWorkbench: {
    padding: 0,
    overflow: "clip",
  },
  componentWorkbench__tabs: {
    paddingInline: space[12],
    minInlineSize: 0,
  },
  componentWorkbench__panel: {
    padding: 0,
    minInlineSize: 0,
  },
  codeView: {
    padding: space[20],
  },
  workbenchReferenceGrid: {
    padding: space[20],
    gap: space[24],
    display: "grid",
    gridTemplateColumns: "minmax(0, 1fr)",
  },
  codeView__heading: {
    gap: space[16],
    alignItems: { default: "start", [breakpoints.md]: "stretch" },
    display: "flex",
    flexDirection: { default: null, [breakpoints.md]: "column" },
    justifyContent: "space-between",
  },
  codeCopy: {
    gap: space[6],
    display: "grid",
    justifyItems: { default: "end", [breakpoints.md]: "start" },
    minInlineSize: { default: size[192], [breakpoints.md]: 0 },
  },
  codeCopy__status: {
    color: color.textMuted,
    fontSize: type.metadataSize,
    minBlockSize: type.metadataLineHeight,
    textAlign: "end",
  },
  paperFixtureStage: {
    padding: space[8],
    color: color.textInk,
    minBlockSize: "100%",
  },
  paperFixtureRow: {
    gap: density.inlineGap,
    alignItems: "center",
    display: "flex",
    flexWrap: "wrap",
  },
  paperFixtureStack: {
    gap: space[16],
    display: "grid",
  },
  workbenchPropsTable: {
    tableLayout: "fixed",
  },
  workbenchPropDefault: {
    marginInline: 0,
    fontSize: fontSize[13],
    marginBottom: space[8],
    marginTop: 0,
  },
  workbenchPropType: {
    display: "block",
    fontSize: fontSize[13],
    fontWeight: "normal",
    marginBlockStart: space[6],
  },
  workbenchApi: {
    padding: space[20],
    gap: space[32],
    display: "grid",
  },
  paperCodeBlock: {
    font: `0.8125rem/1.65 ${font.monospace}`,
    padding: space[16],
    overflow: "auto",
    maxInlineSize: "100%",
    whiteSpace: "pre",
  },
  canvasControls: {
    padding: space[16],
    gap: space[12],
    alignItems: "end",
    display: "grid",
    gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
  },
  canvasControls__fixture: {
    gridColumnEnd: '-1',
    gridColumnStart: '1',
    minInlineSize: 0,
  },
  canvasControls__viewport: {
    gridColumnEnd: '-1',
    gridColumnStart: '1',
    minInlineSize: 0,
  },
  canvasControls__title: {
    overflow: "hidden",
    blockSize: size[1],
    clipPath: "inset(50%)",
    inlineSize: size[1],
    position: "absolute",
    whiteSpace: "nowrap",
  },
  canvasControls__theme: {
    minInlineSize: 0,
  },
  canvasControls__density: {
    minInlineSize: 0,
  },
  canvasStatus: {
    margin: 0,
    paddingInline: space[16],
    color: color.textMuted,
    paddingBottom: space[12],
    paddingTop: 0,
  },
  exampleCanvas__stage: {
    padding: { default: space[8], [breakpoints.md]: space[4] },
    overflow: "auto",
    maxInlineSize: "100%",
  },
  previewSpecimen: {
    margin: space[32],
    padding: space[32],
  },
  previewBreakpoint: {
    borderBlockStartColor: color.ruleSubtle,
    borderBlockStartStyle: "solid",
    borderBlockStartWidth: border.static,
    color: color.statusInformation,
    fontWeight: fontWeight.semibold,
    marginBlockStart: space[24],
    paddingBlockStart: space[16],
  },
  previewBreakpoint__wide: {
    display: { default: "none", [breakpoints.aboveMd]: "inline" },
  },
  searchField: {
    margin: 0,
  },
  searchCount: {
    margin: 0,
    paddingBlock: space[12],
    color: color.textMuted,
  },
  searchResult: {
    gap: space[16],
    paddingBlock: space[14],
    paddingInline: 0,
    textDecoration: "none",
    transition: { default: null, [breakpoints.motionOk]: `background-color ${motion.quick}` },
    backgroundColor: { default: null, ":hover": color.surfaceRaised },
    borderBlockStartColor: color.ruleSubtle,
    borderBlockStartStyle: "solid",
    borderBlockStartWidth: border.static,
    color: color.textInk,
    display: "flex",
    justifyContent: "space-between",
  },
  lifecycleBadge: {
    alignSelf: "start",
    color: color.statusInformation,
    fontSize: type.metadataSize,
  },
  textLink: {
    color: color.textLink,
  },
  previewBreakpoint__narrow: {
    display: { default: null, [breakpoints.aboveMd]: "none" },
  },
  shellSearchTrigger__label: {
    display: { default: null, [breakpoints.md]: "none" },
  },
  shellSearchTrigger__shortcut: {
    display: { default: null, [breakpoints.md]: "none" },
  },
  mobileNavTrigger__label: {
    display: { default: null, [breakpoints.tiny]: "none" },
  },
  catalogueIndex__section: {},
  designHistory: {},
  docsWordmark__image_accent: {},
  docsWordmark__mark_ink: {},
  documentListGroup: {},
  metadataList: {},
  paperAccentRail: {},
  token: {},
})

