import { ReactNode } from 'react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { coldarkDark as theme } from 'react-syntax-highlighter/dist/cjs/styles/prism'

theme['pre[class*="language-"]'].margin = 0
theme['pre[class*="language-"]'].lineHeight = '1.2'
theme['code[class*="language-"]'].fontSize = '0.8em'
theme['code[class*="language-"]'].lineHeight = '1.2'

export const CodeBlock = ({
  code,
  language,
}: {
  code: string
  language?: string
}) => (
  <SyntaxHighlighter
    language={language ?? 'typescript'}
    style={theme}
    showLineNumbers
  >
    {code}
  </SyntaxHighlighter>
)

export const Preview = ({
  children,
  dark = false,
  className = '',
}: {
  children: ReactNode
  dark?: boolean
  className?: string
}) => (
  <div
    className={`not-prose overflow-hidden p-8 relative ${
      dark ? 'bg-black' : 'bg-white'
    } ${className}`}
  >
    {children}
  </div>
)

export const Separator = () => <hr className="my-8" />

export const Title = ({ children }: { children: ReactNode }) => (
  <h2 className="my-2 font-semibold text-xl">{children}</h2>
)
