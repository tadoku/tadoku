import { useContext } from 'react'
import { DisplayPreferencesContext } from './displayPreferences'

export function useDisplayPreferences() {
  const value = useContext(DisplayPreferencesContext)
  if (!value) {
    throw new Error(
      'useDisplayPreferences must be used inside DisplayPreferencesProvider',
    )
  }
  return value
}
