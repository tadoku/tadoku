import '@testing-library/jest-dom/vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { catalogRegistry } from 'paper-ui/catalog'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ExampleCanvas } from '../src/documentation/ExampleCanvas'

afterEach(() => vi.unstubAllGlobals())

describe('ExampleCanvas', () => {
  it('uses an iframe with a real selected viewport width', async () => {
    const user = userEvent.setup()
    render(<ExampleCanvas />)

    const frame = screen.getByTitle('Paper responsive component preview')
    expect(frame).toHaveAttribute('data-preview-width', '1280')

    const phoneButton = screen.getByRole('button', {
      name: 'Phone, 360 pixels',
    })
    const desktopButton = screen.getByRole('button', {
      name: 'Desktop, 1280 pixels',
    })
    expect(phoneButton).toHaveClass('paper-button', 'paper-button--outline')
    expect(desktopButton).toHaveClass('paper-button', 'paper-button--default')

    await user.click(phoneButton)
    expect(frame).toHaveAttribute('data-preview-width', '360')
    expect(frame).toHaveStyle({ inlineSize: '360px' })
    expect(phoneButton).toHaveAttribute('aria-pressed', 'true')
    expect(phoneButton).toHaveClass('paper-button--default')
    expect(desktopButton).toHaveClass('paper-button--outline')
  })

  it('reports independent theme and density settings', async () => {
    const user = userEvent.setup()
    render(<ExampleCanvas />)

    expect(screen.getByLabelText('Preview theme')).toHaveClass('paper-select')
    expect(screen.getByLabelText('Preview density')).toHaveClass('paper-select')

    await user.selectOptions(screen.getByLabelText('Preview theme'), 'dark')
    await user.selectOptions(screen.getByLabelText('Preview density'), 'compact')

    expect(
      screen.getByText('Desktop · 1280px · dark · compact'),
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

    await user.selectOptions(screen.getByLabelText('Preview theme'), 'dark')
    await user.selectOptions(screen.getByLabelText('Preview density'), 'compact')

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

    await user.selectOptions(screen.getByLabelText('Preview theme'), 'dark')
    await user.selectOptions(screen.getByLabelText('Preview density'), 'compact')

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
      expect(screen.getByLabelText('Preview theme')).toHaveValue('light')
      expect(screen.getByLabelText('Preview density')).toHaveValue(
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
    expect(screen.getByText('Phone · 360px · light · comfortable')).toBeInTheDocument()
  })
})
