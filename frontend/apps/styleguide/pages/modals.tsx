import { Showcase } from '@components/Showcase'

import StandardModalExample from '@examples/modals/standard'
import standardCode from '@examples/modals/standard.tsx?raw'

export default function Modals() {
  return (
    <>
      <h1 className="title mb-8">Modal</h1>
      <p>
        In general modals should be used for <strong>short interactions</strong>{' '}
        such as:
      </p>
      <ul className="list">
        <li>Confirmation dialogs</li>
        <li>Single input forms</li>
      </ul>
      <p>
        <strong>Avoid</strong> in cases such as:
      </p>
      <ul className="list !mb-6">
        <li>Complex forms</li>
      </ul>

      <Showcase title="Standard" code={standardCode}>
        <StandardModalExample />
      </Showcase>
    </>
  )
}
