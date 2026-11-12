import { useHydrated } from '@tanstack/react-router'
import { AppliedFilters, Button, FilterMenu, buttonClassName, cn, surface, type FilterMenuHandle, type FilterMenuSection } from 'paper-ui'
import type { Ref } from 'react'
import { filterCount, noFilters, type ProfileFilters } from '../lib/contest-filters'
import type { ActivitySlug } from '../lib/activity'
import { scoreFormat } from '../lib/format'

type Dimension = keyof ProfileFilters

/**
 * The Filters entry point beside the Activity heading. Before hydration it's a native disclosure holding a GET form, one
 * checkbox group per filter, which submits to the first page; once hydrated it's Paper's FilterMenu.
 */
export function ProfileFilterMenu({ menu, sections, action, resultLabel, onChange }: {
  menu: Ref<FilterMenuHandle>
  sections: FilterMenuSection[]
  /** The first page's address, which the form submits to. */
  action: string
  resultLabel: string
  onChange: (next: Partial<ProfileFilters>) => void
}) {
  const hydrated = useHydrated()
  if (hydrated) {
    return (
      <FilterMenu
        ref={menu}
        sections={sections}
        onChange={(id, values) => onChange({ [id]: [...values] })}
        onClear={() => onChange(noFilters)}
        resultLabel={resultLabel}
      />
    )
  }

  const count = sections.reduce((total, section) => total + section.selected.length, 0)
  const listed = sections.map(section => ({
    ...section,
    values: section.values.filter(value => value.points > 0 || section.selected.includes(value.value)),
  })).filter(section => section.values.length)
  return (
    <details className="relative">
      <summary className={buttonClassName({ variant: 'ghost', className: 'list-none [&::-webkit-details-marker]:hidden' })}>
        {count ? `Filters (${count})` : 'Filters'}
      </summary>
      <form
        method="get"
        action={action}
        aria-label="Filters"
        className={cn(
          surface({ tone: 'overlay', elevation: 'floating' }),
          'absolute end-0 top-full z-menu mt-1.5 w-filter-menu overflow-y-auto px-field-x py-2',
        )}
      >
        {listed.map(section => (
          <fieldset key={section.id} className="m-0 border-0 border-b border-rule-subtle p-0 py-2">
            <legend className="mb-1 p-0 text-xs font-bold uppercase tracking-label text-muted">{section.label}</legend>
            {section.values.map(value => (
              <label key={value.value} className="flex min-h-control cursor-pointer items-center gap-inline text-ui">
                <input
                  type="checkbox"
                  name={section.id}
                  value={value.value}
                  defaultChecked={section.selected.includes(value.value)}
                  className="size-check flex-none accent-action"
                />
                <span className="min-w-0 flex-auto wrap-anywhere">{value.label}</span>
                <span className="flex-none text-sm tabular-nums text-muted">{scoreFormat.format(value.points)}</span>
              </label>
            ))}
          </fieldset>
        ))}
        <div className="flex justify-end pt-2">
          <Button type="submit">Apply</Button>
        </div>
      </form>
    </details>
  )
}

/** The applied values as removable chips with Clear all and the filtered count, under the Activity heading. */
export function ProfileAppliedFilters({ filters, labels, resultLabel, onChange, onEmptied }: {
  filters: ProfileFilters
  labels: { language: (code: string) => string; activity: (slug: ActivitySlug) => string }
  resultLabel: string
  onChange: (next: Partial<ProfileFilters>) => void
  /** Called once no filter is left, to return focus to the Filters trigger. */
  onEmptied: () => void
}) {
  const remove = (dimension: Dimension, value: string) => {
    onChange({ [dimension]: filters[dimension].filter(item => item !== value) })
    if (filterCount(filters) === 1) onEmptied()
  }
  const label = (dimension: Dimension, value: string) =>
    dimension === 'language' ? labels.language(value) : dimension === 'activity' ? labels.activity(value as ActivitySlug) : value

  return (
    <AppliedFilters
      variant="removable"
      className="mb-4"
      filters={(['language', 'activity', 'tag'] as const).flatMap(dimension =>
        filters[dimension].map(value => ({ id: `${dimension}:${value}`, label: label(dimension, value), onRemove: () => remove(dimension, value) })))}
      onClear={() => {
        onChange(noFilters)
        onEmptied()
      }}
      resultLabel={resultLabel}
    />
  )
}

