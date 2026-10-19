import { useId } from "react";
import { CheckIcon } from "../../icons";
import { useField, type Option } from "./field";

export interface ToggleSelectProps {
  readonly label: string;
  readonly options: readonly Option<string>[];
  readonly hint?: string;
  readonly disabled?: boolean;
}

/** A connected, optional choice. The form value is [] or [selectedOption.value]. */
export function ToggleSelect({ label, options, hint, disabled }: ToggleSelectProps) {
  const id = `paper-toggle-${useId().replace(/:/gu, "")}`;
  const { field, error } = useField<string[]>();
  const selected: string[] = field.state.value || [];
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
              type="button"
              className="paper-toggle-select__option"
              disabled={disabled || option.disabled}
              aria-pressed={pressed}
              aria-describedby={describedBy}
              onBlur={field.handleBlur}
              onClick={() => field.handleChange(pressed ? [] : [option.value])}
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
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error}</p> : null}
    </fieldset>
  );
}
