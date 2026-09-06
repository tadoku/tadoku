import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { catalogRegistry } from 'paper-ui/catalog'
import { expect, it } from 'vitest'
import { ComponentWorkbench } from '../src/documentation/ComponentWorkbench'

const document = catalogRegistry.documents.find(entry => entry.id === 'component.button')!
const fixtures = catalogRegistry.fixtures.filter(entry => document.fixtureIds.includes(entry.id))

it('highlights source while preserving the exact copyable text', async () => {
  const user = userEvent.setup()
  const { container } = render(<ComponentWorkbench document={document} fixtures={fixtures} />)
  await user.click(screen.getByRole('tab', { name: 'Code' }))
  const code = container.querySelector('pre code')!
  expect(code.textContent).toBe(fixtures[0].code)
  expect(code.querySelectorAll('.token').length).toBeGreaterThan(0)
})

it('keeps API prose out of code blocks and provides an explicit class reference', async () => {
  const user = userEvent.setup()
  render(<ComponentWorkbench document={document} fixtures={[]} />)
  await user.click(screen.getByRole('tab', { name: 'API / Props' }))
  const guidance = screen.getByText(document.api.invalidCombinations[0])
  expect(guidance.closest('code')).toBeNull()
  expect(screen.getByRole('table', { name: 'CSS classes and helpers' })).toBeVisible()
})
