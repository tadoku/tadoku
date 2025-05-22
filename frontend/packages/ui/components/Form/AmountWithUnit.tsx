import React, { HTMLProps } from 'react'
import { FieldValues, useFormContext } from 'react-hook-form'
import { FormElementProps, Option } from './types'

interface AmountWithUnitProps<T extends FieldValues>
  extends FormElementProps<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name' | 'type'> {
  label: string
  units: Option[]
}

export function AmountWithUnit<T extends FieldValues>(
  props: AmountWithUnitProps<T>,
) {
  const { name, label, units, ...inputProps } = props

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
        className="block mb-1 font-medium text-sm text-gray-700"
        htmlFor={`${name}.value`}
      >
        {label}
      </label>
      <div className="flex h-11 overflow-visible input-frame group focus-within:ring-2 focus-within:ring-primary/40 focus-within:border-primary transition ease-in-out">
        <input
          type="number"
          step="any"
          {...inputProps}
          id={`${name}.value`}
          {...register(`${name}.value`, { valueAsNumber: true })}
          className="!border-l-0 !border-t-0 !border-b-0 !border-r border-black/10 focus:!border-black/10 !h-10 !bg-none focus:!ring-0 focus:!outline-none w-full"
        />
        <select
          {...register(`${name}.unit`)}
          className="w-auto !h-10 px-2 pr-8 border-none focus:!ring-0 focus:!outline-none bg-transparent"
        >
          {units.map(unit => (
            <option key={unit.value} value={unit.value}>
              {unit.label}
            </option>
          ))}
        </select>
      </div>
      {hasError ? (
        <p className="text-sm text-red-600 mt-1">{errorMessage}</p>
      ) : null}
    </div>
  )
}
