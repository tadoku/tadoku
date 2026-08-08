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
    const wordmark = screen.getByRole('link', { name: 'Tadoku Paper' })
    const brandMarks = wordmark.querySelectorAll('img')

    expect(search.querySelector('svg')).not.toBeNull()
    expect(browse.querySelector('svg')).not.toBeNull()
    expect(browse.querySelector('.mobile-nav-trigger__label')).not.toBeNull()
    expect(brandMarks).toHaveLength(2)
    expect(styleguideStyles).toContain(
      'inline-size: 2.25rem;\n    block-size: 2.25rem;\n    aspect-ratio: 1;',
    )
    expect([...brandMarks].map((mark) => mark.getAttribute('src'))).toEqual(
      expect.arrayContaining([
        expect.stringContaining('cut-meter.svg'),
        expect.stringContaining('cut-meter-reversed.svg'),
      ]),
    )
    expect(getComputedStyle(wordmark).whiteSpace).toBe('nowrap')

    await user.click(browse)
    expect(
      screen.getByRole('button', { name: 'Close navigation' }).querySelector('svg'),
    ).not.toBeNull()
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

  it('keeps shell element resets out of isolated component previews', () => {
    expect(styleguideStyles).not.toContain('\n  button,\n  input,\n  select {')
    expect(styleguideStyles).not.toContain('\n  button,\n  select {')
    expect(styleguideStyles).not.toContain('\n  button {\n')
    expect(styleguideStyles).not.toContain('\n  code {\n')
    expect(styleguideStyles).toContain('.docs-shell button')
  })

  it('keeps scoped shell resets low-specificity so selected controls retain contrast', () => {
    expect(styleguideStyles).not.toContain('\n  .docs-shell button,\n')
    expect(styleguideStyles).toContain('.docs-shell :where(button, input, select)')
  })

  it('collapses both header actions to icons at the narrow phone floor', () => {
    expect(styleguideStyles).toContain(
      '.shell-search-trigger__label,\n    .mobile-nav-trigger__label,',
    )
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

  it('animates the Browse drawer while preserving reduced-motion safety', async () => {
    const user = userEvent.setup()
    const { container } = render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    const backdrop = container.querySelector('.mobile-nav-backdrop')
    const drawer = container.querySelector<HTMLElement>('.mobile-nav-drawer')

    expect(backdrop).toHaveAttribute('data-open', 'false')
    expect(backdrop).toHaveAttribute('aria-hidden', 'true')
    expect(drawer?.inert).toBe(true)

    await user.click(screen.getByRole('button', { name: 'Browse' }))
    expect(backdrop).toHaveAttribute('data-open', 'true')
    expect(backdrop).toHaveAttribute('aria-hidden', 'false')
    expect(drawer?.inert).toBe(false)

    await user.click(screen.getByRole('button', { name: 'Close navigation' }))
    expect(backdrop).toHaveAttribute('data-open', 'false')
    expect(drawer?.inert).toBe(true)

    expect(styleguideStyles).toContain(
      '@media (prefers-reduced-motion: no-preference) {\n    .mobile-nav-backdrop {\n      transition: opacity var(--paper-motion-standard)',
    )
    expect(styleguideStyles).toContain(
      '.mobile-nav-backdrop[data-open=\'false\'] .mobile-nav-drawer',
    )
  })

  it('uses the translucent semantic scrim for full-screen shell overlays', () => {
    expect(styleguideStyles).toMatch(
      /\.search-backdrop,\n {2}\.mobile-nav-backdrop \{[^}]*background: var\(--paper-color-surface-scrim\);[^}]*-webkit-backdrop-filter: blur\(0\.375rem\);[^}]*backdrop-filter: blur\(0\.375rem\);/s,
    )
    expect(styleguideStyles).toMatch(
      /@media \(forced-colors: active\) \{[^}]*\.search-backdrop,\n {4}\.mobile-nav-backdrop \{[^}]*backdrop-filter: none;/s,
    )
  })

  it('keeps the Browse header fixed while only its native-scroll navigation body moves', () => {
    const { container } = render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    const drawer = container.querySelector('.mobile-nav-drawer')
    const scrollBody = container.querySelector('.mobile-nav-drawer__body')

    expect(drawer).toHaveClass('paper-elevation-showcase')
    expect(scrollBody?.querySelector('.catalogue-nav')).not.toBeNull()
    expect(styleguideStyles).toMatch(
      /\.mobile-nav-drawer\s*\{[^}]*display:\s*flex;[^}]*flex-direction:\s*column;[^}]*overflow:\s*hidden;/s,
    )
    expect(styleguideStyles).toMatch(
      /\.mobile-nav-drawer__body\s*\{[^}]*min-block-size:\s*0;[^}]*overflow-y:\s*auto;/s,
    )
    expect(styleguideStyles).not.toMatch(/scrollbar-width|::-webkit-scrollbar/)
  })

  it('keeps the inactive Cut Meter asset out of the wordmark layout', () => {
    expect(styleguideStyles).toContain('.paper-wordmark :where(img) {')
    expect(styleguideStyles).not.toContain('.paper-wordmark img {')
  })

  it('switches the Cut Meter to its canonical reversed asset in dark mode', () => {
    expect(styleguideStyles).toContain(
      ":root[data-theme='dark'] .paper-wordmark__mark--light",
    )
    expect(styleguideStyles).toContain(
      ":root[data-theme='dark'] .paper-wordmark__mark--dark",
    )
  })
})
