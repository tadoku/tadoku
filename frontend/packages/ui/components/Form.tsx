import React, { HTMLProps } from 'react'
import {
  FieldPath,
  FieldValues,
  RegisterOptions,
  useFormContext,
} from 'react-hook-form'

interface Props<T extends FieldValues> {
  name: FieldPath<T>
  options?: RegisterOptions
}

interface InputProps<T extends FieldValues>
  extends Props<T>,
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

interface TextaAreaProps<T extends FieldValues>
  extends Props<T>,
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

interface CheckboxProps<T extends FieldValues>
  extends Props<T>,
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

export interface Option {
  value: string
  label: string
}

interface SelectProps<T extends FieldValues>
  extends Props<T>,
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

interface RadioSelectProps<T extends FieldValues>
  extends Props<T>,
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

export { AutocompleteInput } from './Form/AutocompleteInput'
export { AutocompleteMultiInput } from './Form/AutocompleteMultiInput'
export { RadioGroup } from './Form/RadioGroup'
export { AmountWithUnit } from './Form/AmountWithUnit'
