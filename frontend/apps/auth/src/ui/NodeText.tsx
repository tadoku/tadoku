import { UiNode, UiNodeTextAttributes } from '@ory/client'
import { UiText } from '@ory/client'

interface Props {
  node: UiNode
  attributes: UiNodeTextAttributes
}

const Content = ({ node, attributes }: Props) => {
  switch (attributes.text.id) {
    case 1050015:
      // This text node contains lookup secrets. Let's make them a bit more beautiful!
      const secrets = (attributes.text.context as any).secrets.map(
        (text: UiText, k: number) => (
          <div key={k}>
            {/* Used lookup_secret has ID 1050014 */}
            <code>{text.id === 1050014 ? 'Used' : text.text}</code>
          </div>
        ),
      )
      return (
        <div>
          <div className="row">{secrets}</div>
        </div>
      )
  }

  return <div>{attributes.text.text}</div>
}

export const NodeText = ({ node, attributes }: Props) => {
  return (
    <>
      <p>{node.meta?.label?.text}</p>
      <Content node={node} attributes={attributes} />
    </>
  )
}
