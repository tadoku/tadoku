import { UiNode } from '@ory/client'
import {
  isUiNodeInputAttributes,
  isUiNodeImageAttributes,
  isUiNodeScriptAttributes,
  isUiNodeTextAttributes,
  isUiNodeAnchorAttributes,
} from '@ory/integrations/ui'
import { FormDispatcher } from './helpers'
import { NodeAnchor } from './NodeAnchor'
import { NodeImage } from './NodeImage'
import { NodeInput } from './NodeInput'
import { NodeScript } from './NodeScript'
import { NodeText } from './NodeText'

interface NodeProps {
  node: UiNode
  disabled: boolean
  dispatchSubmit: FormDispatcher
}

export const Node = ({ node, disabled, dispatchSubmit }: NodeProps) => {
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
        dispatchSubmit={dispatchSubmit}
      />
    )
  }

  return null
}

export default Node
