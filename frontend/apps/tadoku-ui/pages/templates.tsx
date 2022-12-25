import { CodeBlock, Preview, Separator, Title } from '@components/example'

export default function Templates() {
  return (
    <>
      <h1 className="title mb-8">Templates</h1>

      <Title>Vertical stack</Title>
      <Preview>
        <div className="v-stack spaced">
          <div className="bg-red-200">one</div>
          <div className="bg-green-200">two</div>
          <div className="bg-blue-200">three</div>
        </div>
      </Preview>
      <CodeBlock
        language={'html'}
        code={`<div className="v-stack spaced">
  <div className="bg-red-200">one</div>
  <div className="bg-green-200">two</div>
  <div className="bg-blue-200">three</div>
</div>`}
      />

      <Separator />

      <Title>Horizonal stack</Title>
      <Preview>
        <div className="h-stack spaced">
          <div className="bg-red-200 h-24">one</div>
          <div className="bg-green-200">two</div>
          <div className="bg-blue-200">three</div>
        </div>
      </Preview>
      <CodeBlock
        language={'html'}
        code={`<div className="h-stack spaced">
  <div className="bg-red-200 h-24">one</div>
  <div className="bg-green-200">two</div>
  <div className="bg-blue-200">three</div>
</div>`}
      />

      <Separator />

      <Title>Card</Title>
      <Preview className="bg-gray-100">
        <div className="card">
          <p>Some content</p>
        </div>
      </Preview>
      <CodeBlock
        language={'html'}
        code={`<div className="card">
  <p>Some content</p>
</div>`}
      />
    </>
  )
}
