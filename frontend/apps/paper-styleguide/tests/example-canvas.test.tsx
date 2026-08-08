import '@testing-library/jest-dom/vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ExampleCanvas } from '../src/documentation/ExampleCanvas'

afterEach(() => vi.unstubAllGlobals())

describe('ExampleCanvas', () => {
  it('uses an iframe with a real selected viewport width', async () => {
    const user = userEvent.setup()
    render(<ExampleCanvas />)

    const frame = screen.getByTitle('Paper responsive component preview')
    expect(frame).toHaveAttribute('data-preview-width', '1280')

    await user.click(screen.getByRole('button', { name: 'Phone, 360 pixels' }))
    expect(frame).toHaveAttribute('data-preview-width', '360')
    expect(frame).toHaveStyle({ inlineSize: '360px' })
  })

  it('reports independent theme and density settings', async () => {
    const user = userEvent.setup()
    render(<ExampleCanvas />)

    await user.selectOptions(screen.getByLabelText('Preview theme'), 'dark')
    await user.selectOptions(screen.getByLabelText('Preview density'), 'compact')

    expect(
      screen.getByText('Desktop · 1280px · dark · compact'),
    ).toBeInTheDocument()
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
