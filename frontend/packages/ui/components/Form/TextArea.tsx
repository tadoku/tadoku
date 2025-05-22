import React, { HTMLProps } from 'react'
import { FieldValues, useFormContext } from 'react-hook-form'
import { FormElementProps } from './types'

interface TextaAreaProps<T extends FieldValues>
  extends FormElementProps<T>,
    Omit<HTMLProps<HTMLTextAreaElement>, 'name'> {
  label: string
  hint?: string
}

export function TextArea<T extends FieldValues>(props: TextaAreaProps<T>) {
  const { name, label, options, hint, ...inputProps } = props

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
      <textarea
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
