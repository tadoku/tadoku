import { useHydrated } from '@tanstack/react-router'
import { Button, Surface, chooseTheme, isThemeChoice, usePaperForm, type ThemeChoice } from 'paper-ui'
import { useEffect, useState } from 'react'
import { useIdentity } from '../lib/session'
import { ReturnTo } from './AccountForm'

const themes: readonly { readonly value: ThemeChoice; readonly label: string }[] = [
  { value: 'system', label: 'System' },
  { value: 'light', label: 'Light' },
  { value: 'dark', label: 'Dark' },
]

const formId = 'appearance'

/** The choice checked in the server-rendered picker, which a tap before hydration may have changed. */
function checkedTheme() {
  if (typeof document === 'undefined') return
  const form = document.getElementById(formId)
  const value = form instanceof HTMLFormElement ? new FormData(form).get('theme') : undefined
  return isThemeChoice(value) ? value : undefined
}

/**
 * The theme picker, which shows the choice stored in this browser's shared cookie. Before hydration or without
 * JavaScript, **Save theme** posts it to `/actions/theme`; once hydrated, a choice applies as it is picked.
 */
export function Appearance({ next }: { readonly next?: string }) {
  const { webUrl, theme } = useIdentity()
  const hydrated = useHydrated()
  // Hydration removes Save theme, so a choice tapped before it applies now.
  const [tapped] = useState(() => {
    const checked = checkedTheme()
    return checked !== theme ? checked : undefined
  })
  const form = usePaperForm({ defaultValues: { theme: tapped ?? theme } })
  useEffect(() => {
    if (tapped) chooseTheme(tapped, webUrl)
  }, [tapped, webUrl])

  return (
    <Surface className="p-5">
      <form id={formId} method="post" action="/actions/theme" className="flex flex-col items-start gap-4">
        <ReturnTo value={next} />
        <form.AppForm>
          <form.AppField
            name="theme"
            listeners={{
              onChange: ({ value }) => chooseTheme(value, webUrl),
            }}
          >
            {field => (
              <field.RadioSelect
                label="Theme"
                variant="segmented"
                options={themes}
                hint="Saved in this browser for every Tadoku page. System follows your device’s setting."
              />
            )}
          </form.AppField>
        </form.AppForm>
        {hydrated ? null : <Button type="submit">Save theme</Button>}
      </form>
    </Surface>
  )
}
