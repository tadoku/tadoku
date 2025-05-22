import React, { HTMLProps } from 'react'
import { FieldValues, useFormContext } from 'react-hook-form'
import { FormElementProps, Option } from './types'

interface AmountWithUnitProps<T extends FieldValues>
  extends FormElementProps<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name' | 'type'> {
  label: string
  units: Option[]
  unitsLabel?: string
}

export function AmountWithUnit<T extends FieldValues>(
  props: AmountWithUnitProps<T>,
) {
  const { name, label, units, unitsLabel, ...inputProps } = props

  const {
    register,
    formState: { errors },
  } = useFormContext()

  const fieldError = errors?.[name as keyof typeof errors] as any

  const hasError = fieldError !== undefined
  const errorMessage =
    fieldError?.message?.toString() || 'This input is invalid'

  return (
    <div className="w-full">
      <label
        className="label mb-2"
        htmlFor={`${name}-value`}
        id={`${name}-label`}
      >
        <span className="label-text">{label}</span>
      </label>
      <div
        className="flex h-11 overflow-visible input-frame group focus-within:ring-2 focus-within:ring-primary/40 focus-within:border-primary transition ease-in-out"
        role="group"
        aria-labelledby={`${name}-label`}
      >
        <input
          type="number"
          {...inputProps}
          id={`${name}-value`}
          {...register(`${name}-value`, { valueAsNumber: true })}
          className="!border-l-0 !border-t-0 !border-b-0 !border-r border-black/10 focus:!border-black/10 !h-full !bg-none focus:!ring-0 focus:!outline-none w-full"
          aria-invalid={hasError ? 'true' : 'false'}
          aria-describedby={hasError ? `${name}-error` : undefined}
        />
        <select
          {...register(`${name}-unit`)}
          className="w-auto !h-full px-2 pr-8 border-none bg-black/5 focus:!ring-0 focus:!outline-none focus:bg-transparent"
          aria-label={unitsLabel || `Unit for ${label.toLocaleLowerCase()}`}
        >
          {units.map(unit => (
            <option key={unit.value} value={unit.value}>
              {unit.label}
            </option>
          ))}
        </select>
      </div>
      {hasError ? (
        <p
          id={`${name}-error`}
          className="text-sm text-red-600 mt-1"
          role="alert"
        >
          {errorMessage}
        </p>
      ) : null}
    </div>
  )
}
