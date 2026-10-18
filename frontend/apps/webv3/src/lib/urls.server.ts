import { getRequestHeader } from '@tanstack/react-start/server'

const label = '[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?'
const hostPattern = new RegExp(`^(?:${label}\\.)*${label}(?::[0-9]{1,5})?$`, 'i')
const labelPattern = new RegExp(`^${label}$`)

export function resolveRequestUrls(webUrl: string, apiUrl: string, host?: string) {
  const web = new URL(webUrl)
  const api = new URL(apiUrl)
  for (const url of [web, api]) {
    if (!['https:', 'http:'].includes(url.protocol) || url.username || url.password) {
      throw new Error('Invalid Tadoku runtime URL')
    }
  }
  if (host && !hostPattern.test(host)) throw new Error('Invalid Webv3 request host')

  const hostname = host?.toLowerCase().replace(/:\d+$/, '')
  const production = web.hostname === 'tadoku.app'
  const branchRoot = production ? 'next.preview.tadoku.app' : `next.${web.hostname}`
  let branch: string | undefined
  const kratosUrl = new URL(`/kratos`, new URL(`${web.protocol}//account.${web.host}`)).href
  if (hostname?.endsWith(`.${branchRoot}`)) {
    branch = hostname.slice(0, -(branchRoot.length + 1))
    if (!labelPattern.test(branch)) throw new Error('Invalid Webv3 branch host')
    web.hostname = `${branch}.${production ? 'preview.tadoku.app' : web.hostname}`
    api.href = new URL(`${api.pathname}${api.search}`, web).href
  }

  // ponytail: sibling account host; add TADOKU_AUTH_URL only for a non-sibling deployment.
  const authHost = branch
    ? `${branch}.account.${production ? 'preview.tadoku.app' : new URL(webUrl).hostname}`
    : `account.${web.hostname}`
  return {
    webUrl: web.origin, apiUrl: api.href.replace(/\/$/, ''),
    appOrigin: `${web.protocol}//${host || `next.${web.host}`}`,
    authUrl: `${web.protocol}//${authHost}${web.port ? `:${web.port}` : ''}`, kratosUrl,
  }
}

export function requestUrls() {
  return resolveRequestUrls(
    process.env.TADOKU_WEB_URL || 'https://tadoku.dev.lab',
    process.env.TADOKU_API_URL || 'https://tadoku.dev.lab/api/internal',
    getRequestHeader('host'),
  )
}
