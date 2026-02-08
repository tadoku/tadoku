import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'

import VStackExample from '@examples/templates/v-stack'
import vStackCode from '@examples/templates/v-stack.tsx?raw'

import HStackExample from '@examples/templates/h-stack'
import hStackCode from '@examples/templates/h-stack.tsx?raw'

import CardExample from '@examples/templates/card'
import cardCode from '@examples/templates/card.tsx?raw'

export default function Templates() {
  return (
    <>
      <h1 className="title mb-8">Templates</h1>

      <Showcase title="Vertical stack" code={vStackCode}>
        <VStackExample />
      </Showcase>

      <Separator />

      <Showcase title="Horizontal stack" code={hStackCode}>
        <HStackExample />
      </Showcase>

      <Separator />

      <Showcase title="Card" code={cardCode} previewClassName="bg-gray-100">
        <CardExample />
      </Showcase>
    </>
  )
}
