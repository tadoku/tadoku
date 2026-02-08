import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'
import { CodeBlock } from '@components/example'

import ToastExample from '@examples/toasts/example'
import exampleCode from '@examples/toasts/example.tsx?raw'

export default function Toasts() {
  return (
    <>
      <h1 className="title mb-8">Toasts</h1>

      <p>
        Implemented with{' '}
        <a
          href="https://fkhadra.github.io/react-toastify/introduction"
          target="_blank"
          rel="noreferrer"
        >
          react-toast
        </a>
        , please refer to their documentation on how to use it.
      </p>

      <Showcase title="Examples" code={exampleCode}>
        <ToastExample />
      </Showcase>

      <Separator />

      <h2 className="font-semibold text-xl my-2">Setup</h2>
      <p>
        In order to use toasts you need to set them up in your{' '}
        <code>_app.tsx</code>.
      </p>
      <CodeBlock
        language="typescript"
        code={`import ToastContainer from '@components/toasts'
import type { AppProps } from 'next/app'
// Note that it's important that your global stylesheet is loaded last, otherwise the theme will be overridden
import '../styles/globals.css'

export default function App({ Component, pageProps }: AppProps) {
  return (
    <>
      <Component {...pageProps} />
      <ToastContainer />
    </>
  )
}`}
      />
    </>
  )
}
