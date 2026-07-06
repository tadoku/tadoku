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
        value={String(attributes.value ?? '')}
        {...register(attributes.name)}
        onClick={async e => {
          setValue(attributes.name, attributes.value)
          await dispatchSubmit(e)
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
