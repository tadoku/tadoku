import * as stylex from "@stylexjs/stylex";
import {
  useId,
  type InputHTMLAttributes,
  type ReactNode,
  type SelectHTMLAttributes,
  type TextareaHTMLAttributes,
} from "react";
import { ChevronDownIcon } from "../../icons";
import { icon } from "../../foundations/recipes";
import { FieldFrame, FieldMessages, describedBy, useField, type FieldContract, type Option, type OptionGroup } from "./field";
import { field as fieldStyles, input, select as selectStyles } from "./field-styles";
import { choice } from "./choice-styles";

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
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error} disabled={props.disabled}>
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
        {...stylex.props(
          input.base,
          input.textArea,
          props.readOnly && input.readOnly,
          Boolean(error) && input.invalid,
          props.disabled && input.disabled,
        )}
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
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error} disabled={props.disabled}>
      <span {...stylex.props(selectStyles.control)}>
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
          {...stylex.props(input.base, input.select, Boolean(error) && input.invalid, props.disabled && input.disabled)}
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
        <ChevronDownIcon {...stylex.props(icon.compact, selectStyles.icon)} aria-hidden="true" />
      </span>
    </FieldFrame>
  );
}

export interface CheckboxProps
  extends FieldContract,
    Omit<InputHTMLAttributes<HTMLInputElement>, "name" | "type" | "onChange" | "onBlur" | "required"> {
  /** `card` frames the choice as a selectable row, for choices with their own details. */
  readonly variant?: "default" | "card";
  /** A card's secondary line under its label. */
  readonly description?: ReactNode;
  /** A card's trailing detail, such as a value or why it is unavailable. */
  readonly aside?: ReactNode;
}

export function Checkbox({
  label,
  hint,
  required,
  id: providedId,
  variant = "default",
  description,
  aside,
  ...props
}: CheckboxProps) {
  const id = providedId ?? `paper-checkbox-${useId().replace(/:/gu, "")}`;
  const { field, error } = useField<boolean | string[]>();
  const selected = field.state.value;
  const option = typeof props.value === "string" ? props.value : undefined;
  const card = variant === "card";
  const checked = option === undefined ? selected === true : Array.isArray(selected) && selected.includes(option);
  // A card's label element wraps its details too, so the box is named by the label text alone and described by the rest.
  const details = card
    ? [description ? `${id}-description` : undefined, aside ? `${id}-aside` : undefined].filter(Boolean).join(" ")
    : "";
  const external = [details, props["aria-describedby"]].filter(Boolean).join(" ") || undefined;
  return (
    <fieldset {...stylex.props(fieldStyles.choiceGroup)}>
      <label
        {...stylex.props(
          choice.option,
          card && choice.card,
          card && checked && choice.cardChecked,
          card && props.disabled && choice.cardDisabled,
        )}
        htmlFor={id}
      >
        <input
          {...props}
          name={field.name}
          checked={checked}
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
          aria-labelledby={card ? `${id}-label` : undefined}
          aria-describedby={describedBy(id, hint, error, external)}
          {...stylex.props(
            choice.box,
            card && choice.cardBox,
            Boolean(error) && choice.boxInvalid,
            props.disabled && choice.checkboxDisabled,
          )}
        />
        {card ? (
          <>
            <span {...stylex.props(choice.body)}>
              <span {...stylex.props(choice.label, props.disabled && choice.labelDisabled)} id={`${id}-label`}>{label}</span>
              {description ? <span {...stylex.props(choice.detail)} id={`${id}-description`}>{description}</span> : null}
            </span>
            {aside ? <span {...stylex.props(choice.detail, choice.aside)} id={`${id}-aside`}>{aside}</span> : null}
          </>
        ) : <span {...stylex.props(props.disabled && choice.textDisabled)}>{label}</span>}
      </label>
      <FieldMessages id={id} hint={hint} error={error} hintStyle={choice.hint} />
    </fieldset>
  );
}
