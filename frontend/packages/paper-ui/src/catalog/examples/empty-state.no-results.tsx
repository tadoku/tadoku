import { useStore } from '@tanstack/react-form'
import { Button, EmptyState, usePaperForm } from 'paper-ui'

const contests = ['2026 Round 5', 'Japanese novel club', 'Autumn in French']

export default function Example() {
  const methods = usePaperForm({ defaultValues: { query: 'Klingon' } })
  const query = useStore(methods.store, state => state.values['query'])
  const matches = contests.filter(title =>
    title.toLocaleLowerCase().includes(query.trim().toLocaleLowerCase()),
  )
  const clearSearch = () => {
    methods.setFieldValue('query', '')
    document.getElementsByName('query')[0]?.focus()
  }

  return (
    <div className="paper-stack" style={{ maxWidth: '40rem' }}>
      <methods.AppForm>
        <form role="search" onSubmit={event => event.preventDefault()}>
          <methods.AppField name="query">{field => <field.Input label="Find a contest" type="search" />}</methods.AppField>
        </form>
      </methods.AppForm>
      <p role="status">
        {matches.length} {matches.length === 1 ? 'contest' : 'contests'}
      </p>
      {matches.length ? (
        <ul>
          {matches.map(title => (
            <li key={title}>{title}</li>
          ))}
        </ul>
      ) : (
        <EmptyState
          headingLevel={2}
          title={`No contests match “${query.trim()}”.`}
          description="Try another title or language, or clear the search to see every contest."
          actions={
            <Button variant="outline" onClick={clearSearch}>
              Clear search
            </Button>
          }
        />
      )}
    </div>
  )
}
