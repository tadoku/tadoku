import '@testing-library/jest-dom/vitest'
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { catalogRegistry } from 'paper-ui/catalog'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ExampleCanvas } from '../src/documentation/ExampleCanvas'

afterEach(() => vi.unstubAllGlobals())

describe('ExampleCanvas', () => {
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
    expect(within(settings).getAllByRole('radio')).toHaveLength(3)
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
      screen.getByText('Desktop · 1280 px · Dark · Compact'),
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

  it('starts with the phone canvas when the host viewport is narrow', () => {
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: true })))

    render(<ExampleCanvas />)

    expect(screen.getByTitle('Paper responsive component preview')).toHaveAttribute(
      'data-preview-width',
      '360',
    )
    expect(screen.getByText('Phone · 360 px · Light · Comfortable')).toBeInTheDocument()
  })

  it('uses registered fixture viewport dimensions instead of the fallbacks', () => {
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
    expect(frame).toHaveAttribute('width', '412')
    expect(frame).toHaveAttribute('height', '915')
    expect(frame).toHaveAttribute('data-preview-height', '915')
    expect(screen.getByText('Reader · 412 px · Light · Comfortable')).toBeInTheDocument()
  })
})
