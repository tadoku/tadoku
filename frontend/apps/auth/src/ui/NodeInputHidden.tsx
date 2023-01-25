import { useFormContext } from 'react-hook-form'
import { NodeInputProps } from './helpers'

export function NodeInputHidden<T>({ attributes }: NodeInputProps) {
  // Render a hidden input field
  const { register } = useFormContext()

  return (
    <input
      {...register(attributes.name, {
        required: attributes.required,
        value: attributes.value || 'true',
      })}
      type={attributes.type}
    />
  )
}
