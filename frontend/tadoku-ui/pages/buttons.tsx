import { CodeBlock, Preview, Separator, Title } from '@components/example'

export default function Buttons() {
  return (
    <>
      <h1 className="title mb-8">Buttons</h1>

      <Title>Button</Title>
      <Preview>
        <button className="btn">Call to Action</button>
      </Preview>
      <CodeBlock
        language="html"
        code={`<button className="btn">Call to Action</button>`}
      />

      <Separator />
    </>
  )
}
