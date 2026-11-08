import { rootRouteId, useHydrated, useLoaderData, useNavigate, useRouterState } from '@tanstack/react-router'
import { ActionMenu, Button, EmptyState, Navbar, SiteFooter, SkipLink, appFrame, buttonClassName, cn, navbarStyles, type NavbarItem } from 'paper-ui'
import mark from 'paper-ui/assets/brand/cut-meter.svg?url'
import reversedMark from 'paper-ui/assets/brand/cut-meter-reversed.svg?url'
import wordmark from 'paper-ui/assets/brand/wordmark-accent.svg?url'
import reversedWordmark from 'paper-ui/assets/brand/wordmark-reversed.svg?url'
import { Fragment, useState, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  ArrowRightStartOnRectangleIcon,
  ChevronDownIcon,
  Cog6ToothIcon,
  MoonIcon,
  ShieldCheckIcon,
  SunIcon,
  UserIcon,
  iconClassName,
} from 'paper-ui/icons'
import { DiscordIcon, GitHubIcon, TwitterIcon } from './BrandIcons'
import { useAccountUrls, viewerQueryOptions, type Viewer } from '../lib/session'
import { useLinks } from '../lib/links'
import { Announcements } from './Announcements'
import { renderLink } from './router-link'

const themeKey = 'tadoku-theme'

/** Render in <head> before the body so Paper's [data-theme] is set before first paint. */
export const themeBootScript = `try{var t=localStorage.getItem('${themeKey}');document.documentElement.dataset.theme=t==='dark'||t==='light'?t:matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light'}catch(e){}`

type Theme = 'light' | 'dark'

const currentTheme = (): Theme => document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light'

function setTheme(next: Theme) {
  document.documentElement.dataset.theme = next
  try { localStorage.setItem(themeKey, next) } catch { /* Private mode keeps the choice for this page only. */ }
}

const toggleTheme = () => setTheme(currentTheme() === 'dark' ? 'light' : 'dark')

/** The account menu's appearance item. The menu renders only after hydration, so it reads the theme directly. */
function useThemeSwitch() {
  const [theme, setThemeState] = useState<Theme>(currentTheme)
  const next: Theme = theme === 'dark' ? 'light' : 'dark'
  return {
    label: next === 'dark' ? 'Use dark theme' : 'Use light theme',
    icon: next === 'dark' ? <MoonIcon className={iconClassName('compact')} /> : <SunIcon className={iconClassName('compact')} />,
    toggle: () => { setTheme(next); setThemeState(next) },
  }
}

const discordUrl = 'https://discord.gg/AsC9vZs2Ex'
const frame = appFrame()
const brandMark = 'block h-auto w-brand-mark'
const brandWordmark = 'block h-auto w-brand'
const accountActions = [
  'ms-4 flex min-w-0 items-center justify-end gap-3 max-md:flex-wrap',
  'max-md:max-w-account-label max-md:gap-1.6 max-md:text-sm-tight',
].join(' ')
const mobileLink = navbarStyles({ contained: true }).mobileLink()
const mobileButton = cn(mobileLink, 'w-full justify-start')
const logoutFormId = 'app-logout'

export function AppShell({ children }: { children: ReactNode }) {
  const currentPath = useRouterState({ select: state => state.location.pathname })
  const navigating = useRouterState({ select: state => state.status === 'pending' })
  const links = useLinks()
  const { data: viewer } = useQuery(viewerQueryOptions())
  const navigation: NavbarItem[] = [
    { type: 'link', id: 'home', label: 'Home', href: links.home, current: currentPath === links.home },
    { type: 'link', id: 'leaderboard', label: 'Leaderboard', href: links.leaderboard, current: currentPath.startsWith('/leaderboard/') },
    { type: 'link', id: 'contests', label: 'Contests', href: links.contests, current: currentPath.startsWith('/contests/') },
    { type: 'link', id: 'blog', label: 'Blog', href: links.blog, current: currentPath === '/blog' || currentPath.startsWith('/blog/') },
    { type: 'link', id: 'guide', label: 'Guide', href: links.guide, current: currentPath === links.guide },
  ]

  return (
    <div className={frame.root()}>
      <SkipLink href="#main-content">Skip to content</SkipLink>
      <header className={frame.header()}>
        <Navbar
          contained
          brand={<>
            <img className={`${brandMark} dark:hidden`} src={mark} width="40" height="40" alt="" />
            <img className={`${brandMark} hidden dark:block`} src={reversedMark} width="40" height="40" alt="" />
            <img className={`${brandWordmark} dark:hidden`} src={wordmark} width="158" height="29" alt="Tadoku" />
            <img className={`${brandWordmark} hidden dark:block`} src={reversedWordmark} width="158" height="29" alt="Tadoku" />
          </>}
          brandHref="/"
          navigation={navigation}
          isLoading={navigating}
          currentPath={currentPath}
          renderLink={renderLink}
          actions={<div className={accountActions}><AccountActions viewer={viewer} /></div>}
          mobileFooter={viewer?.status === 'signed-in' ? closeMenu => <MobileAccount viewer={viewer} closeMenu={closeMenu} /> : undefined}
        />
        {viewer && 'userId' in viewer ? <form id={logoutFormId} method="post" action="/logout" hidden /> : null}
      </header>
      <main id="main-content" className={frame.content()} tabIndex={-1} aria-busy={navigating}>
        <Announcements />
        {viewer?.status === 'banned' ? <Suspended /> : children}
      </main>
      <SiteFooter
        brand={<>
          <img src={reversedMark} width="32" height="32" alt="" />
          <img src={reversedWordmark} width="126" height="23" alt="" />
        </>}
        brandHref={links.home}
        brandLabel="Tadoku home"
        renderLink={renderLink}
        credit={<>Built &amp; designed by <a href="https://antonve.be">antonve</a></>}
        social={[
          { label: 'Discord', href: discordUrl, icon: <DiscordIcon /> },
          { label: 'GitHub', href: 'https://github.com/tadoku', icon: <GitHubIcon /> },
          { label: 'Twitter', href: 'https://twitter.com/tadoku_app', icon: <TwitterIcon /> },
        ]}
        groups={[
          { title: 'Get started', links: [
            { label: 'Leaderboard', href: links.leaderboard },
            { label: 'Contests', href: links.contests },
            { label: 'Blog', href: links.blog },
          ] },
          { title: 'Learn', links: [
            { label: 'About Tadoku', href: links.page('about') },
            { label: 'Guide', href: links.guide },
            { label: 'Rules', href: links.page('rules') },
            { label: 'Questions & help', href: links.page('faq') },
          ] },
          { title: 'Legal', links: [
            { label: 'Privacy', href: links.page('privacy') },
            { label: 'Contact', href: links.page('contact') },
          ] },
        ]}
      />
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
        <a className="text-link hover:text-link-hover" href={discordUrl}>Contact us on Discord</a>
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
    { id: 'profile', label: 'Profile', href: links.profile(viewer.userId), icon: <UserIcon className={iconClassName('compact')} /> },
    ...(viewer.role === 'admin'
      ? [{ id: 'admin', label: 'Admin', href: adminUrl, icon: <ShieldCheckIcon className={iconClassName('compact')} /> }]
      : []),
    { id: 'settings', label: 'Account settings', href: settings, icon: <Cog6ToothIcon className={iconClassName('compact')} /> },
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
        <a href={login} className={buttonClassName({ variant: 'ghost' })}>Log in</a>
        <a href={register} className={buttonClassName({ className: 'shrink-0 whitespace-nowrap' })}>Sign up</a>
      </>
    case 'signed-in':
      return <>
        <AccountMenu viewer={viewer} />
        <a href={links.logs} className={buttonClassName({ className: 'shrink-0 whitespace-nowrap' })}>Log activity</a>
      </>
  }
}

function AccountMenu({ viewer }: { viewer: SignedIn }) {
  const hydrated = useHydrated()
  return (
    <span className="max-lg:hidden [&_.paper-button>span]:max-w-measure-label [&_.paper-button>span]:truncate">
      {hydrated
        ? <AccountMenuOpen viewer={viewer} />
        : (
          // The menu's trigger as server HTML renders it, disabled until it can open.
          <button type="button" className={buttonClassName({ variant: 'ghost', className: 'disabled:cursor-progress disabled:opacity-full' })} disabled>
            <span>{viewer.displayName}</span>
            <ChevronDownIcon className={iconClassName('compact')} aria-hidden="true" />
          </button>
        )}
    </span>
  )
}

/** The account menu once it can open: the account's pages, the appearance switch, then Log out. */
function AccountMenuOpen({ viewer }: { viewer: SignedIn }) {
  const navigate = useNavigate()
  const theme = useThemeSwitch()
  const items = useAccountLinks(viewer).map(link => ({
    id: link.id,
    label: link.label,
    icon: link.icon,
    onSelect: () => link.href.startsWith('/') ? void navigate({ to: link.href }) : location.assign(link.href),
  }))
  return (
    <ActionMenu
      label={viewer.displayName}
      triggerVariant="ghost"
      items={[
        ...items,
        { id: 'theme', label: theme.label, icon: theme.icon, onSelect: theme.toggle },
        { id: 'logout', label: 'Log out', icon: <ArrowRightStartOnRectangleIcon className={iconClassName('compact')} />, onSelect: logOut },
      ]}
    />
  )
}

function MobileAccount({ viewer, closeMenu }: { viewer: SignedIn; closeMenu: () => void }) {
  const links = useAccountLinks(viewer)
  return (
    <section className="mt-3 border-t border-rule-subtle pt-3" aria-label={`${viewer.displayName} account`}>
      <p className="mb-1 mt-0 px-field-x text-sm font-semibold text-muted wrap-anywhere">{viewer.displayName}</p>
      {links.map(link => (
        <Fragment key={link.id}>
          {renderLink({ href: link.href, className: mobileLink, onClick: closeMenu, children: link.label })}
        </Fragment>
      ))}
      {/* Server-rendered, so CSS shows the label for the current theme instead of reading it during render. */}
      <Button
        variant="ghost"
        className={cn(mobileButton, 'mt-3 border-t border-rule-subtle')}
        onClick={toggleTheme}
        leadingIcon={(
          <>
            <MoonIcon className={iconClassName('compact', 'dark:hidden')} />
            <SunIcon className={iconClassName('compact', 'hidden dark:inline')} />
          </>
        )}
      >
        <span className="dark:hidden">Use dark theme</span>
        <span className="hidden dark:inline">Use light theme</span>
      </Button>
      <Button type="submit" form={logoutFormId} variant="ghost" className={mobileButton}>Log out</Button>
    </section>
  )
}
