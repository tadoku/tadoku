import { getRequestHeader } from '@tanstack/react-start/server'

const label = '[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?'
const hostPattern = new RegExp(`^(?:${label}\\.)*${label}(?::[0-9]{1,5})?$`, 'i')
const labelPattern = new RegExp(`^${label}$`)

export function requestUrls() {
  const web = new URL(process.env.TADOKU_WEB_URL || 'https://tadoku.dev.lab')
  const api = new URL(process.env.TADOKU_API_URL || 'https://tadoku.dev.lab/api/internal')
  for (const url of [web, api]) {
    if (!['https:', 'http:'].includes(url.protocol) || url.username || url.password) {
      throw new Error('Invalid Tadoku runtime URL')
    }
  }

  const host = getRequestHeader('host')
  if (host && !hostPattern.test(host)) throw new Error('Invalid account request host')
  const hostname = host?.toLowerCase().replace(/:\d+$/, '')
  const production = web.hostname === 'tadoku.app'
  const branchRoot = production ? 'authv2.preview.tadoku.app' : `authv2.${web.hostname}`
  const origin = `${web.protocol}//${host?.toLowerCase() || `authv2.${web.host}`}`
  const account = new URL(web.origin)
  account.hostname = `account.${web.hostname}`
  if (hostname?.endsWith(`.${branchRoot}`)) {
    const branch = hostname.slice(0, -(branchRoot.length + 1))
    if (!labelPattern.test(branch)) throw new Error('Invalid account branch host')
    const branchDomain = production ? 'preview.tadoku.app' : web.hostname
    web.hostname = `${branch}.${branchDomain}`
    account.hostname = `${branch}.account.${branchDomain}`
    api.host = web.host
    api.protocol = web.protocol
  }

  return { origin, webUrl: web.origin, accountUrl: account.origin, apiUrl: api.href.replace(/\/$/, '') }
}
