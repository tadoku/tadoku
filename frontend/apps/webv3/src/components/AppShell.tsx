import { Link, useHydrated, useRouterState } from '@tanstack/react-router'
import { Button, Navbar, buttonClassName, type NavbarItem, type NavigationLinkProps } from 'paper-ui'
import wordmark from 'paper-ui/assets/brand/wordmark-accent.svg?url'
import reversedWordmark from 'paper-ui/assets/brand/wordmark-reversed.svg?url'
import type { ReactNode } from 'react'
import { useLinks } from '../lib/links'

const themeKey = 'tadoku-theme'

/** Render in <head> before the body so Paper's [data-theme] is set before first paint. */
export const themeBootScript = `try{var t=localStorage.getItem('${themeKey}');document.documentElement.dataset.theme=t==='dark'||t==='light'?t:matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light'}catch(e){}`

function toggleTheme() {
  const next = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark'
  document.documentElement.dataset.theme = next
  try { localStorage.setItem(themeKey, next) } catch { /* Private mode keeps the choice for this page only. */ }
}

function renderLink({ href, ...props }: NavigationLinkProps) {
  return href.startsWith('/') ? <Link to={href as '/' | '/plan'} {...props} /> : <a href={href} {...props} />
}

export function AppShell({ children }: { children: ReactNode }) {
  const currentPath = useRouterState({ select: state => state.location.pathname })
  const hydrated = useHydrated()
  const links = useLinks()
  const navigation: NavbarItem[] = [
    { type: 'link', id: 'leaderboard', label: 'Leaderboard', href: links.leaderboard },
    { type: 'link', id: 'contests', label: 'Contests', href: links.contests },
    { type: 'link', id: 'guide', label: 'Guide', href: links.guide },
    { type: 'link', id: 'goal', label: 'Daily goal', href: '/plan' },
  ]

  return (
    <div className="app-shell">
      <a className={`${buttonClassName()} app-skip`} href="#main-content">Skip to content</a>
      <header className="app-header">
        <Navbar
          brand={<>
            <img className="app-wordmark app-wordmark--light" src={wordmark} width="158" height="29" alt="Tadoku" />
            <img className="app-wordmark app-wordmark--dark" src={reversedWordmark} width="158" height="29" alt="Tadoku" />
          </>}
          brandHref="/"
          navigation={navigation}
          currentPath={currentPath}
          renderLink={renderLink}
          actions={<a href={links.logs} className={buttonClassName({ className: 'app-header-action' })}>Log activity</a>}
        />
      </header>
      <main id="main-content" className="app-content" tabIndex={-1}>{children}</main>
      <footer className="app-footer">
        <div className="app-footer__inner">
          <nav aria-label="Footer">
            <a className="text-link" href={links.legacyUrl('/pages/about')}>About Tadoku</a>
            <a className="text-link" href={links.guide}>Guide</a>
            <a className="text-link" href="https://github.com/tadoku/tadoku">GitHub</a>
          </nav>
          <Button variant="link" onClick={toggleTheme} disabled={!hydrated}>
            <span className="app-theme-label--light">Use dark theme</span>
            <span className="app-theme-label--dark">Use light theme</span>
          </Button>
        </div>
      </footer>
    </div>
  )
}
