import { Button, Navbar, SkipLink, appFrame } from 'paper-ui'
import markUrl from 'paper-ui/assets/brand/cut-meter.svg?url'
import reversedMark from 'paper-ui/assets/brand/cut-meter-reversed.svg?url'
import wordmark from 'paper-ui/assets/brand/wordmark-accent.svg?url'
import reversedWordmark from 'paper-ui/assets/brand/wordmark-reversed.svg?url'
import type { ReactNode } from 'react'
import type { Identity } from '../lib/session'

const frame = appFrame()
const brandMark = 'block h-auto w-brand-mark'
const brandWordmark = 'block h-auto w-brand'

export function Shell({ identity, children }: { readonly identity: Identity; readonly children: ReactNode }) {
  const account = identity.status === 'ready' ? identity.session.account : null
  return (
    <div className="flex min-h-screen flex-col">
      <SkipLink href="#main">Skip to content</SkipLink>
      <header className={frame.header()}>
        <Navbar
          contained
          brand={<>
            <img className={`${brandMark} dark:hidden`} src={markUrl} width="40" height="40" alt="" />
            <img className={`${brandMark} hidden dark:block`} src={reversedMark} width="40" height="40" alt="" />
            <img className={`${brandWordmark} dark:hidden`} src={wordmark} width="158" height="29" alt="Tadoku" />
            <img className={`${brandWordmark} hidden dark:block`} src={reversedWordmark} width="158" height="29" alt="Tadoku" />
          </>}
          brandHref={identity.webUrl}
          navigation={[
            { type: 'link', id: 'home', label: 'Home', href: `${identity.webUrl}/` },
            { type: 'link', id: 'leaderboard', label: 'Leaderboard', href: `${identity.webUrl}/leaderboard/latest/1` },
            { type: 'link', id: 'contests', label: 'Contests', href: `${identity.webUrl}/contests/official/1` },
            { type: 'link', id: 'blog', label: 'Blog', href: `${identity.webUrl}/blog/1` },
            { type: 'link', id: 'guide', label: 'Guide', href: `${identity.webUrl}/pages/manual` },
          ]}
          actions={account ? (
            <form className="ms-4 flex min-w-0 items-center gap-3" method="post" action="/actions/logout">
              <span className="min-w-0 max-w-account-label truncate font-semibold text-ink max-lg:hidden">{account.display_name}</span>
              <Button variant="ghost" type="submit" className="shrink-0 whitespace-nowrap">Log out</Button>
            </form>
          ) : undefined}
        />
      </header>
      <main
        id="main"
        className="w-full flex-1 px-gutter-account pb-18 pt-main-top-account focus:outline-none"
        tabIndex={-1}
      >
        {children}
      </main>
    </div>
  )
}
