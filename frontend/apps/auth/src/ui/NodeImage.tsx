import { UiNode, UiNodeImageAttributes } from '@ory/client'

interface Props {
  node: UiNode
  attributes: UiNodeImageAttributes
}

export const NodeImage = ({ node, attributes }: Props) => {
  // eslint-disable-next-line @next/next/no-img-element
  return <img src={attributes.src} alt={node.meta.label?.text} />
}
