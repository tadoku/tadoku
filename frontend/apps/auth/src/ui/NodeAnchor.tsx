import { UiNodeAnchorAttributes } from '@ory/kratos-client'
import { followKratosUrl } from '../branch'

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
        followKratosUrl(attributes.href)
      }}
    >
      {attributes.title.text}
    </button>
  )
}
