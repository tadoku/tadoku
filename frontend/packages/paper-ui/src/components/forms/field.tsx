import type { ReactNode } from "react";
import {
  useFormContext,
  useFormState,
  type FieldValues,
  type RegisterOptions,
} from "react-hook-form";

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
  readonly name: string;
  readonly label: string;
  readonly hint?: string;
  readonly required?: boolean;
  readonly rules?: RegisterOptions<FieldValues, string>;
}

export function useField(name: string) {
  const form = useFormContext();
  if (!form) throw new Error("Paper form controls require react-hook-form FormProvider");
  const formState = useFormState({ control: form.control, name });
  const error = form.getFieldState(name, formState).error;
  return { form, error };
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
