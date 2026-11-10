/** Marks the 409 answer to a server function this build doesn't have, as called by a tab from an earlier release. */
export const outdatedBuildHeader = 'X-Tadoku-Outdated-Build'

const reloadedAtKey = 'tadoku-outdated-build-reloaded-at'
const reloadPauseMs = 5 * 60_000
let reloading: Promise<never> | undefined

/**
 * Fetches a server function call from the browser. When the server no longer has the function, the tab reloads once to
 * pick up the current build and the call never settles. A tab that reloaded for this in the last five minutes, or that
 * cannot store when it did, gets the 409 as the call's error instead, so a lasting mismatch cannot reload forever.
 */
export async function serverFunctionFetch(input: RequestInfo | URL, init?: RequestInit) {
  const response = await fetch(input, init)
  if (response.status !== 409 || response.headers.get(outdatedBuildHeader) !== 'true') return response
  if (reloading || claimReload()) {
    reloading ??= new Promise<never>(() => location.reload())
    return reloading
  }
  return response
}

function claimReload() {
  try {
    const last = Number(sessionStorage.getItem(reloadedAtKey))
    if (Date.now() - last < reloadPauseMs) return false
    sessionStorage.setItem(reloadedAtKey, String(Date.now()))
    return true
  } catch {
    return false
  }
}
