import { useContext, type ComponentType, type ReactNode } from 'react'
import { GuideCode } from '../FoundationGuide'
import { PreviewSettings } from './preview-settings'

export function FoundationCanvas({ children }: { children: ReactNode }) {
  const { theme, density } = useContext(PreviewSettings)
  return <div data-theme={theme} data-density={density} className="foundation-guide__preview paper-type-body">{children}</div>
}

export function FoundationExample({ Example, code, label = 'Working example' }: { Example: ComponentType<{ theme?: 'light' | 'dark' }>; code: string; label?: string }) {
  const { theme } = useContext(PreviewSettings)
  return <><h3>{label}</h3><FoundationCanvas><Example theme={theme as 'light' | 'dark'} /></FoundationCanvas><GuideCode code={code} label="Source for this example" /></>
}
