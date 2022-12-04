import { ReactNode } from 'react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { coldarkDark as theme } from 'react-syntax-highlighter/dist/cjs/styles/prism'

export const CodeBlock = ({ code }: { code: string }) => (
  <SyntaxHighlighter language="typescript" style={theme}>
    {code}
  </SyntaxHighlighter>
)

export const Preview = ({
  children,
  dark = false,
}: {
  children: ReactNode
  dark?: boolean
}) => (
  <div
    className={`not-prose overflow-hidden p-8 relative ${
      dark ? 'bg-black' : 'bg-white'
    }`}
  >
    {children}
  </div>
)

export const Separator = () => <hr className="my-8" />

export const Title = ({ children }: { children: ReactNode }) => (
  <h2 className="my-2 font-semibold text-xl">{children}</h2>
)
