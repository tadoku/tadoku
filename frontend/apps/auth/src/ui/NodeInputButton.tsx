import { getNodeLabel } from '@ory/integrations/ui'
import { useFormContext } from 'react-hook-form'

import { NodeInputProps } from './helpers'

export function NodeInputButton<T>({
  node,
  attributes,
  disabled,
  dispatchSubmit,
}: NodeInputProps) {
  const { register, setValue } = useFormContext()

  return (
    <>
      <button
        type="button"
        {...register(attributes.name)}
        onClick={e => {
          // This section is only used for WebAuthn. The script is loaded via a <script> node
          // and the functions are available on the global window level. Unfortunately, there
          // is currently no better way than executing eval / function here at this moment.
          //
          // Please note that we also need to prevent the default action from happening.
          if (attributes.onclick) {
            e.stopPropagation()
            e.preventDefault()
            const run = new Function(attributes.onclick)
            run()
            return
          }

          setValue(attributes.name, attributes.value)
          dispatchSubmit(e)
          setValue(attributes.name, undefined)
        }}
        disabled={attributes.disabled || disabled}
        className="btn"
      >
        {getNodeLabel(node)}
      </button>
    </>
  )
}
