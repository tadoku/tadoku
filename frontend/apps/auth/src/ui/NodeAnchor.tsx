import { UiNodeAnchorAttributes } from '@ory/kratos-client'

interface Props {
  attributes: UiNodeAnchorAttributes
}

export const NodeAnchor = ({ attributes }: Props) => {
  return (
    <button
      type="button"
      onClick={e => {
        e.stopPropagation()
        e.preventDefault()
        window.location.href = attributes.href
      }}
    >
      {attributes.title.text}
    </button>
  )
}
