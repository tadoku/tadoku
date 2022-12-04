import { Logo, LogoInverted } from '@components/branding'
import { CodeBlock, Preview, Separator, Title } from '@components/example'

export default function branding() {
  return (
    <>
      <h1 className="title mb-8">Branding</h1>
      <Title>Logo</Title>
      <Preview>
        <Logo />
      </Preview>
      <CodeBlock code={`<Logo />`} />

      <Separator />

      <Title>Logo for dark backgrounds</Title>
      <Preview dark>
        <LogoInverted />
      </Preview>
      <CodeBlock code={`<LogoInverted />`} />
    </>
  )
}
