import { ReactNode } from 'react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { coldarkDark as theme } from 'react-syntax-highlighter/dist/cjs/styles/prism'

export const CodeBlock = ({ code }: { code: string }) => (
  <SyntaxHighlighter language="typescript" style={theme}>
    {code}
  </SyntaxHighlighter>
)

export const Preview = ({ children }: { children: ReactNode }) => (
  <div className="not-prose bg-white overflow-hidden p-8 relative">
    {children}
  </div>
)

export const Separator = () => <hr className="my-8" />

export const Title = ({ children }: { children: ReactNode }) => (
  <h2 className="my-2 font-semibold text-xl">{children}</h2>
)
