import { useRouteContext } from '@tanstack/react-router'
import { createServerFn } from '@tanstack/react-start'

/** The request's session, or with `false` only its URLs, for a page that must not wait for the identity API. */
export const getSession = createServerFn({ method: 'GET' })
  .inputValidator((read: boolean) => read)
  .handler(async ({ data: read }) => {
    const { readSession } = await import('./session.server')
    return readSession(read)
  })

export type Identity = Awaited<ReturnType<typeof getSession>>
export type ReadyIdentity = Extract<Identity, { status: 'ready' }>

/** The request's identity on screens the root renders only once the session API has answered. */
export function useIdentity(): ReadyIdentity {
  const identity = useRouteContext({ from: '__root__', select: context => context.identity })
  if (identity.status !== 'ready') throw new Error('Account screens render only with a session answer')
  return identity
}
