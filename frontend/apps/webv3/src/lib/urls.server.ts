import { getRequestHeader } from '@tanstack/react-start/server'

const label = '[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?'
const hostPattern = new RegExp(`^(?:${label}\\.)*${label}(?::[0-9]{1,5})?$`, 'i')
const labelPattern = new RegExp(`^${label}$`)

/** The URLs for the current server function or page request. */
export function requestUrls() {
  return urlsForHost(getRequestHeader('host'))
}

/**
 * The Webv2, API, Webv3, account and admin URLs for a request to `host`. A branch's Webv3 host selects that branch's sites.
 * Server routes outside TanStack Start pass their own Host header; everything else uses `requestUrls()`.
 */
export function urlsForHost(host: string | null | undefined) {
  const { web, api } = runtimeOrigins()
  if (host && !hostPattern.test(host)) throw new Error('Invalid Webv3 request host')
  const appOrigin = `${web.protocol}//${host || `next.${web.host}`}`

  // Branches live under preview.tadoku.app in production and under the development domain otherwise.
  const branchDomain = web.hostname === 'tadoku.app' ? 'preview.tadoku.app' : web.hostname
  const branch = host ? branchName(host, branchDomain) : undefined
  // The account and admin apps are the sibling `account.` and `admin.` hosts, as in authv2 and ui/app-urls.
  const sibling = (app: string) =>
    `${web.protocol}//${branch ? `${branch}.${app}.${branchDomain}` : `${app}.${web.hostname}`}${web.port ? `:${web.port}` : ''}`
  if (branch) {
    web.hostname = `${branch}.${branchDomain}`
    api.href = new URL(api.pathname + api.search, web).href
  }

  return {
    webUrl: web.origin,
    apiUrl: api.href.replace(/\/$/, ''),
    appOrigin,
    accountUrl: sibling('account'),
    adminUrl: sibling('admin'),
  }
}

/** The configured web and API URLs, as new objects on every call. Throws an error naming the variable that is invalid. */
export function runtimeOrigins() {
  return {
    web: runtimeUrl('TADOKU_WEB_URL', process.env.TADOKU_WEB_URL || 'https://tadoku.dev.lab'),
    api: runtimeUrl('TADOKU_API_URL', process.env.TADOKU_API_URL || 'https://tadoku.dev.lab/api/internal'),
  }
}

function runtimeUrl(name: string, value: string) {
  const url = URL.canParse(value) ? new URL(value) : undefined
  if (!url || !['https:', 'http:'].includes(url.protocol) || url.username || url.password) {
    throw new Error(`${name} must be an http or https URL without credentials`)
  }
  return url
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
