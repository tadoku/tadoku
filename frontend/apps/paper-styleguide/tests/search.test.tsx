import '@testing-library/jest-dom/vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
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
})
