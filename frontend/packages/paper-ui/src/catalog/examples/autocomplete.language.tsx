import { useStore } from '@tanstack/react-form'
import { Button, Flash, usePaperForm } from 'paper-ui'
import { useState } from 'react'

const LANGUAGES = [
  { id: 'ja', label: 'Japanese' },
  { id: 'zh', label: 'Chinese' },
  { id: 'ko', label: 'Korean' },
]

export default function Example() {
  const [savedValues, setSavedValues] = useState<Record<
    string,
    unknown
  > | null>(null)
  const methods = usePaperForm({
    defaultValues: { autocomplete: { id: 'ja', label: 'Japanese' } }, onSubmit: ({ value: values }) => setSavedValues(values) })
  return (
    <methods.AppForm>
      <methods.Form className="paper-stack" style={{ maxWidth: '32rem' }}>
        <methods.AppField name="autocomplete" validators={{ onChange: ({ value }) => { if (!value) return "Choose an option."; } }}>{field => <field.AutocompleteInput label="Language" required options={LANGUAGES} format={option => option.label} getId={option => option.id} placeholder="Search languages" />}</methods.AppField>
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
