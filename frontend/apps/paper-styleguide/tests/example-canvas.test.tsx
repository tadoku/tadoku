import '@testing-library/jest-dom/vitest'
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { catalogRegistry } from 'paper-ui/catalog'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ExampleCanvas } from '../src/documentation/ExampleCanvas'

afterEach(() => vi.unstubAllGlobals())

describe('ExampleCanvas', () => {
  it('keeps example links in the preview while running their selection callbacks', () => {
    const onSelect = vi.fn()
    render(<ExampleCanvas fixture={{
      ...catalogRegistry.fixtures[0],
      render: () => <a href="/reading/logs" onClick={onSelect}>Reading logs</a>,
    }} />)
    const frame = screen.getByTitle<HTMLIFrameElement>('Paper responsive component preview')
    frame.contentDocument!.body.innerHTML = '<div id="paper-preview-root"></div>'
    fireEvent.load(frame)
    const link = within(frame.contentDocument!.body).getByRole('link', { name: 'Reading logs' })

    expect(fireEvent.click(link)).toBe(false)
    expect(onSelect).toHaveBeenCalledOnce()
    expect(screen.getByRole('status')).toHaveTextContent('Preview destination: /reading/logs')
    expect(link).toHaveAttribute('href', '/reading/logs')
  })

  it('starts each fixture with its own form and result state while preserving preview preferences', async () => {
    const user = userEvent.setup()
    const fixtures = ['select.language', 'select.language.empty'].map((id) => {
      const fixture = catalogRegistry.fixtures.find((candidate) => candidate.id === id)
      if (!fixture) throw new Error(`Missing fixture ${id}`)
      return fixture
    })
    render(<ExampleCanvas fixtures={fixtures} />)
    const frame = screen.getByTitle<HTMLIFrameElement>('Paper responsive component preview')
    const previewDocument = frame.contentDocument!
    previewDocument.body.innerHTML = '<div id="paper-preview-root"></div>'
    fireEvent.load(frame)
    const preview = within(previewDocument.body)

    expect(preview.getByRole('combobox', { name: 'Language' })).toHaveValue('ja')
    await user.click(preview.getByRole('button', { name: 'Save entry' }))
    expect(await preview.findByText('Entry saved')).toBeInTheDocument()
    await user.selectOptions(screen.getByLabelText('Theme'), 'dark')
    await user.selectOptions(screen.getByLabelText('Fixture'), 'select.language.empty')

    expect(preview.getByRole('combobox', { name: 'Language' })).toHaveValue('')
    expect(preview.queryByText('Entry saved')).not.toBeInTheDocument()
    expect(screen.getByLabelText('Theme')).toHaveValue('dark')
    await user.click(preview.getByRole('button', { name: 'Save entry' }))
    await waitFor(() => expect(preview.getByRole('combobox', { name: 'Language' })).toHaveAttribute('aria-invalid', 'true'))
  })

  it('fits the available stage by default and preserves exact widths only on request', async () => {
    const user = userEvent.setup()
    let resize: ResizeObserverCallback = () => {}
    vi.stubGlobal('ResizeObserver', class {
      constructor(callback: ResizeObserverCallback) { resize = callback }
      observe() {}
      disconnect() {}
    })
    render(<ExampleCanvas />)
    const frame = screen.getByTitle('Paper responsive component preview')
    expect(screen.getByRole('radio', { name: 'Fit' })).toBeChecked()
    act(() => resize([{ contentRect: { width: 592 } } as ResizeObserverEntry], {} as ResizeObserver))
    expect(frame).toHaveAttribute('data-preview-width', '592')
    expect(frame).toHaveStyle({ inlineSize: '100%' })
    expect(screen.getByText('Fit · 592 px · Light · Comfortable')).toBeInTheDocument()
    await user.click(screen.getByRole('radio', { name: 'Desktop' }))
    expect(frame).toHaveStyle({ inlineSize: '1280px' })
    act(() => resize([{ contentRect: { width: 320 } } as ResizeObserverEntry], {} as ResizeObserver))
    expect(frame).toHaveAttribute('data-preview-width', '1280')
    await user.click(screen.getByRole('radio', { name: 'Fit' }))
    expect(frame).toHaveAttribute('data-preview-width', '320')
  })

  it('keeps fixture, theme, and density together before the separate viewport control', () => {
    const fixtures = catalogRegistry.fixtures.slice(0, 2)
    expect(fixtures).toHaveLength(2)

    render(<ExampleCanvas fixtures={fixtures} />)

    const settings = screen.getByRole('group', { name: 'Preview settings' })
    expect(within(settings).getByLabelText('Fixture')).toBeInTheDocument()
    expect(
      Array.from(settings.children).map((child) => child.className),
    ).toEqual([
      'canvas-controls__title paper-type-label',
      'canvas-controls__fixture',
      'canvas-controls__theme',
      'canvas-controls__density',
      'canvas-controls__viewport',
    ])
  })

  it('omits the optional fixture control without leaving a placeholder', () => {
    render(<ExampleCanvas />)

    const settings = screen.getByRole('group', { name: 'Preview settings' })
    expect(within(settings).queryByLabelText('Fixture')).not.toBeInTheDocument()
    expect(
      Array.from(settings.children).map((child) => child.className),
    ).toEqual([
      'canvas-controls__title paper-type-label',
      'canvas-controls__theme',
      'canvas-controls__density',
      'canvas-controls__viewport',
    ])
  })

  it('uses an iframe with a real selected viewport width', async () => {
    const user = userEvent.setup()
    render(<ExampleCanvas />)

    const frame = screen.getByTitle('Paper responsive component preview')
    await user.click(screen.getByRole('radio', { name: 'Desktop' }))
    expect(frame).toHaveAttribute('data-preview-width', '1280')

    const phoneOption = screen.getByRole('radio', {
      name: 'Phone',
    })
    const desktopOption = screen.getByRole('radio', {
      name: 'Desktop',
    })
    expect(phoneOption).not.toHaveClass('paper-button')
    expect(desktopOption).toBeChecked()
    expect(phoneOption).not.toHaveAttribute('aria-pressed')

    await user.click(phoneOption)
    expect(frame).toHaveAttribute('data-preview-width', '360')
    expect(frame).toHaveStyle({ inlineSize: '360px' })
    expect(phoneOption).toBeChecked()
    expect(desktopOption).not.toBeChecked()
  })

  it('presents a compact semantic settings region without a display heading', () => {
    render(<ExampleCanvas />)

    const settings = screen.getByRole('group', { name: 'Preview settings' })
    expect(settings).toHaveClass('canvas-controls')
    expect(within(settings).getByLabelText('Theme')).toHaveClass('paper-select')
    expect(within(settings).getByLabelText('Density')).toHaveClass('paper-select')
    expect(within(settings).getByRole('group', { name: 'Viewport' })).toHaveClass(
      'paper-radio-select--segmented',
    )
    expect(within(settings).getAllByRole('radio')).toHaveLength(4)
    expect(settings.querySelector('.paper-button')).toBeNull()
    expect(settings.querySelector('[aria-pressed]')).toBeNull()
    expect(screen.queryByText('Isolated preview')).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Real viewport canvas' })).not.toBeInTheDocument()
  })

  it('reports independent theme and density settings', async () => {
    const user = userEvent.setup()
    render(<ExampleCanvas />)

    expect(screen.getByLabelText('Theme')).toHaveClass('paper-select')
    expect(screen.getByLabelText('Density')).toHaveClass('paper-select')

    await user.selectOptions(screen.getByLabelText('Theme'), 'dark')
    await user.selectOptions(screen.getByLabelText('Density'), 'compact')

    expect(
      screen.getByText('Fit · 360 px · Dark · Compact'),
    ).toBeInTheDocument()
  })

  it('applies default and changed preferences to the isolated document', async () => {
    const user = userEvent.setup()
    render(<ExampleCanvas />)

    const frame = screen.getByTitle<HTMLIFrameElement>(
      'Paper responsive component preview',
    )
    fireEvent.load(frame)

    expect(frame.contentDocument?.documentElement).toHaveAttribute(
      'data-theme',
      'light',
    )
    expect(frame.contentDocument?.documentElement).toHaveAttribute(
      'data-density',
      'comfortable',
    )

    await user.selectOptions(screen.getByLabelText('Theme'), 'dark')
    await user.selectOptions(screen.getByLabelText('Density'), 'compact')

    await waitFor(() => {
      expect(frame.contentDocument?.documentElement).toHaveAttribute(
        'data-theme',
        'dark',
      )
      expect(frame.contentDocument?.documentElement).toHaveAttribute(
        'data-density',
        'compact',
      )
    })
  })

  it('resets preferences that a replacement fixture no longer supports', async () => {
    const user = userEvent.setup()
    const fixture = catalogRegistry.fixtures[0]
    const { rerender } = render(<ExampleCanvas fixture={fixture} />)

    await user.selectOptions(screen.getByLabelText('Theme'), 'dark')
    await user.selectOptions(screen.getByLabelText('Density'), 'compact')

    rerender(
      <ExampleCanvas
        fixture={{
          ...fixture,
          themes: ['light'],
          densities: ['comfortable'],
        }}
      />,
    )

    await waitFor(() => {
      expect(screen.getByLabelText('Theme')).toHaveValue('light')
      expect(screen.getByLabelText('Density')).toHaveValue(
        'comfortable',
      )
    })
  })

  it('starts with the fitting canvas when the host viewport is narrow', () => {
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: true })))

    render(<ExampleCanvas />)

    expect(screen.getByTitle('Paper responsive component preview')).toHaveAttribute(
      'data-preview-width',
      '360',
    )
    expect(screen.getByText('Fit · 360 px · Light · Comfortable')).toBeInTheDocument()
  })

  it('uses registered fixture viewport dimensions instead of the fallbacks', async () => {
    const user = userEvent.setup()
    const fixture = catalogRegistry.fixtures[0]
    render(
      <ExampleCanvas
        fixture={{
          ...fixture,
          viewports: [
            { id: 'reader', label: 'Reader', width: 412, height: 915 },
          ],
        }}
      />,
    )

    const frame = screen.getByTitle('Paper responsive component preview')
    await user.click(screen.getByRole('radio', { name: 'Reader' }))
    expect(frame).toHaveAttribute('width', '412')
    expect(frame).toHaveAttribute('height', '915')
    expect(frame).toHaveAttribute('data-preview-height', '915')
    expect(screen.getByText('Reader · 412 px · Light · Comfortable')).toBeInTheDocument()
  })
})
