import { createFormHookContexts, useStore } from "@tanstack/react-form"
import type { ClipboardEvent, ReactNode } from "react";
import { field as fieldStyles } from "./styles";

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
  const slots = fieldStyles({ invalid: Boolean(error) });
  return (
    <div className={slots.root()}>
      <label className={slots.label()} htmlFor={id} id={`${id}-label`}>
        {label}
        {required ? <span className={slots.required()} aria-hidden="true">*</span> : null}
      </label>
      {children}
      {hint ? <p className={slots.hint()} id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className={slots.error()} id={`${id}-error`} role="alert">{error}</p> : null}
    </div>
  );
}

export function describedBy(id: string, hint?: string, error?: unknown, external?: string): string | undefined {
  return [hint ? `${id}-hint` : undefined, error ? `${id}-error` : undefined, external]
    .filter(Boolean)
    .join(" ") || undefined;
}

/**
 * Enters a number pasted with surrounding whitespace, such as a copied spreadsheet cell, which Firefox would otherwise
 * leave as an empty number input. Call it after the caller's own `onPaste`, which can cancel the paste.
 */
export function enterTrimmedNumber(event: ClipboardEvent<HTMLInputElement>, enter: (value: number) => void) {
  const input = event.currentTarget;
  if (event.defaultPrevented || input.disabled || input.readOnly) return;

  const pasted = event.clipboardData.getData("text/plain");
  const trimmed = pasted.trim();
  if (trimmed === pasted) return;

  const parsed = input.cloneNode() as HTMLInputElement;
  parsed.value = trimmed;
  if (!Number.isFinite(parsed.valueAsNumber)) return;

  event.preventDefault();
  enter(parsed.valueAsNumber);
}
