import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { CatalogIndex, ResolvedCatalogueRoute } from '../src/app/routes'

describe('document surfaces', () => {
  it('uses Paper surfaces for catalogue and design-history cards', () => {
    render(
      <MemoryRouter>
        <CatalogIndex />
      </MemoryRouter>,
    )

    const componentCard = screen.getByRole('link', { name: /ButtonStable/u })
    expect(componentCard).toHaveClass(
      'document-card',
      'paper-surface-card',
      'paper-elevation-floating',
    )
    expect(componentCard).not.toHaveClass('paper-surface-raised')

    const historyCard = screen.getByRole('link', {
      name: /Original refinement audit/u,
    })
    expect(historyCard).toHaveClass(
      'design-history__link',
      'paper-surface-card',
      'paper-elevation-flat',
    )
  })

  it('uses a semantic Paper surface for the not-found state', () => {
    render(
      <MemoryRouter initialEntries={['/outside-the-catalogue']}>
        <ResolvedCatalogueRoute />
      </MemoryRouter>,
    )

    const notFound = screen
      .getByRole('heading', { name: 'This Paper page does not exist.' })
      .closest('section')
    expect(notFound).toHaveClass(
      'not-found',
      'paper-surface-card',
      'paper-elevation-flat',
      'paper-accent-rail',
    )
  })
})
