import * as stylex from "@stylexjs/stylex";
import { forwardRef, useId, type InputHTMLAttributes } from "react";
import { sx, type StyleXProp } from "../../../foundations/props";
import { RequiredMark, useField } from "../field";
import { field as fieldStyles, input } from "../field-styles";

export interface InputProps
  extends Omit<
    InputHTMLAttributes<HTMLInputElement>,
    "name" | "onBlur" | "onChange"
  > {
  readonly label: string;
  readonly hint?: string;
  /** Layout for the field's wrapper, such as its grid placement. */
  readonly xstyle?: StyleXProp;
  /** Styles for the hint line, such as keeping a stored value's spacing. */
  readonly hintXstyle?: StyleXProp;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  {
    label,
    hint,
    id: providedId,
    required,
    readOnly,
    disabled,
    className,
    xstyle,
    hintXstyle,
    ...props
  },
  forwardedRef,
) {
  const { field, error } = useField<string | number>();
  const generatedId = useId();
  const id = providedId ?? `paper-input-${generatedId.replace(/:/gu, "")}`;

  const hintId = hint ? `${id}-hint` : undefined;
  const errorId = error ? `${id}-error` : undefined;
  const describedBy = [hintId, errorId, props["aria-describedby"]].filter(Boolean).join(" ") || undefined;

  return (
    <div {...sx(className, fieldStyles.root, disabled && fieldStyles.disabled, xstyle)}>
      <label {...stylex.props(fieldStyles.label)} htmlFor={id}>
        {label}
        <RequiredMark required={required} />
      </label>
      <input
        {...props}
        name={field.name}
        value={typeof field.state.value === "number" && Number.isNaN(field.state.value) ? "" : field.state.value ?? ""}
        onChange={event => field.handleChange(props.type === "number" ? event.currentTarget.valueAsNumber : event.currentTarget.value)}
        onBlur={field.handleBlur}
        id={id}
        required={required}
        readOnly={readOnly}
        disabled={disabled}
        aria-invalid={error ? true : undefined}
        aria-describedby={describedBy}
        {...stylex.props(input.base, readOnly && input.readOnly, error && input.invalid, disabled && input.disabled)}
        ref={forwardedRef}
      />
      {hint ? (
        <p {...stylex.props(fieldStyles.hint, hintXstyle)} id={hintId}>
          {hint}
        </p>
      ) : null}
      {error ? (
        <p {...stylex.props(fieldStyles.error)} id={errorId} role="alert">
          {error || "Check this field."}
        </p>
      ) : null}
    </div>
  );
});
