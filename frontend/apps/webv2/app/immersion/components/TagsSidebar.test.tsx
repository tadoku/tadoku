// @vitest-environment jsdom

import { render, screen, within } from '@testing-library/react'
import React from 'react'
import { FormProvider, useForm } from 'react-hook-form'
import { afterEach, describe, expect, it } from 'vitest'
import { TagsSidebar } from './TagsSidebar'

const TagsSidebarForm = ({ activityId }: { activityId: number }) => {
  const methods = useForm({ defaultValues: { tags: [] } })

  return (
    <FormProvider {...methods}>
      <TagsSidebar activityId={activityId} />
    </FormProvider>
  )
}

afterEach(() => document.body.replaceChildren())

describe('TagsSidebar', () => {
  it('groups reading modifiers above common tags without changing tag labels', () => {
    render(<TagsSidebarForm activityId={1} />)

    const modifiersHeading = screen.getByText('Score modifiers')
    const commonHeading = screen.getByText('Common tags')
    const modifiers = modifiersHeading.closest('section')
    const common = commonHeading.closest('section')

    expect(
      modifiersHeading.compareDocumentPosition(commonHeading) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
    expect(modifiers).not.toBeNull()
    expect(common).not.toBeNull()
    expect(
      within(modifiers!).getByRole('button', { name: 'manga' }),
    ).toBeTruthy()
    expect(
      within(modifiers!).getByRole('button', { name: 'comic' }),
    ).toBeTruthy()
    expect(
      within(modifiers!).getByRole('button', { name: 'two_column' }),
    ).toBeTruthy()
    expect(within(common!).getByRole('button', { name: 'book' })).toBeTruthy()
  })

  it.each([2, 4])('groups dense as a modifier for activity %s', activityId => {
    render(<TagsSidebarForm activityId={activityId} />)

    const modifiers = screen.getByText('Score modifiers').closest('section')

    expect(modifiers).not.toBeNull()
    expect(
      within(modifiers!).getByRole('button', { name: 'dense' }),
    ).toBeTruthy()
  })

  it('only shows common tags when an activity has no modifiers', () => {
    render(<TagsSidebarForm activityId={3} />)

    expect(screen.queryByText('Score modifiers')).toBeNull()
    expect(screen.getByText('Common tags')).toBeTruthy()
  })
})
