import { useRouteContext } from '@tanstack/react-router'
import { createServerFn } from '@tanstack/react-start'

export const getSession = createServerFn({ method: 'GET' }).handler(async () => {
  const { readSession } = await import('./session.server')
  return readSession()
})

export type Identity = Awaited<ReturnType<typeof getSession>>
export type ReadyIdentity = Extract<Identity, { status: 'ready' }>

/** The request's identity on screens the root renders only once the session API has answered. */
export function useIdentity(): ReadyIdentity {
  const identity = useRouteContext({ from: '__root__', select: context => context.identity })
  if (identity.status !== 'ready') throw new Error('Account screens render only with a session answer')
  return identity
}
