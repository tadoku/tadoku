import { Button, Flash, RadioGroup } from 'paper-ui'
import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function Example() {
  const [savedValues, setSavedValues] = useState<Record<
    string,
    unknown
  > | null>(null)
  const methods = useForm({ defaultValues: { format: 'book' } })
  return (
    <FormProvider {...methods}>
      <form
        className="paper-stack"
        style={{ maxWidth: '32rem' }}
        noValidate
        onSubmit={methods.handleSubmit(values => setSavedValues(values))}
      >
        <RadioGroup
          name="format"
          label="Reading format"
          options={[
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
          ]}
        />
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
            {JSON.stringify(methods.watch(), null, 2)}
          </pre>
        </details>
      </form>
    </FormProvider>
  )
}
