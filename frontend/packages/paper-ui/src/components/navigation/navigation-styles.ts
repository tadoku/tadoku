import * as stylex from "@stylexjs/stylex";
import {
  border,
  breakpoints,
  color,
  density,
  font,
  fontWeight,
  opacity,
  space,
  type,
} from "../../tokens/tokens.stylex";


export const list = stylex.create({
  reset: { margin: 0, padding: 0, listStyle: "none", },
});

export const sidebar = stylex.create({
  root: { gap: density.sidebarGap, display: "grid", },
  section: { gap: space[4], display: "grid", },
  title: {
    marginInline: 0,
    fontFamily: font.body,
    fontSize: type.labelSize,
    lineHeight: type.labelLineHeight,
    marginBlockEnd: space[4],
    marginBlockStart: 0,
    textTransform: "uppercase",
  },
  link: {
    gap: density.inlineGap,
    textDecoration: "none",
    alignItems: "center",
    backgroundColor: {
      ['[aria-current="page"]']: color.surfacePaper,
      default: null,
      [':hover:not([aria-disabled="true"])']: color.surfaceRaised,
    },
    boxSizing: "border-box",
    color: { ['[aria-current="page"]']: color.textInk, default: color.textMuted, [':hover:not([aria-disabled="true"])']: color.textInk, },
    cursor: { ['[aria-disabled="true"]']: "default", default: null, },
    display: "flex",
    fontWeight: { ['[aria-current="page"]']: fontWeight.semibold, default: null, },
    minBlockSize: density.controlHeight,
    opacity: { ['[aria-disabled="true"]']: opacity.inactive, default: null, },
    paddingInlineEnd: density.fieldPaddingInline,
    paddingInlineStart: `calc(${density.fieldPaddingInline} + ${border.accentRail})`,
    position: "relative",
    "::before": {
      insetBlock: 0,
      backgroundColor: { ['[aria-current="page"]']: color.textLink, default: "transparent", },
      content: '""',
      inlineSize: border.accentRail,
      insetInlineStart: 0,
      position: "absolute",
    },
  },
});

export const tabbar = stylex.create({
  /** The rule is an inset shadow so the active tab's bar covers it rather than stacking a second line above it. */
  horizontal: {
    boxShadow: `inset 0 calc(-1 * ${border.static}) 0 ${color.ruleDefault}`,
    display: "flex",
    overflowX: "auto",
  },
  vertical: { boxShadow: `inset ${border.static} 0 0 ${color.ruleDefault}`, display: "grid" },
  link: {
    borderWidth: 0,
    gap: density.inlineGap,
    paddingInline: density.fieldPaddingInline,
    textDecoration: "none",
    alignItems: "center",
    backgroundColor: {
      default: "transparent",
      [':hover:not([aria-disabled="true"]):not([aria-current="page"])']: color.inkWash,
    },
    boxSizing: "border-box",
    color: {
      ['[aria-current="page"]']: color.textInk,
      default: color.textMuted,
      [':hover:not([aria-disabled="true"]):not([aria-current="page"])']: color.textInk,
    },
    cursor: { ['[aria-disabled="true"]']: "default", default: "pointer", },
    display: "flex",
    fontWeight: fontWeight.semibold,
    minBlockSize: density.controlHeight,
    opacity: { ['[aria-disabled="true"]']: opacity.inactive, default: null, },
    position: "relative",
    whiteSpace: "nowrap",
  },
  horizontalLink: { boxShadow: { ['[aria-current="page"]']: `inset 0 calc(-1 * ${border.indicator}) 0 ${color.textLink}`, default: null, } },
  verticalLink: { boxShadow: { ['[aria-current="page"]']: `inset ${border.indicator} 0 0 ${color.textLink}`, default: null, } },
});

export const breadcrumb = stylex.create({
  list: { gap: space[4], alignItems: "center", display: "flex", flexWrap: "wrap", minInlineSize: 0, },
  item: {
    gap: space[6],
    alignItems: "center",
    display: { default: "flex", [breakpoints.md]: "none" },
  },
  /** Phones keep only the current page and its parent. */
  itemKept: { display: "flex" },
  link: { gap: space[6], textDecoration: "underline", alignItems: "center", display: "flex", },
  current: { gap: space[6], alignItems: "center", color: color.textMuted, display: "flex", },
  separator: { marginInline: space[4], color: color.ruleStrong, },
});

export const pagination = stylex.create({
  list: {
    gap: space[4],
    alignItems: "center",
    display: "flex",
    justifyContent: { default: "flex-end", [breakpoints.md]: "space-between" },
  },
  control: {
    borderWidth: 0,
    gap: space[6],
    paddingInline: space[8],
    textDecoration: "none",
    alignItems: "center",
    backgroundColor: { default: "transparent", [':hover:not([aria-disabled="true"]):not([aria-current="page"])']: color.actionNeutralHover },
    boxSizing: "border-box",
    color: { ['[aria-disabled="true"]']: color.textMuted, default: color.textInk, },
    cursor: { ['[aria-current="page"]']: "default", ['[aria-disabled="true"]']: "default", default: "pointer", },
    display: "inline-flex",
    justifyContent: "center",
    minBlockSize: density.controlHeight,
    minInlineSize: density.controlHeight,
  },
  current: { boxShadow: `inset 0 calc(-1 * ${border.indicator}) 0 ${color.textLink}`, fontWeight: fontWeight.semibold },
  /** Phones keep only the current page, as "2 of 5", between Previous and Next. */
  phoneHidden: { display: { default: null, [breakpoints.md]: "none" } },
  of: { display: { default: "none", [breakpoints.md]: "inline" }, fontWeight: fontWeight.regular },
  gap: {
    placeItems: "center",
    color: color.textMuted,
    display: { default: "grid", [breakpoints.md]: "none" },
    inlineSize: density.controlHeight,
  },
});
