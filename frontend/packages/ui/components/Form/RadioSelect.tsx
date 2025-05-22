import React, { HTMLProps } from 'react'
import { FieldValues, useFormContext } from 'react-hook-form'
import { FormElementProps } from './types'
import { Option } from './types'

interface RadioSelectProps<T extends FieldValues>
  extends FormElementProps<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name'> {
  label: string
  values: Option[]
  hint?: string
}

export function RadioSelect<T extends FieldValues>(props: RadioSelectProps<T>) {
  const { name, label, type, options, values, hint, ...selectProps } = props

  const {
    register,
    formState: { errors },
  } = useFormContext()

  const hasError = errors[name] !== undefined
  let errorMessage =
    errors[name]?.message?.toString() || 'This selection is invalid'

  return (
    <div className={`label ${hasError ? 'error' : ''}`}>
      <span className="label-text">
        {label}
        {hint ? (
          <span className="label-hint hidden sm:flex">{hint}</span>
        ) : undefined}
      </span>
      {values.map(({ value, label }) => (
        <label
          className="label-inline"
          htmlFor={`${name}-${value}`}
          key={value}
        >
          <input
            type="radio"
            value={value}
            id={`${name}-${value}`}
            {...selectProps}
            {...register(name, options)}
          />
          <span>{label}</span>
        </label>
      ))}
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </div>
  )
}
