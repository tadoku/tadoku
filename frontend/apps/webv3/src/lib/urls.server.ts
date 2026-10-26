import { getRequestHeader } from '@tanstack/react-start/server'

const label = '[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?'
const hostPattern = new RegExp(`^(?:${label}\\.)*${label}(?::[0-9]{1,5})?$`, 'i')
const labelPattern = new RegExp(`^${label}$`)

/** The URLs for the current server function or page request. */
export function requestUrls() {
  return urlsForHost(getRequestHeader('host'))
}

/**
 * The Webv2, API, Webv3 and account URLs for a request to `host`. A branch's Webv3 host selects that branch's sites.
 * Server routes outside TanStack Start pass their own Host header; everything else uses `requestUrls()`.
 */
export function urlsForHost(host: string | null | undefined) {
  const web = new URL(process.env.TADOKU_WEB_URL || 'https://tadoku.dev.lab')
  const api = new URL(process.env.TADOKU_API_URL || 'https://tadoku.dev.lab/api/internal')
  for (const url of [web, api]) {
    if (!['https:', 'http:'].includes(url.protocol) || url.username || url.password) {
      throw new Error('Invalid Tadoku runtime URL')
    }
  }
  if (host && !hostPattern.test(host)) throw new Error('Invalid Webv3 request host')
  const appOrigin = `${web.protocol}//${host || `next.${web.host}`}`

  // Branches live under preview.tadoku.app in production and under the development domain otherwise.
  const branchDomain = web.hostname === 'tadoku.app' ? 'preview.tadoku.app' : web.hostname
  const branch = host ? branchName(host, branchDomain) : undefined
  // The account app is the sibling `account.` host, as in authv2 and ui/app-urls.
  const accountHost = branch ? `${branch}.account.${branchDomain}` : `account.${web.hostname}`
  if (branch) {
    web.hostname = `${branch}.${branchDomain}`
    api.href = new URL(api.pathname + api.search, web).href
  }

  return {
    webUrl: web.origin,
    apiUrl: api.href.replace(/\/$/, ''),
    appOrigin,
    accountUrl: `${web.protocol}//${accountHost}${web.port ? `:${web.port}` : ''}`,
  }
}

/** The branch a `<branch>.next.<domain>` host selects, if it is one. */
function branchName(host: string, domain: string) {
  const hostname = host.toLowerCase().replace(/:\d+$/, '')
  const root = `.next.${domain}`
  if (!hostname.endsWith(root)) return undefined

  const branch = hostname.slice(0, -root.length)
  if (!labelPattern.test(branch)) throw new Error('Invalid Webv3 branch host')
  return branch
}
