import React, { HTMLProps } from 'react'
import { FieldValues, useFormContext } from 'react-hook-form'
import { FormElementProps } from './types'

interface InputProps<T extends FieldValues>
  extends FormElementProps<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name'> {
  label: string
  hint?: string
}

export function Input<T extends FieldValues>(props: InputProps<T>) {
  const { name, label, type, options, hint, ...inputProps } = props

  const {
    register,
    formState: { errors },
  } = useFormContext()

  const hasError = errors[name] !== undefined
  const errorMessage =
    errors[name]?.message?.toString() || 'This input is invalid'

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">
        {label}
        {hint ? (
          <span className="label-hint hidden sm:flex">{hint}</span>
        ) : undefined}
      </span>
      <input
        type={type}
        id={name}
        {...inputProps}
        className={`input ${inputProps?.className ?? ''}`}
        {...register(name, options)}
      />
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </label>
  )
}
