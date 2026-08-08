import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { catalogRegistry } from 'paper-ui/catalog'
import { DocsShell } from '../src/app/DocsShell'
import { CatalogIndex, ResolvedCatalogueRoute } from '../src/app/routes'
import { DocumentPage } from '../src/documentation/DocumentPage'
import '../src/styles/shell.css'

const testDirectory = dirname(fileURLToPath(import.meta.url))
const shellStyles = readFileSync(
  resolve(testDirectory, '../src/styles/shell.css'),
  'utf8',
)

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
    expect(shellStyles).toContain(
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
    expect(shellStyles).not.toContain('\n  button,\n  input,\n  select {')
    expect(shellStyles).not.toContain('\n  button,\n  select {')
    expect(shellStyles).not.toContain('\n  button {\n')
    expect(shellStyles).not.toContain('\n  code {\n')
    expect(shellStyles).toContain('.docs-shell button')
  })

  it('keeps scoped shell resets low-specificity so selected controls retain contrast', () => {
    expect(shellStyles).not.toContain('\n  .docs-shell button,\n')
    expect(shellStyles).toContain('.docs-shell :where(button, input, select)')
  })

  it('collapses both header actions to icons at the narrow phone floor', () => {
    expect(shellStyles).toContain(
      '.shell-search-trigger__label,\n    .mobile-nav-trigger__label,',
    )
  })

  it('switches the Cut Meter to its canonical reversed asset in dark mode', () => {
    expect(shellStyles).toContain(
      ":root[data-theme='dark'] .paper-wordmark__mark--light",
    )
    expect(shellStyles).toContain(
      ":root[data-theme='dark'] .paper-wordmark__mark--dark",
    )
  })
})
