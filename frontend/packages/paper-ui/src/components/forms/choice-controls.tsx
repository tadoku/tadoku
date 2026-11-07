import * as stylex from "@stylexjs/stylex";
import { useId } from "react";
import { CheckIcon } from "../../icons";
import { a11y } from "../../foundations/recipes";
import { FieldMessages, RequiredMark, describedBy, useField, type FieldContract, type Option } from "./field";
import { field as fieldStyles } from "./field-styles";
import { choice, radioCard, segment } from "./choice-styles";

export const RADIO_SELECT_VARIANTS = ["default", "segmented"] as const;

export type RadioSelectVariant = (typeof RADIO_SELECT_VARIANTS)[number];

export interface RadioSelectProps extends FieldContract {
  readonly options: readonly Option<string>[];
  readonly disabled?: boolean;
  readonly variant?: RadioSelectVariant;
}

export function RadioSelect({
  label,
  hint,
  required,
  options,
  disabled,
  variant = "default",
}: RadioSelectProps) {
  const id = `paper-radio-${useId().replace(/:/gu, "")}`;
  const { field, error } = useField<string>();
  const segmented = variant === "segmented";
  const invalid = Boolean(error);
  return (
    <fieldset
      {...stylex.props(fieldStyles.choiceGroup)}
      aria-describedby={describedBy(id, hint, error)}
      aria-invalid={error ? true : undefined}
      disabled={disabled}
    >
      <legend {...stylex.props(fieldStyles.label, fieldStyles.legend)}>{label}<RequiredMark required={required} /></legend>
      <div {...stylex.props(choice.list, segmented && segment.list)}>
        {options.map((option) => {
          const checked = field.state.value === option.value;
          const optionDisabled = disabled || option.disabled;
          return (
            <label
              {...stylex.props(
                segmented
                  ? [
                      segment.root,
                      checked && segment.checked,
                      optionDisabled && segment.disabled,
                      invalid && (checked ? segment.invalidChecked : segment.invalid),
                    ]
                  : choice.option,
              )}
              key={option.value}
            >
              <input
                name={field.name}
                checked={checked}
                onChange={() => field.handleChange(option.value)}
                onBlur={field.handleBlur}
                type="radio"
                value={option.value}
                disabled={option.disabled}
                required={required}
                aria-invalid={error ? true : undefined}
                {...stylex.props(
                  segmented ? [stylex.defaultMarker(), a11y.visuallyHidden] : [choice.box, invalid && choice.boxInvalid],
                )}
              />
              <span {...stylex.props(!segmented && optionDisabled && choice.textDisabled)}>{option.label}</span>
              {segmented ? (
                <CheckIcon {...stylex.props(segment.check, checked && segment.checkVisible)} aria-hidden="true" />
              ) : null}
            </label>
          );
        })}
      </div>
      <FieldMessages id={id} hint={hint} error={error} />
    </fieldset>
  );
}

export interface RadioGroupOption<Value = string> extends Option<Value> {
  readonly description: string;
}

export interface RadioGroupProps<Value = string>
  extends FieldContract {
  readonly options: readonly RadioGroupOption<Value>[];
}

export function RadioGroup<Value extends string = string>({
  label,
  hint,
  required = true,
  options,
}: RadioGroupProps<Value>) {
  const id = `paper-radio-group-${useId().replace(/:/gu, "")}`;
  const { field, error } = useField<string>();
  return (
    <fieldset
      {...stylex.props(fieldStyles.choiceGroup)}
      aria-describedby={describedBy(id, hint, error)}
      aria-invalid={error ? true : undefined}
    >
      <legend {...stylex.props(fieldStyles.label, fieldStyles.legend)}>{label}<RequiredMark required={required} /></legend>
      <div {...stylex.props(radioCard.list)}>
        {options.map((option) => (
          <label
            {...stylex.props(
              radioCard.root,
              field.state.value === option.value && radioCard.checked,
              Boolean(error) && radioCard.invalid,
              option.disabled && radioCard.disabled,
            )}
            key={option.value}
          >
            <input
              name={field.name}
              checked={field.state.value === option.value}
              onChange={() => field.handleChange(option.value)}
              onBlur={field.handleBlur}
              type="radio"
              value={option.value}
              required={required}
              disabled={option.disabled}
              aria-invalid={error ? true : undefined}
              {...stylex.props(stylex.defaultMarker(), radioCard.input)}
            />
            <span {...stylex.props(radioCard.text)}><strong>{option.label}</strong><small {...stylex.props(radioCard.description)}>{option.description}</small></span>
          </label>
        ))}
      </div>
      <FieldMessages id={id} hint={hint} error={error} />
    </fieldset>
  );
}
