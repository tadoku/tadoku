import { NodeInputProps } from './helpers'

export function NodeInputHidden<T>({ attributes, register }: NodeInputProps) {
  // Render a hidden input field
  return (
    <input
      {...register(attributes.name, {
        required: attributes.required,
        value: attributes.value,
      })}
      type={attributes.type}
    />
  )
}
