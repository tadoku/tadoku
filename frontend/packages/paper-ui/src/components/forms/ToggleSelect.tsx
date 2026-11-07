import { useId } from "react";
import { CheckIcon, iconClassName } from "../../icons";
import { useField, type Option } from "./field";
import { choiceField, field as fieldStyles, toggleSelect } from "./styles";

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

  const group = choiceField({ invalid: Boolean(error) });
  const fieldSlots = fieldStyles();
  const slots = toggleSelect({ invalid: Boolean(error) });
  return (
    <fieldset
      className={group.root({ className: slots.root() })}
      disabled={disabled}
      aria-describedby={describedBy}
      aria-invalid={error ? true : undefined}
    >
      <legend className={fieldSlots.label({ className: [group.legend(), slots.legend()] })}>{label}</legend>
      <div className={slots.options()}>
        {options.map((option) => {
          const pressed = selected.includes(option.value);
          return (
            <button
              key={option.value}
              type="button"
              className={slots.option()}
              disabled={disabled || option.disabled}
              aria-pressed={pressed}
              aria-describedby={describedBy}
              onBlur={field.handleBlur}
              onClick={() => field.handleChange(pressed ? [] : [option.value])}
            >
              <span className={slots.text()}>
                <span>{option.label}</span>{" "}
                {option.description ? <span className={slots.description()}>{option.description}</span> : null}
              </span>
              <CheckIcon className={iconClassName("compact", slots.check())} aria-hidden="true" />
            </button>
          );
        })}
      </div>
      {hint ? <p className={fieldSlots.hint()} id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className={fieldSlots.error()} id={`${id}-error`} role="alert">{error}</p> : null}
    </fieldset>
  );
}
