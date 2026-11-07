import { Button, SkipLink } from 'paper-ui'
import markUrl from 'paper-ui/assets/brand/cut-meter.svg?url'
import reversedMark from 'paper-ui/assets/brand/cut-meter-reversed.svg?url'
import wordmark from 'paper-ui/assets/brand/wordmark-accent.svg?url'
import reversedWordmark from 'paper-ui/assets/brand/wordmark-reversed.svg?url'
import type { ReactNode } from 'react'
import type { Identity } from '../lib/session'

/** Render in <head> before the body so Paper's [data-theme] is set before first paint. */
export const themeBootScript = `try{var t=localStorage.getItem('tadoku-theme');document.documentElement.dataset.theme=t==='dark'||t==='light'?t:matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light'}catch(e){}`

const mark = 'block h-auto w-mark-account max-sm:w-7'
const wordmarkSize = 'block h-auto w-wordmark-account max-sm:w-25'

export function Shell({ identity, children }: { readonly identity: Identity; readonly children: ReactNode }) {
  const account = identity.status === 'ready' ? identity.session.account : null
  return (
    <div className="flex min-h-screen flex-col">
      <SkipLink href="#main">Skip to content</SkipLink>
      <header
        className={[
          'flex min-h-header-account flex-wrap items-center justify-between gap-x-6 gap-y-3 border-b border-rule bg-paper',
          'px-header-x-account py-2 max-sm:flex-nowrap max-sm:gap-3',
        ].join(' ')}
      >
        <a className="inline-flex items-center gap-3.5 text-muted no-underline" href={identity.webUrl}>
          <span className="flex items-center gap-1.75">
            <img className={`${mark} dark:hidden`} src={markUrl} width="35" height="35" alt="" />
            <img className={`${mark} hidden dark:block`} src={reversedMark} width="35" height="35" alt="" />
            <img className={`${wordmarkSize} dark:hidden`} src={wordmark} width="158" height="29" alt="Tadoku" />
            <img className={`${wordmarkSize} hidden dark:block`} src={reversedWordmark} width="158" height="29" alt="Tadoku" />
          </span>
          <span className="border-s border-rule ps-3.5 text-sm leading-metadata font-semibold max-sm:hidden">Account</span>
        </a>
        {account ? (
          <form className="m-0 flex min-w-0 items-center gap-4 max-sm:gap-3" method="post" action="/actions/logout">
            <span className="min-w-0 max-w-account-label truncate text-sm text-ink max-sm:max-w-none">
              {account.display_name}
            </span>
            <Button
              variant="ghost"
              type="submit"
              className="min-h-6 border-0 p-0 text-sm font-regular underline max-sm:shrink-0 max-sm:whitespace-nowrap"
            >
              Log out
            </Button>
          </form>
        ) : null}
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
