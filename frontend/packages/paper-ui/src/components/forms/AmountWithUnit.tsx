import { useId, type InputHTMLAttributes } from 'react';
import { FieldFrame, describedBy, useField, type Option } from './field';

export interface AmountWithUnitProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, 'name' | 'type' | 'value' | 'onChange' | 'onBlur'> {
  readonly label: string;
  readonly units: readonly Option<string>[];
  readonly unitsLabel?: string;
  readonly hint?: string;
  readonly unitField: {
    readonly name: string;
    readonly state: { readonly value: string };
    readonly handleChange: (value: string) => void;
    readonly handleBlur: () => void;
  };
}

export function AmountWithUnit({ label, units, unitsLabel, unitField, hint, required, disabled, ...props }: AmountWithUnitProps) {
  const id = `paper-amount-${useId().replace(/:/gu, '')}`;
  const { field, error } = useField<number | string>();
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error}>
      <div className="paper-compound-field" role="group" aria-labelledby={`${id}-label`}>
        <input
          {...props}
          name={field.name}
          value={typeof field.state.value === 'number' && Number.isNaN(field.state.value) ? '' : field.state.value ?? ''}
          onChange={event => field.handleChange(event.currentTarget.valueAsNumber)}
          onBlur={field.handleBlur}
          required={required}
          disabled={disabled}
          id={id}
          type="number"
          className="paper-input"
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy(id, hint, error, props['aria-describedby'])}
        />
        <select
          name={unitField.name}
          value={unitField.state.value}
          onChange={event => unitField.handleChange(event.currentTarget.value)}
          onBlur={unitField.handleBlur}
          disabled={disabled}
          aria-label={unitsLabel ?? `Unit for ${label.toLocaleLowerCase()}`}
          className="paper-input paper-compound-field__unit"
          aria-describedby={describedBy(id, hint, error, props['aria-describedby'])}
        >
          {units.map(unit => <option key={unit.value} value={unit.value} disabled={unit.disabled}>{unit.label}</option>)}
        </select>
      </div>
    </FieldFrame>
  );
}
