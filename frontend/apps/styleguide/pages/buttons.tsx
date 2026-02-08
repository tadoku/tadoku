import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'

import ButtonVariants from '@examples/buttons/variants'
import variantsCode from '@examples/buttons/variants.tsx?raw'

import DisabledButtons from '@examples/buttons/disabled'
import disabledCode from '@examples/buttons/disabled.tsx?raw'

import ButtonLinks from '@examples/buttons/links'
import linksCode from '@examples/buttons/links.tsx?raw'

import IconButtons from '@examples/buttons/icon-buttons'
import iconButtonsCode from '@examples/buttons/icon-buttons.tsx?raw'

import IconButtonLinks from '@examples/buttons/icon-links'
import iconLinksCode from '@examples/buttons/icon-links.tsx?raw'

import ButtonGroupExample from '@examples/buttons/button-group'
import buttonGroupCode from '@examples/buttons/button-group.tsx?raw'

export default function Buttons() {
  return (
    <>
      <h1 className="title mb-8">Buttons</h1>

      <Showcase title="Buttons" code={variantsCode}>
        <ButtonVariants />
      </Showcase>

      <Separator />

      <Showcase title="Disabled button" code={disabledCode}>
        <DisabledButtons />
      </Showcase>

      <Separator />

      <Showcase title="Button links" code={linksCode}>
        <ButtonLinks />
      </Showcase>

      <Separator />

      <Showcase title="Icon buttons" code={iconButtonsCode}>
        <IconButtons />
      </Showcase>

      <Separator />

      <Showcase title="Icon button links" code={iconLinksCode}>
        <IconButtonLinks />
      </Showcase>

      <Separator />

      <Showcase title="Button group" code={buttonGroupCode}>
        <ButtonGroupExample />
      </Showcase>
    </>
  )
}
