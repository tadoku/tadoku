import * as stylex from "@stylexjs/stylex";
import { createFormHookContexts, useStore } from "@tanstack/react-form"
import type { ReactNode } from "react";
import { field as styles } from "./field-styles";

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
  disabled,
  children,
}: {
  readonly id: string;
  readonly label: string;
  readonly hint?: string;
  readonly error?: string;
  readonly required?: boolean;
  readonly disabled?: boolean;
  readonly children: ReactNode;
}) {
  return (
    <div {...stylex.props(styles.root, disabled && styles.disabled)}>
      <label {...stylex.props(styles.label)} htmlFor={id} id={`${id}-label`}>
        {label}
        <RequiredMark required={required} />
      </label>
      {children}
      <FieldMessages id={id} hint={hint} error={error} />
    </div>
  );
}

export function RequiredMark({ required }: { readonly required?: boolean }) {
  return required ? <span {...stylex.props(styles.required)} aria-hidden="true">*</span> : null;
}

/** The hint and error lines under a control, referenced by `describedBy`. */
export function FieldMessages({ id, hint, error, hintStyle }: {
  readonly id: string;
  readonly hint?: string;
  readonly error?: string;
  readonly hintStyle?: stylex.StyleXStyles;
}) {
  return (
    <>
      {hint ? <p {...stylex.props(styles.hint, hintStyle)} id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p {...stylex.props(styles.error)} id={`${id}-error`} role="alert">{error}</p> : null}
    </>
  );
}

export function describedBy(id: string, hint?: string, error?: unknown, external?: string): string | undefined {
  return [hint ? `${id}-hint` : undefined, error ? `${id}-error` : undefined, external]
    .filter(Boolean)
    .join(" ") || undefined;
}
