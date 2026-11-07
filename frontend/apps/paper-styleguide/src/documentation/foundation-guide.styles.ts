import * as stylex from '@stylexjs/stylex'
import { border, breakpoints, color, density, fontSize, size, space } from 'paper-ui/tokens.stylex'

export const foundationStyles = stylex.create({
  foundationGuide: {
    minInlineSize: 0,
  },
  foundationGuide__grid: {
    gap: space[24],
    display: "grid",
    gridTemplateColumns: `repeat(auto-fit, minmax(min(100%, ${size[240]}), 1fr))`,
  },
  foundationGuide__sample: {
    padding: space[16],
    borderColor: color.ruleDefault,
    borderStyle: "solid",
    borderWidth: border.static,
    backgroundColor: color.surfacePaper,
    color: color.textInk,
    minInlineSize: 0,
  },
  foundationCode: {
    marginBlock: space[24],
    minInlineSize: 0,
  },
  foundationReferenceHint: {
    color: { default: null, [breakpoints.sm]: color.textMuted },
    display: { default: "none", [breakpoints.sm]: "block" },
    fontSize: { default: null, [breakpoints.sm]: fontSize[12] },
    insetInlineStart: { default: null, [breakpoints.sm]: 0 },
    position: { default: null, [breakpoints.sm]: "sticky" },
  },
  foundationReference: {
    marginBlock: space[24],
    overflowX: "auto",
  },
  foundationSpaceScale: {
    padding: 0,
    gap: space[12],
    listStyle: "none",
    display: "grid",
  },
  foundationDensityExample: {
    fontSize: density.bodySize,
    lineHeight: density.bodyLineHeight,
  },
  foundationColorSwatch: {
    borderColor: color.ruleDefault,
    borderStyle: "solid",
    borderWidth: border.static,
    flex: "none",
    blockSize: size[32],
    display: "block",
    inlineSize: size[48],
  },
  foundationTypeList: {
    padding: 0,
    gap: space[16],
    listStyle: "none",
    display: "grid",
  },
  foundationRailHeading: {
    paddingInlineStart: space[16],
  },
  foundationMotionDemo: {
    gap: space[16],
    display: "grid",
  },
  foundationMotionDemo__track: {
    borderColor: color.ruleDefault,
    borderStyle: "solid",
    borderWidth: border.static,
    backgroundColor: color.surfaceCanvas,
    blockSize: size[32],
    position: "relative",
  },
  foundationMotionDemo__marker: {
    insetBlock: space[4],
    backgroundColor: color.actionDefault,
    inlineSize: size[20],
    insetInlineStart: space[4],
    position: "absolute",
    transitionProperty: "inset-inline-start",
    transitionTimingFunction: "ease-out",
  },
  foundationMotionDemo__markerMoved: { insetInlineStart: `calc(100% - ${size[24]})` },
  foundationIconGrid: {
    gap: space[16],
    display: "grid",
    gridTemplateColumns: `repeat(auto-fit, minmax(min(100%, ${size[160]}), 1fr))`,
  },
  foundationGuide__preview: {
    padding: space[16],
    borderColor: color.ruleDefault,
    borderStyle: "solid",
    borderWidth: border.static,
    marginBlock: space[24],
    backgroundColor: color.surfaceCanvas,
    color: color.textInk,
    minInlineSize: 0,
  },
  foundationGuide__settingsNote: {
    color: color.textMuted,
    fontSize: fontSize[14],
  },
  foundationBrandGuide: {},
  foundationColorTable: {},
  paperStack: {},
})

