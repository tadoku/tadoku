import { CodeBlock, Separator } from '@components/example'

export default function Home() {
  return (
    <>
      <h1 className="title mb-4">Getting Started</h1>
      <p className="mb-8">
        The design system powering{' '}
        <a href="https://www.tadoku.app" className="text-link">
          Tadoku
        </a>
        .
      </p>

      <h2 className="subtitle mb-4">Installation</h2>
      <p className="mb-4">
        Add the ui package to your app&apos;s dependencies in{' '}
        <code className="bg-gray-100 px-1 rounded">package.json</code>:
      </p>
      <CodeBlock code={`"dependencies": {
  "ui": "workspace:*"
}`} language="json" />

      <Separator />

      <h2 className="subtitle mb-4">Usage</h2>
      <CodeBlock
        code={`// Import styles in _app.tsx
import 'ui/styles/globals.css'

// Import components
import { Input, Modal, Flash, Logo } from 'ui'

// Buttons use CSS classes, not components
<button className="btn primary">Primary</button>`}
        language="tsx"
      />
    </>
  )
}
