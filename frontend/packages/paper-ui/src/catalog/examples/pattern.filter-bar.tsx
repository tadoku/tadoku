import { useStore } from '@tanstack/react-form'
import {
  AppliedFilters,
  Button,
  EmptyState,
  SectionHeading,
  Table,
  usePaperForm,
} from 'paper-ui'
import { ChevronDownIcon } from 'paper-ui/icons'
import { useId, useRef, useState } from 'react'

interface Standing {
  readonly name: string
  readonly language: string
  readonly activity: string
  readonly score: number
}

const standings: readonly Standing[] = [
  { name: 'Kai', language: 'Japanese', activity: 'Reading', score: 1400 },
  { name: 'Mei', language: 'Japanese', activity: 'Listening', score: 1120 },
  { name: 'Sora', language: 'Spanish', activity: 'Reading', score: 860 },
  { name: 'Anton', language: 'Japanese', activity: 'Reading', score: 450 },
  { name: 'Lina', language: 'Spanish', activity: 'Listening', score: 320 },
]
const noFilters = { language: 'all', activity: 'all', query: '' }

export default function FilterBarExample() {
  const [open, setOpen] = useState(false)
  const toggle = useRef<HTMLButtonElement>(null)
  const panelId = useId()
  const headingId = useId()
  const methods = usePaperForm({ defaultValues: noFilters })
  const {
    language = 'all',
    activity = 'all',
    query = '',
  } = useStore(methods.store, state => state.values)
  const search = query.trim().toLocaleLowerCase()
  const applied = [
    language !== 'all' ? language : '',
    activity !== 'all' ? activity : '',
    search ? `Name: ${query.trim()}` : '',
  ].filter(Boolean)
  const rows = standings.filter(
    row =>
      (language === 'all' || row.language === language) &&
      (activity === 'all' || row.activity === activity) &&
      row.name.toLocaleLowerCase().includes(search),
  )
  const clearFilters = () => {
    methods.reset(noFilters, { keepDefaultValues: true })
    toggle.current?.focus()
  }

  return (
    <section aria-labelledby={headingId} style={{ maxWidth: '44rem' }}>
      <SectionHeading
        level={2}
        id={headingId}
        title="2026 Round 5 standings"
        actions={
          <Button
            ref={toggle}
            variant="ghost"
            aria-expanded={open}
            aria-controls={panelId}
            onClick={() => setOpen(!open)}
            trailingIcon={
              <ChevronDownIcon
                className="paper-icon-compact"
                style={open ? { transform: 'rotate(180deg)' } : undefined}
              />
            }
          >
            Filters{applied.length ? ` (${applied.length})` : ''}
          </Button>
        }
      />
      <div id={panelId} hidden={!open}>
        <methods.AppForm>
          <form
            onSubmit={event => event.preventDefault()}
            style={{
              display: 'grid',
              gap: '0.75rem',
              gridTemplateColumns: 'repeat(auto-fit, minmax(11rem, 1fr))',
              paddingBlockEnd: '1rem',
            }}
          >
            <methods.AppField name="language">
              {field => (
                <field.Select
                  label="Language"
                  options={[
                    { value: 'all', label: 'All languages' },
                    { value: 'Japanese', label: 'Japanese' },
                    { value: 'Spanish', label: 'Spanish' },
                  ]}
                />
              )}
            </methods.AppField>
            <methods.AppField name="activity">
              {field => (
                <field.Select
                  label="Activity"
                  options={[
                    { value: 'all', label: 'All activities' },
                    { value: 'Reading', label: 'Reading' },
                    { value: 'Listening', label: 'Listening' },
                  ]}
                />
              )}
            </methods.AppField>
            <methods.AppField name="query">
              {field => (
                <field.Input label="Find a participant" type="search" />
              )}
            </methods.AppField>
          </form>
        </methods.AppForm>
      </div>
      <div
        className="paper-cluster"
        style={{ justifyContent: 'space-between' }}
      >
        <AppliedFilters filters={applied} onClear={clearFilters} />
        <p className="paper-type-metadata" role="status">
          {rows.length} of {standings.length} participants
        </p>
      </div>
      {rows.length ? (
        <Table
          caption="Standings"
          captionVisibility="screen-reader"
          rows={rows}
          getRowKey={row => row.name}
          minWidth="17rem"
          columns={[
            {
              id: 'participant',
              header: 'Participant',
              rowHeader: true,
              cell: row => row.name,
            },
            { id: 'language', header: 'Language', cell: row => row.language },
            {
              id: 'score',
              header: 'Score',
              align: 'end',
              cell: row => row.score.toLocaleString('en'),
            },
          ]}
        />
      ) : (
        <EmptyState
          variant="plain"
          headingLevel={3}
          title="No participants match these filters."
          description="Filtering and searching never change anyone’s rank."
          actions={
            <Button variant="outline" onClick={clearFilters}>
              Clear filters
            </Button>
          }
        />
      )}
    </section>
  )
}
