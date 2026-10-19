import {
  useId,
  type InputHTMLAttributes,
  type SelectHTMLAttributes,
  type TextareaHTMLAttributes,
} from "react";
import { ChevronDownIcon, iconClassName } from "../../icons";
import { FieldFrame, describedBy, useField, type FieldContract, type Option, type OptionGroup } from "./field";

export interface TextAreaProps
  extends FieldContract,
    Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, "name" | "onChange" | "onBlur" | "required"> {}

export function TextArea({
  label,
  hint,
  required,
  id: providedId,
  ...props
}: TextAreaProps) {
  const id = providedId ?? `paper-textarea-${useId().replace(/:/gu, "")}`;
  const { field, error } = useField<string>();
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error}>
      <textarea
        {...props}
        name={field.name}
        value={field.state.value ?? ""}
        onChange={event => field.handleChange(event.currentTarget.value)}
        onBlur={field.handleBlur}
        id={id}
        required={required}
        aria-invalid={error ? true : undefined}
        aria-describedby={describedBy(id, hint, error, props["aria-describedby"])}
        className="paper-input paper-textarea"
      />
    </FieldFrame>
  );
}

export interface SelectProps
  extends FieldContract,
    Omit<SelectHTMLAttributes<HTMLSelectElement>, "name" | "onChange" | "onBlur" | "required"> {
  readonly options: readonly Option<string>[];
  readonly groups?: readonly OptionGroup<string>[];
  readonly placeholder?: string;
}

export function Select({
  label,
  hint,
  required,
  options,
  groups,
  placeholder,
  id: providedId,
  ...props
}: SelectProps) {
  const id = providedId ?? `paper-select-${useId().replace(/:/gu, "")}`;
  const { field, error } = useField<string>();
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error}>
      <span className="paper-select__control">
        <select
          {...props}
          name={field.name}
          value={field.state.value ?? ""}
          onChange={event => field.handleChange(event.currentTarget.value)}
          onBlur={field.handleBlur}
          id={id}
          required={required}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy(id, hint, error, props["aria-describedby"])}
          className="paper-input paper-select"
        >
          {placeholder ? <option value="">{placeholder}</option> : null}
          {groups
            ? groups.map((group) => (
                <optgroup key={group.label} label={group.label}>
                  {group.options.map((option) => (
                    <option key={option.value} value={option.value} disabled={option.disabled}>
                      {option.label}
                    </option>
                  ))}
                </optgroup>
              ))
            : options.map((option) => (
                <option key={option.value} value={option.value} disabled={option.disabled}>
                  {option.label}
                </option>
              ))}
        </select>
        <ChevronDownIcon
          className={iconClassName("compact", "paper-select__icon")}
          aria-hidden="true"
        />
      </span>
    </FieldFrame>
  );
}

export interface CheckboxProps
  extends FieldContract,
    Omit<InputHTMLAttributes<HTMLInputElement>, "name" | "type" | "onChange" | "onBlur" | "required"> {}

export function Checkbox({
  label,
  hint,
  required,
  id: providedId,
  ...props
}: CheckboxProps) {
  const id = providedId ?? `paper-checkbox-${useId().replace(/:/gu, "")}`;
  const { field, error } = useField<boolean | string[]>();
  const selected = field.state.value;
  const option = typeof props.value === "string" ? props.value : undefined;
  return (
    <fieldset className={`paper-choice-field${error ? " paper-field--invalid" : ""}`}>
      <label className="paper-choice" htmlFor={id}>
        <input
          {...props}
          name={field.name}
          checked={option === undefined ? selected === true : Array.isArray(selected) && selected.includes(option)}
          onChange={event => {
            if (option === undefined) {
              field.handleChange(event.currentTarget.checked);
            } else {
              const values = Array.isArray(selected) ? selected : [];
              field.handleChange(event.currentTarget.checked
                ? [...values, option]
                : values.filter(value => value !== option));
            }
          }}
          onBlur={field.handleBlur}
          id={id}
          type="checkbox"
          required={required}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy(id, hint, error, props["aria-describedby"])}
        />
        <span>{label}</span>
      </label>
      {hint ? <p className="paper-field__hint" id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error}</p> : null}
    </fieldset>
  );
}
