import React, { HTMLProps, ReactNode } from 'react'
import { FieldValues, useFormContext } from 'react-hook-form'
import { FormElementProps } from './types'

interface InputGroupProps<T extends FieldValues>
  extends FormElementProps<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name' | 'type' | 'prefix'> {
  label: string
  type?: string
  prefix?: string
  suffix?: string
  labelAction?: ReactNode
}

export function InputGroup<T extends FieldValues>(
  props: InputGroupProps<T>,
) {
  const { name, label, type = 'number', prefix, suffix, labelAction, options, ...inputProps } = props

  const {
    register,
    formState: { errors },
  } = useFormContext()

  const fieldError = errors?.[name as keyof typeof errors] as any
  const hasError = fieldError !== undefined
  const errorMessage = fieldError?.message?.toString() || 'This input is invalid'

  return (
    <div className="w-full">
      <div className="flex items-center justify-between mb-2">
        <label
          className="label"
          htmlFor={name}
          id={`${name}-label`}
        >
          <span className="label-text">{label}</span>
        </label>
        {labelAction}
      </div>
      <div
        className="flex h-11 overflow-visible input-frame group focus-within:ring-2 focus-within:ring-primary/40 focus-within:border-primary transition ease-in-out cursor-text"
        role="group"
        aria-labelledby={`${name}-label`}
        onClick={() => document.getElementById(name)?.focus()}
      >
        {prefix ? (
          <span className="flex items-center px-3 bg-black/5 whitespace-nowrap">
            {prefix}
          </span>
        ) : null}
        <input
          type={type}
          {...inputProps}
          id={name}
          {...register(name, options)}
          className={`!border-t-0 !border-b-0 !h-full !bg-none focus:!ring-0 focus:!outline-none w-full ${
            prefix ? '!border-l border-black/10 focus:!border-black/10' : '!border-l-0'
          } ${
            suffix ? '!border-r border-black/10 focus:!border-black/10' : '!border-r-0'
          }`}
          aria-invalid={hasError ? 'true' : 'false'}
          aria-describedby={hasError ? `${name}-error` : undefined}
        />
        {suffix ? (
          <span className="flex items-center px-3 bg-black/5 whitespace-nowrap">
            {suffix}
          </span>
        ) : null}
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
