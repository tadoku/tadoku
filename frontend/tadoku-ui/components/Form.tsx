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
      <input
        type={type}
        id={name}
        {...inputProps}
        {...register(name, options)}
      />
      <span className="error">{errorMessage}</span>
    </label>
  )
}
