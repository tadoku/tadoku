import Prism from 'prismjs'
import 'prismjs/components/prism-jsx'
import 'prismjs/components/prism-typescript'
import 'prismjs/components/prism-tsx'
import 'prismjs/components/prism-bash'
import type { ReactNode } from 'react'

function tokens(
  content: string | Prism.Token | (string | Prism.Token)[],
): ReactNode {
  if (typeof content === 'string') return content
  if (Array.isArray(content))
    return content.map((token, index) => (
      <span key={index}>{tokens(token)}</span>
    ))
  return (
    <span className={`token ${content.type}`}>{tokens(content.content)}</span>
  )
}

export function CodeBlock({
  code,
  language = 'tsx',
  label,
}: {
  code: string
  language?: string
  label?: string
}) {
  return (
    <pre className="paper-code-block" tabIndex={0} aria-label={label}>
      <code>
        {tokens(
          Prism.tokenize(
            code,
            Prism.languages[language] ?? Prism.languages.tsx,
          ),
        )}
      </code>
    </pre>
  )
}
