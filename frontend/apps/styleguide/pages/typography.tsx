import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'

import TitleExample from '@examples/typography/title'
import titleCode from '@examples/typography/title.tsx?raw'

import SubtitleExample from '@examples/typography/subtitle'
import subtitleCode from '@examples/typography/subtitle.tsx?raw'

import TextLinkExample from '@examples/typography/text-link'
import textLinkCode from '@examples/typography/text-link.tsx?raw'

export default function Typography() {
  return (
    <>
      <h1 className="title mb-8">Typography</h1>

      <Showcase title="Title" code={titleCode}>
        <TitleExample />
      </Showcase>

      <Separator />

      <Showcase title="Subtitle" code={subtitleCode}>
        <SubtitleExample />
      </Showcase>

      <Separator />

      <Showcase title="Text link" code={textLinkCode}>
        <TextLinkExample />
      </Showcase>
    </>
  )
}
