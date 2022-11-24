import { UiNode } from '@ory/client'
import {
  isUiNodeInputAttributes,
  isUiNodeImageAttributes,
  isUiNodeScriptAttributes,
  isUiNodeTextAttributes,
  isUiNodeAnchorAttributes,
} from '@ory/integrations/ui'
import { UseFormRegister } from 'react-hook-form'
import { NodeAnchor } from './NodeAnchor'
import { NodeImage } from './NodeImage'
import { NodeInput } from './NodeInput'
import { NodeScript } from './NodeScript'
import { NodeText } from './NodeText'

interface NodeProps {
  node: UiNode
  disabled: boolean
  register: UseFormRegister<any>
}

export const Node = ({ node, disabled, register }: NodeProps) => {
  if (isUiNodeImageAttributes(node.attributes)) {
    return <NodeImage node={node} attributes={node.attributes} />
  }

  if (isUiNodeScriptAttributes(node.attributes)) {
    return <NodeScript node={node} attributes={node.attributes} />
  }

  if (isUiNodeTextAttributes(node.attributes)) {
    return <NodeText node={node} attributes={node.attributes} />
  }

  if (isUiNodeAnchorAttributes(node.attributes)) {
    return <NodeAnchor attributes={node.attributes} />
  }

  if (isUiNodeInputAttributes(node.attributes)) {
    return (
      <NodeInput
        node={node}
        disabled={disabled}
        attributes={node.attributes}
        register={register}
      />
    )
  }

  return null
}

export default Node
