import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { catalogRegistry } from 'paper-ui/catalog'
import { describe, expect, it } from 'vitest'
import { DocumentPage } from '../src/documentation/DocumentPage'

const buttonDocument = catalogRegistry.documents.find(
  (document) => document.id === 'component.button',
)!
const inputDocument = catalogRegistry.documents.find(
  (document) => document.id === 'component.input',
)!
const flashDocument = catalogRegistry.documents.find(
  (document) => document.id === 'component.flash',
)!

describe('Stable component documentation', () => {
  it('shows useful release metadata without a project owner field', () => {
    render(<DocumentPage document={buttonDocument} />)

    expect(screen.getByText('Status')).toBeInTheDocument()
    expect(screen.getByText('Version')).toBeInTheDocument()
    expect(screen.getByText('Reviewed')).toBeInTheDocument()
    expect(screen.queryByText('Owner')).not.toBeInTheDocument()
  })

  it('shows a concise, user-facing outline instead of the registry schema', () => {
    render(<DocumentPage document={buttonDocument} />)

    const outline = screen.getByRole('navigation', { name: 'On this page' })
    expect(within(outline).getAllByRole('link').map((link) => link.textContent)).toEqual([
      'Usage',
      'Examples',
      'Variants and states',
      'Behavior',
      'Content guidance',
      'Accessibility',
    ])
    expect(screen.queryByRole('heading', { name: 'Overview' })).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Implementation' })).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'API reference' })).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Migration' })).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Lifecycle' })).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Metadata' })).not.toBeInTheDocument()
  })

  it('keeps the retained Button, Input, and Flash guidance actionable', () => {
    const { rerender } = render(<DocumentPage document={buttonDocument} />)
    expect(screen.getByText(/Use one default button for the primary action/u)).toBeInTheDocument()
    expect(screen.getByText(/type=submit/u)).toBeInTheDocument()

    rerender(<DocumentPage document={inputDocument} />)
    expect(screen.getByText(/read-only value remains focusable/u)).toBeInTheDocument()
    expect(screen.getByText(/error should explain how to fix the value/u)).toBeInTheDocument()

    rerender(<DocumentPage document={flashDocument} />)
    expect(screen.getByText(/Danger is reserved for an urgent failure/u)).toBeInTheDocument()
    expect(screen.getByText(/visible=false removes the message/u)).toBeInTheDocument()
  })

  it('switches Preview, Code, API, and Accessibility with arrow keys', async () => {
    const user = userEvent.setup()
    render(<DocumentPage document={buttonDocument} />)

    const tabs = screen.getByRole('tablist', { name: 'Example views' })
    const preview = within(tabs).getByRole('tab', { name: 'Preview' })
    preview.focus()
    await user.keyboard('{ArrowRight}')
    expect(within(tabs).getByRole('tab', { name: 'Code' })).toHaveFocus()
    expect(screen.getByText(/import \{ Button/u)).toBeInTheDocument()

    await user.keyboard('{ArrowRight}')
    expect(within(tabs).getByRole('tab', { name: 'API / Props' })).toHaveFocus()
    expect(screen.getByRole('heading', { name: 'Public types' })).toBeInTheDocument()

    await user.keyboard('{ArrowRight}')
    expect(within(tabs).getByRole('tab', { name: 'Accessibility' })).toHaveFocus()
    expect(screen.getByRole('heading', { name: 'Keyboard' })).toBeInTheDocument()
  })
})
