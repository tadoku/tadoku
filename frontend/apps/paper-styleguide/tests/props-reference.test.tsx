import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { catalogRegistry } from 'paper-ui/catalog'
import { expect, it } from 'vitest'
import { ComponentWorkbench } from '../src/documentation/ComponentWorkbench'

it('explains actual prop types, defaults and requirements in the API view', async () => {
  const user = userEvent.setup()
  const document = catalogRegistry.documents.find((entry) => entry.id === 'component.button')!
  render(<ComponentWorkbench document={{ ...document, api: { ...document.api, props: [
    { name: 'variant', type: 'ButtonVariant', defaultValue: 'default', description: 'Choose action emphasis.' },
    { name: 'children', type: 'ReactNode', required: true, description: 'Visible action label.' },
  ] } }} fixtures={[]} />)
  await user.click(screen.getByRole('tab', { name: 'API / Props' }))
  const table = screen.getByRole('table', { name: 'Button props' })
  expect(within(table).getByText('ButtonVariant')).toBeVisible()
  expect(within(table).getByText('Choose action emphasis.')).toBeVisible()
  expect(within(table).getByText('Required')).toBeVisible()
})
