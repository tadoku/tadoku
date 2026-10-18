import { useId, type InputHTMLAttributes } from "react";
import { useFormState } from "react-hook-form";
import { FieldFrame, describedBy, useField, type Option } from "./field";

export interface AmountWithUnitProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, "name" | "type"> {
  readonly name: string;
  readonly label: string;
  readonly units: readonly Option<string>[];
  readonly unitsLabel?: string;
  readonly hint?: string;
}

export function AmountWithUnit({
  name,
  label,
  units,
  unitsLabel,
  hint,
  required,
  min,
  max,
  disabled,
  ...props
}: AmountWithUnitProps) {
  const id = `paper-amount-${useId().replace(/:/gu, "")}`;
  const amountName = `${name}Value`;
  const unitName = `${name}Unit`;
  const { form, error: amountError } = useField(amountName);
  const unitState = useFormState({ control: form.control, name: unitName });
  const unitError = form.getFieldState(unitName, unitState).error;
  const error = amountError ?? unitError;
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error?.message?.toString()}>
      <div className="paper-compound-field" role="group" aria-labelledby={`${id}-label`}>
        <input
          {...props}
          {...form.register(amountName, {
            valueAsNumber: true,
            disabled,
            required: required ? "Enter an amount." : undefined,
            min: min === undefined ? undefined : { value: min, message: `Enter at least ${min}.` },
            max: max === undefined ? undefined : { value: max, message: `Enter no more than ${max}.` },
          })}
          required={required}
          min={min}
          max={max}
          disabled={disabled}
          id={id}
          type="number"
          className="paper-input"
          aria-invalid={amountError ? true : undefined}
          aria-describedby={describedBy(id, hint, error, props["aria-describedby"])}
        />
        <select
          {...form.register(unitName, { disabled })}
          disabled={disabled}
          aria-label={unitsLabel ?? `Unit for ${label.toLocaleLowerCase()}`}
          className="paper-input paper-compound-field__unit"
          aria-invalid={unitError ? true : undefined}
          aria-describedby={describedBy(id, hint, error, props["aria-describedby"])}
        >
          {units.map((unit) => <option key={unit.value} value={unit.value} disabled={unit.disabled}>{unit.label}</option>)}
        </select>
      </div>
    </FieldFrame>
  );
}
