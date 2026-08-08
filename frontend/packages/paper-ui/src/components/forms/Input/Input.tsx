import {
  forwardRef,
  useId,
  type ForwardedRef,
  type InputHTMLAttributes,
  type MutableRefObject,
} from "react";
import {
  useFormContext,
  useFormState,
  type FieldValues,
  type RegisterOptions,
} from "react-hook-form";

export interface InputProps
  extends Omit<
    InputHTMLAttributes<HTMLInputElement>,
    "name" | "onBlur" | "onChange"
  > {
  readonly name: string;
  readonly label: string;
  readonly hint?: string;
  readonly rules?: RegisterOptions<FieldValues, string>;
}

function assignRef(
  ref: ForwardedRef<HTMLInputElement>,
  node: HTMLInputElement | null,
): void {
  if (typeof ref === "function") ref(node);
  else if (ref) (ref as MutableRefObject<HTMLInputElement | null>).current = node;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  {
    name,
    label,
    hint,
    rules,
    id: providedId,
    required,
    readOnly,
    disabled,
    className,
    ...props
  },
  forwardedRef,
) {
  const form = useFormContext();
  const generatedId = useId();
  const id = providedId ?? `paper-input-${generatedId.replace(/:/gu, "")}`;

  if (!form) {
    throw new Error("Paper Input must be rendered inside react-hook-form FormProvider");
  }

  const registration = form.register(name, {
    ...rules,
    required: rules?.required ?? (required ? "This field is required." : undefined),
  });
  const subscribedFormState = useFormState({ control: form.control, name });
  const error = form.getFieldState(name, subscribedFormState).error;
  const hintId = hint ? `${id}-hint` : undefined;
  const errorId = error ? `${id}-error` : undefined;
  const describedBy = [hintId, errorId].filter(Boolean).join(" ") || undefined;

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
      {hint ? (
        <p className="paper-field__hint" id={hintId}>
          {hint}
        </p>
      ) : null}
      <input
        {...props}
        {...registration}
        id={id}
        required={required}
        readOnly={readOnly}
        disabled={disabled}
        aria-invalid={error ? true : undefined}
        aria-describedby={describedBy}
        className="paper-input"
        ref={(node) => {
          registration.ref(node);
          assignRef(forwardedRef, node);
        }}
      />
      {error ? (
        <p className="paper-field__error" id={errorId} role="alert">
          {error.message?.toString() || "Check this field."}
        </p>
      ) : null}
    </div>
  );
});
