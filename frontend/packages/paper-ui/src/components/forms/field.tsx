import { createFormHookContexts, useStore } from "@tanstack/react-form"
import type { ReactNode } from "react"

export const { fieldContext, formContext, useFieldContext, useFormContext } = createFormHookContexts();

export interface Option<Value = string> {
  readonly value: Value;
  readonly label: string;
  readonly description?: string;
  readonly disabled?: boolean;
}

export interface OptionGroup<Value = string> {
  readonly label: string;
  readonly options: readonly Option<Value>[];
}

export interface FieldContract {
  readonly label: string;
  readonly hint?: string;
  readonly required?: boolean;
}

export function useField<Value>() {
  const field = useFieldContext<Value>();
  const form = useFormContext();
  const submitted = useStore(form.store, state => state.submissionAttempts > 0);
  const errors = useStore(field.store, state => state.meta.errors);
  const showErrors = submitted || Boolean(field.state.meta.errorMap.onSubmit) ||
    (field.state.meta.isTouched && Boolean(field.options.validators?.onBlur));
  const first = showErrors ? errors[0] : undefined;
  const error = typeof first === "string" ? first : first?.message;
  return { field, error };
}

export function FieldFrame({
  id,
  label,
  hint,
  error,
  required,
  children,
}: {
  readonly id: string;
  readonly label: string;
  readonly hint?: string;
  readonly error?: string;
  readonly required?: boolean;
  readonly children: ReactNode;
}) {
  return (
    <div className={`paper-field${error ? " paper-field--invalid" : ""}`}>
      <label className="paper-field__label" htmlFor={id} id={`${id}-label`}>
        {label}
        {required ? <span className="paper-field__required" aria-hidden="true">*</span> : null}
      </label>
      {children}
      {hint ? <p className="paper-field__hint" id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error}</p> : null}
    </div>
  );
}

export function describedBy(id: string, hint?: string, error?: unknown, external?: string): string | undefined {
  return [hint ? `${id}-hint` : undefined, error ? `${id}-error` : undefined, external]
    .filter(Boolean)
    .join(" ") || undefined;
}
