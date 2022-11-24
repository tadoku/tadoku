import { getNodeLabel } from '@ory/integrations/ui'

import { NodeInputProps } from './helpers'

export function NodeInputSubmit<T>({
  node,
  attributes,
  disabled,
  register,
}: NodeInputProps) {
  return (
    <>
      <button
        type="submit"
        {...register(attributes.name)}
        value={attributes.value || ''}
        disabled={attributes.disabled || disabled}
      >
        {getNodeLabel(node)}
      </button>
    </>
  )
}
