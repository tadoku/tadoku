import { CodeBlock, Preview, Separator, Title } from '@components/example'

export default function Typography() {
  return (
    <>
      <h1 className="title mb-8">Typography</h1>
      <Title>Title</Title>
      <Preview>
        <h1 className="title">The quick brown fox jumps over the lazy dog</h1>
      </Preview>
      <CodeBlock
        language={'html'}
        code={`<h1 className="title">
  The quick brown fox jumps over the lazy dog
</h1>`}
      />

      <Separator />

      <Title>Subtitle</Title>
      <Preview>
        <h1 className="subtitle">
          The quick brown fox jumps over the lazy dog
        </h1>
      </Preview>
      <CodeBlock
        language={'html'}
        code={`<h1 className="subtitle">
  The quick brown fox jumps over the lazy dog
</h1>`}
      />

      <Separator />

      <Title>Text link</Title>
      <Preview>
        <a className="text-link" href="#">
          A text link
        </a>
      </Preview>
      <CodeBlock
        language={'html'}
        code={`<a className="text-link" href="#">
  A text link
</a>`}
      />
    </>
  )
}
