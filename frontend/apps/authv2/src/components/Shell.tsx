import { Button, buttonClassName } from 'paper-ui'
import wordmark from 'paper-ui/assets/brand/wordmark-accent.svg?url'
import reversedWordmark from 'paper-ui/assets/brand/wordmark-reversed.svg?url'
import type { ReactNode } from 'react'
import type { Identity } from '../lib/session'

/** Render in <head> before the body so Paper's [data-theme] is set before first paint. */
export const themeBootScript = `try{var t=localStorage.getItem('tadoku-theme');document.documentElement.dataset.theme=t==='dark'||t==='light'?t:matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light'}catch(e){}`

export function Shell({ identity, children }: { readonly identity: Identity; readonly children: ReactNode }) {
  const account = identity.status === 'ready' ? identity.session.account : null
  return (
    <div className="account-shell">
      <a className={buttonClassName({ className: 'account-skip' })} href="#main">Skip to content</a>
      <header className="account-header">
        <a className="account-brand" href={identity.webUrl}>
          <img className="account-wordmark account-wordmark--light" src={wordmark} width="158" height="29" alt="Tadoku" />
          <img className="account-wordmark account-wordmark--dark" src={reversedWordmark} width="158" height="29" alt="Tadoku" />
          <span className="account-brand__label">Account</span>
        </a>
        {account ? (
          <form className="account-session" method="post" action="/actions/logout">
            <span className="account-session__name">{account.display_name}</span>
            <Button variant="ghost" type="submit">Log out</Button>
          </form>
        ) : null}
      </header>
      <main id="main" className="account-main" tabIndex={-1}>{children}</main>
    </div>
  )
}
