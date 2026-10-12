import { Button, EmptyState, Input } from 'paper-ui'
import { FormProvider, useForm, useWatch } from 'react-hook-form'

const contests = ['2026 Round 5', 'Japanese novel club', 'Autumn in French']

export default function Example() {
  const methods = useForm({ defaultValues: { query: 'Klingon' } })
  const query = useWatch({ control: methods.control, name: 'query' })
  const matches = contests.filter(title =>
    title.toLocaleLowerCase().includes(query.trim().toLocaleLowerCase()),
  )
  const clearSearch = () => {
    methods.setValue('query', '')
    methods.setFocus('query')
  }

  return (
    <div className="paper-stack" style={{ maxWidth: '40rem' }}>
      <FormProvider {...methods}>
        <form role="search" onSubmit={event => event.preventDefault()}>
          <Input name="query" label="Find a contest" type="search" />
        </form>
      </FormProvider>
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
