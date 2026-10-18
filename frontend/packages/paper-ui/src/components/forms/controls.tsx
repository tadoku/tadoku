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
  name,
  label,
  hint,
  rules,
  required,
  id: providedId,
  ...props
}: TextAreaProps) {
  const id = providedId ?? `paper-textarea-${useId().replace(/:/gu, "")}`;
  const { form, error } = useField(name);
  const registration = form.register(name, {
    ...rules,
    disabled: props.disabled,
    required: rules?.required ?? (required ? "This field is required." : undefined),
  });
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error?.message?.toString()}>
      <textarea
        {...props}
        {...registration}
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
  name,
  label,
  hint,
  rules,
  required,
  options,
  groups,
  placeholder,
  id: providedId,
  ...props
}: SelectProps) {
  const id = providedId ?? `paper-select-${useId().replace(/:/gu, "")}`;
  const { form, error } = useField(name);
  const registration = form.register(name, {
    ...rules,
    disabled: props.disabled,
    required: rules?.required ?? (required ? "Choose an option." : undefined),
  });
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error?.message?.toString()}>
      <span className="paper-select__control">
        <select
          {...props}
          {...registration}
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
  name,
  label,
  hint,
  rules,
  required,
  id: providedId,
  ...props
}: CheckboxProps) {
  const id = providedId ?? `paper-checkbox-${useId().replace(/:/gu, "")}`;
  const { form, error } = useField(name);
  const registration = form.register(name, {
    ...rules,
    disabled: props.disabled,
    required: rules?.required ?? (required ? "Select this option to continue." : undefined),
  });
  return (
    <fieldset className={`paper-choice-field${error ? " paper-field--invalid" : ""}`}>
      <label className="paper-choice" htmlFor={id}>
        <input
          {...props}
          {...registration}
          id={id}
          type="checkbox"
          required={required}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy(id, hint, error, props["aria-describedby"])}
        />
        <span>{label}</span>
      </label>
      {hint ? <p className="paper-field__hint" id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error.message?.toString()}</p> : null}
    </fieldset>
  );
}
