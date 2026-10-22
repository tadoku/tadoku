import getConfig from 'next/config'
import { createContext, ReactNode, useContext } from 'react'

export interface AppUrls {
  branch?: string
  homeUrl: string
  accountUrl: string
  adminUrl: string
  apiEndpoint: string
}

const labelPattern = /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/

function branchMapper(config: Record<string, unknown>, host?: string) {
  const bases = ['homeUrl', 'accountUrl', 'adminUrl']
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

/** `accountUrl` is always the `account.` sibling of `homeUrl`, where the account app is served. */
export function resolveAppUrls(config: Record<string, unknown>, host?: string): AppUrls {
  if (typeof config.homeUrl !== 'string') throw new Error('Missing homeUrl')
  const account = new URL(config.homeUrl)
  account.hostname = `account.${account.hostname}`
  const { branch, prefix } = branchMapper({ ...config, accountUrl: account.origin }, host)
  if (typeof config.apiEndpoint !== 'string') throw new Error('Missing apiEndpoint')
  return {
    ...(branch ? { branch } : {}),
    homeUrl: prefix(config.homeUrl),
    accountUrl: prefix(account.origin),
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
    apiEndpoint: (typeof window === 'undefined' ? process.env.NEXT_SERVER_API_ENDPOINT : undefined)
      ?? serverRuntimeConfig?.apiEndpoint
      ?? publicRuntimeConfig.apiEndpoint,
  }, host).apiEndpoint
}

const AppUrlsContext = createContext<AppUrls | undefined>(undefined)

export function AppUrlsProvider({ urls, children }: { urls: AppUrls; children: ReactNode }) {
  return <AppUrlsContext.Provider value={urls}>{children}</AppUrlsContext.Provider>
}

export function useAppUrls(): AppUrls {
  return useContext(AppUrlsContext) ?? browserAppUrls()
}
