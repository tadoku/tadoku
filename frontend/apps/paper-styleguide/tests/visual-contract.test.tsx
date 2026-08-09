import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { catalogRegistry } from 'paper-ui/catalog'
import { DocsShell } from '../src/app/DocsShell'
import { CatalogIndex, ResolvedCatalogueRoute } from '../src/app/routes'
import { DocumentPage } from '../src/documentation/DocumentPage'
import '../src/styles/shell.css'
import { styleguideStyles } from './style-sources'

describe('responsive visual contract', () => {
  it('reserves inline space between the home hero accent rail and its copy', () => {
    render(
      <MemoryRouter>
        <CatalogIndex />
      </MemoryRouter>,
    )

    const hero = screen
      .getByRole('heading', { name: 'Paper makes the interface legible.' })
      .closest('header')

    expect(hero).not.toBeNull()
    expect(getComputedStyle(hero!).paddingInlineStart).toBe('1rem')
  })

  it('reserves the same rail inset on document heroes', () => {
    const document = catalogRegistry.documents.find(
      (candidate) => candidate.id === 'component.button',
    )

    expect(document).toBeDefined()
    render(<DocumentPage document={document!} />)

    const hero = screen.getByRole('heading', { name: 'Button' }).closest('header')

    expect(hero).not.toBeNull()
    expect(getComputedStyle(hero!).paddingInlineStart).toBe('1rem')
  })

  it('reserves the same rail inset on the not-found state', () => {
    render(
      <MemoryRouter initialEntries={['/outside-the-catalogue']}>
        <ResolvedCatalogueRoute />
      </MemoryRouter>,
    )

    const notFound = screen
      .getByRole('heading', { name: 'This Paper page does not exist.' })
      .closest('section')

    expect(notFound).not.toBeNull()
    expect(getComputedStyle(notFound!).paddingInlineStart).toBe('1rem')
  })

  it('uses the approved Paper icon grammar in the responsive header', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>
          <p>Page content</p>
        </DocsShell>
      </MemoryRouter>,
    )

    const search = screen.getByRole('button', { name: 'Search Paper' })
    const browse = screen.getByRole('button', { name: 'Browse' })
    const wordmark = screen.getByRole('link', { name: 'Tadoku' })
    const brandMarks = wordmark.querySelectorAll('img')

    expect(search.querySelector('svg')).not.toBeNull()
    expect(browse.querySelector('svg')).not.toBeNull()
    expect(browse).toHaveClass('paper-button--ghost')
    expect(browse.querySelector('.mobile-nav-trigger__label')).not.toBeNull()
    expect(brandMarks).toHaveLength(2)
    expect(brandMarks[0]).toHaveAttribute(
      'src',
      expect.stringContaining('wordmark-accent.svg'),
    )
    expect(brandMarks[1]).toHaveAttribute(
      'src',
      expect.stringContaining('wordmark-reversed.svg'),
    )
    for (const mark of brandMarks) {
      expect(mark).toHaveAttribute('width', '126.4')
      expect(mark).toHaveAttribute('height', '23.2')
      expect(mark).toHaveAttribute('aria-hidden', 'true')
    }

    await user.click(browse)
    expect(
      screen.getByRole('button', { name: 'Close navigation' }).querySelector('svg'),
    ).not.toBeNull()
  })

  it('renders the legacy wordmark at the original styleguide scale', () => {
    expect(styleguideStyles).toMatch(
      /\.docs-wordmark\s*\{(?=[^}]*inline-size:\s*126\.4px;)(?=[^}]*block-size:\s*23\.2px;)[^}]*\}/su,
    )
    expect(styleguideStyles).not.toMatch(/\.docs-wordmark\s*\{[^}]*transform:/su)
  })

  it('leaves the search trigger visual recipe to the public Paper Button', () => {
    const searchSkinRules = [
      ...styleguideStyles.matchAll(
        /([^{}]*\.shell-search-trigger[^{}]*)\{([^}]*)\}/gu,
      ),
    ]
      .map((match) => ({
        selector: match[1].trim(),
        declarations: match[2].trim().replace(/\s+/gu, ' '),
      }))
      .filter(({ declarations }) => declarations !== 'display: none;')

    expect(searchSkinRules).toEqual([])
  })

  it('uses the public Paper Navbar and Sidebar recipes for the application shell', () => {
    render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    const headerNavigation = screen.getByRole('navigation', {
      name: 'Main navigation',
    })
    const catalogueNavigation = screen.getByRole('navigation', {
      name: 'Paper catalogue',
    })

    expect(headerNavigation).toHaveClass('paper-navbar')
    expect(headerNavigation.closest('header')).toHaveClass('docs-navbar')
    expect(headerNavigation.querySelector('.paper-navbar__actions')).not.toBeNull()
    expect(headerNavigation).toContainElement(
      screen.getByRole('button', { name: 'Search Paper' }),
    )
    expect(headerNavigation).toContainElement(
      screen.getByRole('button', { name: 'Browse' }),
    )
    expect(screen.queryByRole('button', { name: 'Open menu' })).not.toBeInTheDocument()
    expect(document.querySelector('.docs-header')).toBeNull()
    expect(catalogueNavigation).toHaveClass('paper-sidebar')
  })

  it('keeps the desktop catalogue on a light Paper surface with a subtle end rule', () => {
    expect(styleguideStyles).toMatch(
      /\.docs-sidebar\s*\{[^}]*border-inline-end:\s*var\(--paper-border-static-width\) solid var\(--paper-color-rule-subtle\);[^}]*background:\s*var\(--paper-color-surface-paper\);/su,
    )
  })

  it('uses the compact Paper density for the desktop catalogue', () => {
    render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    const catalogueNavigation = screen.getByRole('navigation', {
      name: 'Paper catalogue',
    })

    expect(catalogueNavigation.closest('.docs-sidebar')).toHaveAttribute(
      'data-density',
      'compact',
    )
  })

  it('groups the catalogue index by document kind', () => {
    render(
      <MemoryRouter>
        <CatalogIndex />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { name: 'Foundations' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Components' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Governance' })).toBeInTheDocument()
  })

  it('leaves native element and form-control foundations to Paper', () => {
    expect(styleguideStyles).not.toMatch(
      /\.docs-shell\s+:where\([^)]*(?:button|input|select)/u,
    )
    expect(styleguideStyles).not.toMatch(
      /(?:^|[\s>,])(?:button|input|select)(?:\b|[\s:[.#>])/mu,
    )
  })

  it('only hides search metadata responsively without reskinning the Button', () => {
    const compactStyles = styleguideStyles.slice(
      styleguideStyles.indexOf('@media (max-width: 48rem)'),
      styleguideStyles.indexOf('@media (max-width: 23rem)'),
    )

    expect(compactStyles).toMatch(
      /\.shell-search-trigger__label,\s*\.shell-search-trigger__shortcut\s*\{\s*display:\s*none;\s*\}/su,
    )
    expect(compactStyles).not.toMatch(/\.shell-search-trigger\s*\{/u)
  })

  it('collapses the catalogue sidebar into Browse at mid-size widths', () => {
    const midSizeStyles = styleguideStyles.slice(
      styleguideStyles.indexOf('@media (max-width: 64rem)'),
      styleguideStyles.indexOf('@media (max-width: 48rem)'),
    )

    expect(midSizeStyles).toContain(
      '.docs-shell {\n      display: block;\n    }',
    )
    expect(midSizeStyles).toContain(
      '.docs-sidebar {\n      display: none;\n    }',
    )
    expect(midSizeStyles).toContain(
      '.mobile-nav-trigger {\n      display: inline-flex;\n    }',
    )
  })

  it('delegates the Browse overlay and focus return to Paper Drawer', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    const browse = screen.getByRole('button', { name: 'Browse' })
    await user.click(browse)
    expect(screen.getByRole('dialog', { name: 'Browse Paper' })).toHaveClass(
      'paper-drawer',
    )
    expect(document.querySelector('.paper-drawer__backdrop')).not.toBeNull()

    await user.click(screen.getByRole('button', { name: 'Close navigation' }))
    expect(screen.queryByRole('dialog', { name: 'Browse Paper' })).not.toBeInTheDocument()
    expect(browse).toHaveFocus()
  })

  it('uses Paper overlay scrims for search and Browse', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    await user.click(screen.getByRole('button', { name: 'Browse' }))
    expect(document.querySelector('.paper-drawer__backdrop')).not.toBeNull()
    await user.click(screen.getByRole('button', { name: 'Close navigation' }))

    await user.click(screen.getByRole('button', { name: 'Search Paper' }))
    expect(document.querySelector('.paper-modal__backdrop')).not.toBeNull()
  })

  it('keeps the Browse header fixed while Paper Drawer owns the scrolling body', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    await user.click(screen.getByRole('button', { name: 'Browse' }))
    const drawer = screen.getByRole('dialog', { name: 'Browse Paper' })
    const scrollBody = drawer.querySelector('.paper-drawer__body')

    expect(drawer).toHaveClass('paper-elevation-showcase')
    expect(scrollBody?.querySelector('.paper-sidebar')).not.toBeNull()
    expect(styleguideStyles).not.toMatch(/scrollbar-width|::-webkit-scrollbar/)
  })

  it('switches the wordmark to its exact legacy reversed asset in dark mode', () => {
    expect(styleguideStyles).toContain(
      ":root[data-theme='dark'] .docs-wordmark__image--accent",
    )
    expect(styleguideStyles).toContain(
      ":root[data-theme='dark'] .docs-wordmark__image--reversed",
    )
    expect(styleguideStyles).not.toContain('content: url(')
  })
})
