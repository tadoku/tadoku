import { useStore } from '@tanstack/react-form'
import { Button, Flash, usePaperForm } from 'paper-ui'
import { useState } from 'react'

export default function Example() {
  const [savedValues, setSavedValues] = useState<Record<
    string,
    unknown
  > | null>(null)
  const methods = usePaperForm({ defaultValues: { format: 'book' }, onSubmit: ({ value: values }) => setSavedValues(values) })
  return (
    <methods.AppForm>
      <methods.Form className="paper-stack" style={{ maxWidth: '32rem' }}>
        <methods.AppField name="format" validators={{ onChange: ({ value }) => { if (!value) return "Choose an option."; } }}>{field => <field.RadioGroup label="Reading format" options={[
            {
              value: 'book',
              label: 'Book',
              description: 'Printed or digital text',
            },
            { value: 'audio', label: 'Audio', description: 'Narrated content' },
            {
              value: 'other',
              label: 'Other',
              description: 'Another eligible format',
              disabled: true,
            },
          ]} />}</methods.AppField>
        <div className="paper-cluster">
          <Button type="submit">Save entry</Button>
          <Button
            variant="outline"
            onClick={() => {
              methods.reset()
              setSavedValues(null)
            }}
          >
            Reset
          </Button>
        </div>
        {savedValues ? (
          <Flash variant="success" title="Entry saved">
            <pre
              style={{
                whiteSpace: 'pre-wrap',
                overflowWrap: 'anywhere',
                margin: 0,
              }}
            >
              {JSON.stringify(savedValues, null, 2)}
            </pre>
          </Flash>
        ) : null}
        <details>
          <summary>Current form values</summary>
          <pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>
            {JSON.stringify(useStore(methods.store, state => state.values), null, 2)}
          </pre>
        </details>
      </methods.Form>
    </methods.AppForm>
  )
}
