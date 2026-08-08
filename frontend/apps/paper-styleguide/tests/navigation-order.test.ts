import { describe, expect, it } from 'vitest'
import type { CatalogDocument } from 'paper-ui/catalog'
import { buildNavigationGroups } from '../src/app/catalogue'

function document(
  name: string,
  category: string,
  kind: CatalogDocument['kind'] = 'component',
): CatalogDocument {
  return {
    id: `${kind}.${name.toLocaleLowerCase().replace(/ /gu, '-')}`,
    route: `/${kind}/${name.toLocaleLowerCase().replace(/ /gu, '-')}`,
    name,
    kind,
    category,
  } as CatalogDocument
}

describe('catalogue navigation learning order', () => {
  it('progresses from foundations and common components to specialized material', () => {
    const documents = [
      document('Contributing', 'governance', 'governance'),
      document('Table', 'data-display'),
      document('Logging', 'patterns', 'pattern'),
      document('Modal', 'overlays'),
      document('Navbar', 'navigation'),
      document('Flash', 'feedback'),
      document('Input', 'forms'),
      document('Button', 'actions'),
      document('Color', 'foundations', 'foundation'),
    ]

    expect(buildNavigationGroups(documents).map(({ id }) => id)).toEqual([
      'foundation',
      'actions',
      'forms',
      'feedback',
      'navigation',
      'overlays',
      'data-display',
      'pattern',
      'governance',
    ])
  })

  it('teaches common form controls before specialized combinations', () => {
    const documents = [
      document('AmountWithUnit', 'forms'),
      document('TagsInput', 'forms'),
      document('DatePicker', 'forms'),
      document('Autocomplete', 'forms'),
      document('Checkbox', 'forms'),
      document('Input', 'forms'),
      document('TextArea', 'forms'),
      document('Select', 'forms'),
    ]

    expect(
      buildNavigationGroups(documents)[0]?.documents.map(({ name }) => name),
    ).toEqual([
      'Input',
      'TextArea',
      'Select',
      'Checkbox',
      'Autocomplete',
      'TagsInput',
      'AmountWithUnit',
      'DatePicker',
    ])
  })

  it('orders unknown groups and documents deterministically', () => {
    const documents = [
      document('Zeta', 'future-z'),
      document('Beta', 'future-a'),
      document('Alpha', 'future-a'),
    ]

    const groups = buildNavigationGroups(documents)

    expect(groups.map(({ id }) => id)).toEqual(['future-a', 'future-z'])
    expect(groups[0]?.documents.map(({ name }) => name)).toEqual([
      'Alpha',
      'Beta',
    ])
  })
})
