const key = 'tadoku-theme'

/** Render in <head> before the body so `[data-theme]` is set before first paint. The shared cookie wins, so every application agrees. */
export const themeBootScript = `try{var c=document.cookie.match(/(?:^|; )${key}=(dark|light)/),t=c?c[1]:localStorage.getItem('${key}');document.documentElement.dataset.theme=t==='dark'||t==='light'?t:matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light'}catch(e){}`

export type ThemeName = 'light' | 'dark'

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

/** Remembers the theme in a cookie that `otherOrigin`, a sibling application, can read. */
export function shareTheme(theme: ThemeName, otherOrigin: string) {
  const domain = sharedDomain(location.hostname, new URL(otherOrigin).hostname)
  const scope = domain ? `; domain=${domain}` : ''
  const secure = location.protocol === 'https:' ? '; secure' : ''
  document.cookie = `${key}=${theme}; path=/; max-age=31536000; samesite=lax${scope}${secure}`
}

/** Applies a theme chosen in this application and remembers it here and for `otherOrigin`. */
export function chooseTheme(theme: ThemeName, otherOrigin: string) {
  document.documentElement.dataset.theme = theme
  try { localStorage.setItem(key, theme) } catch { /* Private mode keeps the choice for this page only. */ }
  shareTheme(theme, otherOrigin)
}

/** The theme this browser stored for this application, if any. */
export function storedTheme(): ThemeName | null {
  try {
    const value = localStorage.getItem(key)
    return value === 'dark' || value === 'light' ? value : null
  } catch {
    return null
  }
}
