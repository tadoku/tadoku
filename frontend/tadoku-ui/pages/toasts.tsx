import { CodeBlock, Preview, Separator, Title } from '@components/example'
import { toast } from 'react-toastify'

export default function Toasts() {
  return (
    <>
      <h1 className="title mb-8">Toasts</h1>
      <Title>Examples</Title>
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
      <Preview>
        <div className="h-stack">
          <button className="btn" onClick={() => toast.info('Info toast')}>
            Info toast
          </button>
          <button
            className="btn"
            onClick={() => toast.success('Success toast')}
          >
            Success toast
          </button>
          <button
            className="btn"
            onClick={() => toast.warning('Warning toast')}
          >
            Warning toast
          </button>
          <button className="btn" onClick={() => toast.error('Error toast')}>
            Error toast
          </button>
        </div>
      </Preview>
      <CodeBlock
        language="typescript"
        code={`import { toast } from 'react-toastify'

const ToastExample = () => (
  <div className="h-stack">
    <button className="btn" onClick={() => toast.info('Info toast')}>
      Info toast
    </button>
    <button
      className="btn"
      onClick={() => toast.success('Success toast')}
    >
      Success toast
    </button>
    <button
      className="btn"
      onClick={() => toast.warning('Warning toast')}
    >
      Warning toast
    </button>
    <button className="btn" onClick={() => toast.error('Error toast')}>
      Error toast
    </button>
  </div>
)`}
      />

      <Separator />

      <Title>Setup</Title>
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
