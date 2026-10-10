import getConfig from 'next/config'
import { branchFlowUrl, branchUrl, browserAppUrls } from 'ui/app-urls'

// Use for every browser navigation to a URL that Kratos built from its production ui_urls.
export function followKratosUrl(url: string) {
  window.location.href = branchUrl(getConfig().publicRuntimeConfig, window.location.host, url)
}

// Call with a flow fetched by id before rendering it; true means the browser is moving to the flow's branch host.
export function forwardBranchFlow(flow: { return_to?: string; request_url?: string }): boolean {
  const url = branchFlowUrl(getConfig().publicRuntimeConfig, window.location.href, flow)
  if (url) window.location.replace(url)
  return url !== undefined
}

// Pass as returnTo when creating a flow so Kratos links from emails can find their way back to this branch.
export function flowReturnTo(returnTo: string | string[] | undefined): string | undefined {
  if (returnTo) return String(returnTo)
  const { branch, authUiUrl } = browserAppUrls()
  return branch ? authUiUrl : undefined
}
