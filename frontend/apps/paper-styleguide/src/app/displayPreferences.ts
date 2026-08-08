import { createContext } from 'react'

export type PaperTheme = 'light' | 'dark'
export type PaperDensity = 'comfortable' | 'compact'

export interface DisplayPreferencesValue {
  theme: PaperTheme
  density: PaperDensity
  setTheme: (theme: PaperTheme) => void
  setDensity: (density: PaperDensity) => void
}

export const DisplayPreferencesContext =
  createContext<DisplayPreferencesValue | null>(null)
