import { PageHeader, Surface } from 'paper-ui'
import { Route } from '../routes/__root'

export function AccountSkeleton() {
  const { identity } = Route.useRouteContext()
  return <>
    <PageHeader title="Your Tadoku account" lead="Account pages are being prepared." />
    <Surface className="account-status">
      {identity.status === 'unavailable'
        ? <p role="status">Account service is unavailable. Try again later.</p>
        : <p>{identity.session.account ? `Welcome, ${identity.session.account.display_name}.` : 'You are not signed in.'}</p>}
      <a className="text-link" href={new URL('/login', identity.accountUrl).href}>Open existing account app</a>
    </Surface>
  </>
}
