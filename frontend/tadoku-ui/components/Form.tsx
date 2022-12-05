import { HTMLInputTypeAttribute } from 'react'
import { RegisterOptions, UseFormRegister } from 'react-hook-form'

export function Input({
  name,
  label,
  register,
  type,
  options,
}: {
  name: string
  label: string
  register: UseFormRegister<any>
  type: HTMLInputTypeAttribute
  options: RegisterOptions
}) {
  return (
    <label className="label" htmlFor={name}>
      <span className="label-text">{label}</span>
      <input type={type} id={name} {...register(name, options)} />
    </label>
  )
}
