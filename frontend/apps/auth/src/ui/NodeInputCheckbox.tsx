import { useFormContext } from 'react-hook-form'

import { NodeInputProps } from './helpers'

export function NodeInputCheckbox<T>({
  node,
  attributes,
  disabled,
}: NodeInputProps) {
  const { register } = useFormContext()

  // Render a checkbox
  return (
    <div
      style={{
        backgroundColor: node.messages.find(({ type }) => type === 'error')
          ? 'red'
          : 'inherit',
      }}
    >
      <label htmlFor={attributes.name}>
        {node.meta.label?.text ?? attributes.name}
      </label>
      <input
        {...register(attributes.name)}
        type="checkbox"
        id={attributes.name}
        disabled={attributes.disabled || disabled}
      />
      <p>{node.messages.map(({ text }) => text).join('\n')}</p>
    </div>
  )
}
