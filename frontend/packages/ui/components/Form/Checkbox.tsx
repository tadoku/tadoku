import React, { HTMLProps } from 'react'
import { FieldValues, useFormContext } from 'react-hook-form'
import { FormElementProps } from './types'

interface CheckboxProps<T extends FieldValues>
  extends FormElementProps<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name' | 'type'> {
  label: string
  hint?: string
}

export function Checkbox<T extends FieldValues>(props: CheckboxProps<T>) {
  const { name, label, options, hint, ...inputProps } = props

  const {
    register,
    formState: { errors },
  } = useFormContext()

  const hasError = errors[name] !== undefined
  const errorMessage =
    errors[name]?.message?.toString() || 'This input is invalid'

  return (
    <div className={`label ${hasError ? 'error' : ''}`}>
      <label className="label-inline" htmlFor={name}>
        <input
          type="checkbox"
          id={name}
          {...inputProps}
          className={`input ${inputProps?.className ?? ''}`}
          {...register(name, options)}
        />
        <span>{label}</span>
      </label>
      {hint ? <span className="label-hint">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </div>
  )
}
