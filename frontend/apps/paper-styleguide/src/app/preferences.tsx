import {
  type PropsWithChildren,
  useEffect,
  useMemo,
  useState,
} from 'react'
import {
  DisplayPreferencesContext,
  type PaperDensity,
  type PaperTheme,
} from './displayPreferences'

export function DisplayPreferencesProvider({ children }: PropsWithChildren) {
  const [theme, setTheme] = useState<PaperTheme>('light')
  const [density, setDensity] = useState<PaperDensity>('comfortable')

  useEffect(() => {
    const root = document.documentElement
    const previousTheme = root.dataset.theme
    const previousDensity = root.dataset.density
    root.dataset.theme = theme
    root.dataset.density = density

    return () => {
      if (previousTheme) root.dataset.theme = previousTheme
      else delete root.dataset.theme
      if (previousDensity) root.dataset.density = previousDensity
      else delete root.dataset.density
    }
  }, [density, theme])

  const value = useMemo(
    () => ({ theme, density, setTheme, setDensity }),
    [density, theme],
  )

  return (
    <DisplayPreferencesContext.Provider value={value}>
      {children}
    </DisplayPreferencesContext.Provider>
  )
}
