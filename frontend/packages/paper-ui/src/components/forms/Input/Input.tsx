import { useField } from "../field"
import { forwardRef, useId, type InputHTMLAttributes } from "react"

export interface InputProps
  extends Omit<
    InputHTMLAttributes<HTMLInputElement>,
    "name" | "onBlur" | "onChange"
  > {
  readonly label: string;
  readonly hint?: string;
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
    <div
      className={[
        "paper-field",
        error ? "paper-field--invalid" : "",
        disabled ? "paper-field--disabled" : "",
        className ?? "",
      ]
        .filter(Boolean)
        .join(" ")}
    >
      <label className="paper-field__label" htmlFor={id}>
        {label}
        {required ? (
          <span className="paper-field__required" aria-hidden="true">
            *
          </span>
        ) : null}
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
        className="paper-input"
        ref={forwardedRef}
      />
      {hint ? (
        <p className="paper-field__hint" id={hintId}>
          {hint}
        </p>
      ) : null}
      {error ? (
        <p className="paper-field__error" id={errorId} role="alert">
          {error || "Check this field."}
        </p>
      ) : null}
    </div>
  );
});
