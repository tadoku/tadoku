const key = 'tadoku-theme'
const darkQuery = '(prefers-color-scheme: dark)'

/**
 * Render in <head> before the body so `[data-theme]` is set before first paint, and kept in step with the system
 * while the choice is `system`. The shared cookie wins, so every application agrees.
 */
export const themeBootScript = `try{var m=matchMedia('${darkQuery}'),a=function(){var c=document.cookie.match(/(?:^|; )${key}=(dark|light|system)/),t=c?c[1]:localStorage.getItem('${key}');document.documentElement.dataset.theme=t==='dark'||t==='light'?t:m.matches?'dark':'light'};a();m.addEventListener('change',a)}catch(e){}`

export type ThemeName = 'light' | 'dark'

/** A theme preference: a fixed theme, or whichever the system prefers. */
export type ThemeChoice = ThemeName | 'system'

export const isThemeChoice = (value: unknown): value is ThemeChoice => value === 'light' || value === 'dark' || value === 'system'

/** The trailing host labels two hosts share, when that is a real domain rather than a single label such as `localhost`. */
function sharedDomain(host: string, otherHost: string) {
  const labels = host.split('.').reverse()
  const others = otherHost.split('.').reverse()
  const shared: string[] = []
  for (const [index, label] of labels.entries()) {
    if (label !== others[index]) break
    shared.unshift(label)
  }
  return shared.length > 1 ? shared.join('.') : undefined
}

/**
 * The cookie, for `Set-Cookie` or `document.cookie`, that stores `choice` for `host` and for `otherOrigin`, a sibling
 * application. `secure` when `host` is served over HTTPS.
 */
export function themeCookie(choice: ThemeChoice, host: string, otherOrigin: string, secure: boolean) {
  const domain = sharedDomain(host, new URL(otherOrigin).hostname)
  return `${key}=${choice}; path=/; max-age=31536000; samesite=lax${domain ? `; domain=${domain}` : ''}${secure ? '; secure' : ''}`
}

/** The choice a `Cookie` header or `document.cookie` holds, read as the boot script reads it. */
export function cookieTheme(cookies: string | undefined): ThemeChoice | undefined {
  const value = cookies?.match(new RegExp(`(?:^|;\\s*)${key}=([^;]*)`))?.[1]
  return isThemeChoice(value) ? value : undefined
}

function shareTheme(choice: ThemeChoice, otherOrigin: string) {
  document.cookie = themeCookie(choice, location.hostname, otherOrigin, location.protocol === 'https:')
}

/** Applies a choice made in this application and remembers it here and for `otherOrigin`. */
export function chooseTheme(choice: ThemeChoice, otherOrigin: string) {
  document.documentElement.dataset.theme = choice === 'system' ? (matchMedia(darkQuery).matches ? 'dark' : 'light') : choice
  try { localStorage.setItem(key, choice) } catch { /* Private mode keeps the choice for this page only. */ }
  shareTheme(choice, otherOrigin)
}

/** Shares a theme stored before the cookie existed. An existing cookie is newer, so it is left alone. */
export function adoptStoredTheme(otherOrigin: string) {
  if (cookieTheme(document.cookie)) return
  try {
    const value = localStorage.getItem(key)
    if (isThemeChoice(value)) shareTheme(value, otherOrigin)
  } catch { /* Private mode has nothing stored to share. */ }
}
