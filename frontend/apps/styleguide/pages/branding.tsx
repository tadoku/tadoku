import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'

import LogoExample from '@examples/branding/logo'
import logoCode from '@examples/branding/logo.tsx?raw'

import LogoInvertedExample from '@examples/branding/logo-inverted'
import logoInvertedCode from '@examples/branding/logo-inverted.tsx?raw'

export default function Branding() {
  return (
    <>
      <h1 className="title mb-8">Branding</h1>

      <Showcase title="Logo" code={logoCode}>
        <LogoExample />
      </Showcase>

      <Separator />

      <Showcase title="Logo for dark backgrounds" code={logoInvertedCode} dark>
        <LogoInvertedExample />
      </Showcase>
    </>
  )
}
