import * as stylex from '@stylexjs/stylex'
import Prism from 'prismjs'
import 'prismjs/components/prism-jsx'
import 'prismjs/components/prism-typescript'
import 'prismjs/components/prism-tsx'
import 'prismjs/components/prism-bash'
import type { ReactNode } from 'react'
import { shellStyles } from '../styles/shell.styles'
import { color } from 'paper-ui/tokens.stylex'

const tokenStyles = stylex.create({
  comment: { color: color.textMuted },
  keyword: { color: color.actionDefault },
  tag: { color: color.actionDefault },
  string: { color: color.statusSuccess },
  'attr-value': { color: color.statusSuccess },
  function: { color: color.textLink },
  'class-name': { color: color.textLink },
  number: { color: color.textLink },
  'attr-name': { color: color.statusDanger },
  operator: { color: color.statusDanger },
})

function tokens(
  content: string | Prism.Token | (string | Prism.Token)[],
): ReactNode {
  if (typeof content === 'string') return content
  if (Array.isArray(content))
    return content.map((token, index) => (
      <span key={index}>{tokens(token)}</span>
    ))
  return (
    <span {...stylex.props(tokenStyles[content.type as keyof typeof tokenStyles])} data-token={content.type}>{tokens(content.content)}</span>
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
    <pre {...stylex.props(shellStyles.paperCodeBlock)} tabIndex={0} aria-label={label}>
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
