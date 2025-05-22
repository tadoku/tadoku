import React, { HTMLProps } from 'react'
import { FieldValues, useFormContext } from 'react-hook-form'
import { FormElementProps, Option } from './types'

interface SelectProps<T extends FieldValues>
  extends FormElementProps<T>,
    Omit<HTMLProps<HTMLSelectElement>, 'name'> {
  label: string
  values: Option[]
  hint?: string
}

export function Select<T extends FieldValues>(props: SelectProps<T>) {
  const { name, label, type, options, values, hint, ...selectProps } = props

  const {
    register,
    formState: { errors },
  } = useFormContext()

  const hasError = errors[name] !== undefined
  const errorMessage =
    errors[name]?.message?.toString() || 'This selection is invalid'

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">
        {label}
        {hint ? (
          <span className="label-hint hidden sm:flex">{hint}</span>
        ) : undefined}
      </span>
      <select
        {...selectProps}
        className={`input ${selectProps?.className ?? ''}`}
        {...register(name, options)}
      >
        {values.map(({ value, label }) => (
          <option value={value} key={value}>
            {label}
          </option>
        ))}
      </select>
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </label>
  )
}
