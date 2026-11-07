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
    <pre className="paper-code-block max-w-full overflow-auto whitespace-pre p-4 font-mono text-fine leading-165 [&&_code]:text-ink [&&_code]:wrap-normal [&_.token.comment]:text-muted [&_.token.keyword]:text-action [&_.token.tag]:text-action [&_.token.string]:text-success [&_.token.attr-value]:text-success [&_.token.function]:text-link [&_.token.class-name]:text-link [&_.token.number]:text-link [&_.token.attr-name]:text-danger [&_.token.operator]:text-danger" tabIndex={0} aria-label={label}>
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
