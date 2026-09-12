import { useId } from "react";
import { useController, type FieldValues, type RegisterOptions } from "react-hook-form";
import { CheckIcon } from "../../icons";
import type { Option } from "./controls";

export interface ToggleSelectProps {
  readonly name: string;
  readonly label: string;
  readonly options: readonly Option<string>[];
  readonly hint?: string;
  readonly disabled?: boolean;
  readonly rules?: RegisterOptions<FieldValues, string>;
}

/** A connected, optional choice. The form value is [] or [selectedOption.value]. */
export function ToggleSelect({ name, label, options, hint, disabled, rules }: ToggleSelectProps) {
  const id = `paper-toggle-${useId().replace(/:/gu, "")}`;
  const { field, fieldState: { error } } = useController({ name, rules, defaultValue: [] });
  const selected: string[] = field.value || [];
  const firstEnabled = options.find((option) => !option.disabled)?.value;
  const describedBy = [hint ? `${id}-hint` : "", error ? `${id}-error` : ""].filter(Boolean).join(" ") || undefined;

  return (
    <fieldset
      className={`paper-choice-field paper-toggle-select${error ? " paper-field--invalid" : ""}`}
      disabled={disabled}
      aria-describedby={describedBy}
      aria-invalid={error ? true : undefined}
    >
      <legend className="paper-field__label">{label}</legend>
      <div className="paper-toggle-select__options">
        {options.map((option) => {
          const pressed = selected.includes(option.value);
          return (
            <button
              key={option.value}
              ref={option.value === firstEnabled ? field.ref : undefined}
              type="button"
              className="paper-toggle-select__option"
              disabled={disabled || option.disabled}
              aria-pressed={pressed}
              aria-describedby={describedBy}
              onBlur={field.onBlur}
              onClick={() => field.onChange(pressed ? [] : [option.value])}
            >
              <span className="paper-toggle-select__text">
                <span>{option.label}</span>{" "}
                {option.description ? <span className="paper-toggle-select__description">{option.description}</span> : null}
              </span>
              <CheckIcon className="paper-icon-compact paper-toggle-select__check" aria-hidden="true" />
            </button>
          );
        })}
      </div>
      {hint ? <p className="paper-field__hint" id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error.message?.toString()}</p> : null}
    </fieldset>
  );
}
