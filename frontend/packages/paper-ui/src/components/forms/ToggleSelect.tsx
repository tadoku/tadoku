import * as stylex from "@stylexjs/stylex";
import { useId } from "react";
import { CheckIcon } from "../../icons";
import { icon } from "../../foundations/recipes";
import { border, breakpoints, color, density, font, fontWeight, fontSize, lineHeight, size, space, type, zIndex } from "../../tokens/tokens.stylex";
import { FieldMessages, useField, type Option } from "./field";
import { field as fieldStyles } from "./field-styles";

export interface ToggleSelectProps {
  readonly label: string;
  readonly options: readonly Option<string>[];
  readonly hint?: string;
  readonly disabled?: boolean;
}


const styles = stylex.create({
  root: { containerType: "inline-size", minInlineSize: 0 },
  legend: { marginBlockEnd: space[8] },
  options: {
    borderColor: { default: color.ruleFieldEdge, [breakpoints.forcedColors]: "CanvasText" },
    borderStyle: "solid",
    borderWidth: border.static,
    gap: border.static,
    backgroundColor: { default: color.ruleDefault, [breakpoints.forcedColors]: "CanvasText" },
    borderBlockEndWidth: border.fieldEdge,
    display: "flex",
    flexDirection: { default: null, [breakpoints.containerNarrow]: "column" },
    flexWrap: "wrap",
  },
  optionsInvalid: { borderColor: { default: color.statusDanger, [breakpoints.forcedColors]: "CanvasText" } },
  option: {
    borderRadius: 0,
    borderWidth: 0,
    gap: space[8],
    paddingBlock: density.fieldPaddingBlock,
    paddingInline: density.fieldPaddingInline,
    alignItems: "center",
    backgroundColor: {
      default: color.surfacePaper,
      [breakpoints.forcedColors]: "Canvas",
      ":hover:not(:disabled)": color.toggleHover,
    },
    boxSizing: "border-box",
    color: { default: color.textInk, [breakpoints.forcedColors]: { default: "CanvasText", ":disabled": "GrayText" }, ":disabled": color.textMuted, },
    cursor: { default: "pointer", ":disabled": "default" },
    display: "flex",
    flexBasis: { default: size[136], [breakpoints.containerNarrow]: "auto" },
    flexGrow: 1,
    flexShrink: 1,
    fontFamily: font.body,
    fontSize: fontSize[14],
    fontWeight: fontWeight.semibold,
    justifyContent: "space-between",
    lineHeight: lineHeight.snug,
    minBlockSize: `max(${size[44]}, ${density.controlHeight})`,
    minInlineSize: 0,
    outlineColor: {
      default: null,
      ":focus-visible": {
        default: color.focusRing,
        [breakpoints.forcedColors]: "Highlight",
      },
    },
    outlineOffset: {
      default: null,
      ":focus-visible": { default: `calc(-1 * ${border.focusRing})`, [breakpoints.forcedColors]: border.focusOffset },
    },
    outlineStyle: {
      default: null,
      ":focus-visible": {
        default: "solid",
        [breakpoints.forcedColors]: "solid",
      },
    },
    outlineWidth: {
      default: null,
      ":focus-visible": {
        default: border.focusRing,
        [breakpoints.forcedColors]: border.focusRing,
      },
    },
    position: { default: null, ":focus-visible": "relative" },
    textAlign: "start",
    zIndex: { default: null, ":focus-visible": zIndex.raised },
  },
  pressed: {
    backgroundColor: { default: color.actionSoft, [breakpoints.forcedColors]: "Highlight" },
    color: { default: color.textInk, [breakpoints.forcedColors]: "HighlightText" },
    forcedColorAdjust: { default: null, [breakpoints.forcedColors]: "none" },
  },
  text: {
    gap: { default: space[2], [breakpoints.containerNarrow]: space[8] },
    alignItems: { default: null, [breakpoints.containerNarrow]: "baseline" },
    display: { default: "grid", [breakpoints.containerNarrow]: "flex" },
    minInlineSize: 0,
    overflowWrap: "anywhere",
  },
  textDisabled: { color: { default: color.textMuted, [breakpoints.forcedColors]: "GrayText" } },
  description: {
    color: color.textMuted,
    fontSize: type.metadataSize,
    fontWeight: fontWeight.regular,
    lineHeight: type.metadataLineHeight,
  },
  descriptionPressed: { color: { default: color.textMuted, [breakpoints.forcedColors]: "inherit" } },
  check: { flexShrink: 0, visibility: "hidden" },
  checkPressed: { visibility: "visible" },
});

/** A connected, optional choice. The form value is [] or [selectedOption.value]. */
export function ToggleSelect({ label, options, hint, disabled }: ToggleSelectProps) {
  const id = `paper-toggle-${useId().replace(/:/gu, "")}`;
  const { field, error } = useField<string[]>();
  const selected: string[] = field.state.value || [];
  const describedBy = [hint ? `${id}-hint` : "", error ? `${id}-error` : ""].filter(Boolean).join(" ") || undefined;
  return (
    <fieldset
      {...stylex.props(fieldStyles.choiceGroup, styles.root)}
      disabled={disabled}
      aria-describedby={describedBy}
      aria-invalid={error ? true : undefined}
    >
      <legend {...stylex.props(fieldStyles.label, fieldStyles.legend, styles.legend)}>{label}</legend>
      <div {...stylex.props(styles.options, Boolean(error) && styles.optionsInvalid)}>
        {options.map((option) => {
          const pressed = selected.includes(option.value);
          const optionDisabled = disabled || option.disabled;
          return (
            <button
              key={option.value}
              type="button"
              {...stylex.props(styles.option, pressed && styles.pressed)}
              disabled={optionDisabled}
              aria-pressed={pressed}
              aria-describedby={describedBy}
              onBlur={field.handleBlur}
              onClick={() => field.handleChange(pressed ? [] : [option.value])}
            >
              <span {...stylex.props(styles.text, optionDisabled && styles.textDisabled)}>
                <span>{option.label}</span>{" "}
                {option.description ? (
                  <span {...stylex.props(styles.description, pressed && styles.descriptionPressed)}>{option.description}</span>
                ) : null}
              </span>
              <CheckIcon {...stylex.props(icon.compact, styles.check, pressed && styles.checkPressed)} aria-hidden="true" />
            </button>
          );
        })}
      </div>
      <FieldMessages id={id} hint={hint} error={error} />
    </fieldset>
  );
}
