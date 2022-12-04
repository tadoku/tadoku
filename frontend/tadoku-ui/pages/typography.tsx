import { CodeBlock, Preview, Separator, Title } from '@components/example'

export default function Typography() {
  return (
    <>
      <h1 className="title">Typography</h1>
      <Title>Title</Title>
      <Preview>
        <h1 className="title">The quick brown fox jumps over the lazy dog</h1>
      </Preview>
      <CodeBlock
        code={`<h1 className="title">The quick brown fox jumps over the lazy dog</h1>`}
      />

      <Separator />

      <Title>Subtitle</Title>
      <Preview>
        <h1 className="subtitle">
          The quick brown fox jumps over the lazy dog
        </h1>
      </Preview>
      <CodeBlock
        code={`<h1 className="subtitle">The quick brown fox jumps over the lazy dog</h1>`}
      />
    </>
  )
}
