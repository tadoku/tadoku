import { useId } from "react";
import { CheckIcon, iconClassName } from "../../icons";
import { describedBy, useField, type FieldContract, type Option } from "./field";

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
  return (
    <fieldset
      className={`paper-choice-field paper-radio-select paper-radio-select--${variant}${error ? " paper-field--invalid" : ""}`}
      aria-describedby={describedBy(id, hint, error)}
      aria-invalid={error ? true : undefined}
      disabled={disabled}
    >
      <legend className="paper-field__label">{label}{required ? <span className="paper-field__required" aria-hidden="true">*</span> : null}</legend>
      <div className="paper-choice-list">
        {options.map((option) => (
          <label className={segmented ? "paper-radio-select__segment" : "paper-choice"} key={option.value}>
            <input
              name={field.name}
              checked={field.state.value === option.value}
              onChange={() => field.handleChange(option.value)}
              onBlur={field.handleBlur}
              type="radio"
              value={option.value}
              disabled={option.disabled}
              required={required}
              aria-invalid={error ? true : undefined}
            />
            <span className={segmented ? "paper-radio-select__label" : undefined}>{option.label}</span>
            {segmented ? (
              <CheckIcon
                className={iconClassName("compact", "paper-radio-select__check")}
                aria-hidden="true"
              />
            ) : null}
          </label>
        ))}
      </div>
      {hint ? <p className="paper-field__hint" id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error}</p> : null}
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
      className={`paper-choice-field${error ? " paper-field--invalid" : ""}`}
      aria-describedby={describedBy(id, hint, error)}
      aria-invalid={error ? true : undefined}
    >
      <legend className="paper-field__label">{label}{required ? <span className="paper-field__required" aria-hidden="true">*</span> : null}</legend>
      <div className="paper-radio-cards">
        {options.map((option) => (
          <label className="paper-radio-card" key={option.value}>
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
            />
            <span><strong>{option.label}</strong><small>{option.description}</small></span>
          </label>
        ))}
      </div>
      {hint ? <p className="paper-field__hint" id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error}</p> : null}
    </fieldset>
  );
}
