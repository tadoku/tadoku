import { CodeBlock, Preview, Separator, Title } from '@components/example'

export default function Buttons() {
  return (
    <>
      <h1 className="title mb-8">Buttons</h1>

      <Title>Primary button</Title>
      <Preview>
        <button className="btn primary">Submit</button>
      </Preview>
      <CodeBlock
        language="html"
        code={`<button className="btn primary">Submit</button>`}
      />

      <Separator />

      <Title>Secondary button</Title>
      <Preview>
        <button className="btn secondary">Call to Action</button>
      </Preview>
      <CodeBlock
        language="html"
        code={`<button className="btn secondary">Call to Action</button>`}
      />

      <Separator />

      <Title>Tertiary button (default)</Title>
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
