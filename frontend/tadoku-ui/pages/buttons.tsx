import { CodeBlock, Preview, Separator, Title } from '@components/example'
import {
  PencilSquareIcon,
  TrashIcon,
  ArrowDownTrayIcon,
  ChevronLeftIcon,
  UserIcon,
} from '@heroicons/react/24/solid'

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

      <Separator />

      <Title>Button links</Title>
      <Preview>
        <div className="space-x-3">
          <a href="#" className="btn primary">
            Primary
          </a>
          <a href="#" className="btn secondary">
            Secondary
          </a>
          <a href="#" className="btn">
            Tertiary (default)
          </a>
          <a href="#" className="btn danger">
            Danger
          </a>
          <a href="#" className="btn ghost">
            Ghost
          </a>
        </div>
      </Preview>
      <CodeBlock
        language="html"
        code={`<div className="space-x-3">
  <a href="#" className="btn primary">Primary</a>
  <a href="#" className="btn secondary">Secondary</a>
  <a href="#" className="btn">Tertiary (default)</a>
  <a href="#" className="btn danger">Danger</a>
  <a href="#" className="btn ghost">Ghost</a>
</div>`}
      />

      <Separator />

      <Title>Icon buttons</Title>
      <Preview>
        <div className="h-stack">
          <button className="btn primary">
            <PencilSquareIcon />
            Primary
          </button>
          <button className="btn secondary">
            <UserIcon />
            Secondary
          </button>
          <button className="btn">
            <ArrowDownTrayIcon />
            Tertiary
          </button>
          <button className="btn danger">
            <TrashIcon />
            Danger
          </button>
          <button className="btn ghost">
            <ChevronLeftIcon />
            Ghost
          </button>
        </div>
      </Preview>
      <CodeBlock
        language="html"
        code={`<div className="h-stack">
  <button className="btn primary">
    <PencilSquareIcon />
    Primary
  </button>
  <button className="btn secondary">
    <TrashIcon />
    Secondary
  </button>
  <button className="btn">
    <ArrowDownTrayIcon />
    Tertiary
  </button>
  <button className="btn danger">
    <TrashIcon />
    Danger
  </button>
  <button className="btn ghost">
    <ChevronLeftIcon />
    Ghost
  </button>
</div>`}
      />

      <Separator />

      <Title>Icon button links</Title>
      <Preview>
        <div className="space-x-3">
          <a href="#" className="btn primary">
            <PencilSquareIcon />
            Primary
          </a>
          <a href="#" className="btn secondary">
            <TrashIcon />
            Secondary
          </a>
          <a href="#" className="btn">
            <ArrowDownTrayIcon />
            Tertiary
          </a>
          <a href="#" className="btn danger">
            <TrashIcon />
            Danger
          </a>
          <a href="#" className="btn ghost">
            <ChevronLeftIcon />
            Ghost
          </a>
        </div>
      </Preview>
      <CodeBlock
        language="html"
        code={`<div className="space-x-3">
  <a href="#" className="btn primary">
    <PencilSquareIcon />
    Primary
  </a>
  <a href="#" className="btn secondary">
    <TrashIcon />
    Secondary
  </a>
  <a href="#" className="btn">
    <ArrowDownTrayIcon />
    Tertiary
  </a>
  <a href="#" className="btn danger">
    <TrashIcon />
    Danger
  </a>
  <a href="#" className="btn ghost">
    <ChevronLeftIcon />
    Ghost
  </a>
</div>`}
      />
    </>
  )
}
