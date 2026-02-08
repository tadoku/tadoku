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
      <CodeBlock code={`pnpm add ui`} language="bash" />

      <Separator />

      <h2 className="subtitle mb-4">Usage</h2>
      <CodeBlock
        code={`// Import components
import { Button, Input, Modal, Flash } from 'ui'

// Import styles in _app.tsx
import 'ui/styles/globals.css'`}
        language="tsx"
      />
    </>
  )
}
