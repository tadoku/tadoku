import { Logo, LogoInverted } from 'ui'
import { CodeBlock, Preview, Separator, Title } from '@components/example'

export default function Branding() {
  return (
    <>
      <h1 className="title mb-8">Branding</h1>
      <Title>Logo</Title>
      <Preview>
        <Logo />
      </Preview>
      <CodeBlock
        code={`import { Logo } from 'ui'

const example = () => <Logo />`}
      />

      <Separator />

      <Title>Logo for dark backgrounds</Title>
      <Preview dark>
        <LogoInverted />
      </Preview>
      <CodeBlock
        code={`import { LogoInverted } from 'ui'

const example = () => <LogoInverted />`}
      />
    </>
  )
}
