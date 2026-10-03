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

export function resolveAppUrls(config: Record<string, unknown>, host?: string): AppUrls {
  const bases = ['homeUrl', 'authUiUrl', 'adminUrl'] as const
  const urls = Object.fromEntries(bases.map(key => {
    if (typeof config[key] !== 'string') throw new Error(`Missing ${key}`)
    const url = new URL(config[key] as string)
    if (!['https:', 'http:'].includes(url.protocol)) throw new Error(`Invalid ${key}`)
    return [key, url]
  })) as Record<typeof bases[number], URL>
  if (typeof config.apiEndpoint !== 'string') throw new Error('Missing apiEndpoint')
  const hostname = host?.toLowerCase().replace(/:\d+$/, '')
  const baseHosts = bases.map(key => urls[key].hostname)
  let branch: string | undefined
  if (hostname && !baseHosts.includes(hostname)) {
    for (const base of baseHosts) {
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
    if (!baseHosts.includes(url.hostname)) return value
    url.hostname = `${branch}.${url.hostname}`
    return url.toString().replace(/\/$/, value.endsWith('/') ? '/' : '')
  }
  return {
    ...(branch ? { branch } : {}),
    homeUrl: prefix(config.homeUrl as string),
    authUiUrl: prefix(config.authUiUrl as string),
    adminUrl: prefix(config.adminUrl as string),
    apiEndpoint: prefix(config.apiEndpoint),
  }
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
