import { useRef, useEffect } from 'react'
import { EditorView, basicSetup } from 'codemirror'
import { EditorState, Extension } from '@codemirror/state'
import { ViewUpdate } from '@codemirror/view'
import { placeholder as cmPlaceholder } from '@codemirror/view'

interface Props {
  value: string
  onChange: (value: string) => void
  extensions: Extension[]
  placeholder?: string
}

export function CodeEditor({ value, onChange, extensions, placeholder }: Props) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const onChangeRef = useRef(onChange)
  onChangeRef.current = onChange

  // Create the editor once — value is intentionally excluded; the sync effect below handles updates
  useEffect(() => {
    if (!containerRef.current) return

    const view = new EditorView({
      state: EditorState.create({
        doc: value,
        extensions: [
          basicSetup,
          ...extensions,
          ...(placeholder ? [cmPlaceholder(placeholder)] : []),
          EditorView.updateListener.of((update: ViewUpdate) => {
            if (update.docChanged) {
              onChangeRef.current(update.state.doc.toString())
            }
          }),
          EditorView.theme({
            '&': {
              border: '1px solid #d1d5db',
              minHeight: '400px',
              fontSize: '0.875rem',
            },
            '&.cm-focused': {
              outline: '2px solid #6366f1',
              outlineOffset: '-1px',
            },
            '.cm-scroller': {
              fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
            },
            '.cm-content': {
              minHeight: '380px',
            },
          }),
        ],
      }),
      parent: containerRef.current,
    })
    viewRef.current = view

    return () => {
      view.destroy()
      viewRef.current = null
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [extensions, placeholder])

  // Sync external value changes into the editor
  useEffect(() => {
    const view = viewRef.current
    if (!view) return

    const current = view.state.doc.toString()
    if (current !== value) {
      view.dispatch({
        changes: { from: 0, to: current.length, insert: value },
      })
    }
  }, [value])

  return <div ref={containerRef} />
}
