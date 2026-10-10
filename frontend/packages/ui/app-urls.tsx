import getConfig from 'next/config'
import { createContext, ReactNode, useContext } from 'react'

export interface AppUrls {
  branch?: string
  homeUrl: string
  authUiUrl: string
  adminUrl: string
  apiEndpoint: string
}

const labelPattern = /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/

function branchMapper(config: Record<string, unknown>, host?: string) {
  const bases = ['homeUrl', 'authUiUrl', 'adminUrl'] as const
  const urls = Object.fromEntries(bases.map(key => {
    if (typeof config[key] !== 'string') throw new Error(`Missing ${key}`)
    const url = new URL(config[key] as string)
    if (!['https:', 'http:'].includes(url.protocol)) throw new Error(`Invalid ${key}`)
    return [key, url]
  })) as Record<typeof bases[number], URL>
  const hostname = host?.toLowerCase().replace(/:\d+$/, '')
  const baseHosts = bases.map(key => urls[key].hostname)
  const branchHosts = baseHosts.map(base =>
    ['tadoku.app', 'account.tadoku.app', 'admin.tadoku.app'].includes(base)
      ? base.replace('tadoku.app', 'preview.tadoku.app')
      : base,
  )
  let branch: string | undefined
  if (hostname && !baseHosts.includes(hostname)) {
    for (const base of branchHosts) {
      if (!hostname.endsWith(`.${base}`)) continue
      const label = hostname.slice(0, -(base.length + 1))
      if (labelPattern.test(label)) {
        branch = label
        break
      }
    }
  }
  const prefix = (value: string) => {
    if (!branch || value.startsWith('/')) return value
    const url = new URL(value)
    const index = baseHosts.indexOf(url.hostname)
    if (index === -1) return value
    url.hostname = `${branch}.${branchHosts[index]}`
    return url.toString().replace(/\/$/, value.endsWith('/') ? '/' : '')
  }
  return { branch, prefix }
}

export function resolveAppUrls(config: Record<string, unknown>, host?: string): AppUrls {
  const { branch, prefix } = branchMapper(config, host)
  if (typeof config.apiEndpoint !== 'string') throw new Error('Missing apiEndpoint')
  return {
    ...(branch ? { branch } : {}),
    homeUrl: prefix(config.homeUrl as string),
    authUiUrl: prefix(config.authUiUrl as string),
    adminUrl: prefix(config.adminUrl as string),
    apiEndpoint: prefix(config.apiEndpoint),
  }
}

// Maps an absolute production app URL, such as a Kratos redirect_browser_to, to the sibling branch host of `host`.
// URLs under config.kratosPublicEndpoint stay on the shared Kratos host.
export function branchUrl(config: Record<string, unknown>, host: string | undefined, url: string): string {
  const kratos = config.kratosPublicEndpoint
  if (typeof kratos === 'string' && (url === kratos || url.startsWith(`${kratos.replace(/\/$/, '')}/`))) return url
  return branchMapper(config, host).prefix(url)
}

// Returns the same page on the flow's branch account host when a Kratos flow fetched on `currentUrl` was started
// from a different branch, as identified by its return_to or the return_to in its request_url. Only branch hosts of
// the configured apps qualify, and a page already on the flow's branch returns undefined.
export function branchFlowUrl(
  config: Record<string, unknown>,
  currentUrl: string,
  flow: { return_to?: string; request_url?: string },
): string | undefined {
  const parse = (value?: string | null) => {
    try {
      return value ? new URL(value) : undefined
    } catch {
      return undefined
    }
  }

  const returnTo = parse(flow.return_to) ?? parse(parse(flow.request_url)?.searchParams.get('return_to'))
  const target = returnTo && resolveAppUrls(config, returnTo.host)
  const current = new URL(currentUrl)
  if (!target?.branch || target.branch === resolveAppUrls(config, current.host).branch) return undefined

  const account = new URL(target.authUiUrl)
  current.protocol = account.protocol
  current.host = account.host
  return current.toString()
}

export function appUrlsForHost(host?: string): AppUrls {
  return resolveAppUrls(getConfig().publicRuntimeConfig, host)
}

export function browserAppUrls(): AppUrls {
  return appUrlsForHost(typeof window === 'undefined' ? undefined : window.location.host)
}

export function serverApiEndpointForHost(host?: string): string {
  const { publicRuntimeConfig, serverRuntimeConfig } = getConfig()
  return resolveAppUrls({
    ...publicRuntimeConfig,
    apiEndpoint: serverRuntimeConfig?.apiEndpoint ?? publicRuntimeConfig.apiEndpoint,
  }, host).apiEndpoint
}

const AppUrlsContext = createContext<AppUrls | undefined>(undefined)

export function AppUrlsProvider({ urls, children }: { urls: AppUrls; children: ReactNode }) {
  return <AppUrlsContext.Provider value={urls}>{children}</AppUrlsContext.Provider>
}

export function useAppUrls(): AppUrls {
  return useContext(AppUrlsContext) ?? browserAppUrls()
}
