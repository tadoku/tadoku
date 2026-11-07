import * as stylex from "@stylexjs/stylex";
import { useId, type InputHTMLAttributes } from "react";
import { ChevronDownIcon } from "../../icons";
import { icon } from "../../foundations/recipes";
import { border, breakpoints, color, density, size } from "../../tokens/tokens.stylex";
import { FieldFrame, describedBy, useField, type Option } from "./field";
import { input, select } from "./field-styles";

export interface AmountWithUnitProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, 'name' | 'type' | 'value' | 'onChange' | 'onBlur'> {
  readonly label: string;
  readonly units: readonly Option<string>[];
  readonly unitsLabel?: string;
  readonly hint?: string;
  readonly unitField: {
    readonly name: string;
    readonly state: { readonly value: string };
    readonly handleChange: (value: string) => void;
    readonly handleBlur: () => void;
  };
}


const styles = stylex.create({
  group: {
    borderColor: { default: color.ruleFieldEdge, [breakpoints.forcedColors]: "CanvasText" },
    borderStyle: "solid",
    borderWidth: border.static,
    backgroundColor: { default: color.surfacePaper, [breakpoints.forcedColors]: "Canvas" },
    borderBlockEndWidth: 0,
    boxSizing: "border-box",
    display: "grid",
    gridTemplateColumns: "minmax(0, 1fr) auto",
    minBlockSize: density.controlHeight,
    outlineColor: {
      default: null,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "Highlight",
      },
      [stylex.when.descendant(":focus-visible")]: color.focusHalo,
    },
    outlineOffset: 0,
    outlineStyle: {
      default: null,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "solid",
      },
      [stylex.when.descendant(":focus-visible")]: "solid",
    },
    outlineWidth: {
      default: null,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: border.forcedFocus,
      },
      [stylex.when.descendant(":focus-visible")]: border.focusHalo,
    },
  },
  groupInvalid: {
    borderColor: { default: color.statusDanger, [breakpoints.forcedColors]: "Mark" },
    outlineColor: {
      default: null,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "Highlight",
      },
      [stylex.when.descendant(":focus-visible")]: color.dangerHalo,
    },
    outlineStyle: {
      default: null,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: "solid",
      },
      [stylex.when.descendant(":focus-visible")]: "solid",
    },
    outlineWidth: {
      default: null,
      [breakpoints.forcedColors]: {
        default: null,
        [stylex.when.descendant(":focus-visible")]: border.forcedFocus,
      },
      [stylex.when.descendant(":focus-visible")]: border.focusHalo,
    },
  },
  amount: {
    borderWidth: 0,
    backgroundColor: "transparent",
    borderBlockEndColor: {
      default: color.ruleFieldEdge,
      [breakpoints.forcedColors]: "CanvasText",
      ":focus-visible": { default: color.focusRing, [breakpoints.forcedColors]: "Highlight" },
    },
    borderBlockEndWidth: border.fieldEdge,
    minBlockSize: `calc(${density.controlHeight} - (${border.static} * 2))`,
    outlineStyle: {
      default: null,
      ":focus-visible": "none",
    },
  },
  amountInvalid: {
    borderBlockEndColor: { default: color.statusDanger, [breakpoints.forcedColors]: "Mark", ":focus-visible": color.statusDanger, },
  },
  unit: {
    blockSize: `calc(${density.controlHeight} - ${border.static})`,
    borderInlineStartColor: { default: color.ruleDefault, [breakpoints.forcedColors]: "CanvasText" },
    borderInlineStartStyle: "solid",
    borderInlineStartWidth: border.static,
    boxSizing: "border-box",
    inlineSize: "auto",
    minInlineSize: size[96],
  },
  unitSelect: {
    borderWidth: 0,
    backgroundColor: "transparent",
    blockSize: "100%",
    borderBlockEndColor: { default: color.ruleFieldEdge, ":hover:not(:disabled)": color.ruleStrong },
    borderBlockEndWidth: border.fieldEdge,
    minBlockSize: 0,
  },
});

export function AmountWithUnit({ label, units, unitsLabel, unitField, hint, required, disabled, ...props }: AmountWithUnitProps) {
  const id = `paper-amount-${useId().replace(/:/gu, '')}`;
  const { field, error } = useField<number | string>();
  const invalid = Boolean(error);
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error} disabled={disabled}>
      <div {...stylex.props(styles.group, invalid && styles.groupInvalid)} role="group" aria-labelledby={`${id}-label`}>
        <input
          {...props}
          name={field.name}
          value={typeof field.state.value === 'number' && Number.isNaN(field.state.value) ? '' : field.state.value ?? ''}
          onChange={event => field.handleChange(event.currentTarget.valueAsNumber)}
          onBlur={field.handleBlur}
          required={required}
          disabled={disabled}
          id={id}
          type="number"
          {...stylex.props(
            stylex.defaultMarker(),
            input.base,
            disabled && input.disabled,
            styles.amount,
            invalid && styles.amountInvalid,
          )}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy(id, hint, error, props['aria-describedby'])}
        />
        <span {...stylex.props(select.control, styles.unit)}>
          <select
            name={unitField.name}
            value={unitField.state.value}
            onChange={event => unitField.handleChange(event.currentTarget.value)}
            onBlur={unitField.handleBlur}
            disabled={disabled}
            aria-label={unitsLabel ?? `Unit for ${label.toLocaleLowerCase()}`}
            {...stylex.props(input.base, input.select, disabled && input.disabled, styles.unitSelect)}
            aria-describedby={describedBy(id, hint, error, props['aria-describedby'])}
          >
            {units.map(unit => <option key={unit.value} value={unit.value} disabled={unit.disabled}>{unit.label}</option>)}
          </select>
          <ChevronDownIcon {...stylex.props(icon.compact, select.icon)} aria-hidden="true" />
        </span>
      </div>
    </FieldFrame>
  );
}
