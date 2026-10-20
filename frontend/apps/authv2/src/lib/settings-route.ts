import { redirect, type ParsedLocation } from '@tanstack/react-router'
import { continuation, withReturnTo } from './identity'
import { accountSearch } from './search'
import type { Identity } from './session'

/** Settings route options shared by `/`, the identity API's recovery target, and `/settings`. */
export const settingsRoute = {
  validateSearch: accountSearch,
  head: () => ({ meta: [{ title: 'Account settings · Tadoku' }] }),
  beforeLoad: ({ context, location }: { context: { identity: Identity }; location: ParsedLocation }) => {
    const { identity } = context
    if (identity.status !== 'ready' || identity.session.account) return
    const next = continuation(accountSearch(location.search).return_to, identity.origin)
    const here = withReturnTo(`${identity.origin}${location.pathname}`, next)
    throw redirect({ href: next && continuation(here, identity.origin) ? withReturnTo('/login', here) : '/login' })
  },
}
