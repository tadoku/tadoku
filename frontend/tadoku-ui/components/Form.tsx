import { HTMLInputTypeAttribute, HTMLProps } from 'react'
import {
  FieldPath,
  FieldValues,
  FormState,
  RegisterOptions,
  UseFormRegister,
} from 'react-hook-form'

interface Props<T extends FieldValues> {
  name: FieldPath<T>
  register: UseFormRegister<T>
  formState: FormState<T>
  options: RegisterOptions
}

interface InputProps<T extends FieldValues>
  extends Props<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name'> {
  label: string
}

export function Input<T extends FieldValues>(props: InputProps<T>) {
  const { name, register, formState, label, type, options, ...inputProps } =
    props
  const hasError = formState.errors[name] !== undefined
  const errorMessage =
    formState.errors[name]?.message?.toString() ?? 'This field is invalid'

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">{label}</span>
      <input type={type} id={name} />
      <span className="error">{errorMessage}</span>
    </label>
  )
}

interface Option {
  value: string
  label: string
}

interface SelectProps<T extends FieldValues>
  extends Props<T>,
    Omit<HTMLProps<HTMLSelectElement>, 'name'> {
  label: string
  values: Option[]
}

export function Select<T extends FieldValues>(props: SelectProps<T>) {
  const {
    name,
    register,
    formState,
    label,
    type,
    options,
    values,
    ...selectProps
  } = props
  const hasError = formState.errors[name] !== undefined
  const errorMessage =
    formState.errors[name]?.message?.toString() ?? 'This field is invalid'

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">{label}</span>
      <select {...selectProps} {...register(name, options)}>
        {values.map(({ value, label }) => (
          <option value={value}>{label}</option>
        ))}
      </select>
      <span className="error">{errorMessage}</span>
    </label>
  )
}
