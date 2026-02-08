import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'

import FlashStandard from '@examples/flash/standard'
import standardCode from '@examples/flash/standard.tsx?raw'

import FlashLinks from '@examples/flash/links'
import linksCode from '@examples/flash/links.tsx?raw'

import FlashIcons from '@examples/flash/icons'
import iconsCode from '@examples/flash/icons.tsx?raw'

export default function FlashPreview() {
  return (
    <>
      <h1 className="title mb-8">Flash messages</h1>

      <Showcase title="Standard" code={standardCode}>
        <FlashStandard />
      </Showcase>

      <Separator />

      <Showcase title="Link variation" code={linksCode}>
        <FlashLinks />
      </Showcase>

      <Separator />

      <Showcase title="Icon variation" code={iconsCode}>
        <FlashIcons />
      </Showcase>
    </>
  )
}
