import { getNodeLabel } from '@ory/integrations/ui'

import { NodeInputProps } from './helpers'

export function NodeInputCheckbox<T>({
  node,
  attributes,
  disabled,
  register,
}: NodeInputProps) {
  // Render a checkbox
  return (
    <div
      style={{
        backgroundColor: node.messages.find(({ type }) => type === 'error')
          ? 'red'
          : 'inherit',
      }}
    >
      <label htmlFor={attributes.name}>{getNodeLabel(node)}</label>
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
