import * as stylex from "@stylexjs/stylex";
import {
  border,
  breakpoints,
  color,
  density,
  font,
  fontWeight,
  lineHeight,
  opacity,
  radius,
  size,
  space,
  type,
  zIndex,
} from "../../tokens/tokens.stylex";

/** The label, hint and error around a control, and the grid that stacks them. */
export const field = stylex.create({
  root: {
    gap: space[6],
    alignContent: "start",
    boxSizing: "border-box",
    display: "grid",
    inlineSize: "100%",
    minInlineSize: 0,
  },
  disabled: { color: color.textMuted },
  label: {
    color: color.textInk,
    fontSize: type.labelSize,
    fontWeight: fontWeight.semibold,
    lineHeight: type.labelLineHeight,
  },
  required: { color: color.statusDanger, marginInlineStart: space[4] },
  hint: {
    margin: 0,
    color: color.textMuted,
    fontSize: type.metadataSize,
    lineHeight: type.metadataLineHeight,
  },
  error: {
    margin: 0,
    color: color.statusDanger,
    fontSize: type.metadataSize,
    fontWeight: fontWeight.semibold,
    lineHeight: type.metadataLineHeight,
  },
  /** A fieldset that groups choices like a field, with its legend in the grid. */
  choiceGroup: {
    margin: 0,
    padding: 0,
    borderWidth: 0,
    gap: space[8],
    display: "grid",
    minInlineSize: 0,
  },
  /** A floated legend joins the grid, so the field gap separates it from its options like a label. */
  legend: { padding: 0, float: "inline-start", inlineSize: "100%", },
});


/** The text field box shared by inputs, text areas and selects. */
export const input = stylex.create({
  base: {
    margin: 0,
    borderColor: {
      default: color.ruleFieldEdge,
      [":focus-visible"]: color.focusRing,
      [":hover:not(:disabled):not([readonly])"]: color.ruleStrong,
    },
    borderRadius: radius.none,
    borderStyle: "solid",
    borderWidth: border.static,
    paddingBlock: density.fieldPaddingBlock,
    paddingInline: density.fieldPaddingInline,
    backgroundColor: color.surfacePaper,
    borderBlockEndColor: {
      default: color.ruleFieldEdge,
      [":focus-visible"]: { default: color.focusRing, [breakpoints.forcedColors]: "Highlight" },
      [":hover:not(:disabled):not([readonly])"]: color.ruleStrong,
    },
    borderBlockEndWidth: border.fieldEdge,
    boxShadow: { default: null, [":focus-visible"]: { default: null, [breakpoints.forcedColors]: "none" } },
    boxSizing: "border-box",
    color: color.textInk,
    fontFamily: font.body,
    fontSize: density.bodySize,
    inlineSize: "100%",
    lineHeight: lineHeight.snug,
    minBlockSize: density.controlHeight,
    outlineColor: {
      default: null,
      [":focus-visible"]: color.focusHalo,
      [breakpoints.forcedColors]: {
        default: null,
        [":focus-visible"]: "Highlight",
      },
    },
    outlineOffset: { default: null, [":focus-visible"]: 0 },
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
      [":focus-visible"]: border.focusHalo,
      [breakpoints.forcedColors]: {
        default: null,
        [":focus-visible"]: border.forcedFocus,
      },
    },
    position: { default: null, [":focus-visible"]: "relative" },
    zIndex: { default: null, [":focus-visible"]: zIndex.raised },
    // Native date and time segments otherwise add padding inside the control padding. The rule does not list this
    // Chromium pseudo-element, which StyleX compiles as written.
    // eslint-disable-next-line @stylexjs/valid-styles
    "::-webkit-datetime-edit-fields-wrapper": { padding: 0 },
    "::placeholder": { color: color.textMuted, opacity: opacity.visible },
  },
  textArea: { lineHeight: density.bodyLineHeight, minBlockSize: size[88], resize: "vertical" },
  select: {
    appearance: "none",
    paddingInlineEnd: `calc(${density.fieldPaddingInline} * 2 + ${size[16]})`,
  },
  readOnly: { backgroundColor: color.surfaceRaised, borderBlockEndColor: color.ruleDefault },
  invalid: {
    borderColor: color.statusDanger,
    borderBlockEndColor: color.statusDanger,
    outlineColor: {
      default: null,
      [":focus-visible"]: color.dangerHalo,
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
      [":focus-visible"]: border.focusHalo,
      [breakpoints.forcedColors]: {
        default: null,
        [":focus-visible"]: border.forcedFocus,
      },
    },
  },
  /** Disabled fields stay legible: muted text at full opacity, and no hover. */
  disabled: {
    borderColor: color.ruleDefault,
    borderBlockEndColor: color.ruleDefault,
    color: color.textMuted,
    cursor: "default",
  },
});

export const select = stylex.create({
  control: { display: "block", inlineSize: "100%", position: "relative" },
  icon: {
    color: color.textMuted,
    insetBlockStart: "50%",
    insetInlineEnd: density.fieldPaddingInline,
    pointerEvents: "none",
    position: "absolute",
    transform: "translateY(-50%)",
    zIndex: zIndex.focused,
  },
});
