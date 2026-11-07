import * as stylex from '@stylexjs/stylex'
import { rootRouteId, useHydrated, useLoaderData, useNavigate, useRouterState } from '@tanstack/react-router'
import {
  ActionMenu,
  Button,
  EmptyState,
  Navbar,
  SiteFooter,
  ThemeVariant,
  applyTheme,
  buttonStyles,
  iconSizes,
  linkStyles,
  navbarMobileLinkStyles,
  siteFooterCreditLink,
  type NavbarItem,
} from 'paper-ui'
import { border, breakpoints, color, density, fontSize, fontWeight, measure, opacity, page, ratio, size, space, zIndex } from 'paper-ui/tokens.stylex'
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
} from 'paper-ui/icons'
import { DiscordIcon, GitHubIcon, TwitterIcon } from './BrandIcons'
import { useAccountUrls, viewerQueryOptions, type Viewer } from '../lib/session'
import { useLinks } from '../lib/links'
import { Announcements } from './Announcements'
import { renderLink } from './router-link'

/** The stored theme, shared with authv2 and Webv2. */
export const themeKey = 'tadoku-theme'

type Theme = 'light' | 'dark'

const currentTheme = (): Theme => document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light'

function setTheme(next: Theme) {
  applyTheme(next)
  try { localStorage.setItem(themeKey, next) } catch { /* Private mode keeps the choice for this page only. */ }
}

const toggleTheme = () => setTheme(currentTheme() === 'dark' ? 'light' : 'dark')

/** The account menu's appearance item. The menu renders only after hydration, so it reads the theme directly. */
function useThemeSwitch() {
  const [theme, setThemeState] = useState<Theme>(currentTheme)
  const next: Theme = theme === 'dark' ? 'light' : 'dark'
  return {
    label: next === 'dark' ? 'Use dark theme' : 'Use light theme',
    icon: next === 'dark' ? <MoonIcon {...stylex.props(iconSizes.compact)} /> : <SunIcon {...stylex.props(iconSizes.compact)} />,
    toggle: () => { setTheme(next); setThemeState(next) },
  }
}

const discordUrl = 'https://discord.gg/AsC9vZs2Ex'

const styles = stylex.create({
  shell: { display: 'flex', flexDirection: 'column', minHeight: '100vh' },
  skip: {
    insetBlockStart: space[12],
    insetInlineStart: space[12],
    position: 'fixed',
    transform: { default: `translateY(${ratio.offscreen})`, ':focus': 'none' },
    zIndex: zIndex.modal,
  },
  header: {
    backgroundColor: color.surfacePaper,
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: 'solid',
    borderBlockEndWidth: border.static,
  },
  navbar: {
    marginInline: 'auto',
    paddingInline: page.gutter,
    borderBlockEndWidth: 0,
    boxSizing: 'border-box',
    maxInlineSize: `calc(${page.contentWidth} + ${page.gutter} * 2)`,
  },
  navbarInner: { gap: { default: density.inlineGap, [breakpoints.md]: space[8] }, paddingBlock: space[8], paddingInline: 0 },
  brand: { flexShrink: { default: null, [breakpoints.md]: 0 } },
  brandLink: { gap: `calc(${page.brandWidth} * 0.056)` },
  actions: { marginInlineStart: { default: null, [breakpoints.md]: 'auto' } },
  mobilePanel: { paddingInline: 0 },
  mark: { blockSize: 'auto', display: 'block', inlineSize: `calc(${page.brandWidth} * 0.282)` },
  wordmark: { blockSize: 'auto', display: 'block', inlineSize: page.brandWidth },
  content: {
    flex: '1',
    marginInline: 'auto',
    paddingInline: page.gutter,
    boxSizing: 'border-box',
    inlineSize: '100%',
    maxInlineSize: `calc(${page.contentWidth} + ${page.gutter} * 2)`,
    minInlineSize: 0,
    outlineStyle: { default: null, ':focus': 'none' },
    paddingBlockEnd: { default: space[64], [breakpoints.tablet]: space[48], [breakpoints.phone]: space[40] },
    paddingBlockStart: { default: space[48], [breakpoints.tablet]: space[32], [breakpoints.phone]: space[28] },
  },
  accountActions: {
    gap: { default: space[12], [breakpoints.sm]: space[6] },
    alignItems: 'center',
    display: 'flex',
    flexWrap: 'wrap',
    fontSize: { default: null, [breakpoints.sm]: fontSize[13] },
    justifyContent: 'flex-end',
    marginInlineStart: space[16],
    maxInlineSize: { default: null, [breakpoints.sm]: size[192] },
  },
  loginLink: {
    color: color.textInk,
    fontWeight: fontWeight.semibold,
    textDecorationLine: { default: 'none', ':focus-visible': 'underline', ':hover': 'underline' },
  },
  headerAction: { flexShrink: 0, whiteSpace: 'nowrap' },
  desktopMenu: { display: { default: null, [breakpoints.md]: 'none' } },
  accountName: { overflow: 'hidden', maxInlineSize: measure.title, textOverflow: 'ellipsis', whiteSpace: 'nowrap', },
  /** The trigger waits only for hydration; it keeps its colours instead of looking unavailable. */
  awaitingHydration: { cursor: { default: null, ':disabled': 'progress' }, opacity: { default: null, ':disabled': opacity.visible } },
  mobileAccount: {
    borderBlockStartColor: color.ruleSubtle,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
    marginBlockStart: space[12],
    paddingBlockStart: space[12],
  },
  mobileAccountName: {
    marginInline: 0,
    paddingInline: density.fieldPaddingInline,
    color: color.textMuted,
    fontSize: fontSize[14],
    fontWeight: fontWeight.semibold,
    marginBlockEnd: space[4],
    marginBlockStart: 0,
    overflowWrap: 'anywhere',
  },
  mobileButton: { justifyContent: 'flex-start', minBlockSize: size[44], width: '100%' },
  /** The appearance switch is a row in the phone menu's account section, set apart from the account's pages. */
  themeRow: {
    borderBlockStartColor: color.ruleSubtle,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
    marginBlockStart: space[12],
  },
})
const logoutFormId = 'app-logout'

export function AppShell({ children }: { children: ReactNode }) {
  const currentPath = useRouterState({ select: state => state.location.pathname })
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
    <div {...stylex.props(styles.shell)}>
      <a {...stylex.props(...buttonStyles(), styles.skip)} href="#main-content">Skip to content</a>
      <header {...stylex.props(styles.header)}>
        <Navbar
          brand={<>
            <ThemeVariant
              light={<img {...stylex.props(styles.mark)} src={mark} width="40" height="40" alt="" />}
              dark={<img {...stylex.props(styles.mark)} src={reversedMark} width="40" height="40" alt="" />}
            />
            <ThemeVariant
              light={<img {...stylex.props(styles.wordmark)} src={wordmark} width="158" height="29" alt="Tadoku" />}
              dark={<img {...stylex.props(styles.wordmark)} src={reversedWordmark} width="158" height="29" alt="Tadoku" />}
            />
          </>}
          brandHref="/"
          navigation={navigation}
          currentPath={currentPath}
          renderLink={renderLink}
          actions={<div {...stylex.props(styles.accountActions)}><AccountActions viewer={viewer} /></div>}
          mobileFooter={viewer?.status === 'signed-in' ? closeMenu => <MobileAccount viewer={viewer} closeMenu={closeMenu} /> : undefined}
          xstyle={styles.navbar}
          innerXstyle={styles.navbarInner}
          brandXstyle={styles.brand}
          brandLinkXstyle={styles.brandLink}
          actionsXstyle={styles.actions}
          mobilePanelXstyle={styles.mobilePanel}
        />
        {viewer && 'userId' in viewer ? <form id={logoutFormId} method="post" action="/logout" hidden /> : null}
      </header>
      <main id="main-content" {...stylex.props(styles.content)} tabIndex={-1}>
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
        credit={<>Built &amp; designed by <a href="https://antonve.be" {...stylex.props(siteFooterCreditLink)}>antonve</a></>}
        social={[
          { label: 'Discord', href: discordUrl, icon: <DiscordIcon /> },
          { label: 'GitHub', href: 'https://github.com/tadoku', icon: <GitHubIcon /> },
          { label: 'Twitter', href: 'https://twitter.com/tadoku_app', icon: <TwitterIcon /> },
        ]}
        groups={[
          {
            title: 'Get started',
            links: [
              { id: 'home', label: 'Home', href: links.home },
              { id: 'leaderboard', label: 'Leaderboard', href: links.leaderboard },
              { id: 'contests', label: 'Contests', href: links.contests },
              { id: 'blog', label: 'Blog', href: links.blog },
            ],
          },
          {
            title: 'Learn',
            links: [
              { id: 'about', label: 'About Tadoku', href: links.page('about') },
              { id: 'guide', label: 'Guide', href: links.guide },
              { id: 'rules', label: 'Rules', href: links.page('rules') },
              { id: 'faq', label: 'Questions & help', href: links.page('faq') },
            ],
          },
          {
            title: 'Legal',
            links: [
              { id: 'privacy', label: 'Privacy', href: links.page('privacy') },
              { id: 'contact', label: 'Contact', href: links.page('contact') },
            ],
          },
        ]}
        renderLink={renderLink}
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
        <a {...stylex.props(linkStyles.base)} href={discordUrl}>Contact us on Discord</a>
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
    { id: 'profile', label: 'Profile', href: links.profile(viewer.userId), icon: <UserIcon {...stylex.props(iconSizes.compact)} /> },
    ...(viewer.role === 'admin'
      ? [{ id: 'admin', label: 'Admin', href: adminUrl, icon: <ShieldCheckIcon {...stylex.props(iconSizes.compact)} /> }]
      : []),
    { id: 'settings', label: 'Account settings', href: settings, icon: <Cog6ToothIcon {...stylex.props(iconSizes.compact)} /> },
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
        <a href={login} {...stylex.props(linkStyles.plain, styles.loginLink)}>Log in</a>
        <a href={register} {...stylex.props(...buttonStyles({ variant: 'outline' }))}>Sign up</a>
      </>
    case 'signed-in':
      return <>
        <AccountMenu viewer={viewer} />
        <a href={links.logs} {...stylex.props(...buttonStyles(), styles.headerAction)}>Log activity</a>
      </>
  }
}

function AccountMenu({ viewer }: { viewer: SignedIn }) {
  const hydrated = useHydrated()
  return (
    <span {...stylex.props(styles.desktopMenu)}>
      {hydrated
        ? <AccountMenuOpen viewer={viewer} />
        : (
          // The menu's trigger as server HTML renders it, disabled until it can open.
          <button type="button" {...stylex.props(...buttonStyles({ variant: 'ghost' }), styles.awaitingHydration)} disabled>
            <span {...stylex.props(styles.accountName)}>{viewer.displayName}</span>
            <ChevronDownIcon {...stylex.props(iconSizes.compact)} aria-hidden="true" />
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
      triggerLabelXstyle={styles.accountName}
      items={[
        ...items,
        { id: 'theme', label: theme.label, icon: theme.icon, onSelect: theme.toggle },
        { id: 'logout', label: 'Log out', icon: <ArrowRightStartOnRectangleIcon {...stylex.props(iconSizes.compact)} />, onSelect: logOut },
      ]}
    />
  )
}

function MobileAccount({ viewer, closeMenu }: { viewer: SignedIn; closeMenu: () => void }) {
  const links = useAccountLinks(viewer)
  return (
    <section {...stylex.props(styles.mobileAccount)} aria-label={`${viewer.displayName} account`}>
      <p {...stylex.props(styles.mobileAccountName)}>{viewer.displayName}</p>
      {links.map(link => (
        <Fragment key={link.id}>
          {renderLink({ href: link.href, ...stylex.props(...navbarMobileLinkStyles), onClick: closeMenu, children: link.label })}
        </Fragment>
      ))}
      {/* Server-rendered, so the theme variant shows the label for the current theme without reading it. */}
      <Button variant="ghost" xstyle={[navbarMobileLinkStyles, styles.mobileButton, styles.themeRow]} onClick={toggleTheme}>
        <ThemeVariant
          light={<><MoonIcon {...stylex.props(iconSizes.compact)} aria-hidden="true" /><span>Use dark theme</span></>}
          dark={<><SunIcon {...stylex.props(iconSizes.compact)} aria-hidden="true" /><span>Use light theme</span></>}
        />
      </Button>
      <Button type="submit" form={logoutFormId} variant="ghost" xstyle={[navbarMobileLinkStyles, styles.mobileButton]}>Log out</Button>
    </section>
  )
}
