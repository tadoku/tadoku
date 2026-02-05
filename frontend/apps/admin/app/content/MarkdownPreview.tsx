import { Component, ReactNode } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

interface ErrorBoundaryProps {
  children: ReactNode
  fallback?: ReactNode
}

interface ErrorBoundaryState {
  hasError: boolean
}

class MarkdownErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError() {
    return { hasError: true }
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback ?? (
        <p className="text-sm text-slate-500 italic">Unable to render preview.</p>
      )
    }
    return this.props.children
  }
}

interface Props {
  content: string
  className?: string
}

export function MarkdownPreview({ content, className }: Props) {
  return (
    <MarkdownErrorBoundary>
      <div className={`auto-format ${className ?? ''}`}>
        <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
      </div>
    </MarkdownErrorBoundary>
  )
}
