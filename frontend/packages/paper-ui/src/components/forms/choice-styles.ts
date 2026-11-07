import * as stylex from "@stylexjs/stylex";
import {
  border,
  breakpoints,
  color,
  density,
  fontWeight,
  lineHeight,
  opacity,
  radius,
  size,
  space,
  type,
  zIndex,
} from "../../tokens/tokens.stylex";


/** Checkboxes and radios: a native box beside its label, or a card row. */
export const choice = stylex.create({
  list: { columnGap: space[20], display: "flex", flexWrap: "wrap", rowGap: space[12] },
  option: { gap: density.inlineGap, alignItems: "start", cursor: "pointer", display: "inline-flex", },
  box: {    marginInline: 0,
    accentColor: color.actionDefault,
    blockSize: size[18],
    boxShadow: {
      default: null,
      [":focus-visible"]: { default: `inset 0 0 0 ${border.static} ${color.focusRing}`, [breakpoints.forcedColors]: "none" },
    },
 boxSizing: "border-box",
    flexShrink: 0,
 fontFamily: "inherit",
 fontSize: "inherit",
    inlineSize: size[18],
    marginBlockEnd: 0,
    marginBlockStart: space[3],
    outlineColor: {
      default: null,
      [":focus-visible"]: color.focusRing,
      [breakpoints.forcedColors]: {
        default: null,
        [":focus-visible"]: "Highlight",
      },
    },
    outlineOffset: { default: null, [":focus-visible"]: { default: border.focusOffset, [breakpoints.forcedColors]: 0 } },
    outlineStyle: {
      default: null,
      [":focus-visible"]: "solid",
      [breakpoints.forcedColors]: {
        default: null,
        [":focus-visible"]: "solid",
      },
    },
    outlineWidth: {
      default: null,
      [":focus-visible"]: border.focusRing,
      [breakpoints.forcedColors]: {
        default: null,
        [":focus-visible"]: border.forcedFocus,
      },
    },
  },
  boxInvalid: {
    boxShadow: {
      default: null,
      [":focus-visible"]: { default: `inset 0 0 0 ${border.static} ${color.statusDanger}`, [breakpoints.forcedColors]: "none" },
    },
    outlineColor: {
      default: null,
      [":focus-visible"]: color.statusDanger,
      [breakpoints.forcedColors]: {
        default: null,
        [":focus-visible"]: "Highlight",
      },
    },
    outlineStyle: {
      default: null,
      [":focus-visible"]: "solid",
      [breakpoints.forcedColors]: {
        default: null,
        [":focus-visible"]: "solid",
      },
    },
    outlineWidth: {
      default: null,
      [":focus-visible"]: border.focusRing,
      [breakpoints.forcedColors]: {
        default: null,
        [":focus-visible"]: border.forcedFocus,
      },
    },
  },
  /**
   * Browsers fade a disabled box to about 1.3:1; an outline a step lighter than the field edge (about 3.5:1) keeps
   * it reading as a control without matching an enabled box.
   */
  checkboxDisabled: {
    borderRadius: radius.small,
    outlineColor: color.disabledCheckEdge,
    outlineOffset: `calc(-1 * ${border.static})`,
    outlineStyle: "solid",
    outlineWidth: border.static,
  },
  textDisabled: { color: color.textMuted, cursor: "default" },
  /** A checkbox hint sits under its label text, so it reads as that option's, not the next one's. */
  hint: {
    marginBlockStart: `calc(-1 * ${space[4]})`,
    paddingInlineStart: `calc(${size[18]} + ${density.inlineGap})`,
  },
  card: {
    borderColor: color.ruleFieldEdge,
    borderStyle: "solid",
    borderWidth: border.static,
    paddingBlock: space[12],
    paddingInline: space[16],
    alignItems: "center",
    backgroundColor: { default: color.surfacePaper, ":hover": color.actionNeutralHover },
    boxSizing: "border-box",
    columnGap: density.inlineGap,
    display: "grid",
    gridTemplateColumns: { default: "auto minmax(0, 1fr) auto", [breakpoints.phone]: "auto minmax(0, 1fr)" },
    inlineSize: "100%",
    rowGap: space[4],
  },
  // A checked card keeps the hover wash while the pointer is over it.
  cardChecked: {
    borderColor: color.textLink,
    backgroundColor: { default: color.actionSoft, ":hover": color.actionNeutralHover },
    boxShadow: `inset 0 calc(-1 * ${border.static}) 0 ${color.textLink}`,
  },
  cardDisabled: {
    borderColor: color.ruleDefault,
    backgroundColor: color.surfacePaper,
    cursor: "default",
  },
  cardBox: { margin: 0 },
  body: { gap: space[2], display: "grid", minInlineSize: 0, },
  label: { fontWeight: fontWeight.semibold, overflowWrap: "anywhere" },
  labelDisabled: { color: color.textMuted, fontWeight: fontWeight.regular },
  detail: { color: color.textMuted, fontSize: type.metadataSize },
  aside: {
    gridColumn: { default: null, [breakpoints.phone]: 2 },
    fontVariantNumeric: "tabular-nums",
    textAlign: { default: "end", [breakpoints.phone]: "start" },
  },
});

/** RadioSelect's segmented variant: joined segments with a check on the selected one. */
export const segment = stylex.create({
  list: {
    gap: 0,
    alignItems: "stretch",
    display: "inline-flex",
    flexWrap: "wrap",
    inlineSize: "fit-content",
    maxInlineSize: "100%",
  },
  root: {
    borderColor: {
      default: color.ruleFieldEdge,
      [stylex.when.descendant(":focus-visible")]: color.focusRing,
      ":hover": color.ruleStrong,
      [breakpoints.forcedColors]: { default: "CanvasText", [stylex.when.descendant(":focus-visible")]: "Highlight" },
    },
    borderStyle: "solid",
    borderWidth: border.static,
    paddingBlock: density.fieldPaddingBlock,
    paddingInline: density.fieldPaddingInline,
    alignItems: "center",
    backgroundColor: {
      default: color.surfacePaper,
      ":hover": color.actionNeutralHover,
      [breakpoints.forcedColors]: "Canvas",
    },
    borderBlockEndWidth: border.fieldEdge,
    boxSizing: "border-box",
    color: { default: color.textInk, [breakpoints.forcedColors]: "CanvasText" },
    cursor: "pointer",
    display: "inline-grid",
    flexBasis: size[112],
    flexGrow: 1,
    flexShrink: 1,
    fontSize: density.bodySize,
    fontWeight: fontWeight.semibold,
    forcedColorAdjust: { default: null, [breakpoints.forcedColors]: "none" },
    gridTemplateColumns: "1fr auto",
    lineHeight: lineHeight.snug,
    marginBlockStart: `calc(-1 * ${border.static})`,
    marginInlineStart: { default: `calc(-1 * ${border.static})`, ":first-child": 0 },
    minBlockSize: density.controlHeight,
    minInlineSize: density.radioMinWidth,
    outlineColor: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: color.focusHalo,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "Highlight",
      },
    },
    outlineOffset: 0,
    outlineStyle: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: "solid",
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "solid",
      },
    },
    outlineWidth: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: border.focusHalo,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: border.forcedFocus,
      },
    },
    position: "relative",
    zIndex: { default: null, [stylex.when.descendant(":focus-visible")]: zIndex.focused },
  },
  checked: {
    borderColor: {
      default: color.textLink,
      [stylex.when.descendant(":focus-visible")]: color.focusRing,
      [breakpoints.forcedColors]: "Highlight",
    },
    backgroundColor: { default: color.actionSoft, [breakpoints.forcedColors]: "Canvas" },
    zIndex: { default: zIndex.raised, [stylex.when.descendant(":focus-visible")]: zIndex.focused },
  },
  disabled: {
    borderColor: { default: color.ruleFieldEdge, [breakpoints.forcedColors]: "CanvasText" },
    backgroundColor: { default: color.surfacePaper, [breakpoints.forcedColors]: "Canvas" },
    color: { default: color.textMuted, [breakpoints.forcedColors]: "GrayText" },
    cursor: "default",
  },
  invalid: {
    borderColor: {
      default: color.statusDanger,
      ":hover": color.ruleStrong,
      [breakpoints.forcedColors]: "Mark",
    },
    outlineColor: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: color.dangerHalo,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "Highlight",
      },
    },
    outlineStyle: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: "solid",
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "solid",
      },
    },
    outlineWidth: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: border.focusHalo,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: border.forcedFocus,
      },
    },
  },
  invalidChecked: { borderColor: { default: color.statusDanger, [breakpoints.forcedColors]: "Mark" } },
  check: {
    flex: "none",
    blockSize: size[16],
    color: { default: color.textLink, [breakpoints.forcedColors]: "Highlight" },
    inlineSize: size[16],
    marginInlineStart: density.inlineGap,
    opacity: 0,
    verticalAlign: "middle",
  },
  checkVisible: { opacity: opacity.visible },
});

/** RadioGroup: a grid of cards, each a radio with its label over a description. */
export const radioCard = stylex.create({
  list: { gap: space[10], display: "grid", gridTemplateColumns: `repeat(auto-fit, minmax(${size[192]}, 1fr))`, },
  root: {
    padding: space[14],
    borderColor: {
      default: color.ruleFieldEdge,
      [stylex.when.descendant(":focus-visible")]: color.focusRing,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "Highlight",
      },
    },
    borderStyle: "solid",
    borderWidth: border.static,
    gap: space[12],
    backgroundColor: color.surfacePaper,
    cursor: "pointer",
    display: "grid",
    gridTemplateColumns: "auto 1fr",
    outlineColor: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: color.focusHalo,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "Highlight",
      },
    },
    outlineOffset: 0,
    outlineStyle: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: "solid",
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "solid",
      },
    },
    outlineWidth: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: border.focusHalo,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: border.forcedFocus,
      },
    },
    position: { default: null, [stylex.when.descendant(":focus-visible")]: "relative" },
    zIndex: { default: null, [stylex.when.descendant(":focus-visible")]: zIndex.raised },
  },
  checked: { borderColor: color.textLink, borderBlockEndWidth: border.fieldEdge, },
  invalid: {
    outlineColor: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: color.dangerHalo,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "Highlight",
      },
    },
    outlineStyle: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: "solid",
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "solid",
      },
    },
    outlineWidth: {
      default: null,
      [stylex.when.descendant(":focus-visible")]: border.focusHalo,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: border.forcedFocus,
      },
    },
  },
  disabled: { color: color.textMuted, cursor: "default" },
  input: { accentColor: color.actionDefault, marginBlockStart: space[4], outlineStyle: { default: null, ":focus-visible": "none" }, },
  text: { gap: space[3], display: "grid", },
  description: { color: color.textMuted },
});
