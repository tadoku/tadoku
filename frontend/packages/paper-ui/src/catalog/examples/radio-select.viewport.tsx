import { useStore } from '@tanstack/react-form'
import { Button, Flash, usePaperForm } from 'paper-ui'
import { useState } from 'react'

export default function Example() {
  const [savedValues, setSavedValues] = useState<Record<
    string,
    unknown
  > | null>(null)
  const methods = usePaperForm({
    defaultValues: { viewport: 'desktop' },
    onSubmit: ({ value: values }) => setSavedValues(values),
  })
  const values = useStore(methods.store, state => state.values)

  return (
    <methods.AppForm>
      <methods.Form className="paper-stack" style={{ maxWidth: '32rem' }}>
        <methods.AppField
          name="viewport"
          validators={{
            onChange: ({ value }) => {
              if (!value) return 'Choose an option.'
            },
          }}
        >
          {field => (
            <field.RadioSelect
              label="Preview size"
              hint="Choose one viewport for the isolated preview."
              variant="segmented"
              required
              options={[
                { value: 'phone', label: 'Phone' },
                { value: 'tablet', label: 'Tablet' },
                { value: 'desktop', label: 'Desktop' },
              ]}
            />
          )}
        </methods.AppField>
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
            {JSON.stringify(values, null, 2)}
          </pre>
        </details>
      </methods.Form>
    </methods.AppForm>
  )
}
