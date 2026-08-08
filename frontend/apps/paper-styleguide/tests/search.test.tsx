import '@testing-library/jest-dom/vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { catalogRegistry } from 'paper-ui/catalog'
import { CatalogueSearch } from '../src/app/CatalogueSearch'

describe('CatalogueSearch', () => {
  it('opens from slash, focuses search, filters, and closes with Escape', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <CatalogueSearch documents={catalogRegistry.documents} />
      </MemoryRouter>,
    )

    await user.keyboard('/')
    const input = screen.getByRole('searchbox', { name: 'Search documents' })
    expect(input).toHaveFocus()

    await user.type(input, 'color')
    expect(screen.getByText(/result/u)).toBeInTheDocument()

    await user.keyboard('{Escape}')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Search Paper/u })).toHaveFocus()
  })

  it('opens with Control-K', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <CatalogueSearch documents={catalogRegistry.documents} />
      </MemoryRouter>,
    )

    await user.keyboard('{Control>}k{/Control}')
    expect(screen.getByRole('dialog', { name: 'Search Paper' })).toBeVisible()
  })

  it('uses the Paper modal and form input and resets the query after dismissal', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <CatalogueSearch documents={catalogRegistry.documents} />
      </MemoryRouter>,
    )

    const trigger = screen.getByRole('button', { name: 'Search Paper' })
    expect(trigger).toHaveClass('paper-button', 'paper-button--outline')

    await user.click(trigger)
    const dialog = screen.getByRole('dialog', { name: 'Search Paper' })
    const input = screen.getByRole('searchbox', { name: 'Search documents' })
    expect(dialog).toHaveClass('paper-modal')
    expect(input).toHaveClass('paper-input')

    await user.type(input, 'button')
    await user.click(screen.getByRole('button', { name: 'Close search' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()

    await user.click(trigger)
    expect(screen.getByRole('searchbox', { name: 'Search documents' })).toHaveValue('')
  })

  it('closes and resets when a result navigates', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter initialEntries={['/search']}>
        <CatalogueSearch documents={catalogRegistry.documents} />
        <Routes>
          <Route path="*" element={<p>Route content</p>} />
        </Routes>
      </MemoryRouter>,
    )

    await user.keyboard('/')
    await user.type(screen.getByRole('searchbox'), 'button')
    await user.click(within(screen.getByRole('dialog')).getAllByRole('link')[0])
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Search Paper' }))
    expect(screen.getByRole('searchbox')).toHaveValue('')
  })

  it('dismisses on outside press through Paper Modal', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <CatalogueSearch documents={catalogRegistry.documents} />
      </MemoryRouter>,
    )

    const trigger = screen.getByRole('button', { name: 'Search Paper' })
    await user.click(trigger)
    const backdrop = document.querySelector<HTMLElement>(
      '.paper-modal__backdrop',
    )
    expect(backdrop).not.toBeNull()
    await user.click(backdrop!)

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('supports ArrowUp, ArrowDown, Home, and End across search results', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <CatalogueSearch documents={catalogRegistry.documents} />
      </MemoryRouter>,
    )

    await user.keyboard('/')
    await user.type(screen.getByRole('searchbox'), 'button')
    const results = within(screen.getByRole('dialog')).getAllByRole('link')
    expect(results).toHaveLength(2)

    await user.keyboard('{ArrowUp}')
    expect(results[1]).toHaveFocus()
    await user.keyboard('{Home}')
    expect(results[0]).toHaveFocus()
    await user.keyboard('{End}')
    expect(results[1]).toHaveFocus()
    await user.keyboard('{ArrowDown}')
    expect(results[0]).toHaveFocus()
    await user.keyboard('{ArrowUp}')
    expect(results[1]).toHaveFocus()
  })
})
