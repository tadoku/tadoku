import * as stylex from '@stylexjs/stylex'
import { Button, buttonStyles, linkStyles, ThemeVariant } from 'paper-ui'
import mark from 'paper-ui/assets/brand/cut-meter.svg?url'
import reversedMark from 'paper-ui/assets/brand/cut-meter-reversed.svg?url'
import wordmark from 'paper-ui/assets/brand/wordmark-accent.svg?url'
import reversedWordmark from 'paper-ui/assets/brand/wordmark-reversed.svg?url'
import type { ReactNode } from 'react'
import type { Identity } from '../lib/session'
import { shellStyles } from './styles'

/** The stored theme, shared with Webv3 and Webv2. */
export const themeKey = 'tadoku-theme'

export function Shell({ identity, children }: { readonly identity: Identity; readonly children: ReactNode }) {
  const account = identity.status === 'ready' ? identity.session.account : null
  return (
    <div {...stylex.props(shellStyles.shell)}>
      <a {...stylex.props(...buttonStyles(), shellStyles.skip)} href="#main">Skip to content</a>
      <header {...stylex.props(shellStyles.header)}>
        <a {...stylex.props(linkStyles.plain, shellStyles.brand)} href={identity.webUrl}>
          <span {...stylex.props(shellStyles.lockup)}>
            <ThemeVariant
              light={<img {...stylex.props(shellStyles.mark)} src={mark} width="35" height="35" alt="" />}
              dark={<img {...stylex.props(shellStyles.mark)} src={reversedMark} width="35" height="35" alt="" />}
            />
            <ThemeVariant
              light={<img {...stylex.props(shellStyles.wordmark)} src={wordmark} width="158" height="29" alt="Tadoku" />}
              dark={<img {...stylex.props(shellStyles.wordmark)} src={reversedWordmark} width="158" height="29" alt="Tadoku" />}
            />
          </span>
          <span {...stylex.props(shellStyles.brandLabel)}>Account</span>
        </a>
        {account ? (
          <form {...stylex.props(shellStyles.session)} method="post" action="/actions/logout">
            <span {...stylex.props(shellStyles.sessionName)}>{account.display_name}</span>
            <Button variant="ghost" type="submit" xstyle={shellStyles.logout}>Log out</Button>
          </form>
        ) : null}
      </header>
      <main id="main" {...stylex.props(shellStyles.main)} tabIndex={-1}>{children}</main>
    </div>
  )
}
