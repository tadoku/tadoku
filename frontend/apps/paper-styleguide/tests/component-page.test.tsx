import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { catalogRegistry, REQUIRED_COMPONENT_SECTION_KEYS } from 'paper-ui/catalog'
import { describe, expect, it } from 'vitest'
import { DocumentPage } from '../src/documentation/DocumentPage'

const buttonDocument = catalogRegistry.documents.find(
  (document) => document.id === 'component.button',
)!

describe('Stable component documentation', () => {
  it('renders the complete instructional sequence from registry metadata', () => {
    render(<DocumentPage document={buttonDocument} />)

    for (const key of REQUIRED_COMPONENT_SECTION_KEYS) {
      const section = buttonDocument.sections?.required[key]
      expect(section).toBeDefined()
      expect(screen.getByRole('heading', { name: section!.heading })).toBeInTheDocument()
    }
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
