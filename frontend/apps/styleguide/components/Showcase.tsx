import { ReactNode, useState } from 'react'
import { CodeBlock, Preview } from './example'
import { CodeBracketIcon, CubeIcon } from '@heroicons/react/20/solid'

interface ShowcaseProps {
  title: string
  children: ReactNode
  code: string
  language?: string
  dark?: boolean
  previewClassName?: string
}

export function Showcase({
  title,
  children,
  code,
  language = 'tsx',
  dark,
  previewClassName,
}: ShowcaseProps) {
  const [showCode, setShowCode] = useState(false)

  return (
    <div>
      <div className="flex justify-between items-center my-2">
        <h2 className="font-semibold text-xl">{title}</h2>
        <button
          className="btn ghost flex items-center gap-1.5"
          onClick={() => setShowCode(!showCode)}
        >
          {showCode ? (
            <>
              <CubeIcon className="w-4 h-4" />
              Preview
            </>
          ) : (
            <>
              <CodeBracketIcon className="w-4 h-4" />
              Code
            </>
          )}
        </button>
      </div>
      {showCode ? (
        <div className="max-h-[32rem] overflow-auto">
          <CodeBlock code={code} language={language} />
        </div>
      ) : (
        <Preview dark={dark} className={previewClassName}>
          {children}
        </Preview>
      )}
    </div>
  )
}
