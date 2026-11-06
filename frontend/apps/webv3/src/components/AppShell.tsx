import { Link, rootRouteId, useHydrated, useLoaderData, useNavigate, useRouterState } from '@tanstack/react-router'
import { ActionMenu, Button, EmptyState, Navbar, buttonClassName, type NavbarItem } from 'paper-ui'
import mark from 'paper-ui/assets/brand/cut-meter.svg?url'
import reversedMark from 'paper-ui/assets/brand/cut-meter-reversed.svg?url'
import wordmark from 'paper-ui/assets/brand/wordmark-accent.svg?url'
import reversedWordmark from 'paper-ui/assets/brand/wordmark-reversed.svg?url'
import { Fragment, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronDownIcon, iconClassName } from 'paper-ui/icons'
import { useAccountUrls, viewerQueryOptions, type Viewer } from '../lib/session'
import { useLinks } from '../lib/links'
import { Announcements } from './Announcements'
import { renderLink } from './router-link'

const themeKey = 'tadoku-theme'

/** Render in <head> before the body so Paper's [data-theme] is set before first paint. */
export const themeBootScript = `try{var t=localStorage.getItem('${themeKey}');document.documentElement.dataset.theme=t==='dark'||t==='light'?t:matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light'}catch(e){}`

function toggleTheme() {
  const next = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark'
  document.documentElement.dataset.theme = next
  try { localStorage.setItem(themeKey, next) } catch { /* Private mode keeps the choice for this page only. */ }
}

const discordUrl = 'https://discord.gg/AsC9vZs2Ex'
const logoutFormId = 'app-logout'

export function AppShell({ children }: { children: ReactNode }) {
  const currentPath = useRouterState({ select: state => state.location.pathname })
  const hydrated = useHydrated()
  const links = useLinks()
  const { data: viewer } = useQuery(viewerQueryOptions())
  const navigation: NavbarItem[] = [
    { type: 'link', id: 'home', label: 'Home', href: links.home, current: currentPath === links.home },
    { type: 'link', id: 'leaderboard', label: 'Leaderboard', href: links.leaderboard, current: currentPath.startsWith('/leaderboard/') },
    { type: 'link', id: 'contests', label: 'Contests', href: links.contests, current: currentPath.startsWith('/contests/') },
    { type: 'link', id: 'blog', label: 'Blog', href: links.blog, current: currentPath === links.blog || currentPath.startsWith('/blog/') },
    { type: 'link', id: 'guide', label: 'Guide', href: links.guide, current: currentPath === links.guide },
  ]

  return (
    <div className="app-shell">
      <a className={`${buttonClassName()} app-skip`} href="#main-content">Skip to content</a>
      <header className="app-header">
        <Navbar
          brand={<>
            <img className="app-mark app-mark--light" src={mark} width="40" height="40" alt="" />
            <img className="app-mark app-mark--dark" src={reversedMark} width="40" height="40" alt="" />
            <img className="app-wordmark app-wordmark--light" src={wordmark} width="158" height="29" alt="Tadoku" />
            <img className="app-wordmark app-wordmark--dark" src={reversedWordmark} width="158" height="29" alt="Tadoku" />
          </>}
          brandHref="/"
          navigation={navigation}
          currentPath={currentPath}
          renderLink={renderLink}
          actions={<div className="app-account-actions"><AccountActions viewer={viewer} /></div>}
          mobileFooter={viewer?.status === 'signed-in' ? closeMenu => <MobileAccount viewer={viewer} closeMenu={closeMenu} /> : undefined}
        />
        {viewer && 'userId' in viewer ? <form id={logoutFormId} method="post" action="/logout" hidden /> : null}
      </header>
      <main id="main-content" className="app-content" tabIndex={-1}>
        <Announcements />
        {viewer?.status === 'banned' ? <Suspended /> : children}
      </main>
      <footer className="app-footer">
        <div className="app-footer__inner">
          <nav aria-label="Footer">
            <Link className="text-link" to={links.home}>Home</Link>
            <Link className="text-link" to={links.leaderboard}>Leaderboard</Link>
            <Link className="text-link" to={links.contests}>Contests</Link>
            <Link className="text-link" to={links.blog}>Blog</Link>
            <Link className="text-link" to={links.page('about')}>About Tadoku</Link>
            <Link className="text-link" to={links.guide}>Guide</Link>
            <Link className="text-link" to={links.page('rules')}>Rules</Link>
            <Link className="text-link" to={links.page('faq')}>Questions &amp; help</Link>
            <Link className="text-link" to={links.page('privacy')}>Privacy</Link>
            <Link className="text-link" to={links.page('contact')}>Contact</Link>
            <a className="text-link" href="https://github.com/tadoku">GitHub</a>
            <a className="text-link" href={discordUrl}>Discord</a>
            <a className="text-link" href="https://twitter.com/tadoku_app">Twitter</a>
          </nav>
          <div className="app-footer__meta">
            <span>Built by <a className="text-link" href="https://antonve.be">antonve</a></span>
            <Button variant="link" onClick={toggleTheme} disabled={!hydrated}>
              <span className="app-theme-label--light">Use dark theme</span>
              <span className="app-theme-label--dark">Use light theme</span>
            </Button>
          </div>
        </div>
      </footer>
    </div>
  )
}

/** Webv2 replaces every page with this for a suspended account. */
function Suspended() {
  return (
    <EmptyState
      headingLevel={1}
      title="Your account is suspended"
      description="If you believe this is an error, contact us on Discord."
      actions={<>
        <a className="text-link" href={discordUrl}>Contact us on Discord</a>
        <Button type="submit" form={logoutFormId}>Log out</Button>
      </>}
    />
  )
}

type SignedIn = Extract<Viewer, { status: 'signed-in' }>

/** The account's destinations, in Webv2's order: profile, admin for administrators, then account settings. */
function useAccountLinks(viewer: SignedIn) {
  const links = useLinks()
  const { adminUrl } = useLoaderData({ from: rootRouteId })
  const { settings } = useAccountUrls()
  return [
    { id: 'profile', label: 'Profile', href: links.profile(viewer.userId) },
    ...(viewer.role === 'admin' ? [{ id: 'admin', label: 'Admin', href: adminUrl }] : []),
    { id: 'settings', label: 'Account settings', href: settings },
  ]
}

const logOut = () => (document.getElementById(logoutFormId) as HTMLFormElement).requestSubmit()

function AccountActions({ viewer }: { viewer: Viewer | undefined }) {
  const { login, register } = useAccountUrls()
  const links = useLinks()

  switch (viewer?.status) {
    case undefined:
    case 'banned':
      return null
    case 'unavailable':
      return <span>Account service unavailable</span>
    case 'anonymous':
      return <>
        <a href={login} className="text-link">Log in</a>
        <a href={register} className={buttonClassName({ variant: 'outline' })}>Sign up</a>
      </>
    case 'signed-in':
      return <>
        <AccountMenu viewer={viewer} />
        <a href={links.logs} className={buttonClassName({ className: 'app-header-action' })}>Log activity</a>
      </>
  }
}

function AccountMenu({ viewer }: { viewer: SignedIn }) {
  const hydrated = useHydrated()
  const navigate = useNavigate()
  const items = useAccountLinks(viewer).map(link => ({
    id: link.id,
    label: link.label,
    onSelect: () => link.href.startsWith('/') ? void navigate({ to: link.href }) : location.assign(link.href),
  }))
  return (
    <span className="app-account-menu--desktop">
      {hydrated
        ? <ActionMenu label={viewer.displayName} triggerVariant="ghost" items={[...items, { id: 'logout', label: 'Log out', onSelect: logOut }]} />
        : (
          // The menu's trigger as server HTML renders it, disabled until it can open.
          <button type="button" className={buttonClassName({ variant: 'ghost' })} disabled>
            <span>{viewer.displayName}</span>
            <ChevronDownIcon className={iconClassName('compact')} aria-hidden="true" />
          </button>
        )}
    </span>
  )
}

function MobileAccount({ viewer, closeMenu }: { viewer: SignedIn; closeMenu: () => void }) {
  const links = useAccountLinks(viewer)
  return (
    <section className="app-mobile-account" aria-label={`${viewer.displayName} account`}>
      <p className="app-mobile-account__name">{viewer.displayName}</p>
      {links.map(link => (
        <Fragment key={link.id}>
          {renderLink({ href: link.href, className: 'paper-navbar__mobile-link', onClick: closeMenu, children: link.label })}
        </Fragment>
      ))}
      <Button type="submit" form={logoutFormId} variant="ghost" className="paper-navbar__mobile-link">Log out</Button>
    </section>
  )
}
