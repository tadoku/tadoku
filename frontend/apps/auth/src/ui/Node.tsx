import { UiNode } from '@ory/kratos-client'
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
  if (node.attributes.node_type === 'img') {
    return <NodeImage node={node} attributes={node.attributes} />
  }

  if (node.attributes.node_type === 'script') {
    return <NodeScript node={node} attributes={node.attributes} />
  }

  if (node.attributes.node_type === 'text') {
    return <NodeText node={node} attributes={node.attributes} />
  }

  if (node.attributes.node_type === 'a') {
    return <NodeAnchor attributes={node.attributes} />
  }

  if (node.attributes.node_type === 'input') {
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
