import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { catalogRegistry } from 'paper-ui/catalog'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { CatalogueSearch } from '../src/app/CatalogueSearch'
import { DocsShell } from '../src/app/DocsShell'
import { CatalogIndex, ResolvedCatalogueRoute } from '../src/app/routes'
import { DocumentPage } from '../src/documentation/DocumentPage'
import { ComponentWorkbench } from '../src/documentation/ComponentWorkbench'
import { ExampleCanvas } from '../src/documentation/ExampleCanvas'

const buttonDocument = catalogRegistry.documents.find(
  (document) => document.id === 'component.button',
)!
const buttonFixture = catalogRegistry.fixtures.find((fixture) =>
  buttonDocument.fixtureIds.includes(fixture.id),
)!
const loggingPatternDocument = catalogRegistry.documents.find(
  (document) => document.id === 'pattern.logging',
)!

const originalScrollIntoView = HTMLElement.prototype.scrollIntoView

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
  if (originalScrollIntoView) {
    HTMLElement.prototype.scrollIntoView = originalScrollIntoView
  } else {
    Reflect.deleteProperty(HTMLElement.prototype, 'scrollIntoView')
  }
  Reflect.deleteProperty(navigator, 'clipboard')
})

describe('catalogue experience stability', () => {
  it('copies the registered example and announces successful feedback', async () => {
    const user = userEvent.setup()
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText },
    })
    render(<DocumentPage document={buttonDocument} />)

    await user.click(screen.getByRole('tab', { name: 'Code' }))
    await user.click(screen.getByRole('button', { name: 'Copy code' }))

    expect(writeText).toHaveBeenCalledWith(buttonFixture.code)
    expect(screen.getByRole('status')).toHaveTextContent('Code copied')
  })

  it('keeps all workbench panels stable while switching views', async () => {
    const user = userEvent.setup()
    render(<DocumentPage document={buttonDocument} />)

    const panels = screen.getAllByRole('tabpanel', { hidden: true })
    expect(panels).toHaveLength(4)
    const previewFrame = screen.getByTitle('Paper responsive component preview')
    const workbench = screen.getByRole('region', {
      name: `${buttonDocument.name} examples`,
    })
    expect(workbench).toHaveClass('paper-surface-card')
    expect(screen.getByRole('tablist', { name: 'Example views' })).toHaveClass(
      'paper-tabs__list',
    )

    await user.click(screen.getByRole('tab', { name: 'Code' }))
    await user.click(screen.getByRole('tab', { name: 'Preview' }))

    expect(screen.getByTitle('Paper responsive component preview')).toBe(
      previewFrame,
    )
  })

  it('uses Paper controls and resets a removed fixture selection', async () => {
    const user = userEvent.setup()
    const fixtures = catalogRegistry.fixtures.filter((fixture) =>
      buttonDocument.fixtureIds.includes(fixture.id),
    )
    expect(fixtures.length).toBeGreaterThan(1)

    const { rerender } = render(
      <ComponentWorkbench document={buttonDocument} fixtures={fixtures} />,
    )
    const fixtureSelect = screen.getByLabelText('Fixture')
    expect(fixtureSelect).toHaveClass('paper-select')

    await user.selectOptions(fixtureSelect, fixtures[1].id)
    expect(screen.getByTitle('Paper responsive component preview')).toHaveAttribute(
      'data-fixture-id',
      fixtures[1].id,
    )

    rerender(
      <ComponentWorkbench document={buttonDocument} fixtures={[fixtures[0]]} />,
    )
    rerender(
      <ComponentWorkbench document={buttonDocument} fixtures={fixtures} />,
    )

    expect(screen.getByTitle('Paper responsive component preview')).toHaveAttribute(
      'data-fixture-id',
      fixtures[0].id,
    )

    await user.click(screen.getByRole('tab', { name: 'Code' }))
    expect(screen.getByRole('button', { name: 'Copy code' })).toHaveClass(
      'paper-button',
      'paper-button--outline',
    )
  })

  it('exposes exact iframe content widths for every viewport control', async () => {
    const user = userEvent.setup()
    render(<ExampleCanvas />)

    await user.click(screen.getByRole('button', { name: 'Tablet, 768 pixels' }))
    const frame = screen.getByTitle('Paper responsive component preview')
    expect(frame).toHaveAttribute('width', '768')
    expect(frame).toHaveStyle({ inlineSize: '768px', boxSizing: 'content-box' })
  })

  it('moves from search input to the first result with ArrowDown', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <CatalogueSearch documents={catalogRegistry.documents} />
      </MemoryRouter>,
    )

    await user.keyboard('/')
    await user.type(screen.getByRole('searchbox'), 'button')
    await user.keyboard('{ArrowDown}')

    expect(
      within(screen.getByRole('dialog')).getAllByRole('link')[0],
    ).toHaveFocus()
  })

  it('focuses mobile navigation and preserves its accessible close path', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    await user.click(screen.getByRole('button', { name: 'Browse' }))
    expect(screen.getByRole('button', { name: 'Close navigation' })).toHaveFocus()
  })

  it('closes mobile navigation when the viewport crosses into desktop layout', async () => {
    const user = userEvent.setup()
    const listeners = new Set<(event: MediaQueryListEvent) => void>()
    let collapsed = true
    const desktopBoundary = {
      get matches() {
        return collapsed
      },
      media: '(max-width: 64rem)',
      onchange: null,
      addEventListener: vi.fn(
        (_: string, listener: EventListenerOrEventListenerObject) => {
          listeners.add(listener as (event: MediaQueryListEvent) => void)
        },
      ),
      removeEventListener: vi.fn(
        (_: string, listener: EventListenerOrEventListenerObject) => {
          listeners.delete(listener as (event: MediaQueryListEvent) => void)
        },
      ),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    } as MediaQueryList
    vi.stubGlobal('matchMedia', vi.fn(() => desktopBoundary))

    const { container } = render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    const browse = screen.getByRole('button', { name: 'Browse' })
    const backdrop = container.querySelector('.mobile-nav-backdrop')
    const drawer = container.querySelector<HTMLElement>('.mobile-nav-drawer')
    await user.click(browse)
    expect(backdrop).toHaveAttribute('data-open', 'true')

    collapsed = false
    listeners.forEach((listener) =>
      listener({ matches: false } as MediaQueryListEvent),
    )

    await waitFor(() => expect(browse).toHaveAttribute('aria-expanded', 'false'))
    expect(backdrop).toHaveAttribute('data-open', 'false')
    expect(backdrop).toHaveAttribute('aria-hidden', 'true')
    expect(drawer?.inert).toBe(true)
  })

  it('keeps display preferences out of catalogue navigation', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    expect(screen.queryByLabelText('Display preferences')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Browse' }))
    expect(screen.queryByLabelText('Display preferences')).not.toBeInTheDocument()
  })

  it('keeps lifecycle status out of desktop and mobile catalogue navigation', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    const desktopNavigation = screen.getByRole('navigation', {
      name: 'Paper catalogue',
    })
    expect(within(desktopNavigation).queryByText('Stable')).not.toBeInTheDocument()
    expect(
      within(desktopNavigation).queryByText('Experimental'),
    ).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Browse' }))
    const navigations = screen.getAllByRole('navigation', {
      name: 'Paper catalogue',
    })
    expect(navigations).toHaveLength(2)
    expect(within(navigations[1]).queryByText('Stable')).not.toBeInTheDocument()
    expect(
      within(navigations[1]).queryByText('Experimental'),
    ).not.toBeInTheDocument()
  })

  it('scrolls direct hash routes to their documented section', async () => {
    const scrollIntoView = vi.fn()
    HTMLElement.prototype.scrollIntoView = scrollIntoView

    render(
      <MemoryRouter initialEntries={[`${buttonDocument.route}#accessibility`]}>
        <Routes>
          <Route path="*" element={<ResolvedCatalogueRoute />} />
        </Routes>
      </MemoryRouter>,
    )

    await waitFor(() => expect(scrollIntoView).toHaveBeenCalled())
    expect(document.getElementById('accessibility')).not.toBeNull()
  })

  it('keeps local outline anchors and design history discoverable', () => {
    const { unmount } = render(<DocumentPage document={buttonDocument} />)
    expect(
      screen.getByRole('link', { name: 'Accessibility' }),
    ).toHaveAttribute('href', '#accessibility')
    unmount()

    render(
      <MemoryRouter>
        <CatalogIndex />
      </MemoryRouter>,
    )
    const history = screen.getByRole('heading', { name: 'Design history' })
      .closest('section')
    expect(history).not.toBeNull()
    expect(within(history!).getAllByRole('link')).toHaveLength(5)
    expect(
      within(history!).getByRole('link', { name: /Original refinement audit/u }),
    ).toBeInTheDocument()
    expect(
      within(history!).getByRole('link', { name: /Visual studies/u }),
    ).toBeInTheDocument()
  })

  it('indexes product patterns and experiments as distinct catalogue sections', () => {
    render(
      <MemoryRouter>
        <CatalogIndex />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { name: 'Patterns' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Experiments' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /LoggingStable/u })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Logging v2Experimental/u })).toBeInTheDocument()
  })

  it('renders a registered fixture for a product-pattern document', () => {
    render(<DocumentPage document={loggingPatternDocument} />)

    expect(screen.getByTitle('Paper responsive component preview')).toHaveAttribute(
      'data-fixture-id',
      'pattern.logging-summary',
    )
  })

  it('keeps component registry metadata out of the user-facing guide', () => {
    render(<DocumentPage document={buttonDocument} />)

    expect(screen.queryByRole('link', { name: buttonDocument.sourcePath })).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Metadata' })).not.toBeInTheDocument()
  })
})
