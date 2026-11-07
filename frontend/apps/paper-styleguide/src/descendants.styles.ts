import * as stylex from '@stylexjs/stylex'
import { border, breakpoints, color, decoration, density, font, fontSize, fontWeight, lineHeight, measure, motion, size, space, type } from 'paper-ui/tokens.stylex'

export const descendantStyles = stylex.create({
  canvasControlsU_uPaperField: {
    minInlineSize: 0,
  },
  localOutlineU_uOl: {
    margin: 0,
    padding: 0,
    listStyle: "none",
  },
  catalogueIndexU_uSection: {
    marginBlockStart: { default: null, ":not(:first-child)": space[56], [breakpoints.md]: { default: null, ":not(:first-child)": space[40] } },
  },
  catalogueIndexU_uSectionU_uH2: {
    marginBlockEnd: space[16],
    marginBlockStart: 0,
  },
  designHistoryU_uP: {
    color: color.textMuted,
    maxInlineSize: size[736],
  },
  designHistoryU_uLinksU_uSpan: {
    color: color.textMuted,
    display: "block",
    marginBlockStart: space[6],
  },
  catalogueIndexU_uHeroU_uH1: {
    marginBlockEnd: space[16],
    marginBlockStart: space[5],
  },
  documentHeroU_uH1: {
    marginBlockEnd: space[16],
    marginBlockStart: space[5],
  },
  notFoundU_uH1: {
    marginBlockEnd: space[16],
    marginBlockStart: space[5],
  },
  catalogueIndexU_uHeroU_uP: {
    color: { default: null, ":last-child": color.textMuted },
    fontSize: { default: null, ":last-child": type.componentSize },
    lineHeight: { default: null, ":last-child": lineHeight.loose },
  },
  documentCardU_uSmall: {
    color: color.textMuted,
    float: "inline-end",
  },
  documentCardU_uP: {
    color: color.textMuted,
    display: { default: null, [breakpoints.md]: "none" },
    marginBlockEnd: 0,
  },
  documentSection: {
    marginBlockStart: { default: null, ":not(:first-child)": space[48] },
  },
  metadataStripU_uDt: {
    color: color.textMuted,
    fontSize: type.metadataSize,
  },
  metadataListU_uDt: {
    color: color.textMuted,
    fontSize: type.metadataSize,
  },
  metadataStripU_uDd: {
    marginInline: 0,
    marginBottom: 0,
    marginTop: space[3],
  },
  metadataListU_uDd: {
    marginInline: 0,
    marginBottom: 0,
    marginTop: space[3],
  },
  metadataListU_uDiv: {
    gap: { default: space[16], [breakpoints.md]: 0 },
    paddingBlock: space[12],
    borderBlockEndColor: color.ruleSubtle,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.static,
    display: "grid",
    gridTemplateColumns: { default: `${size[144]} minmax(0, 1fr)`, [breakpoints.md]: "1fr" },
  },
  docsShellU_uCode: {
    color: color.textLink,
    overflowWrap: "anywhere",
  },
  documentListGroupU_uH3: {
    marginBlockEnd: { default: space[8] },
    marginBlockStart: { default: 0 },
  },
  documentListGroupU_uUl: {
    marginBlock: 0,
    paddingInlineStart: space[20],
  },
  documentListGroupU_uLi: {
    marginBlockStart: { default: null, ":not(:first-child)": space[7] },
  },
  foundationPrinciplesU_uGridU_uDiv: {
    gap: space[6],
    alignContent: "start",
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: "solid",
    borderBlockStartWidth: border.static,
    display: "grid",
    paddingBlockStart: space[12],
  },
  foundationDensityGridU_uDiv: {
    gap: space[6],
    alignContent: "start",
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: "solid",
    borderBlockStartWidth: border.static,
    display: "grid",
    paddingBlockStart: space[12],
  },
  foundationPrinciplesU_uGridU_uSpan: {
    color: color.textMuted,
  },
  foundationBrandU_uFigcaption: {
    color: color.textMuted,
  },
  foundationSwatchesU_uLi: {
    gap: space[4],
    display: "grid",
    minInlineSize: 0,
  },
  foundationTypeScaleU_uDiv: {
    gap: space[16],
    paddingBlock: space[12],
    alignItems: "baseline",
    borderBlockEndColor: color.ruleSubtle,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.static,
    display: "grid",
    gridTemplateColumns: "minmax(0, 1fr) auto",
  },
  foundationDensityGridU_uCode: {
    display: "block",
  },
  foundationElevationsU_uCode: {
    display: "block",
  },
  foundationIconsU_uCode: {
    gridColumn: 2,
    display: "block",
  },
  foundationLayoutU_uFlowU_uDiv: {
    padding: space[16],
    borderColor: color.ruleDefault,
    borderStyle: "solid",
    borderWidth: border.static,
  },
  foundationLayoutU_uFlowU_uP: {
    marginBlockEnd: 0,
    maxInlineSize: measure.narrow,
  },
  foundationLayoutU_uClusterU_uSpan: {
    paddingBlock: space[5],
    paddingInline: space[9],
  },
  foundationShapesU_uPaperAccentRail: {
    gap: space[4],
    paddingBlock: space[12],
    paddingInline: space[16],
    display: "grid",
  },
  foundationShapesU_uDl: {
    gap: space[16],
    display: "grid",
    gridTemplateColumns: `repeat(auto-fit, minmax(min(100%, ${size[176]}), 1fr))`,
    marginBlockEnd: 0,
  },
  foundationShapesU_uDd: {
    marginInline: 0,
    marginBottom: 0,
    marginTop: space[4],
  },
  foundationElevationsU_uAny: {
    alignContent: "space-between",
    display: "grid",
    minBlockSize: size[112],
  },
  foundationIconsU_uDiv: {
    alignItems: "center",
    columnGap: space[12],
    display: "grid",
    gridTemplateColumns: "auto 1fr",
    rowGap: space[4],
  },
  foundationIconsU_uSpan: {
    gridColumn: 2,
  },
  foundationMotionU_uDiv: {
    gap: space[12],
    alignItems: "center",
    display: "grid",
    gridTemplateColumns: `minmax(${size[144]}, 1fr) minmax(${size[128]}, 2fr) auto auto`,
  },
  foundationMotionU_uTrackU_uSpan: {
    backgroundColor: color.actionDefault,
    blockSize: size[20],
    display: "block",
    inlineSize: size[20],
    transform: `translateX(calc(100% + ${space[32]}))`,
    transitionProperty: "transform",
    transitionTimingFunction: "ease-out",
  },
  foundationMotionU_uTrack_quickU_uSpan: {
    transitionDuration: motion.quick,
  },
  foundationMotionU_uTrack_standardU_uSpan: {
    transitionDuration: motion.standard,
  },
  foundationMotionU_uTrack_deliberateU_uSpan: {
    transitionDuration: motion.deliberate,
  },
  foundationBrandU_uFigure: {
    margin: 0,
    padding: space[16],
    gap: space[12],
    color: color.textInk,
    display: "grid",
  },
  foundationBrandU_uImg: {
    display: "block",
    maxBlockSize: size[80],
  },
  localOutlineU_uH2: {
    marginBlockStart: 0,
    marginInlineStart: density.fieldPaddingInline,
  },
  documentSectionU_uH2: {
    marginBlockEnd: space[20],
    marginBlockStart: 0,
  },
  documentSectionU_uH3: {
    marginBlockEnd: space[10],
    marginBlockStart: space[28],
  },
  documentSectionU_uP: {
    marginBlockEnd: space[16],
    marginBlockStart: 0,
  },
  documentListGroup: {
    marginBlockStart: { default: null, ":not(:first-child)": space[28] },
  },
  documentListGridU_uDocumentListGroup: {
    marginBlockStart: 0,
  },
  codeViewU_uH3: {
    marginBlockStart: 0,
  },
  workbenchReferenceGridU_uH4: {
    marginInline: 0,
    marginBlockStart: 0,
    marginBottom: space[8],
    marginTop: 0,
  },
  codeViewU_uPre: {
    padding: space[16],
    overflow: "auto",
    backgroundColor: color.surfaceCanvas,
    borderInlineStartColor: { default: border.accentRail },
    borderInlineStartStyle: "solid",
    color: color.textInk,
    fontFamily: font.monospace,
    fontSize: fontSize[14],
    lineHeight: lineHeight.relaxed,
    marginBlockEnd: 0,
    maxBlockSize: size[512],
    whiteSpace: "pre",
  },
  workbenchReferenceGridU_uSection: {
    padding: 0,
    borderWidth: 0,
  },
  paperFixtureStageU_uForm: {
    gap: space[16],
    display: "grid",
  },
  workbenchApiU_uCode: {
    overflowWrap: "normal",
    wordBreak: "normal",
  },
  workbenchApiU_uWorkbenchReferenceGrid: {
    padding: 0,
  },
  workbenchReferenceGridU_uUl: {
    margin: 0,
    paddingInlineStart: space[20],
  },
  workbenchReferenceGridU_uLi: {
    marginBlockStart: { default: null, ":not(:first-child)": space[6] },
    overflowWrap: "anywhere",
  },
  workbenchReferenceGridU_uCode: {
    overflowWrap: "normal",
    wordBreak: "normal",
  },
  docsShellU_uPaperCodeBlockU_uCode: {
    color: color.textInk,
    overflowWrap: "normal",
  },
  paperCodeBlockU_uTokenComment: {
    color: color.textMuted,
  },
  paperCodeBlockU_uTokenKeyword: {
    color: color.actionDefault,
  },
  paperCodeBlockU_uTokenTag: {
    color: color.actionDefault,
  },
  paperCodeBlockU_uTokenString: {
    color: color.statusSuccess,
  },
  paperCodeBlockU_uTokenAttrValue: {
    color: color.statusSuccess,
  },
  paperCodeBlockU_uTokenFunction: {
    color: color.textLink,
  },
  paperCodeBlockU_uTokenClassName: {
    color: color.textLink,
  },
  paperCodeBlockU_uTokenNumber: {
    color: color.textLink,
  },
  paperCodeBlockU_uTokenAttrName: {
    color: color.statusDanger,
  },
  paperCodeBlockU_uTokenOperator: {
    color: color.statusDanger,
  },
  exampleCanvasU_uIframe: {
    borderWidth: 0,
    display: "block",
    maxInlineSize: "none",
  },
  previewSpecimenU_uH3: {
    marginBlock: space[6],
  },
  searchResultU_uStrong: {
    display: "block",
  },
  searchResultU_uSmall: {
    color: color.textMuted,
    display: "block",
    marginBlockStart: space[3],
  },
  foundationGuideU_uH3: {
    marginBlockEnd: space[12],
    marginBlockStart: { default: space[24] },
  },
  foundationGuideU_uPaperStackU_uH3: {
    marginBlock: 0,
  },
  foundationGuideU_uA: {
    color: color.textLink,
    textUnderlineOffset: decoration.offset,
  },
  foundationGuideU_uP: {
    maxInlineSize: measure.body,
  },
  foundationGuideU_uCode: {
    overflowWrap: "anywhere",
  },
  foundationGuideU_uSampleU_uAny: {
    marginBlockEnd: { default: null, ":last-child": 0 },
    marginBlockStart: { default: null, ":first-child": 0 },
  },
  foundationCodeU_uH3: {
    margin: 0,
    flexBasis: size[240],
    flexGrow: 1,
    flexShrink: 1,
  },
  foundationCodeU_uPre: {
    padding: space[16],
    borderColor: color.ruleDefault,
    borderStyle: "solid",
    borderWidth: border.static,
    overflow: "auto",
    backgroundColor: color.surfaceCanvas,
    color: color.textInk,
    fontSize: fontSize[13],
    lineHeight: lineHeight.loose,
    maxInlineSize: "100%",
  },
  foundationCodeU_uPreU_uCode: {
    overflowWrap: "normal",
  },
  foundationCodeU_uAny: {
    fontSize: { "[role='status']": fontSize[14], default: null, },
  },
  foundationReferenceU_uTable: {
    borderCollapse: "collapse",
    fontSize: fontSize[14],
    inlineSize: "100%",
    minInlineSize: size[512],
    textAlign: "start",
  },
  foundationReferenceU_uCaption: {
    fontWeight: fontWeight.bold,
    marginBlockEnd: space[12],
    textAlign: "start",
  },
  foundationReferenceU_uTh: {
    padding: space[12],
    paddingBlock: { default: null, [breakpoints.compactPhone]: space[8] },
    paddingInline: { default: null, [breakpoints.compactPhone]: space[4] },
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.static,
    fontWeight: fontWeight.semibold,
    minInlineSize: size[128],
    overflowWrap: "anywhere",
    textAlign: "start",
    verticalAlign: "top",
  },
  foundationReferenceU_uTd: {
    padding: space[12],
    paddingBlock: { default: null, [breakpoints.compactPhone]: space[8] },
    paddingInline: { default: null, [breakpoints.compactPhone]: space[4] },
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.static,
    overflowWrap: "anywhere",
    textAlign: "start",
    verticalAlign: "top",
  },
  foundationSpaceScaleU_uLi: {
    gap: space[8],
    alignItems: "center",
    display: "grid",
    fontSize: { default: null, [breakpoints.compactPhone]: fontSize[14] },
    gridTemplateColumns: { default: `${size[144]} ${size[56]} 1fr`, [breakpoints.compactPhone]: `${size[128]} ${size[48]} 1fr` },
  },
  foundationSpaceScaleU_uSpan: {
    backgroundColor: color.actionDefault,
    blockSize: size[16],
    display: "block",
  },
  foundationDensityExampleU_uForm: {
    gap: density.inlineGap,
    display: "grid",
  },
  foundationTypeListU_uLi: {
    borderBlockEndColor: color.ruleSubtle,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.static,
    overflowWrap: "anywhere",
    paddingBlockEnd: space[16],
  },
  foundationTypeListU_uCode: {
    display: "block",
    fontSize: fontSize[12],
    marginBlockStart: space[8],
  },
  foundationBrandGuideU_uImg: {
    blockSize: "auto",
    display: "block",
    maxInlineSize: "100%",
  },
  foundationBrandGuideU_uFigure: {
    margin: 0,
  },
  foundationBrandGuideU_uFigcaption: {
    marginBlockStart: 0,
  },
  foundationIconGridU_uDiv: {
    gap: space[8],
    display: "grid",
  },
  foundationColorTableU_uTh: {
    minInlineSize: { default: null, ":nth-child(2)": size[64], ":nth-child(3)": size[64] },
  },
  foundationColorTableU_uTdU_uP: {
    marginBlockEnd: 0,
    marginBlockStart: space[6],
  },
  foundationColorTableU_uCode: {
    fontSize: fontSize[13],
  },
})
