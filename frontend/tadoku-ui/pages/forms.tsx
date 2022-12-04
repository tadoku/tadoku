import { CodeBlock, Preview, Separator, Title } from '@components/example'

export default function Forms() {
  return (
    <>
      <h1 className="title mb-8">Forms</h1>

      <Title>Input</Title>
      <Preview>
        <label className="label">
          <span className="label-text">First name</span>
          <input type="text" placeholder="John Doe" />
        </label>
      </Preview>
      <CodeBlock
        language="html"
        code={`<label className="label">
  <span className="label-text">First name</span>
  <input type="text" placeholder="John Doe" />
</label>`}
      />

      <Separator />

      <Title>Textarea</Title>
      <Preview>
        <label className="label">
          <span className="label-text">Message</span>
          <textarea placeholder="Dolor sit amet..." />
        </label>
      </Preview>
      <CodeBlock
        language="html"
        code={`<label className="label">
  <span className="label-text">Message</span>
  <textarea placeholder="Dolor sit amet..." />
</label>`}
      />

      <Separator />

      <Title>Select</Title>
      <Preview>
        <label className="label">
          <span className="label-text">Choose a color</span>
          <select>
            <option value="#ff0000" selected>
              Red
            </option>
            <option value="#00ff00">Green</option>
            <option value="#0000ff">Blue</option>
          </select>
        </label>
      </Preview>
      <CodeBlock
        language="html"
        code={`<label className="label">
  <span className="label-text">Choose a color</span>
  <select>
    <option value="#ff0000" selected>
      Red
    </option>
    <option value="#00ff00">Green</option>
    <option value="#0000ff">Blue</option>
  </select>
</label>`}
      />

      <Separator />

      <Title>Radio button</Title>
      <Preview></Preview>
      <CodeBlock code={``} />

      <Separator />

      <Title>Error message</Title>
      <Preview></Preview>
      <CodeBlock code={``} />
    </>
  )
}
