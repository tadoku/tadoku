import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { TableOfContents } from '../src/documentation/TableOfContents'

describe('TableOfContents', () => {
  it('uses Paper linked vertical navigation without inventing a current section', () => {
    render(
      <TableOfContents
        items={[
          { id: 'usage', label: 'Usage' },
          { id: 'accessibility', label: 'Accessibility' },
        ]}
      />,
    )

    expect(screen.getByRole('heading', { name: 'On this page' })).toBeVisible()
    const outline = screen.getByRole('navigation', { name: 'On this page' })
    expect(outline).toHaveClass('paper-tabbar', 'paper-tabbar--vertical')
    expect(outline.querySelector('.paper-tabbar__list--vertical')).not.toBeNull()
    expect(within(outline).getByRole('link', { name: 'Usage' })).toHaveAttribute(
      'href',
      '#usage',
    )
    expect(
      within(outline).getByRole('link', { name: 'Accessibility' }),
    ).toHaveAttribute('href', '#accessibility')
    expect(
      within(outline).queryByRole('link', { current: 'page' }),
    ).not.toBeInTheDocument()
  })
})
