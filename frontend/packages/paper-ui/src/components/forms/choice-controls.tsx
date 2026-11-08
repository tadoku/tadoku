import { useId } from "react";
import { CheckIcon, iconClassName } from "../../icons";
import { describedBy, useField, type FieldContract, type Option } from "./field";
import { choice, choiceField, field as fieldStyles, radioCard, segment } from "./styles";

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
  const group = choiceField({ invalid });
  const fieldSlots = fieldStyles();
  const choiceSlots = choice({ invalid });
  const segmentSlots = segment({ invalid });
  return (
    <fieldset
      className={group.root({ className: ["paper-radio-select", `paper-radio-select--${variant}`] })}
      aria-describedby={describedBy(id, hint, error)}
      aria-invalid={error ? true : undefined}
      disabled={disabled}
    >
      <legend className={fieldSlots.label({ className: group.legend() })}>{label}{required ? <span className={fieldSlots.required()} aria-hidden="true">*</span> : null}</legend>
      <div className={group.list({ className: segmented ? segmentSlots.list() : undefined })}>
        {options.map((option) => (
          <label className={segmented ? segmentSlots.root() : choiceSlots.root()} key={option.value}>
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
              className={segmented ? segmentSlots.input() : choiceSlots.input()}
            />
            <span className={segmented ? "paper-radio-select__label" : choiceSlots.text()}>{option.label}</span>
            {segmented ? (
              <CheckIcon
                className={iconClassName("compact", segmentSlots.check())}
                aria-hidden="true"
              />
            ) : null}
          </label>
        ))}
      </div>
      {hint ? <p className={fieldSlots.hint()} id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className={fieldSlots.error()} id={`${id}-error`} role="alert">{error}</p> : null}
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
  const group = choiceField({ invalid: Boolean(error) });
  const fieldSlots = fieldStyles();
  const card = radioCard({ invalid: Boolean(error) });
  return (
    <fieldset
      className={group.root()}
      aria-describedby={describedBy(id, hint, error)}
      aria-invalid={error ? true : undefined}
    >
      <legend className={fieldSlots.label({ className: group.legend() })}>{label}{required ? <span className={fieldSlots.required()} aria-hidden="true">*</span> : null}</legend>
      <div className={card.list()}>
        {options.map((option) => (
          <label className={card.root()} key={option.value}>
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
              className={card.input()}
            />
            <span className={card.text()}><strong>{option.label}</strong><small className={card.description()}>{option.description}</small></span>
          </label>
        ))}
      </div>
      {hint ? <p className={fieldSlots.hint()} id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className={fieldSlots.error()} id={`${id}-error`} role="alert">{error}</p> : null}
    </fieldset>
  );
}
