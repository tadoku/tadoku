import { CodeBlock, Preview, Separator, Title } from '@components/example'

export default function Buttons() {
  return (
    <>
      <h1 className="title mb-8">Buttons</h1>

      <Title>Buttons</Title>
      <Preview>
        <div className="h-stack">
          <button className="btn primary">Primary</button>
          <button className="btn secondary">Secondary</button>
          <button className="btn">Tertiary (default)</button>
          <button className="btn danger">Danger</button>
          <button className="btn ghost">Ghost</button>
        </div>
      </Preview>
      <CodeBlock
        language="html"
        code={`<div className="h-stack">
  <button className="btn primary">Primary</button>
  <button className="btn secondary">Secondary</button>
  <button className="btn">Tertiary (default)</button>
  <button className="btn danger">Danger</button>
  <button className="btn ghost">Ghost</button>
</div>`}
      />

      <Separator />

      <Title>Disabled button</Title>
      <Preview>
        <div className="h-stack">
          <button className="btn primary" disabled>
            Primary
          </button>
          <button className="btn secondary" disabled>
            Secondary
          </button>
          <button className="btn" disabled>
            Tertiary (default)
          </button>
          <button className="btn danger" disabled>
            Danger
          </button>
          <button className="btn ghost" disabled>
            Ghost
          </button>
        </div>
      </Preview>
      <CodeBlock
        language="html"
        code={`<div className="h-stack">
  <button className="btn primary" disabled>Primary</button>
  <button className="btn secondary" disabled>Secondary</button>
  <button className="btn" disabled>Tertiary (default)</button>
  <button className="btn danger" disabled>Danger</button>
  <button className="btn ghost" disabled>Ghost</button>
</div>`}
      />
    </>
  )
}
