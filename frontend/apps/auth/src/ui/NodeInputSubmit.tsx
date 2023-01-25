import { getNodeLabel } from '@ory/integrations/ui'
import { useFormContext } from 'react-hook-form'

import { NodeInputProps } from './helpers'

export function NodeInputSubmit<T>({
  node,
  attributes,
  disabled,
  dispatchSubmit,
}: NodeInputProps) {
  const { register, setValue } = useFormContext()

  return (
    <>
      <button
        type="submit"
        {...register(attributes.name)}
        onClick={e => {
          setValue(attributes.name, attributes.value)
          dispatchSubmit(e)
          setValue(attributes.name, undefined)
        }}
        disabled={attributes.disabled || disabled}
        className="btn primary"
      >
        {getNodeLabel(node).replaceAll('Sign in', 'Log in')}
      </button>
    </>
  )
}
