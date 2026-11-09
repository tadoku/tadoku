import type { QueryClient, QueryKey } from '@tanstack/react-query'

/** A page whose main read failed. `src/server.ts` answers it with 503; the nearest error component renders it. */
export class PageUnavailableError extends Error {
  override name = 'PageUnavailableError'
}

/**
 * Fails the page because its main read failed. Call it from a loader with that read's query: dropping the cached
 * failure keeps it out of the hydrated cache, so Try again reads it afresh.
 */
export function pageUnavailable(queryClient: QueryClient, query: { queryKey: QueryKey }): never {
  queryClient.removeQueries({ queryKey: query.queryKey, exact: true })
  throw new PageUnavailableError('The page’s main read failed')
}
