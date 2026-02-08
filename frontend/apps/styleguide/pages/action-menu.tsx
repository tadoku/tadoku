import { Showcase } from '@components/Showcase'

import ActionMenuExample from '@examples/action-menu/example'
import exampleCode from '@examples/action-menu/example.tsx?raw'

export default function Page() {
  return (
    <>
      <h1 className="title mb-8">Action Menu</h1>

      <Showcase title="Example" code={exampleCode}>
        <ActionMenuExample />
      </Showcase>
    </>
  )
}
